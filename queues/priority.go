package queues
/*
 *  Filename:    priority.go
 *  Package:     queues
 *  Author:      Bryan Matsuo <bmatsuo@soe.ucsc.edu>
 *  Created:     Wed Jul  6 22:18:57 PDT 2011
 *  Description: 
 */
import (
    "fmt"
    "container/heap"
)

type PrioritizedTask interface {
    Task
    Key() float64
    SetKey(float64)
}
type PTask struct {
    F func(int64)
    P float64
}
func (pt PTask) Func() func(int64) {
    return pt.F
}
func (pt PTask) Key() float64 {
    return pt.P
}
func (pt *PTask) SetKey(k float64) {
    pt.P = k
}

type pQueue struct {
    elements []RegisteredTask
}
func newPQueue() *pQueue {
    var h = new(pQueue)
    h.elements = make([]RegisteredTask, 0, 5)
    return h
}
func (h *pQueue) GetPTask(i int) PrioritizedTask {
    if n := len(h.elements) ; i < 0 || i >= n {
        panic("badindex")
    }
    return h.elements[i].Task().(PrioritizedTask)
}
func (h *pQueue) Len() int {
    return len(h.elements)
}
func (h *pQueue) Less(i, j int) bool {
    return h.GetPTask(i).Key() < h.GetPTask(j).Key()
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
        h.elements = append(h.elements, x.(RegisteredTask))
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
func (h *pQueue) FindId(id int64) (int, RegisteredTask) {
    for i, elm := range h.elements {
        if elm.Id() == id {
            return i, elm
        }
    }
    return -1, nil
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
    switch task.Task().(type) {
    case PrioritizedTask:
        heap.Push(pq.h, task)
    default:
        panic(fmt.Sprintf("nokey %v", task))
    }
}
func (pq *PriorityQueue) SetKey(id int64, k float64) {
    var i, task = pq.h.FindId(id)
    if i < 0 {
        return
    }
    heap.Remove(pq.h, i)
    task.Task().(PrioritizedTask).SetKey(k)
    heap.Push(pq.h, task)
}
