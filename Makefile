
GOPATH := ${PWD}
export GOPATH
VERSION ?= "dev"
LDFLAGS=-ldflags "-s -extldflags \"--static\" -w -X main.version=${VERSION}"

build = CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build ${LDFLAGS} -o dist/bwlog_${VERSION}_$(1)_$(2) \
	&& bzip2 -f dist/bwlog_${VERSION}_$(1)_$(2)

bwlog: bwlog.go
	CGO_ENABLED=0 go get github.com/axllent/gitrel github.com/rakyll/statik github.com/gorilla/websocket
	rm -rf statik
	bin/statik -src=web/ -f
	CGO_ENABLED=1 go build ${LDFLAGS} -o bwlog

clean:
	rm -rf bin dist pkg src statik bwlog

release:
	rm -f dist/bwlog_${VERSION}_*
	go get github.com/axllent/gitrel github.com/rakyll/statik github.com/gorilla/websocket
	rm -rf statik
	bin/statik -src=web/ -f
	$(call build,darwin,386)
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
