package heapcache

type itemsHeap []*wrapper

func (h *itemsHeap) Len() int {
	return len(*h)
}

func (h *itemsHeap) Less(i, j int) bool {
	return (*h)[i].item.Less((*h)[j].item)
}

func (h *itemsHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
	(*h)[i].index = i
	(*h)[j].index = j
}

func (h *itemsHeap) Push(value interface{}) {
	item := value.(*wrapper)
	item.index = len(*h)
	*h = append(*h, item)
}

func (h *itemsHeap) Pop() interface{} {
	n := len(*h)
	item := (*h)[n-1]
	item.index = -1 // for safety
	*h = (*h)[0 : n-1]
	return item
}
