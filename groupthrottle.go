package groupthrottle

import (
	"time"
)

type addParams struct {
	Key  string
	Item interface{}
}

// GroupThrottle delays invoking a callback to process items until a specified
// amount of time has elapsed since an item has been added.
type GroupThrottle struct {
	stop     chan bool
	stopped  chan bool
	add      chan *addParams
	remove   chan string
	flush    chan bool
	callback func(...interface{})
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
		itemList := []interface{}{}
		for _, i := range items {
			itemList = append(itemList, i)
		}
		go func() {
			g.callback(itemList)
		}()
		items = make(map[string]interface{})
		timer = nil
	}
}

// New creates a new GroupThrottle that will invoke the specified callback
// function after the specified delay.
func New(callback func(...interface{}), delay time.Duration) *GroupThrottle {
	g := &GroupThrottle{
		stop:    make(chan bool),
		stopped: make(chan bool),
		add:     make(chan *addParams),
		remove:  make(chan string),
		flush:   make(chan bool),
		delay:   delay,
	}
	go g.run()
	return g
}

// Add adds an item to the throttle. If an item already exists with the
// specified key, it is replaced.
func (g *GroupThrottle) Add(key string, item interface{}) {
	g.add <- &addParams{
		Key:  key,
		Item: item,
	}
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
