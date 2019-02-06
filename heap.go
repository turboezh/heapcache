package heapcache

type (
	itemsHeap struct {
		less Less
		Heap []*wrapper
	}
)

func newHeap(capacity int, less Less) *itemsHeap {
	return &itemsHeap{
		less: less,
		Heap: make([]*wrapper, 0, capacity),
	}
}

func (h *itemsHeap) Len() int {
	return len(h.Heap)
}

func (h *itemsHeap) Less(i, j int) bool {
	return h.less(h.Heap[i].value, h.Heap[j].value)
}

func (h *itemsHeap) Swap(i, j int) {
	h.Heap[i], h.Heap[j] = h.Heap[j], h.Heap[i]
	h.Heap[i].index = i
	h.Heap[j].index = j
}

func (h *itemsHeap) Push(value interface{}) {
	item := value.(*wrapper)
	item.index = len(h.Heap)
	h.Heap = append(h.Heap, item)
}

func (h *itemsHeap) Pop() interface{} {
	n := len(h.Heap)
	item := h.Heap[n-1]
	item.index = -1 // for safety
	h.Heap = h.Heap[0 : n-1]
	return item
}
