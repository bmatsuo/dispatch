package queues
/*
 *  Filename:    priority.go
 *  Package:     queues
 *  Author:      Bryan Matsuo <bmatsuo@soe.ucsc.edu>
 *  Created:     Wed Jul  6 22:18:57 PDT 2011
 *  Description: 
 */
import (
    //"os"
    "fmt"
    "container/heap"
    "container/vector"
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
        panic(fmt.Sprintf("nokey %s", task.Task().Type()))
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

//  A priority queue based on the "container/vector" package.
//  Ideally, an array-based priority queue implementation should have
//  fast dequeues and slow enqueues. I fear the vector.Vector class
//  gives slow equeues and slow dequeues.
type VectorPriorityQueue struct {
    v *vector.Vector
}

func NewVectorPriorityQueue() *VectorPriorityQueue {
    var vpq = new(VectorPriorityQueue)
    vpq.v = new(vector.Vector)
    return vpq
}

func (vpq *VectorPriorityQueue) Len() int {
    return vpq.v.Len()
}
type etypeStopIter struct {
}
func (e etypeStopIter) String() string {
    return "STOPITER"
}
func (vpq *VectorPriorityQueue) Enqueue(task RegisteredTask) {
    switch task.Task().(type) {
    case PrioritizedTask:
        break
    default:
        panic(fmt.Sprintf("nokey %s", task.Task().Type()))
    }
    var i int
    defer func() {
        if r := recover(); r != nil {
            switch r.(type) {
            case etypeStopIter:
                break
            default:
                panic(r)
            }
        }
        vpq.v.Insert(i, task)
    } ()
    vpq.v.Do(func (telm interface{}) {
        if task.Task().(PrioritizedTask).Key() > telm.(RegisteredTask).Task().(PrioritizedTask).Key() {
            i++
        } else {
            panic(etypeStopIter{})
        }
    })
}
func (vpq *VectorPriorityQueue) Dequeue() RegisteredTask {
    var head = vpq.v.At(0).(RegisteredTask)
    vpq.v.Delete(0)
    return head
}
func (vpq *VectorPriorityQueue) SetKey(id int64, k float64) {
    var i int
    defer func() {
        if r := recover(); r != nil {
            switch r.(type) {
            case etypeStopIter:
                var rtask = vpq.v.At(i).(RegisteredTask)
                vpq.v.Delete(i)
                rtask.Task().(PrioritizedTask).SetKey(k)
                vpq.Enqueue(rtask)
            default:
                panic(r)
            }
        }
    } ()
    vpq.v.Do(func (telm interface{}) {
        if telm.(RegisteredTask).Id() != id {
            i++
        } else {
            panic(etypeStopIter{})
        }
    })
}

type ArrayPriorityQueue struct {
    v          []RegisteredTask
    head, tail int
}

func NewArrayPriorityQueue() *ArrayPriorityQueue {
    var apq = new(ArrayPriorityQueue)
    apq.v = make([]RegisteredTask, 10)
    return apq
}

func (apq *ArrayPriorityQueue) Len() int {
    return apq.tail - apq.head
}

func (apq *ArrayPriorityQueue) Enqueue(task RegisteredTask) {
    var key = task.Task().(PrioritizedTask).Key()
    var insertoffset = registeredTaskSearch(
            apq.v[apq.head:apq.tail],
            func(t RegisteredTask) bool {
                return t.Task().(PrioritizedTask).Key() < key
            })
    if apq.tail != len(apq.v) {
        for j := apq.tail ; j > insertoffset ; j-- {
            apq.v[j] = apq.v[j-1]
        }
        apq.v[insertoffset] = task
        apq.tail++
        return
    }
    var newv = apq.v
    if apq.head < len(apq.v)/2 {
        newv = make([]RegisteredTask, 2* len(apq.v))
    }
    var i, j int
    j = 0
    for i = apq.head ; i < apq.tail ; i++ {
        if apq.v[i].Task().(PrioritizedTask).Key() > key {
            break
        } else {
            newv[j] = apq.v[i]
            apq.v[i] = nil
        }
        j++
    }
    //fmt.Fprintf(os.Stderr, "Length %d index %d\n", len(newv), j)
    newv[j] = task
    j++
    for ; i < apq.tail ; i++ {
        newv[j] = apq.v[i]
        apq.v[i] = nil
        j++
    }
    apq.v = newv
    apq.head = 0
    apq.tail = j
}

func (apq *ArrayPriorityQueue) Dequeue() RegisteredTask {
    if apq.Len() == 0 {
        panic("empty")
    }
    var task = apq.v[apq.head]
    apq.v[apq.head] = nil
    apq.head++
    return task
}

func (apq *ArrayPriorityQueue) SetKey(id int64, k float64) {
}
