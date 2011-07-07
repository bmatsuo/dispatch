// Copyright 2011, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
/*
 *  Filename:    godirs.go
 *  Author:      Bryan Matsuo <bmatsuo@soe.ucsc.edu>
 *  Created:     Tue Jul  5 22:13:49 PDT 2011
 *  Description: 
 *  Usage:       godirs [options] ARGUMENT ...
 */

//  Package dispatch provides goroutine dispatch and concurrency limiting.
//  It provides an object Dispatch which is a queueing system for concurrent
//  functions. It implements a dynamic limit on the number of routines that
//  it runs simultaneously. It also uses a Queue interface, allowing for
//  alternate queue implementations.
//
//  See github.com/bmatsuo/dispatch/queues for more about the Queue interface.
//
//  See github.com/bmatsuo/dispatch/examples for usage examples.
package dispatch
import (
    "sync"
    "github.com/bmatsuo/dispatch/queues"
)

//  A Dispatch is an automated function dispatch queue with a limited
//  number of concurrent gorountines.
type Dispatch struct {
    // The maximum number of goroutines can be changed while the queue is
    // processing.
    MaxGo      int

    // Handle waiting when the limit of concurrent goroutines has been reached.
    waitingToRun bool
    //nextWake     chan bool
    nextWait     *sync.WaitGroup

    // Handle waiting when function queue is empty.
    waitingOnQ   bool
    restart      *sync.WaitGroup

    // Manage the Start()'ing of a Dispatch, avoiding race conditions.
    startLock    *sync.Mutex
    started      bool

    // Handle goroutine-safe queue operations.
    qLock        *sync.Mutex
    queue        queues.Queue

    // Handle goroutine-safe limiting and identifier operations.
    pLock        *sync.Mutex
    processing   int         // Number of QueueTasks running
    idcount      int64       // pid counter

    // Handle stopping of the Start() method.
    kill         chan bool
}

//  Create a new queue object with a specified limit on concurrency.
func New(maxroutines int) *Dispatch {
    return NewCustom(maxroutines, queues.NewFIFO())
}
func NewCustom(maxroutines int, queue queues.Queue) *Dispatch {
    var rl = new(Dispatch)
    rl.startLock = new(sync.Mutex)
    rl.qLock     = new(sync.Mutex)
    rl.pLock     = new(sync.Mutex)
    rl.restart   = new(sync.WaitGroup)
    rl.kill      = make(chan bool)
    //rl.nextWake  = make(chan bool)
    rl.nextWait  = new(sync.WaitGroup)
    rl.queue     = queue
    rl.MaxGo     = maxroutines
    rl.idcount   = 0
    return rl
}

//  Goroutines called from a Dispatch are given an int identifier unique
//  to that routine.
type StdTask struct {
    F func(id int64)
}
func (dt StdTask) Type() string {
    return "StdTask"
}
func (dt StdTask) SetFunc(f func(id int64)) {
    dt.F = f
}
func (dt StdTask) Func() func(id int64) {
    return dt.F
}
type dispatchTaskWrapper struct {
    id int64
    t  queues.Task
}
func (dtw dispatchTaskWrapper) Func() func(id int64) {
    return dtw.t.Func()
}
func (dtw dispatchTaskWrapper) Id() int64 {
    return dtw.id
}
func (dtw dispatchTaskWrapper) Task() queues.Task {
    return dtw.t
}

//  Enqueue a task for execution as a goroutine.
func (gq *Dispatch) Enqueue(t queues.Task) int64 {
    // Wrap the function so it works with the goroutine limiting code.
    var f = t.Func()
    var dtFunc = func (id int64) {
        // Run the given function.
        f(id)

        // Decrement the process counter.
        gq.pLock.Lock()
        gq.processing--
        var procWaiting = gq.waitingToRun
        if procWaiting {
            gq.waitingToRun = false
        }
        gq.pLock.Unlock()

        // Start any waiting process.
        if procWaiting {
            //gq.nextWake<-true
            gq.nextWait.Done()
        }
    }
    t.SetFunc(dtFunc)

    // Lock the queue and enqueue a new task.
    gq.qLock.Lock()
    gq.idcount++
    var id = gq.idcount
    gq.queue.Enqueue(dispatchTaskWrapper{id, t})
    var loopWaiting = gq.waitingOnQ
    if loopWaiting {
        gq.waitingOnQ = false
    }
    gq.qLock.Unlock()

    // Restart the Start() loop if it was deemed necessary.
    if loopWaiting {
        gq.restart.Done()
    }

    return id
}

//  Stop the queue after gq.Start() has been called. Any goroutines which
//  have not already been dequeued will not be executed until gq.Start()
//  is called again.
func (gq *Dispatch) Stop() {
    // Lock out Start() and queue ops for the entire call.
    gq.startLock.Lock()
    defer gq.startLock.Unlock()
    gq.qLock.Lock()
    defer gq.qLock.Unlock()

    if !gq.started {
        return
    }

    // Clear channel flags and close channels, stoping further processing.
    close(gq.kill)
    gq.started = false
    if gq.waitingOnQ {
        gq.waitingOnQ = false
        gq.restart.Done()
    }
    if gq.waitingToRun {
        gq.waitingToRun = false
        gq.nextWait.Done()
    }
    //close(gq.nextWake)
}

//  Start the next task in the queue. It's assumed that the queue is non-
//  empty. Furthermore, there should only be one goroutine in this method
//  (for this object) at a time. Both conditions are enforced in
//  gq.Start(), which calls gq.next() exclusively.
func (gq *Dispatch) next() {
    for true {
        // Attempt to start processing the file.
        gq.pLock.Lock()
        if gq.processing >= gq.MaxGo {
            gq.waitingToRun = true
            gq.nextWait.Add(1)
            gq.pLock.Unlock()
            gq.nextWait.Wait()
            /*
            var cont, ok =<-gq.nextWake
            if !ok {
                gq.nextWake = make(chan bool)
                return
            }
            if !cont {
                return
            }
            */
            continue
        }
        // Keep the books and reset wait time before unlocking.
        gq.waitingToRun = false
        gq.processing++
        gq.pLock.Unlock()

        // Get an element from the queue.
        gq.qLock.Lock()
        var wrapper = gq.queue.Dequeue().(queues.RegisteredTask)
        gq.qLock.Unlock()

        // Begin processing and asyncronously return.
        //var task = taskelm.Value.(dispatchTaskWrapper)
        var task = wrapper.Func()
        go task(wrapper.Id())
        return
    }
}

//  Start executing goroutines. Don't stop until gq.Stop() is called.
func (gq *Dispatch) Start() {
    // Avoid multiple gq.Start() methods and avoid race conditions.
    gq.startLock.Lock()
    if gq.started {
        panic("already started")
    }
    gq.started = true
    gq.startLock.Unlock()


    // Recreate any channels that were closed by a previous Stop().
    var inited = false
    for !inited {
        select {
        case _, okKill :=<-gq.kill:
            if !okKill {
                gq.kill = make(chan bool)
            }
        /*
        case _, okWake :=<-gq.nextWake:
            if !okWake {
                gq.nextWake = make(chan bool)
            }
        */
        default:
            inited = true
        }
    }

    // Process the queue
    for true {
        select {
        case die, ok :=<-gq.kill:
            // If something came out of this channel, we must stop.
            if !ok {
                // Recreate the channel on a closure.
                gq.kill = make(chan bool)
                return
            }
            if die {
                return
            }
        default:
            // Check the queue size and determine if we need to wait.
            gq.qLock.Lock()
            if gq.waitingOnQ = gq.queue.Len() == 0 ; gq.waitingOnQ {
                gq.restart.Add(1)
            }
            gq.qLock.Unlock()

            if gq.waitingOnQ {
                // Wait for a restart signal from gq.Enqueue
                gq.restart.Wait()
            } else {
                // Process the head of the queue and start the loop again.
                gq.next()
            }
        }
    }
}
