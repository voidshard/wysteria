#!/bin/bash

#
# Builds wysteria:local, then kicks off test integration suite
#

# set fail on error
set -eu

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# generic test ssl certs
# > Nb. we have to set different paths to these for the test client & the server (docker container)
# > plz don't actually use these ..
export WYS_MIDDLEWARE_CLIENT_SSL_KEY=${DIR}/../docker/images/test.key
export WYS_MIDDLEWARE_CLIENT_SSL_CERT=${DIR}/../docker/images/test.crt
export WYS_MIDDLEWARE_SSL_KEY=/usr/ssl/test.key
export WYS_MIDDLEWARE_SSL_CERT=/usr/ssl/test.crt

# Args:
#   $1 -> name of service (ie, we'll use docker-compose up <service>)
#   $2 -> name of test
#   $3 -> some time to wait to allow things to spin up - this isn't ideal but works
dotest () {
    echo "> running test:" $2

    # start test service(s)
    cd ${DIR}

    # write out the docker-compose.yml file
    envsubst < docker-compose.yml.template > docker-compose.yml
    echo "> generated docker-compose.yml"

    docker-compose up -d $1

    # sleep for a bit to allow things to start up
    sleep $3

    # cd in to integration tests root
    cd ${DIR}/integration

    # permit failures (we don't want 'go test' to be able to stop us running)
    set +e

    # throw it over to go test
    go test -v

    # set fail on error
    set -eu

    # tear down test services
    cd ${DIR}
    echo "> stopping test:" $2
    docker-compose down

    # backup docker-compose file to docker-compose.yml.<name>
    # this is useful for testing when things are breaking
    # mv docker-compose.yml{,.$2}

    # remove the docker compose file
    rm -v docker-compose.yml
}

echo "--------------------- BUILD -----------------------------"

${DIR}/../docker/images/local/build.sh

echo "--------------------- TESTS -----------------------------"
#
# Essentially, we're now going to stand up & tear down wysteria in a whole host of different configurations.
# We'll use each of the middleware, database and searchbase options, with and without SSL.
#
# We start with the most primitive (everything embedded) and slowly add more services. As we go on we have to introduce
# sleep calls to wait for docker containers to spin up & become ready. Elastic in particular seems *very* slow to come
# alive. In order, we test:
#
# mware | ssl | search  |  db
# ------------------------------
# grpc  |  F  |  bleve  | bolt
# grpc  |  T  |  bleve  | bolt
# grpc  |  F  ] elastic | bolt
# grpc  |  F  ] elastic | mongo
# nats  |  F  ] elastic | mongo
#
# Eagle eyed readers will note that this isn't *every* *possible* combination, but it *does* test pretty much
# all of the server, database, searchbase, transport & client code.
#

echo "> using grpc"
echo "> using bleve"
echo "> using boltdb"
export WYS_MIDDLEWARE_CLIENT_DRIVER=grpc
export WYS_MIDDLEWARE_CLIENT_CONFIG=localhost:31000
echo "> disabling SSL"
export WYS_MIDDLEWARE_SSL_ENABLE=false
dotest "sololocal" "grpc_bleve_boltdb_nossl" 0

echo "> enabling SSL"
export WYS_MIDDLEWARE_SSL_ENABLE=true
dotest "sololocal" "grpc_bleve_boltdb_ssl" 0

echo "> disabling SSL"
echo "> using elasticsearch"
export WYS_MIDDLEWARE_SSL_ENABLE=false
dotest "elastictest" "grpc_elastic_boltdb_nossl" 15

echo "> using mongo"
dotest "mongotest" "grpc_elastic_mongo_nossl" 15

echo "> using nats.io"
export WYS_MIDDLEWARE_CLIENT_DRIVER=nats
export WYS_MIDDLEWARE_CLIENT_CONFIG=nats://localhost:4222
echo "> disabling SSL"
export WYS_MIDDLEWARE_SSL_ENABLE=false
dotest "local" "nats_elastic_mongo_nossl" 20

echo "--------------------- EXIT -----------------------------"


