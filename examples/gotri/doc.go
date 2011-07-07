/*
 *  Filename:    doc.go
 *  Author:      Bryan Matsuo <bmatsuo@soe.ucsc.edu>
 *  Created:     Tue Jul  5 22:13:49 PDT 2011
 *  Description: Documentation of the dispatch example 'gotri'
 *  Usage:       godoc github.com/bmatsuo/dispatch/examples/gotri
 */

/*
Gotri uses package dispatch to simulate management of a limited resource.
This is a systems problem, but it is phrased in friendlier terms. The
setting for the problem is that of a triathalon. This has been adapted
from a problem in Algorithm Design [2005] by Kleinberg and Tardos.

Imagine there a are a total of n althetes in a participating in a
triathalon. The triathalon consists of a bicycling segment, followed by a
swimming segment and a running segment.

Each athlete knows how long it will take them to complete the bicycling,
swimming, and running segments individually. But, the problem is that
the triathalon has the restriction that only k swimmers can be in the water
at a time for safety concerns.

The athletes are not competitive, and would simply like the triathalon to
finish in a timely manner ('finishing' meaning everying crossing the finish
line).

A FIFO queing of the athletes waiting to swim is not optimal.

One better solution to queue swimmers by the remaining time the need to
finish the triathalon. That is when Bob, a slow athlete, arrives at the
water Carol, a much faster athlete, will allow Bob to swim before she does.

I don't think this solution is not optimal, but it is a decent heuristic to
compare against a FIFO solution.

Usage:

    gotri [options]

Arguments:

    None

Options:

    -n=N
            Set the number of athletes to N.

    -k=K
            Set the max number of simultaneous swimmers.

    -b=BTIME
            Set the max number of seconds it takes an althete to finish
            the biking segment of the triathalon.

    -s=STIME
            Set the max number of seconds it takes an althete to finish
            the swimming segment of the triathalon.

    -r=RTIME
            Set the max number of seconds it takes an althete to finish
            the running segment of the triathalon.

    -f
            Use a FIFO queue instead of the preposed solution.

    -v
            Verbose output

*/
package documentation
