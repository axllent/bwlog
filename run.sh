export GOPATH=${PWD}
# Regenerate static files
bin/statik -src=http/ -f
go run *.go
