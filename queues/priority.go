package queues
/*
 *  Filename:    priority.go
 *  Package:     queues
 *  Author:      Bryan Matsuo <bmatsuo@soe.ucsc.edu>
 *  Created:     Wed Jul  6 22:18:57 PDT 2011
 *  Description: 
 */
import (
    "container/heap"
)

type PrioritizedTask interface {
    RegisteredTask
    Key() float64
}
type pQueue struct {
    elements []PrioritizedTask
}
func newPQueue() *pQueue {
    var h = new(pQueue)
    h.elements = make([]PrioritizedTask, 0, 5)
    return h
}
func (h *pQueue) Len() int {
    return len(h.elements)
}
func (h *pQueue) Less(i, j int) bool {
    if n := len(h.elements) ; i < 0 || i >= n || j < 0 || j >= n {
        panic("badindex")
    }
    return h.elements[i].Key() < h.elements[j].Key()
}
func (h *pQueue) Swap(i, j int) {
    if n := len(h.elements) ; i < 0 || i >=n || j < 0 || j >= n {
        panic("badindex")
    }
    var tmp = h.elements[i]
    h.elements[i] = h.elements[j]
    h.elements[j] = tmp
}
func (h *pQueue) Push(x interface{}) {
    switch x.(type) {
    case PrioritizedTask:
        h.elements = append(h.elements, x.(PrioritizedTask))
    default:
        panic("badtype")
    }
}
func (h *pQueue) Pop() interface{} {
    if len(h.elements) <= 0 {
        panic("empty")
    }
    var head = h.elements[0]
    h.elements = h.elements[1:]
    return head
}

type PriorityQueue struct {
    h  *pQueue
}

func NewPriorityQueue() *PriorityQueue {
    var pq = new(PriorityQueue)
    pq.h = newPQueue()
    // No need to call heap.Init(pq.h) on an empty heap.
    return pq
}

func (pq *PriorityQueue) Len() int {
    return pq.h.Len()
}
func (pq *PriorityQueue) Dequeue() RegisteredTask {
    if pq.Len() <= 0 {
        panic("empty")
    }
    return heap.Pop(pq.h).(RegisteredTask)
}
func (pq *PriorityQueue) Enqueue(task RegisteredTask) {
    heap.Push(pq.h, task)
}
