package fifo

import (
	tmplog "log"
	"testing"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

type AA struct {
	value int
}

func TestNew(t *testing.T) {
	q := NewFixedFifo(10)

	// enqueue elements
	for i := 1; i < 20; i++ {
		item := &AA{
			value: i,
		}
		q.Enqueue(item)
	}

	//peek elements
	for i := 7; i < 11; i++ {
		peeks := q.Peek(i)
		tmplog.Print(i, peeks)
	}
}
