# Monitor de Seguridad Distribuido

Frontend, gateway con load balancing, tres tipos de backend y PostgreSQL:

```text
                         ┌──> events-backend ×3
client → frontend → gateway ──> detection-backend ×3
                         └──> incidents-backend ×3
                                   ↓
                               PostgreSQL
```

## Structure

```text
computo-distribuido/
├── docker-compose.yml
├── schema.sql
├── setup.bat
├── setup.sh
├── frontend/
├── gateway/
├── events-backend/
├── detection-backend/
└── incidents-backend/
```

## Run it

### Windows

```cmd
setup.bat
```

### Linux

Distribuciones basadas en Debian/Ubuntu:

```bash
chmod +x setup.sh
./setup.sh
```

### Manual

```bash
docker compose up --build
```

Después del primer build:

```bash
docker compose up
```

Abrir:

```text
http://localhost:8080
```

## How it works

**Frontend**
- Único servicio expuesto al host en `localhost:8080`.
- Muestra eventos, alertas e incidentes mediante `html/template`.
- Permite filtrar la información mostrada.
- Permite crear eventos, alertas e incidentes.
- Se comunica internamente con el gateway; las rutas de los backends no se exponen directamente.

**Gateway**
- Lee `routes.json` como service discovery estático.
- Agrupa varias instancias por prefijo.
- Distribuye peticiones mediante Round Robin.
- Comprueba `/heartbeat` de cada instancia periódicamente.
- Omite backends marcados como no saludables.
- Devuelve `503` si no existe ninguna instancia saludable para un servicio.
- Limita las peticiones concurrentes mediante un canal usado como semáforo.

**Events Backend**
- Tres instancias idénticas.
- Registra y consulta eventos de seguridad.
- Rutas:

```text
GET  /events
GET  /events/critical
POST /events
GET  /heartbeat
```

**Detection Backend**
- Tres instancias idénticas.
- Registra y consulta alertas relacionadas con eventos.
- Determina la severidad de una alerta a partir de su regla.
- Rutas:

```text
GET  /detection
GET  /detection/critical
POST /detection
GET  /heartbeat
```

**Incidents Backend**
- Tres instancias idénticas.
- Registra y consulta incidentes relacionados con alertas.
- Los nuevos incidentes comienzan con estado `Abierto`.
- Rutas:

```text
GET  /incidents
GET  /incidents/open
POST /incidents
GET  /heartbeat
```

**PostgreSQL**
- Base de datos compartida por las instancias.
- Tablas principales:

```text
eventos → alertas → incidentes
```

- Los datos persisten mediante el volumen `postgres_data`.

**Load balancing**

Cada servicio posee tres instancias:

```text
events-backend-1
events-backend-2
events-backend-3

detection-backend-1
detection-backend-2
detection-backend-3

incidents-backend-1
incidents-backend-2
incidents-backend-3
```

El gateway realiza Round Robin únicamente entre instancias saludables.

Al detener una instancia, el heartbeat termina marcándola como no saludable y deja de recibir tráfico.