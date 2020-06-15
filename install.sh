#!/usr/bin/env sh

./pleasew build && \
	cp plz-out/bin/wollemi $(go env GOPATH)/bin/wollemi && \
	chmod 755 $GOPATH/bin/wollemi
