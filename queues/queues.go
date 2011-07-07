// Copyright 2011, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
/*
 *  Filename:    queue.go
 *  Package:     dispatch
 *  Author:      Bryan Matsuo <bmatsuo@soe.ucsc.edu>
 *  Created:     Wed Jul  6 17:30:20 PDT 2011
 *  Description: 
 */

//  Package queues defines the Queue interface used in package dispatch,
//  and several Queue implementations.
package queues
import (
    "container/list"
)

//  A Task is the interface satisfied by objects passed to a Dispatch.
type Task interface {
    Func() func (id int64)
}
//  A Task given to a Dispatch is given a unique id and becomes a
//  RegisteredTask.
type RegisteredTask interface {
    Task() Task
    Func() func (id int64)
    Id()   int64
}
//  A Queue is a queue for RegisteredTasks, used by a Dispatch.
type Queue interface {
    Enqueue(task RegisteredTask)  // Insert a DispatchTask
    Dequeue() RegisteredTask      // Remove the next task.
    Len() int                     // Number of items to be processed.
    SetKey(int64, float64)        // Set a task's key (priority queues).
}

//  A simple linked-list First In First Out (FIFO) Queue.
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
//  Does nothing. See Queue.
func (dq *FIFO) SetKey(id int64, k float64) { }

//  A simple linked-list Last In First Out (LIFO) Queue.
type LIFO struct {
    waiting      *list.List  // A list with values of type func(int)
}
//  Create a new LIFO.
func NewLIFO() *LIFO {
    var q = new(LIFO)
    q.waiting = list.New()
    return q
}

//  See Queue.
func (dq *LIFO) Len() int {
    return dq.waiting.Len()
}
//  See Queue.
func (dq *LIFO) Enqueue(task RegisteredTask) {
    dq.waiting.PushFront(task)
}
//  See Queue.
func (dq *LIFO) Dequeue() RegisteredTask {
    var taskelm = dq.waiting.Front()
    dq.waiting.Remove(taskelm)
    return taskelm.Value.(RegisteredTask)
}
//  Does nothing. See Queue.
func (dq *LIFO) SetKey(id int64, k float64) { }
