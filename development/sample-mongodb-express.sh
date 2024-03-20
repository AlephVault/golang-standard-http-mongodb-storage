#!/bin/bash
docker run -e ME_CONFIG_MONGODB_URL=mongodb://admin:p455w0rd@localhost:27017 \
           -e ME_CONFIG_BASICAUTH_USERNAME="sample" \
           -e ME_CONFIG_BASICAUTH_PASSWORD="sample" \
           --rm --network host mongo-express