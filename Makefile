ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif
VERSION ?= "dev"
LDFLAGS=-ldflags "-s -w -X github.com/axllent/bwlog/cmd.Version=${VERSION}"

build = GOOS=$(1) GOARCH=$(2) go build ${LDFLAGS} -o dist/bwlog_${VERSION}_$(1)_$(2) \
	&& bzip2 -f dist/bwlog_${VERSION}_$(1)_$(2)

main-build: *.go
	go get github.com/gobuffalo/packr/packr
	${GOPATH}/bin/packr
	go build ${LDFLAGS} -o bwlog

clean:
	rm -rf dist bwlog *-packr.go

release:
	rm -f dist/bwlog_${VERSION}_*
	go get github.com/gobuffalo/packr/packr
	${GOPATH}/bin/packr
	$(call build,darwin,amd64)
	$(call build,freebsd,386)
	$(call build,freebsd,amd64)
	$(call build,freebsd,arm)
	$(call build,linux,386)
	$(call build,linux,amd64)
	$(call build,linux,arm)
	$(call build,linux,arm64)
	$(call build,netbsd,386)
	$(call build,netbsd,amd64)
	$(call build,netbsd,arm)
	$(call build,openbsd,386)
	$(call build,openbsd,amd64)
	$(call build,openbsd,arm)
	$(call build,solaris,amd64)
