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
)

//  A Task is the interface satisfied by objects passed to a Dispatch.
type Task interface {
    SetFunc(func (id int64))
    Func() func (id int64)
    Type() string                 // Used mostly for debugging
}
//  A Task given to a Dispatch is given a unique id and becomes a
//  RegisteredTask.
type RegisteredTask interface {
    Task() Task
    Func() func (id int64)
    Id()   int64
}

func registeredTaskSearch(rts []RegisteredTask, less func(t RegisteredTask)bool) int {
    var (
        low  = 0
        high = len(rts)
        mid  = (high-low+1)/2
        t    RegisteredTask
    )
    /*
    if high == 0 {
        return 0
    }
    if less(rts[0]) {
        return 0
    }
    if !less(rts[high]) {
        return high
    }
    */
    for low < high {
        t = rts[mid]
        var leftSide = less(t)
        switch leftSide {
        case true:
            high = mid
        case false:
            low = mid
        }
        mid = low + (high-low+1)/2
    }
    return low
}

//  A Queue is a queue for RegisteredTasks, used by a Dispatch.
type Queue interface {
    Enqueue(task RegisteredTask)  // Insert a DispatchTask
    Dequeue() RegisteredTask      // Remove the next task.
    Len() int                     // Number of items to be processed.
    SetKey(int64, float64)        // Set a task's key (priority queues).
}

//  A naive First In First Out (FIFO) Queue.
type FIFO struct {
    head, tail  int
    length      int
    circ        []RegisteredTask
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

//  See Queue.
func (dq *FIFO) Len() int {
    return dq.length
}
//  See Queue.
func (dq *FIFO) Enqueue(task RegisteredTask) {
    var n = len(dq.circ)
    if dq.length == len(dq.circ) {
        // Copy the circular slice into a new slice with twice the length.
        var tmp = dq.circ
        dq.circ = make([]RegisteredTask, 2*n)
        for i := 0 ; i < n ; i++ {
            var j = (dq.head+i)%n
            dq.circ[i] = tmp[j]
            tmp[j] = nil
        }
        dq.head = 0
        dq.tail = n
    }
    dq.circ[dq.tail] = task
    dq.tail = (dq.tail+1)%n
    dq.length++
}
//  See Queue.
func (dq *FIFO) Dequeue() RegisteredTask {
    if dq.length == 0 {
        panic("empty")
    }
    var task = dq.circ[dq.head]
    dq.head = (dq.head+1)%dq.length
    dq.length--
    return task
}
//  Does nothing. See Queue.
func (dq *FIFO) SetKey(id int64, k float64) { }

//  A naive Last In First Out (LIFO) Queue (also known as a stack).
type LIFO struct {
    top    int
    stack   []RegisteredTask
}
//  Create a new LIFO.
func NewLIFO() *LIFO {
    var q = new(LIFO)
    q.stack = make([]RegisteredTask, 10)
    q.top = 0
    return q
}

//  See Queue.
func (dq *LIFO) Len() int {
    return dq.top
}
//  See Queue.
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
//  See Queue.
func (dq *LIFO) Dequeue() RegisteredTask {
    if dq.top == 0 {
        panic("empty")
    }
    dq.top--
    var task = dq.stack[dq.top]
    dq.stack[dq.top] = nil
    return task
}
//  Does nothing. See Queue.
func (dq *LIFO) SetKey(id int64, k float64) { }
