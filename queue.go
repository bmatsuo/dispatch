package dispatch
/*
 *  Filename:    queue.go
 *  Package:     dispatch
 *  Author:      Bryan Matsuo <bmatsuo@soe.ucsc.edu>
 *  Created:     Wed Jul  6 17:30:20 PDT 2011
 *  Description: 
 */
import (
    "container/list"
)

//  A Queue is a queue for RegisteredTasks, used by a Dispatch.
type Queue interface {
    Enqueue(task RegisteredTask)  // Insert a DispatchTask
    Dequeue() RegisteredTask      // Remove the next task.
    Len() int                     // Number of items to be processed.
}

//  A simple linked-list FIFO satisfying the Queue interface.
type FIFO struct {
    waiting      *list.List  // A list with values of type func(int)
}
//  Create a new FIFO.
func NewFIFO() *FIFO {
    var q = new(FIFO)
    q.waiting = list.New()
    return q
}

//  See Queue.
func (dq *FIFO) Len() int {
    return dq.waiting.Len()
}
//  See Queue.
func (dq *FIFO) Enqueue(task RegisteredTask) {
    dq.waiting.PushBack(task)
}
//  See Queue.
func (dq *FIFO) Dequeue() RegisteredTask {
    var taskelm = dq.waiting.Front()
    dq.waiting.Remove(taskelm)
    return taskelm.Value.(RegisteredTask)
}
