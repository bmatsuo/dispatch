// Copyright 2011, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main
/*
 *  Filename:    daisy.go
 *  Author:      Bryan Matsuo <bmatsuo@soe.ucsc.edu>
 *  Created:     Sun Jul 31 14:07:46 PDT 2011
 *  Description: 
 *  Usage:       daisy [options] ARGUMENT ...
 */
import (
    "github.com/bmatsuo/dispatch"
    "github.com/bmatsuo/container/bucket"
    "time"
    "sync"
    "fmt"
    "os"
)

//  A simple channel wrapper for use by the ChannelBucket object.
type ChannelElement struct {
    id  int
    c   chan bool
}

func (ce *ChannelElement) GetToken() {
    <-ce.c
}
func (ce *ChannelElement) InsertToken() {
    ce.c <- true
}

//  Hold a re-usable set of bool channels for signalling
type ChannelBucket struct {
    b *bucket.Bucket
}

func NewChannelBucket(n int) *ChannelBucket {
    cb := new(ChannelBucket)
    cb.b = bucket.New(n)
    cb.b.Init(func(i int) interface{} {
        return make(chan bool, 1)
    })
    return cb
}
func (cb *ChannelBucket) Retain() ChannelElement {
    id, v := cb.b.Retain()
    return ChannelElement{id, v.(chan bool)}

}
func (cb *ChannelBucket) Release(ce ChannelElement) {
    cb.b.Release(ce.id)
}

//  Create a link (function) that executes a given function f after the
//  previous link has finished executing.
func DaisyLink(f func(), start, stop ChannelElement, cb *ChannelBucket, wg *sync.WaitGroup) func(int64) {
    return func(id int64) {
        <-start.c
        cb.Release(start)
        f()
        stop.c <- true
        wg.Done()
    }
}
//  Given an id and index and output slice, generate a function to perform
//  That part of the workflow.
func ExampleDaisyFunc(id, i int, outPtr *[]int) func() {
    return func() {
        fmt.Fprintf(os.Stderr, "\r%d %d   \t", id, i)
        (*outPtr) = append(*outPtr, i)
        time.Sleep(opt.linkdelay)
    }
}
//  Start computation of a workflow of 'length' order dependent functions.
//  This simple example creates a list of numbers [0:length].
//      Assertions:
//          (1) Being run asynchronously.
//          (2) d.Start() has already been called.
func ExampleDaisyChain(length, id int, out [][]int, d *dispatch.Dispatch, cb *ChannelBucket, wg *sync.WaitGroup) {
    var (
        c1  = cb.Retain()
        c2  = cb.Retain()
    )
    c1.InsertToken()
    wg.Add(length)
    for i := 0; i < length; i++ {
        d.Enqueue(&dispatch.StdTask{DaisyLink(ExampleDaisyFunc(id, i, &(out[id])), c1, c2, cb, wg)})
        c1 = c2
        c2 = cb.Retain()
        time.Sleep(opt.chaindelay)
    }
    wg.Done()
}

// Run n daisy-chained k-task workflows simultaneously
func Example(n, k int) [][]int {
    numChan := n*opt.maxgo + 1 // > (# chains)*(# threads) [indep. of chain length]
    d := dispatch.New(opt.maxgo)
    cb := NewChannelBucket(numChan)
    wg := new(sync.WaitGroup)
    out := make([][]int, n)

    t1 := time.Nanoseconds()
    go d.Start()
    wg.Add(n)
    for id := 0; id < n; id++ {
        go ExampleDaisyChain(k, id, out, d, cb, wg)
    }
    wg.Wait()
    d.Stop()
    t2 := time.Nanoseconds()

    deltat := t2 - t1
    sec := deltat / 1e9
    frac := (deltat % 1e9) / 1e6

    fmt.Printf("Number of jobs %d; Maximum queue length %d; %d.%-3ds", n*k, d.MaxLength(), sec, frac)
    return out
}

func main() {
    ParseFlags()
    // Options are now stored in the global variable opt.

    out := Example(opt.chains, opt.length)
    fmt.Println()
    if opt.verbose {
        fmt.Println(out)
    }
}
