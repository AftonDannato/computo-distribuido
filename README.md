# Security Event Monitor

Monitor de eventos de ciberseguridad desarrollado en Go, PostgreSQL y Docker.

```text
client → middleware (:8080) → worker → model → PostgreSQL
```

## Structure

```text
computo-distribuido/
├── controllers/
│   └── event_controller.go
├── models/
│   └── event_model.go
├── templates/
│   └── index.html
├── main.go
├── schema.sql
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum
```

## How it works

### Middleware (`main.go`)

- Recibe las peticiones HTTP mediante `http.ServeMux`.
- Vincula estáticamente cada ruta con el channel de su servicio correspondiente.
- Utiliza un worker por servicio, ejecutado como goroutine.
- Cada petición genera un channel de respuesta propio, incluido dentro de un `Job`.
- El worker procesa el trabajo y devuelve un `Result` mediante dicho channel.


### Model (`models/event_model.go`)

- Contiene la conexión a PostgreSQL inyectada como dependencia.
- `GetAllEvents()` obtiene todos los eventos registrados.
- `GetCriticalEvents()` obtiene los eventos con severidad `Alta` o `Crítica`.
- `scanEvents()` convierte las filas obtenidas de PostgreSQL en estructuras `Event`.

### PostgreSQL

- Se ejecuta dentro de Docker.
- `schema.sql` crea la tabla `eventos`.
- También inserta registros iniciales para realizar pruebas.

### Controllers

El directorio `controllers/` pertenece a la implementación MVC original del proyecto.

Actualmente las consultas son procesadas directamente por los workers mediante el modelo, por lo que estos controladores se conservan únicamente como parte de la versión anterior.

## Run it

Solo se requiere Docker y Docker Compose.

```bash
git clone https://github.com/AftonDannato/computo-distribuido.git
cd computo-distribuido
docker compose up --build
```

La aplicación estará disponible en:

```text
http://localhost:8080
```

### Obtener todos los eventos

```bash
curl localhost:8080/api/events
```

### Obtener eventos de severidad Alta o Crítica

```bash
curl localhost:8080/api/critical
```

También pueden probarse directamente desde el navegador:

```text
http://localhost:8080/api/events
http://localhost:8080/api/critical
```

Al realizar las peticiones, los logs permiten observar qué worker recibió y respondió cada trabajo.

### Detener el servicio

```bash
docker compose down
```

Para eliminar también el volumen de PostgreSQL:

```bash
docker compose down -v
```

Al eliminar el volumen, `schema.sql` volverá a ejecutarse en la siguiente inicialización.