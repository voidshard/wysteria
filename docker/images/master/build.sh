#!/bin/bash

#
# Builds docker pulling master from github.com/voidshard/wysteria
#

# fail on error
set -eu

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# set CWD to DIR, just incase
cd ${DIR}

# make us a temp build dir
mkdir ./build/

# whack in all the wysteria files we'll need
cp Dockerfile ./build/
cp ${DIR}/../start.sh ./build/
cp ${DIR}/../wysteria-server.ini.template ./build/

# move to build dir
cd ${DIR}/build

# no fail on err (we want to rm the build file no matter what docker does)
set +e

# call docker to do it's thing
docker build -t wysteria/wysteria:master ./

# fail on error
set -eu

# finally, we can remove the tmp build dir
rm -r ${DIR}/build
