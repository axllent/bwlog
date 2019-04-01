export GOPATH=${PWD}
if [ ! -d "src/" ]; then
	go get \
	github.com/rakyll/statik \
	github.com/bvinc/go-sqlite-lite/sqlite3 \
	github.com/gorilla/websocket
fi
# Regenerate static files
bin/statik -src=http/ -f
go run *.go $@
