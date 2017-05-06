package groupthrottle

import (
	"errors"
	"reflect"
	"time"
)

var (
	ErrCallbackNotFunc = errors.New("param must be a function")
	ErrParamCount      = errors.New("callback must have a single param")
	ErrParamType       = errors.New("callback param must be a slice")
	ErrInvalidType     = errors.New("type does not match callback")
)

type addItem struct {
	Key  string
	Item interface{}
}

// GroupThrottle delays invoking a callback to process items until a specified
// amount of time has elapsed since an item has been added.
type GroupThrottle struct {
	stop     chan bool
	stopped  chan bool
	add      chan *addItem
	remove   chan string
	flush    chan bool
	callback reflect.Value
	itemType reflect.Type
	delay    time.Duration
}

func (g *GroupThrottle) run() {
	defer close(g.stopped)
	var (
		items = make(map[string]interface{})
		timer <-chan time.Time
	)
	for {
		select {
		case p := <-g.add:
			items[p.Key] = p.Item
			timer = time.After(g.delay)
			continue
		case k := <-g.remove:
			delete(items, k)
			if len(items) == 0 {
				timer = nil
			}
			continue
		case <-g.flush:
		case <-timer:
		case <-g.stop:
			return
		}
		itemList := reflect.MakeSlice(reflect.SliceOf(g.itemType), 0, 0)
		for _, i := range items {
			itemList = reflect.Append(itemList, reflect.ValueOf(i))
		}
		go func() {
			g.callback.Call([]reflect.Value{itemList})
		}()
		items = make(map[string]interface{})
		timer = nil
	}
}

// New creates a new GroupThrottle that will invoke the specified callback
// function after the specified delay.
func New(callback interface{}, delay time.Duration) (*GroupThrottle, error) {
	callbackType := reflect.TypeOf(callback)
	if callbackType.Kind() != reflect.Func {
		return nil, ErrCallbackNotFunc
	}
	if callbackType.NumIn() != 1 {
		return nil, ErrParamCount
	}
	paramType := callbackType.In(0)
	if paramType.Kind() != reflect.Slice {
		return nil, ErrParamType
	}
	g := &GroupThrottle{
		stop:     make(chan bool),
		stopped:  make(chan bool),
		add:      make(chan *addItem),
		remove:   make(chan string),
		flush:    make(chan bool),
		callback: reflect.ValueOf(callback),
		itemType: paramType.Elem(),
		delay:    delay,
	}
	go g.run()
	return g, nil
}

// Add adds an item to the throttle. If an item already exists with the
// specified key, it is replaced.
func (g *GroupThrottle) Add(key string, item interface{}) error {
	if !reflect.TypeOf(item).AssignableTo(g.itemType) {
		return ErrInvalidType
	}
	g.add <- &addItem{
		Key:  key,
		Item: item,
	}
	return nil
}

// Remove removes an item from the throttle if it exists.
func (g *GroupThrottle) Remove(key string) {
	g.remove <- key
}

// Flush causes the callback to be invoked immediately with all pending items.
func (g *GroupThrottle) Flush() {
	g.flush <- true
}

// Close shuts down the GroupThrottle. The function is not invoked with pending
// items. Use Flush() before Close() to ensure all pending items are processed.
func (g *GroupThrottle) Close() {
	close(g.stop)
	<-g.stopped
}
