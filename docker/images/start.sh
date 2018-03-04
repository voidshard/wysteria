#!/bin/bash

#
# This script is copied into a docker image at build time (along with the wysteria-server.ini.template).
#  It writes ENV vars given to Docker into the ini.template file & remove empty / null lines before starting
#  wysteria. Thus we can change the config file at `docker run` time -> yaaaaay!
#

# fail on error
set -eu

# write the WYS_* env variables into our config file
envsubst < wysteria-server.ini.template > ./wysteria-server.ini

# remove lines containing the null string (INI parser doesn't seem to like placeholder lines)
sed -i '/null/d' ./wysteria-server.ini

# kick off wysteria proper
./server $@
