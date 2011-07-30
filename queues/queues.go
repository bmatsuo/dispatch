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

import ()

//  A Task is the interface satisfied by objects passed to a Dispatch.
type Task interface {
    SetFunc(func(id int64))
    Func() func(id int64)
    Type() string // Used mostly for debugging
}
//  A Task given to a Dispatch is given a unique id and becomes a
//  RegisteredTask.
type RegisteredTask interface {
    Task() Task
    Func() func(id int64)
    Id() int64
}

//  A Queue is a queue for RegisteredTasks, used by a Dispatch. Queue
//  objects can be priority queues or not, but they all must implement
//  a method SetKey(...). For non-priority queues, that method should
//  just return immediately.
//
//  To avoid race conditions, when Queue methods are called by a Dispatch,
//  the Dispatch locks the queue and prevents any other methods from being
//  called on it. This is something to think about when creating/choosing
//  a Queue implementation.
type Queue interface {
    Enqueue(task RegisteredTask) // Insert a task
    Dequeue() RegisteredTask     // Remove the next task.
    Len() int                    // Number of items waiting for processing.
    SetKey(int64, float64)       // Set a task's key (priority queues).
}

//  A First In First Out (FIFO) Queue implemented as a circular slice.
type FIFO struct {
    head, tail int
    length     int
    circ       []RegisteredTask
}

//  Create a new FIFO.
func NewFIFO() *FIFO {
    var q = new(FIFO)
    q.circ = make([]RegisteredTask, 10)
    q.head = 0
    q.tail = 0
    q.length = 0
    return q
}

func (dq *FIFO) Len() int {
    return dq.length
}

//  Add a task in O(1) amortized time.
func (dq *FIFO) Enqueue(task RegisteredTask) {
    var n = len(dq.circ)
    if dq.length == len(dq.circ) {
        // Copy the circular slice into a new slice with twice the length.
        var tmp = dq.circ
        dq.circ = make([]RegisteredTask, 2*n)
        mid := copy(dq.circ, tmp[dq.head:])
        end := copy(dq.circ[mid:], tmp[:dq.tail])
        dq.head = 0
        dq.tail = mid + end // This should be equal to n.
    }
    dq.circ[dq.tail] = task
    dq.tail = (dq.tail + 1) % n
    dq.length++
}

//  Dequeue a task in O(1) time.
func (dq *FIFO) Dequeue() RegisteredTask {
    if dq.length == 0 {
        panic("empty")
    }
    var task = dq.circ[dq.head]
    var zero RegisteredTask
    dq.circ[dq.head] = zero
    dq.head = (dq.head + 1) % dq.length
    dq.length--
    return task
}

//  Does nothing. See Queue.
func (dq *FIFO) SetKey(id int64, k float64) {}

//  A Last In First Out (LIFO) Queue (also known as a stack) implemented
//  with a slice.
type LIFO struct {
    top   int
    stack []RegisteredTask
}

//  Create a new LIFO.
func NewLIFO() *LIFO {
    var q = new(LIFO)
    q.stack = make([]RegisteredTask, 10)
    q.top = 0
    return q
}

func (dq *LIFO) Len() int {
    return dq.top
}

//  Enqueue (push) a task on the LIFO in O(1) amortized time.
func (dq *LIFO) Enqueue(task RegisteredTask) {
    var n = len(dq.stack)
    if dq.top == n {
        var tmpstack = dq.stack
        dq.stack = make([]RegisteredTask, 2*n)
        copy(dq.stack, tmpstack)
    }
    dq.stack[dq.top] = task
    dq.top++
}

//  Dequeue (pop) a task off the LIFO in O(1) time.
func (dq *LIFO) Dequeue() RegisteredTask {
    if dq.top == 0 {
        panic("empty")
    }
    dq.top--
    var task = dq.stack[dq.top]
    var zero RegisteredTask
    dq.stack[dq.top] = zero
    return task
}

//  Does nothing. See Queue.
func (dq *LIFO) SetKey(id int64, k float64) {}
