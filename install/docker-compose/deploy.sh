#!/bin/bash

set -euo pipefail

# path.
CONFIG_PATH=./etc
DATA_VOLUMES_PATH=./data

# env file.
ENV_FILE=./bk-nodemgr.env
GENERATE_TOOL=./generate.sh

# docker compose.
DOCKER_COMPOSE_FILE=bk-nodemgr.yml
DOCKER_COMPOSE_SYNC_FILE=bk-nodemgr-apigw-sync.yml
DOCKER_COMPOSE_TEMPLATE_PATH=./templates/service
DOCKER_COMPOSE_TEMPLATE_FILE=bk-nodemgr.yml.tpl
DOCKER_COMPOSE_TEMPLATE_SYNC_FILE=bk-nodemgr-apigw-sync.yml.tpl

# config.
CONFIG_TEMPLATE_PATH=./templates/config

# modules.
MODULES=("application" "backend" "file")

function usage() {
    echo "Usage: $0 {install|uninstall|clean}"
    exit 1
}

function generate() {
    # check config.
    if [ ! -f "$GENERATE_TOOL" ]; then
        echo "Config generate tool($GENERATE_TOOL) dose not exist!)"
        exit 1
    fi

    # check docker compose template.
    if [ ! -f "$DOCKER_COMPOSE_TEMPLATE_PATH/$DOCKER_COMPOSE_TEMPLATE_FILE" ]; then
        echo "Config template env($DOCKER_COMPOSE_TEMPLATE_PATH/$DOCKER_COMPOSE_TEMPLATE_FILE) dose not exist!)"
        exit 2
    fi

    # check docker compose sync template.
    if [ ! -f "$DOCKER_COMPOSE_TEMPLATE_PATH/$DOCKER_COMPOSE_TEMPLATE_SYNC_FILE" ]; then
        echo "Config template env($DOCKER_COMPOSE_TEMPLATE_PATH/$DOCKER_COMPOSE_TEMPLATE_SYNC_FILE) dose not exist!)"
        exit 2
    fi

    for module in "${MODULES[@]}"; do
        if [ ! -f "$CONFIG_TEMPLATE_PATH/bk-nodemgr-$module.yml.tpl" ]; then
            echo "Config template($CONFIG_TEMPLATE_PATH/bk-nodemgr-$module.yml.tpl) dose not exist!)"
            exit 2
        fi

        # try to generate configs.
        bash $GENERATE_TOOL -e $ENV_FILE -t $CONFIG_TEMPLATE_PATH/bk-nodemgr-$module.yml.tpl > $CONFIG_PATH/bk-nodemgr-$module.yml

        if [ $? -ne 0 ]; then
            echo "Failed to generate config for module $module"
            exit 3
        fi
    done

    echo "Successfully generated configs to $CONFIG_PATH"

    # try to generate docker compose yml.
    bash $GENERATE_TOOL -e $ENV_FILE -t $DOCKER_COMPOSE_TEMPLATE_PATH/$DOCKER_COMPOSE_TEMPLATE_FILE > $DOCKER_COMPOSE_FILE
    echo "Successfully generated docker-compose file to $DOCKER_COMPOSE_FILE"

    # try to generate docker compose yml.
    bash $GENERATE_TOOL -e $ENV_FILE -t $DOCKER_COMPOSE_TEMPLATE_PATH/$DOCKER_COMPOSE_TEMPLATE_SYNC_FILE > $DOCKER_COMPOSE_SYNC_FILE
    echo "Successfully generated docker-compose file to $DOCKER_COMPOSE_SYNC_FILE"

    # generate mongo keyfile.
    if [ ! -f "./mongo_keyfile" ]; then
        openssl rand -base64 756 > mongo_keyfile
        echo "Successfully generated mongo keyfile to mongo_keyfile"
    fi
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

function install_sync() {
    generate

    # try to start services.
    docker compose -f $DOCKER_COMPOSE_SYNC_FILE up -d
}

function uninstall_sync() {
    read -p "Are you sure you want to uninstall_sync ? (Y/N): " choice
    choice=${choice^^}

    if [[ "$choice" != "Y" ]]; then
        exit 1
    fi

    docker compose -f $DOCKER_COMPOSE_SYNC_FILE down
}

function clean() {
    read -p "Are you sure you want to clean all data ? (Y/N): " choice
    choice=${choice^^}

    if [[ "$choice" != "Y" ]]; then
        exit 1
    fi

    docker compose -f $DOCKER_COMPOSE_FILE down

    rm $CONFIG_PATH/*
    rm -r $DATA_VOLUMES_PATH/*
    rm -f $DOCKER_COMPOSE_FILE
    rm -f $DOCKER_COMPOSE_SYNC_FILE
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
    install-sync)
        install_sync
        ;;
    uninstall-sync)
        uninstall_sync
        ;;
    clean)
        clean
        ;;
    *)
        echo "Invalid operation: $1"
        usage
        ;;
esac
