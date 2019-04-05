
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

release:
	if [ ! -d "src/" ]; then go get github.com/rakyll/statik github.com/bvinc/go-sqlite-lite/sqlite3 github.com/gorilla/websocket; fi
	mkdir -p dist
	rm -f dist/bwlog_${VERSION}_*

	echo "Building binaries for ${VERSION}"

	echo "- linux-amd64"
	rm -rf src/github.com/bvinc/go-sqlite-lite/
	GOOS=linux GOARCH=amd64 go get github.com/bvinc/go-sqlite-lite/sqlite3
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o dist/bwlog_${VERSION}_linux_amd64
	bzip2 dist/bwlog_${VERSION}_linux_amd64

	echo "- linux-386"
	rm -rf src/github.com/bvinc/go-sqlite-lite/
	GOOS=linux GOARCH=386 go get github.com/bvinc/go-sqlite-lite/sqlite3
	GOOS=linux GOARCH=386 go build ${LDFLAGS} -o dist/bwlog_${VERSION}_linux_386
	bzip2 dist/bwlog_${VERSION}_linux_386
