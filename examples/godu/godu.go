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
    "log"
    //"strings"
    "github.com/bmatsuo/dispatch"
    "path/filepath"
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
    fs.BoolVar(&(opt.beClever), "c", false, "Use the clever Dispatch.")
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

type Walker struct {
    gq     *dispatch.Dispatch
    done   chan bool
    paths  []string
    sizes  []int64
    lock   *sync.Mutex
    wg     *sync.WaitGroup
}
func NewWalker() *Walker {
    var w    = new(Walker)
    w.gq     = dispatch.New(20)
    w.lock   = new(sync.Mutex)
    w.wg     = new(sync.WaitGroup)
    w.done   = make(chan bool)
    w.paths  = make([]string, 0, 1)
    w.sizes  = make([]int64, 0, 1)
    return w
}
func (w *Walker) VisitFile(path string, info *os.FileInfo) {
    var f = func (id int64) {
        log.Print("stating %s", path)
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
    w.gq.Enqueue(dispatch.StdTask{f})
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
