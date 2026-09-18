@echo off
echo Comprobando Docker...

where docker >nul 2>nul

if errorlevel 1 (
    echo Docker no esta instalado.

    where winget >nul 2>nul
    if errorlevel 1 (
        echo Winget no esta disponible.
        echo Instala Docker Desktop manualmente.
        pause
        exit /b 1
    )

    echo Instalando Docker Desktop...
    winget install -e --id Docker.DockerDesktop

    echo Docker Desktop fue instalado.
    echo Abre Docker Desktop y vuelve a ejecutar este script.
    pause
    exit /b 0
)

docker info >nul 2>nul
if errorlevel 1 (
    echo Docker esta instalado pero no esta ejecutandose.
    echo Abre Docker Desktop y vuelve a ejecutar este script.
    pause
    exit /b 1
)

echo Comprobando Docker Compose...

docker compose version >nul 2>nul
if errorlevel 1 (
    echo Docker Compose no esta disponible.
    pause
    exit /b 1
)

echo Construyendo y levantando el proyecto...

docker compose up --build