
GOPATH := ${PWD}
export GOPATH
TAG=`git describe --tags`
VERSION ?= `git describe --tags`
LDFLAGS=-ldflags "-s -w -X main.version=${VERSION}"

bwlog: bwlog.go
	if [ ! -d "src/github.com/bvinc/" ]; then go get github.com/bvinc/go-sqlite-lite/sqlite3; fi
	go build ${LDFLAGS}

clean:
	rm -rf pkg src bwlog

