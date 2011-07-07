# Modified the basic makefiles referred to from the
# Go home page.
#
# Copyright 2009 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.inc

TARG=dispatch
GOFILES=\
        dispatch.go\

include $(GOROOT)/src/Make.pkg

exmake: force
	bash -c 'for ex in examples/*; do cd $$ex && gomake && cd ../..; done'
extest: force
	bash -c 'for ex in examples/*; do cd $$ex && gomake test && cd ../..; done'
exinstall: force
	bash -c 'for ex in examples/*; do cd $$ex && gomake install && cd ../..; done'
exclean: force
	bash -c 'for ex in examples/*; do cd $$ex && gomake clean && cd ../..; done'
exnuke: force
	bash -c 'for ex in examples/*; do cd $$ex && gomake nuke && cd ../..; done'
qmake: force
	cd queues && gomake
qtest: force
	cd queues && gomake test
qinstall: force
	cd queues && gomake install
qclean: force
	cd queues && gomake clean
qnuke: force
	cd queues && gomake nuke

allclean: clean qclean exclean force

force: ;
