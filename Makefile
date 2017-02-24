GOOS = windows
export GOOS
GOARCH ?= 386
export GOARCH

ifeq ($(GOARCH),386)
	suffix = 32
endif
ifeq ($(GOARCH),amd64)
	suffix = 64
endif

.PHONY: all depends

all: winipset_$(GOOS)_$(GOARCH).zip

depends:: $(GOPATH)/pkg/$(GOOS)_$(GOARCH)/github.com/lxn/walk.a
$(GOPATH)/pkg/$(GOOS)_$(GOARCH)/github.com/lxn/walk.a:
	go get github.com/lxn/walk

depends:: $(GOPATH)/pkg/$(GOOS)_$(GOARCH)/golang.org/x/text/encoding/japanese.a
$(GOPATH)/pkg/$(GOOS)_$(GOARCH)/golang.org/x/text/encoding/japanese.a:
	go get golang.org/x/text/encoding/japanese

$(GOPATH)/bin/rsrc:
	env -u GOOS -u GOARCH go get github.com/akavel/rsrc

rsrc.syso: winipset.manifest $(GOPATH)/bin/rsrc
	$(GOPATH)/bin/rsrc -manifest winipset.manifest -o rsrc.syso

winipset$(suffix).exe: rsrc.syso *.go depends
	go build -ldflags="-s -H windowsgui -X main.version=$(shell git describe --tags)" -o $@

winipset_$(GOOS)_$(GOARCH).zip: winipset$(suffix).exe
	7z a $@ $<
