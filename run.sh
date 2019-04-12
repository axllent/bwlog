export GOPATH=${PWD}

CGO_ENABLED=0 go get \
github.com/axllent/gitrel \
github.com/rakyll/statik \
github.com/gorilla/websocket

# Regenerate static files
bin/statik -src=web/ -f
CGO_ENABLED=0 go run *.go $@
