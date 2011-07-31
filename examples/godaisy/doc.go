/*
 *  Filename:    doc.go
 *  Author:      Bryan Matsuo <bmatsuo@soe.ucsc.edu>
 *  Created:     Sun Jul 31 14:07:46 PDT 2011
 *  Description: 
 *  Usage:       godoc github.com/bmatsuo/daisy
 */

/*
daisy executes n parallel daisy-chained composite procedures, each of
length k. There are artifical delays in the various aspects of the program
to simulate real world timing conditions.

If G routines are allowed to run simulatiously, then n*G+1 communication
channels are used to synchronize the daisy-chain's links 'atomic' executions.

Usage:

    daisy [options]

Arguments:

    None

Options:

    -G=5
            Maximum number of parallel routines.

    -cd=1000000
            Delay (ns) creating each link.

    -k=30
            Length of each daisy-chain.

    -ld=5000000
            Delay (ns) in each link.

    -n=10
            Number of daisy-chains.

    -v=false
            Verbose program output.

*/
package documentation
