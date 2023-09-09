#/bin/bash

dir=$(cd $(dirname $(dirname $0)); pwd)
scp $dir/env/$1/.env $dir/
scp $dir/env/$1/config.yml $dir/conf/
scp $dir/env/$1/supervisor.conf $dir/