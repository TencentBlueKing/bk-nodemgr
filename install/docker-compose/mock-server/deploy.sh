#!/bin/bash

set -euo pipefail

# path.
CONFIG_PATH=./etc

# env file.
ENV_FILE=./mock-server.env
GENERATE_TOOL=../generate.sh
GEN_MOCK_CONFIG=../../../test/tools/gen-mock-config.sh

# docker compose.
DOCKER_COMPOSE_FILE=mock-server.yml
DOCKER_COMPOSE_TEMPLATE_PATH=./templates/service
DOCKER_COMPOSE_TEMPLATE_FILE=mock-server.yml.tpl

# config.
CONFIG_FILE=$CONFIG_PATH/mock-server.yml
CONFIG_TEMPLATE_PATH=./templates/config
CONFIG_TEMPLATE_FILE=mock-server.yml.tpl

function usage() {
    echo "Usage: $0 {generate|install|uninstall|clean}"
    exit 1
}

function generate() {
    # check config.
    if [ ! -f "$GENERATE_TOOL" ]; then
        echo "Config generate tool($GENERATE_TOOL) does not exist!"
        exit 1
    fi

    # check docker compose template.
    if [ ! -f "$DOCKER_COMPOSE_TEMPLATE_PATH/$DOCKER_COMPOSE_TEMPLATE_FILE" ]; then
        echo "Docker compose template($DOCKER_COMPOSE_TEMPLATE_PATH/$DOCKER_COMPOSE_TEMPLATE_FILE) does not exist!"
        exit 2
    fi

    # check config template.
    if [ ! -f "$CONFIG_TEMPLATE_PATH/$CONFIG_TEMPLATE_FILE" ]; then
        echo "Config template($CONFIG_TEMPLATE_PATH/$CONFIG_TEMPLATE_FILE) does not exist!"
        exit 2
    fi

    # try to generate config.
    bash $GENERATE_TOOL -e $ENV_FILE -t $CONFIG_TEMPLATE_PATH/$CONFIG_TEMPLATE_FILE > $CONFIG_FILE

    if [ $? -ne 0 ]; then
        echo "Failed to generate config for mock-server"
        exit 3
    fi

    echo "Successfully generated config to $CONFIG_FILE"

    # merge mock data into config via gen-mock-config.sh.
    [ -f "$GEN_MOCK_CONFIG" ] || { echo "Mock config tool not found: $GEN_MOCK_CONFIG"; exit 1; }
    bash $GEN_MOCK_CONFIG $CONFIG_FILE $CONFIG_FILE
    echo "Successfully merged mock data into config"

    # try to generate docker compose yml.
    bash $GENERATE_TOOL -e $ENV_FILE -t $DOCKER_COMPOSE_TEMPLATE_PATH/$DOCKER_COMPOSE_TEMPLATE_FILE > $DOCKER_COMPOSE_FILE
    echo "Successfully generated docker-compose file to $DOCKER_COMPOSE_FILE"
}

function install() {
    generate

    # try to start services.
    docker compose -f $DOCKER_COMPOSE_FILE up -d
}

function uninstall() {
    read -p "Are you sure you want to uninstall ? (Y/N): " choice
    choice=${choice^^}

    if [[ "$choice" != "Y" ]]; then
        exit 1
    fi

    docker compose -f $DOCKER_COMPOSE_FILE down
}

function clean() {
    read -p "Are you sure you want to clean all data ? (Y/N): " choice
    choice=${choice^^}

    if [[ "$choice" != "Y" ]]; then
        exit 1
    fi

    docker compose -f $DOCKER_COMPOSE_FILE down

    rm -f $CONFIG_PATH/*
    rm -f $DOCKER_COMPOSE_FILE
}

if [ "$#" -ne 1 ]; then
    usage
fi

case $1 in
    generate)
        generate
        ;;
    install)
        install
        ;;
    uninstall)
        uninstall
        ;;
    clean)
        clean
        ;;
    *)
        echo "Invalid operation: $1"
        usage
        ;;
esac
