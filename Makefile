
GOPATH := ${PWD}
export GOPATH
VERSION ?= "dev"
LDFLAGS=-ldflags "-s -w -X main.version=${VERSION}"

bwlog: bwlog.go
	if [ ! -d "src/" ]; then go get github.com/rakyll/statik github.com/bvinc/go-sqlite-lite/sqlite3 github.com/gorilla/websocket; fi
	bin/statik -src=http/ -f
	go build ${LDFLAGS}

clean:
	rm -rf pkg src bwlog
