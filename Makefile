
GOPATH := ${PWD}
export GOPATH
VERSION ?= "dev"
LDFLAGS=-ldflags "-s -w -X main.version=${VERSION}"

build = GOOS=$(1) GOARCH=$(2) go get github.com/rakyll/statik github.com/bvinc/go-sqlite-lite/sqlite3 github.com/gorilla/websocket && \
	bin/statik -src=http/ -f && \
	GOOS=$(1) GOARCH=$(2) go build ${LDFLAGS} -o dist/bwlog_${VERSION}_$(1)_$(2) \
	&& bzip2 -f dist/bwlog_${VERSION}_$(1)_$(2)

bwlog: bwlog.go
	go get github.com/rakyll/statik github.com/bvinc/go-sqlite-lite/sqlite3 github.com/gorilla/websocket
	bin/statik -src=http/ -f
	go build ${LDFLAGS} -o bwlog

clean:
	rm -rf bin pkg src bwlog

release:
	rm -f dist/bwlog_${VERSION}_*
	$(call build,linux,amd64)
	# $(call build,windows,amd64)
