// Copyright 2011, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main
/*
 *  Filename:    godirs.go
 *  Author:      Bryan Matsuo <bmatsuo@soe.ucsc.edu>
 *  Created:     Tue Jul  5 22:13:49 PDT 2011
 *  Description: 
 *  Usage:       godirs [options] ARGUMENT ...
 */
import (
    "os"
    "fmt"
    "flag"
    "sync"
    "time"
    //"log"
    //"strings"
    "path/filepath"
    "container/list"
)
const (
    stdDelay = 50e3
)

type Options struct {
    list     func(string) ([]string,[]int64)
    beClever bool
    dir      string
    verbose  bool
}
var opt = Options{}
func SetupFlags() *flag.FlagSet {
    var fs = flag.NewFlagSet("godirs", flag.ExitOnError)
    fs.BoolVar(&(opt.beClever), "c", false, "Use the clever GoQueue.")
    fs.BoolVar(&(opt.verbose), "v", false, "Verbose program output.")
    return fs
}
func VerifyFlags(fs *flag.FlagSet) {
    if fs.NArg() < 1 {
        fmt.Fprintf(os.Stderr, "missing DIR argument.")
    }
}
func ParseFlags() {
    var fs = SetupFlags()
    fs.Parse(os.Args[1:])
    VerifyFlags(fs)
    opt.dir = fs.Arg(0)
    if opt.beClever {
        opt.list = WalkerList
    } else {
        opt.list = StupidWalkerList
    }
}

//  A GoQueue is a function queue that allows limiting number of concurrent
//  goroutines.
type GoQueue struct {
    // The maximum number of goroutines can be changed while the queue is
    // processing.
    MaxGo      int

    // Handle waiting when the limit of concurrent goroutines has been reached.
    waitingToRun bool
    nextWake     chan bool

    // Handle waiting when function queue is empty.
    waitingOnQ   bool
    restart      chan bool

    // Manage the Start()'ing of a GoQueue, avoiding race conditions.
    startLock    *sync.Mutex
    started      bool

    // Handle goroutine-safe queue operations.
    qLock        *sync.Mutex
    waiting      *list.List  // A list with values of type func(int)

    // Handle goroutine-safe limiting and identifier operations.
    pLock        *sync.Mutex
    processing   int         // Number of QueueTasks running
    idcount      int64       // pid counter

    // Handle stopping of the Start() method.
    kill         chan bool
}

//  Create a new queue object with a specified limit on concurrency.
func NewGoQueue(maxroutines int) *GoQueue {
    var rl = new(GoQueue)
    rl.startLock = new(sync.Mutex)
    rl.qLock     = new(sync.Mutex)
    rl.pLock     = new(sync.Mutex)
    rl.restart   = make(chan bool)
    rl.kill      = make(chan bool)
    rl.nextWake  = make(chan bool)
    rl.waiting   = list.New()
    rl.MaxGo     = maxroutines
    rl.idcount   = 0
    return rl
}

//  Goroutines called from a GoQueue are given an int identifier unique
//  to that routine.
type GQTask func(id int64)
type goQueueTask struct {
    id int64
    f  GQTask
}

//  Enqueue a task for execution as a goroutine.
func (gq *GoQueue) Enqueue(f GQTask) int64 {
    // Wrap the function so it works with the goroutine limiting code.
    var gqtFunc = func (id int64) {
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
            gq.nextWake<-true
        }
    }

    // Lock the queue and enqueue a new task.
    gq.qLock.Lock()
    gq.idcount++
    var id = gq.idcount
    gq.waiting.PushBack(goQueueTask{id, gqtFunc})
    var loopWaiting = gq.waitingOnQ
    if loopWaiting {
        gq.waitingOnQ = false
    }
    gq.qLock.Unlock()

    // Restart the Start() loop if it was deemed necessary.
    if loopWaiting {
        gq.restart<-true
    }

    return id
}

//  Stop the queue after gq.Start() has been called. Any goroutines which
//  have not already been dequeued will not be executed until gq.Start()
//  is called again.
func (gq *GoQueue) Stop() {
    // Lock out Start() and queue ops for the entire call.
    gq.startLock.Lock()
    defer gq.startLock.Unlock()
    gq.qLock.Lock()
    defer gq.qLock.Unlock()

    if !gq.started {
        return
    }

    // Clear channel flags and close channels, stoping further processing.
    gq.started = false
    gq.waitingToRun = false
    gq.waitingOnQ = false
    close(gq.restart)
    close(gq.kill)
    close(gq.nextWake)
}

//  Start the next task in the queue. It's assumed that the queue is non-
//  empty. Furthermore, there should only be one goroutine in this method
//  (for this object) at a time. Both conditions are enforced in
//  gq.Start(), which calls gq.next() exclusively.
func (gq *GoQueue) next() {
    for true {
        // Attempt to start processing the file.
        gq.pLock.Lock()
        if gq.processing >= gq.MaxGo {
            // Too many threads, wait and try again.
            gq.waitingToRun = true
            gq.pLock.Unlock()
            var cont, ok =<-gq.nextWake
            if !ok {
                gq.nextWake = make(chan bool)
                return
            }
            if !cont {
                return
            }
            continue
        }
        // Keep the books and reset wait time before unlocking.
        gq.waitingToRun = false
        gq.processing++
        gq.pLock.Unlock()

        // Get an element from the queue.
        gq.qLock.Lock()
        var taskelm = gq.waiting.Front()
        gq.waiting.Remove(taskelm)
        gq.qLock.Unlock()

        // Begin processing and asyncronously return.
        var task = taskelm.Value.(goQueueTask)
        go task.f(task.id)
        return
    }
}

//  Start executing goroutines. Don't stop until gq.Stop() is called.
func (gq *GoQueue) Start() {
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
        case _, okRestart :=<-gq.restart:
            if !okRestart {
                gq.restart = make(chan bool)
            }
        case _, okWake :=<-gq.nextWake:
            if !okWake {
                gq.restart = make(chan bool)
            }
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
            gq.waitingOnQ = gq.waiting.Len() == 0
            gq.qLock.Unlock()

            if !gq.waitingOnQ {
                // Process the head of the queue and start the loop again.
                gq.next()
                continue
            }

            // Wait for a restart signal from gq.Enqueue
            var cont, ok =<-gq.restart
            if !ok {
                gq.restart = make(chan bool)
                return
            }
            if !cont {
                return
            }
        }
    }
}

type Walker struct {
    gq  *GoQueue
    done   chan bool
    paths  []string
    sizes  []int64
    lock   *sync.Mutex
    wg     *sync.WaitGroup
}
func NewWalker() *Walker {
    var w    = new(Walker)
    w.gq     = NewGoQueue(20)
    w.lock   = new(sync.Mutex)
    w.wg     = new(sync.WaitGroup)
    w.done   = make(chan bool)
    w.paths  = make([]string, 0, 1)
    w.sizes  = make([]int64, 0, 1)
    return w
}
func (w *Walker) VisitFile(path string, info *os.FileInfo) {
    var f = func (id int64) {
        var stat, err = os.Stat(path)
        if err != nil {
            panic(err)
        }
        time.Sleep(stdDelay)
        w.lock.Lock()
        w.sizes = append(w.sizes, stat.Size)
        w.paths = append(w.paths, path)
        w.lock.Unlock()
        w.wg.Done()
    }
    w.wg.Add(1)
    w.gq.Enqueue(f)
}
func (w *Walker) VisitDir(path string, f *os.FileInfo) bool {
    return true
}
func WalkerList(dir string) ([]string, []int64) {
    var w = NewWalker()
    var errors = make(chan os.Error)
    go func() {
        for e := range errors {
            panic("Walk error: " + e.String())
        }
    } ()
    go w.gq.Start()
    filepath.Walk(dir, w, errors)
    w.wg.Wait()
    return w.paths, w.sizes
}

type StupidWalker struct {
    paths  []string
    sizes  []int64
}
func NewStupidWalker() *StupidWalker {
    var sw    = new(StupidWalker)
    sw.paths  = make([]string, 0, 1)
    sw.sizes  = make([]int64, 0, 1)
    return sw
}
func (sw *StupidWalker) VisitFile(path string, info *os.FileInfo) {
    var stat, err = os.Stat(path)
    if err != nil {
        panic(err)
    }
    time.Sleep(stdDelay)
    sw.sizes = append(sw.sizes, stat.Size)
    sw.paths = append(sw.paths, path)
}
func (sw *StupidWalker) VisitDir(path string, f *os.FileInfo) bool {
    return true
}
func StupidWalkerList(dir string) ([]string, []int64) {
    var sw = NewStupidWalker()
    var errors = make(chan os.Error)
    var done = make(chan bool)
    go func() {
        for e := range errors {
            panic("Walk error: " + e.String())
        }
        done<-true
    } ()
    filepath.Walk(dir, sw, errors)
    close(errors)
    <-done
    return sw.paths, sw.sizes
}

func main() {
    ParseFlags()
    var paths, sizes = opt.list(opt.dir)
    for i, path := range paths {
        fmt.Printf("\t%50s %d\n", path, sizes[i])
    }
}

/*
type recLister struct {
    BaseWait   int64
    MaxProc    int
    waittime   int64
    maxwait    int64
    processing int
    mutex      *sync.Mutex
    errors     chan os.Error
    done       chan bool
    paths      []string
    sizes      []int64
}
func newRecLister() *recLister {
    var rl = new(recLister)
    rl.mutex    = new(sync.Mutex)
    rl.errors   = make(chan os.Error)
    rl.done     = make(chan bool)
    rl.paths    = make([]string, 0, 1)
    rl.sizes    = make([]int64, 0, 1)
    rl.MaxProc  = 5
    rl.BaseWait = 10
    rl.waittime = rl.BaseWait
    rl.maxwait  = 500e6
    return rl
}
func (rl *recLister) VisitDir(path string, f *os.FileInfo) bool {
    //rl.paths = append(rl.paths, path)
    return true
}
func (rl *recLister) visitFile(path string, f *os.FileInfo) {
    var stat, err = os.Stat(path)
    if err != nil {
        panic(err)
    }
    rl.sizes = append(rl.sizes, stat.Size)
    rl.paths = append(rl.paths, path)
    //time.Sleep(5e9)
    rl.mutex.Lock()
    rl.processing--
    rl.mutex.Unlock()
}
func (rl *recLister) VisitFile(path string, f *os.FileInfo) {
    for true {
        // Attempt to start processing the file.
        rl.mutex.Lock()
        if rl.processing >= rl.MaxProc {
            // Too many threads, wait and try again.
            rl.waittime <<= 2
            if rl.waittime > rl.maxwait {
                rl.waittime = rl.maxwait
            }
            rl.mutex.Unlock()
            time.Sleep(rl.waittime)
            continue
        }
        // Keep the books and reset wait time before unlocking and processing.
        rl.processing++
        rl.waittime = rl.BaseWait
        rl.mutex.Unlock()
        go rl.visitFile(path, f)
        return
    }
}

func RecFileList(dir string) ([]string, []int64) {
    var rl = newRecLister()
    go func() {
        for e := range rl.errors {
            panic("Walk error: " + e.String())
        }
        rl.done <- true
    } ()
    filepath.Walk(dir, rl, rl.errors)
    close(rl.errors)
    <-rl.done
    return rl.paths, rl.sizes
}
*/
