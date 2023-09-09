#/bin/bash
dir=$(cd $(dirname $(dirname $0)); pwd)
go build -mod=vendor -o $dir/bin/mix $dir/main.go