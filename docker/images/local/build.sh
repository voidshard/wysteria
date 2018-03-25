#!/bin/bash

#
# Builds docker from local files
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
cp ${DIR}/../test.crt ./build/
cp ${DIR}/../test.key ./build/
cp ${DIR}/../test.csr ./build/
cp ${DIR}/../wysteria-server.ini.template ./build/
cp -r ${DIR}/../../../common ./build/
cp -r ${DIR}/../../../server ./build/
cp -r ${DIR}/../../../client ./build/
cp -r ${DIR}/../../../glide.lock ./build/
cp -r ${DIR}/../../../glide.yaml ./build/

# cd into the build dir
cd ./build/

# no fail on err (we want to rm the build file no matter what docker does)
set +e

# call docker to do it's thing
docker build -t wysteria/wysteria:local ./

# fail on error
set -eu

# finally, we can remove the tmp build dir
rm -r ${DIR}/build
