# Modified the basic makefiles referred to from the
# Go home page.
#
# Copyright 2009 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.inc

TARG=queues
GOFILES=\
		queue.go\

include $(GOROOT)/src/Make.pkg

exinstall: force
	bash -c 'for ex in examples/*; do cd $$ex && gomake install && cd ../..; done'

exnuke: force
	bash -c 'for ex in examples/*; do cd $$ex && gomake nuke && cd ../..; done'

force: ;
