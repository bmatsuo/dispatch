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
    "flag"
)

type Options struct {
    verbose bool
}
var opt = Options{}
func SetupFlags() *flag.FlagSet {
    var fs = flag.NewFlagSet("godirs", flag.ExitOnError)
    fs.BoolVar(&(opt.verbose), "v", false, "Verbose program output.")
    return fs
}
func VerifyFlags() {
}
func ParseFlags() {
    var fs = SetupFlags()
    fs.Parse(os.Args[1:])
}

func main() {
    ParseFlags()
}
