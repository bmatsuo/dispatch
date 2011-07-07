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

//  A DefaultQueue is a FIFO Queue that is implemented
//  using a linked list.
type DefaultQueue struct {
    waiting      *list.List  // A list with values of type func(int)
}
//  Create a new DefaultQueue.
func NewQueue() *DefaultQueue {
    var q = new(DefaultQueue)
    q.waiting = list.New()
    return q
}

//  See Queue.
func (dq *DefaultQueue) Len() int {
    return dq.waiting.Len()
}
//  See Queue.
func (dq *DefaultQueue) Enqueue(task RegisteredTask) {
    dq.waiting.PushBack(task)
}
//  See Queue.
func (dq *DefaultQueue) Dequeue() RegisteredTask {
    var taskelm = dq.waiting.Front()
    dq.waiting.Remove(taskelm)
    return taskelm.Value.(RegisteredTask)
}
