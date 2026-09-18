#!/bin/bash

set -e

echo "Comprobando Docker..."

if ! command -v docker >/dev/null 2>/dev/null; then
    echo "Docker no esta instalado."

    if ! command -v apt >/dev/null 2>/dev/null; then
        echo "Este script requiere una distribucion basada en Debian/Ubuntu."
        exit 1
    fi

    echo "Instalando Docker y Docker Compose..."

    sudo apt update
    sudo apt install -y docker.io docker-compose

    echo "Iniciando Docker..."

    sudo systemctl enable --now docker
fi

echo "Comprobando que Docker este ejecutandose..."

if ! sudo docker info >/dev/null 2>/dev/null; then
    echo "Docker esta instalado pero no esta ejecutandose."
    exit 1
fi

echo "Comprobando Docker Compose..."

if ! sudo docker compose version >/dev/null 2>/dev/null; then
    echo "Docker Compose no esta disponible."
    exit 1
fi

echo "Construyendo y levantando el proyecto..."

sudo docker compose up --build