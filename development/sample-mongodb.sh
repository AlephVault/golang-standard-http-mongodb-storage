#!/bin/bash
BASEDIR=$(pwd)/$(dirname "$(dirname $0)")
# Python code: import os; client = pymongo.MongoClient(host="localhost", username="admin", password="p455w0rd")
docker run --name mongodb-dev --rm -p 27017:27017 \
           -e MONGO_INITDB_ROOT_USERNAME=admin \
           -e MONGO_INITDB_ROOT_PASSWORD=p455w0rd \
           -v $BASEDIR/.tmp/mongo:/data/db mongo:latest
