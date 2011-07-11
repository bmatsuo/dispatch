*dispatch version 0.1_02*

Package dispatch provides goroutine dispatch and concurrency limiting

About dispatch
=============

Package dispatch provides an object Dispatch which is a queueing system for
concurrent functions. It implements a dynamic limit on the number of
routines it is runs simultaneously. It also implements an interface Queue,
allowing for alternate queue implementations.

Performance
===========

Generally, you run concurrent processes for increased program speed. So,
you would like a dispatch method to be fast. However, the general purpose
nature of the dispatch package causes some necessary bloating underlying
the methods of Dispatch objects. For concurrent tasks which are doing 
actual work for more than a few hundred nanoseconds, this should not be
very noticeable.

Hovever, if you have very high performance expectations, you may be better
off writing your own lean and mean goroutine dispatcher that is suited for
your individual purposes.

Dependencies
=============

You must have Go installed (http://golang.org/). 

Documentation
=============
Installation
-------------

Use goinstall to install dispatch

    goinstall github.com/bmatsuo/dispatch

Examples
--------

You can usage examples by checking out the examples subdirectory. You can
run the compile all the examples to run the locally with the command

    cd $GOROOT/src/pkg/github.com/bmatsuo/dispatch && gomake exinstall && cd -

This installs all the examples. So, you can for instance run ```godu```
simply with the command

    godu

When you are done, remove the examples with the command

    cd $GOROOT/src/pkg/github.com/bmatsuo/dispatch && gomake exnuke && cd -


General Documentation
---------------------

Use godoc to vew the documentation for dispatch

    godoc github.com/bmatsuo/dispatch

Or alternatively, use a godoc http server

    godoc -http=:6060

Then, visit the following URLs for complete dispatch documentation

* http://localhost:6060/pkg/github.com/bmatsuo/dispatch

* http://localhost:6060/pkg/github.com/bmatsuo/dispatch/queues

* http://localhost:6060/pkg/github.com/bmatsuo/dispatch/examples

Author
======

Bryan Matsuo <bmatsuo@soe.ucsc.edu>

Copyright & License
===================

Copyright (c) 2011, Bryan Matsuo.
All rights reserved.

Use of this source code is governed by a BSD-style license that can be
found in the LICENSE file.
