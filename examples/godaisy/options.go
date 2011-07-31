package main
/*
 *  Filename:    options.go
 *  Author:      Bryan Matsuo <bmatsuo@soe.ucsc.edu>
 *  Created:     Sun Jul 31 14:07:46 PDT 2011
 *  Description: Parse arguments and options from the command line.
 */
import (
    "os"
    "flag"
)

type Options struct {
    maxgo      int
    linkdelay  int64
    chaindelay int64
    chains     int
    length     int
    verbose    bool
}

var opt = Options{}

func SetupFlags() *flag.FlagSet {
    var fs = flag.NewFlagSet("daisy", flag.ExitOnError)
    fs.IntVar(&(opt.maxgo), "G", 5, "Maximum number of parallel routines.")
    fs.IntVar(&(opt.chains), "n", 10, "Number of daisy-chains.")
    fs.IntVar(&(opt.length), "k", 30, "Length of each daisy-chain.")
    fs.Int64Var(&(opt.linkdelay), "ld", 5e6, "Delay (ns) in each link.")
    fs.Int64Var(&(opt.chaindelay), "cd", 1e6, "Delay (ns) creating each link.")
    fs.BoolVar(&(opt.verbose), "v", false, "Verbose program output.")
    return fs
}
func VerifyFlags(fs *flag.FlagSet) {
}
func ParseFlags() {
    var fs = SetupFlags()
    fs.Parse(os.Args[1:])
    VerifyFlags(fs)
    // Process the verified options...
}
