#!/bin/bash

ORG=a PORT=8080 CONFIG_FILE=$PWD/artifacts/network-config-local.json MIDDLEWARE_CONFIG_FILE=$PWD/middleware/map.json WEB_DIR=$PWD/www node ../NSD/server/index
