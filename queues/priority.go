package queues
/*
 *  Filename:    priority.go
 *  Package:     queues
 *  Author:      Bryan Matsuo <bmatsuo@soe.ucsc.edu>
 *  Created:     Wed Jul  6 22:18:57 PDT 2011
 *  Description: 
 */
import (
    "sort"
    "fmt"
    "container/heap"
    "container/vector"
)

//  A PrioritizedTask is a Task that also has key (float64). Generally,
//  a lower key means higher priority.
type PrioritizedTask interface {
    Task
    Key() float64
    SetKey(float64)
}
//  A structure that satisfies the PrioritizedTask interface (and thus
//  Task aswell).
type PTask struct {
    F func(int64)
    P float64
}
func (pt *PTask) Type() string {
    return "PTask"
}
func (pt *PTask) SetFunc(f func(int64)) {
    pt.F = f
}
func (pt *PTask) Func() func(int64) {
    return pt.F
}
func (pt *PTask) Key() float64 {
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
    switch x.(RegisteredTask).Task().(type) {
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

//  A heap-based priority queue. This implementation of a priority queue
//  is ideal for many situations involving a priority queue. However, other
//  priority queue implementations exist, each with their strengths and
//  weaknesses. See ArrayPriorityQueue and VectorPriorityQueue.
type PriorityQueue struct {
    h  *pQueue
}

//  Create a new heap-based priority queue.
func NewPriorityQueue() *PriorityQueue {
    var pq = new(PriorityQueue)
    pq.h = newPQueue()
    // No need to call heap.Init(pq.h) on an empty heap.
    return pq
}

//  The number of items in the queue.
func (pq *PriorityQueue) Len() int {
    return pq.h.Len()
}

//  Remove a task from the queue with runtime O(log(n)).
func (pq *PriorityQueue) Dequeue() RegisteredTask {
    if pq.Len() <= 0 {
        panic("empty")
    }
    return heap.Pop(pq.h).(RegisteredTask)
}

//  Add a task to the queue with runtime O(log(n)). The Task() method
//  of task must satisfy the PrioritizedTask interface, or a runtime
//  panic is thrown.
func (pq *PriorityQueue) Enqueue(task RegisteredTask) {
    switch task.Task().(type) {
    case PrioritizedTask:
        heap.Push(pq.h, task)
    default:
        panic(fmt.Sprintf("nokey %s", task.Task().Type()))
    }
}

//  Set a task's key with runtime O(n).
func (pq *PriorityQueue) SetKey(id int64, k float64) {
    var i, task = pq.h.FindId(id)
    if i < 0 {
        return
    }
    heap.Remove(pq.h, i)
    task.Task().(PrioritizedTask).SetKey(k)
    heap.Push(pq.h, task)
}

//  A priority queue based on the "container/vector" package. This priority
//  queue implementation has fast dequeues and slow enqueues. 
type VectorPriorityQueue struct {
    head   int
    hmax   int
    v *vector.Vector
}

// Create a new VectorPriorityQueue.
func NewVectorPriorityQueue() *VectorPriorityQueue {
    var vpq = new(VectorPriorityQueue)
    vpq.v = new(vector.Vector)
    vpq.hmax = 1
    return vpq
}

func (vpq *VectorPriorityQueue) Len() int {
    return vpq.v.Len() - vpq.head
}

//  Add a task to the priority queue in O(n) time. This is done with a
//  O(log(n)) binary search and an insert operation.
func (vpq *VectorPriorityQueue) Enqueue(task RegisteredTask) {
    switch task.Task().(type) {
    case PrioritizedTask:
        break
    default:
        panic(fmt.Sprintf("nokey %s", task.Task().Type()))
    }
    var key = task.Task().(PrioritizedTask).Key()
    var insertoffset = sort.Search(vpq.Len(), func(i int) bool {
            if vpq.v.At(vpq.head+i).(RegisteredTask).Task().(PrioritizedTask).Key() >= key {
                return true
            }
            return false })
    vpq.v.Insert(vpq.head+insertoffset, task)
}

//  Remove the task with the smallest key in O(1) amortized time.
func (vpq *VectorPriorityQueue) Dequeue() RegisteredTask {
    var front = vpq.v.At(vpq.head).(RegisteredTask)
    vpq.head++
    if vpq.head >= vpq.hmax {
        vpq.v.Cut(0, vpq.head)
        vpq.hmax *= 2
        vpq.head = 0
    }
    return front
}

//  Change the value of a task's key in O(n) time. This performs search,
//  delete, and enqueue operations. Hence, this is not a fast method.
func (vpq *VectorPriorityQueue) SetKey(id int64, k float64) {
    var (
        n    = vpq.Len()
        i    int
        task RegisteredTask
    )
    for i = vpq.head ; i < n ; i++ {
        task = vpq.v.At(i).(RegisteredTask)
        if task.Id() == id {
            break
        }
    }
    if i < n {
        vpq.v.Delete(i)
        task.Task().(PrioritizedTask).SetKey(k)
        vpq.Enqueue(task)
    }
}


//  An array-based priority queue with a constant time dequeue and a
//  linear time equeue. It should slightly outperform a
//  VectorPriorityQueue, but will likely be removed from the library
//  because it is requires more maintenance.
type ArrayPriorityQueue struct {
    v          []RegisteredTask
    head, tail int
}


//  Create a new array-based priority queue.
func NewArrayPriorityQueue() *ArrayPriorityQueue {
    var apq = new(ArrayPriorityQueue)
    apq.v = make([]RegisteredTask, 10)
    return apq
}

//  The number of items in the queue.
func (apq *ArrayPriorityQueue) Len() int {
    return apq.tail - apq.head
}

//  Add a task to the queue with runtime O(n) (on average n/2 + log_2(n))
func (apq *ArrayPriorityQueue) Enqueue(task RegisteredTask) {
    var key = task.Task().(PrioritizedTask).Key()
    var n = apq.Len()
    var insertoffset = sort.Search(
            n,
            func(i int)bool{
                return apq.v[apq.head+i].Task().(PrioritizedTask).Key() >= key } )
    if apq.tail != len(apq.v) {
        for j := apq.tail ; j > apq.head+insertoffset ; j-- {
            apq.v[j] = apq.v[j-1]
        }
        apq.v[apq.head+insertoffset] = task
        apq.tail++
        return
    }
    var newv = apq.v
    if apq.head <= len(apq.v)/2 {
        newv = make([]RegisteredTask, 2* len(apq.v))
    }
    copy(newv, apq.v[apq.head:apq.head+insertoffset])
    newv[insertoffset] = task
    copy(newv[insertoffset+1:], apq.v[apq.head+insertoffset:apq.tail])
    for i := apq.head ; i < apq.tail ; i++ {
        apq.v[i] = nil
    }
    apq.v = newv
    apq.head = 0
    apq.tail = n+1
}

//  Remove the next task with a runtime O(1).
func (apq *ArrayPriorityQueue) Dequeue() RegisteredTask {
    if apq.Len() == 0 {
        panic("empty")
    }
    var task = apq.v[apq.head]
    apq.v[apq.head] = nil
    apq.head++
    return task
}

//  Add a task to the queue with runtime O(n).
func (apq *ArrayPriorityQueue) SetKey(id int64, k float64) {
}
