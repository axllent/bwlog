export GOPATH=${PWD}

CGO_ENABLED=1 go get \
github.com/rakyll/statik \
github.com/bvinc/go-sqlite-lite/sqlite3 \
github.com/gorilla/websocket

# Regenerate static files
bin/statik -src=http/ -f
CGO_ENABLED=1 go run *.go $@
