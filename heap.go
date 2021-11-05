package eventbus

import (
	"container/heap"
	"sync"
)

type node struct {
	value    interface{}
	priority int
	index    int
}

type store struct {
	lookup map[string]*node
	queue  []string
	key    func(x interface{}) string
}

var (
	_ = heap.Interface(&store{})
)

// Len ...
func (s store) Len() int { return len(s.queue) }

// Less ...
func (s store) Less(i, j int) bool {
	return s.lookup[s.queue[i]].priority > s.lookup[s.queue[j]].priority
}

// Swap ...
func (s store) Swap(i, j int) {
	s.queue[i], s.queue[j] = s.queue[j], s.queue[i]

	it := s.lookup[s.queue[i]]
	it.index = i

	it = s.lookup[s.queue[j]]
	it.index = j
}

// Push ...
func (s *store) Push(x interface{}) {
	it, _ := x.(*node)
	it.index = len(s.queue)

	s.queue = append(s.queue, s.key(it.value))
	s.lookup[s.key(it.value)] = it
}

// Pop ...
func (s *store) Pop() interface{} {
	size := len(s.queue)
	it := s.lookup[s.queue[size-1]]

	// delete
	delete(s.lookup, s.queue[size-1])
	s.queue = s.queue[0 : size-1]
	return it
}

// Heap include heap store structure
type Heap struct {
	locker sync.RWMutex
	s      *store
}

// Add if it exists, update the data, otherwise create a new node and push to the heap
func (h *Heap) Add(x interface{}, priority int) {
	h.locker.Lock()
	defer h.locker.Unlock()

	if it, ok := h.s.lookup[h.s.key(x)]; ok {
		it.priority = priority
		heap.Fix(h.s, it.index)
		return
	}

	it := &node{
		value:    x,
		priority: priority,
	}

	heap.Push(h.s, it)
}

// Delete ...
func (h *Heap) Delete(x interface{}) {
	// if the data exists, delete the data
	// if the data does not exist, do nothing
	h.locker.Lock()
	defer h.locker.Unlock()

	if it, ok := h.s.lookup[h.s.key(x)]; ok {
		_ = heap.Remove(h.s, it.index)
		return
	}
}

// Pop ...
func (h *Heap) Pop() interface{} {
	// return the element with the highest priority,
	// or empty if the queue is empty
	// and at the same time, delete the element from the queue
	h.locker.Lock()
	defer h.locker.Unlock()

	if len(h.s.queue) == 0 {
		return nil
	}

	it := heap.Pop(h.s)
	if it == nil {
		return nil
	}

	return it.(*node).value
}

// Peek ...
func (h *Heap) Peek() interface{} {
	h.locker.RLock()
	defer h.locker.RUnlock()

	if len(h.s.queue) > 0 {
		return h.s.lookup[h.s.queue[0]].value
	}

	return nil
}

// Size ...
func (h *Heap) Size() int {
	return len(h.s.queue)
}

// NewHeap return heap pointer
func NewHeap(key func(x interface{}) string) *Heap {
	return &Heap{
		s: &store{
			lookup: make(map[string]*node),
			key:    key,
		},
	}
}
