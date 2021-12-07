package fifo

import (
	"container/list"
	"sync"
)

type Queue interface {
	Len() int
	Cap() int

	Enqueue(interface{})
	Peek(int) []interface{}
}

// FixedFifo (First In First Out) concurrent queue
type FixedFifo struct {
	cap     int
	rwmutex sync.RWMutex
	arr     *list.List
}

func (q *FixedFifo) Len() int {
	return q.arr.Len()
}
func (q *FixedFifo) Cap() int {
	return q.cap
}

// Get the first n elements
func (q *FixedFifo) Peek(n int) []interface{} {
	size := n
	if q.Len() < n {
		size = q.Len() // Maximum the size of the queue
	}
	items := make([]interface{}, size)
	item := q.arr.Front()
	for i := 0; i < size; i++ {
		items[i] = item.Value
		item = item.Next()
	}
	return items
}

// NewFixedFifo returns a new FixedFifo concurrent queue
func NewFixedFifo(cap int) FixedFifo {
	return FixedFifo{
		cap: cap,
		arr: list.New(),
	}
}

// Enqueue enqueues an element.
func (q *FixedFifo) Enqueue(value interface{}) {
	// lock the object to enqueue the element into the list
	q.rwmutex.Lock()
	defer q.rwmutex.Unlock()

	// enqueue the element
	if q.arr.Len() >= q.cap {
		q.arr.Remove(q.arr.Back()) // Remove the last element
	}
	q.arr.PushFront(value)
}
