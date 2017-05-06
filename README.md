## go-groupthrottle

[![GoDoc](https://godoc.org/github.com/nathan-osman/go-groupthrottle?status.svg)](https://godoc.org/github.com/nathan-osman/go-groupthrottle)
[![MIT License](http://img.shields.io/badge/license-MIT-9370d8.svg?style=flat)](http://opensource.org/licenses/MIT)

This package helps avoid repeatedly calling functions by grouping parameters.

For example, suppose that you had a function that generates a special index based on the pictures in a directory. To keep the index up to date, the function watches for new files being added and regenerates the index. This works great until someone pastes 100 new pictures in the directory; suddenly the function is invoked 100 times.

go-groupthrottle eliminates this problem by grouping or "buffering" the addition of new items, invoking the function once a predetermined amount of time has elapsed with no new additions. Items have keys associated with them as well, so if an item is removed before the function is invoked, it will not be passed to the function.

### Basic Usage

Begin by importing the package and creating a `GroupThrottle`. A function and an interval is provided which determines how much time must elapse with no new additions before the function is invoked. The callback function must take a slice as its only parameter. The slice type will determine what type of items may be added.

    import "github.com/nathan-osman/go-groupthrottle"

    func process(items []*Item) {
        // do something with the items
    }

    t, err := groupthrottle.New(process, 10*time.Second)
    if err != nil {
        // handle error
    }
    defer t.Close()

To add items, use the `Add()` method and supply a key for the item. In the example described above, the filename would be a suitable key name. `Add()` will return an error if the item type does not match the slice type in the callback's parameter.

    i := &Item{
        Field1: 1,
        Field2: 2,
    }
    t.Add("mykey", i)

To remove an item before the timer expires and the function is invoked, use the `Remove()` method.

    t.Remove("mykey")

To immediately invoke the function with all of the pending items, use the `Flush()` method.

    t.Flush()
