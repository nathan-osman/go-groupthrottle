package groupthrottle

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

const (
	Key1 = "key1"
	Key2 = "key2"
	Key3 = "key3"

	Value1 = "value1"
	Value2 = "value2"
	Value3 = "value3"
)

func TestAdd(t *testing.T) {
	c := make(chan []string)
	g, err := New(func(items []string) {
		c <- items
	}, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	g.Add(Key1, Value1)
	time.Sleep(25 * time.Millisecond)
	g.Add(Key2, Value2)
	time.Sleep(25 * time.Millisecond)
	g.Add(Key3, Value3)
	time.Sleep(25 * time.Millisecond)
	select {
	case <-c:
		t.Fatal("should not be able to receive an item")
	default:
	}
	select {
	case items := <-c:
		sort.Strings(items)
		correctItems := []string{Value1, Value2, Value3}
		if !reflect.DeepEqual(items, correctItems) {
			t.Fatalf("%#v != %#v", items, correctItems)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("timeout exceeded")
	}
}

func TestRemove(t *testing.T) {
	c := make(chan bool)
	g, err := New(func(items []string) {
		close(c)
	}, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	g.Add(Key1, Value1)
	g.Remove(Key1)
	select {
	case <-c:
		t.Fatal("item not removed")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestInvalidType(t *testing.T) {
	g, err := New(func(items []string) {}, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	if err := g.Add(Key1, 42); err != ErrInvalidType {
		t.Fatalf("%#v != %#v", err, ErrInvalidType)
	}
}
