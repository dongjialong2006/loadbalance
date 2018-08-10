package types

import (
	"sync"
)

type Heap struct {
	sync.RWMutex
	max   int
	queue []Conn
}

func NewHeap(max int) *Heap {
	return &Heap{
		max:   max,
		queue: make([]Conn, 0),
	}
}

func (q Heap) less(i, j int) bool {
	return q.queue[i].Weight() < q.queue[j].Weight()
}

func (q Heap) swap(i, j int) {
	q.queue[i], q.queue[j] = q.queue[j], q.queue[i]
}

func (q *Heap) Pop() Conn {
	n := len(q.queue) - 1
	if n < 0 {
		return nil
	}

	q.Lock()
	defer q.Unlock()
	q.swap(0, n)
	conn := q.queue[n]
	if nil == conn || !conn.Enable() {
		return nil
	}
	conn.AddRef()

	q.down(0, n)

	return conn
}

func (q *Heap) Fix(i int) {
	q.Lock()
	if !q.down(i, len(q.queue)) {
		q.up(i)
	}
	q.Unlock()
}

func (q *Heap) Push(conn Conn) {
	if q.max > len(q.queue) {
		if conn.Enable() {
			q.Lock()
			q.queue = append(q.queue, conn)
			q.up(len(q.queue) - 1)
			q.Unlock()
		}
	}
}

func (q *Heap) HeapInit() {
	q.Lock()
	n := len(q.queue)
	for i := n/2 - 1; i >= 0; i-- {
		q.down(i, n)
	}
	q.Unlock()
	return
}

func (q *Heap) up(j int) {
	for {
		i := (j - 1) / 2
		if i == j || !q.less(j, i) {
			break
		}
		q.swap(i, j)
		j = i
	}
	return
}

func (q *Heap) down(i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 {
			break
		}
		j := j1
		if j2 := j1 + 1; j2 < n && q.less(j2, j1) {
			j = j2
		}
		if !q.less(j, i) {
			break
		}
		q.swap(i, j)
		i = j
	}
	return i > i0
}
