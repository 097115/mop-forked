# Copyright (c) 2013-2019 by Michael Dvorkin and contributors. All Rights Reserved.
# Use of this source code is governed by a MIT-style license that can
# be found in the LICENSE file.

VERSION = 1.0.0-forked-20220825

build:
	go build -x -o ./mop-$(shell go env GOOS)-$(shell go env GOARCH)

install:
	go install -x

buildall:
	GOOS=darwin  GOARCH=amd64 go build $(GOFLAGS) -o ./mop-$(VERSION)-osx-64
	GOOS=freebsd GOARCH=amd64 go build $(GOFLAGS) -o ./mop-$(VERSION)-freebsd-64
	GOOS=linux   GOARCH=386   go build $(GOFLAGS) -o ./mop-$(VERSION)-linux-32
	GOOS=linux   GOARCH=amd64 go build $(GOFLAGS) -o ./mop-$(VERSION)-linux-64
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o ./mop-$(VERSION)-windows-64.exe
	GOOS=windows GOARCH=386   go build $(GOFLAGS) -o ./mop-$(VERSION)-windows-32.exe

.PHONY: build buildall
