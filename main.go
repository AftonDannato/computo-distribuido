package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/AftonDannato/computo-distribuido/models"
)

// Estructura que representa las tareas que el middleware envía.
// Contiene un canal por el que el worker devolverá la respuesta
type Job struct {
	Response chan<- Result
}

// Estructura que representa la respuesta del worker, en el que devuelve
// un arreglo de eventos de ciberseguridad que obtenemos de la DB.
// En caso de error, este también se devuelve.
type Result struct {
	Events []models.Event
	Err    error
}

// Worker encargada de responder a las peticiones hechas al servicio de "events"
func allEventsWorker(allEventsQueue <-chan Job, model *models.EventModel) {
	for job := range allEventsQueue {
		log.Println("allEventsWorker recibió petición")
		events, err := model.GetAllEvents()

		job.Response <- Result{
			Events: events,
			Err:    err,
		}
		log.Println("allEventsWorker respondió petición")
	}
}

// Worker encargada de responder a las peticiones hechas al servicio de "critical"
func criticalEventsWorker(criticalEventsQueue <-chan Job, model *models.EventModel) {
	for job := range criticalEventsQueue {
		log.Println("criticalEventsWorker recibió petición")
		events, err := model.GetCriticalEvents()

		job.Response <- Result{
			Events: events,
			Err:    err,
		}
		log.Println("criticalEventsWorker respondió petición")
	}
}

// Función que se encarga de fabricar las funciones especificas para manejar las peticiones
// que pueden recibir los distintos tipos de backends
func handleRequest(queue chan<- Job) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := make(chan Result)

		job := Job{
			Response: response,
		}

		queue <- job

		result := <-response

		if result.Err != nil {
			http.Error(w, "Error al obtener los eventos", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(result.Events); err != nil {
			http.Error(w, "Error al generar la respuesta", http.StatusInternalServerError)
		}
	}
}

// Implementación del middleware/multiplexor, en la que vinculamos estaticamente los workers con
// los canales de sus servicios correspondientes, y hacemos que el middleware "aprenda" a quien mandar
// las distintas requests disponibles
func middleware(eventModel *models.EventModel) http.Handler {
	allEventsQueue := make(chan Job, 10)
	criticalEventsQueue := make(chan Job, 10)

	go allEventsWorker(allEventsQueue, eventModel)
	go criticalEventsWorker(criticalEventsQueue, eventModel)

	router := http.NewServeMux()

	router.HandleFunc("GET /api/events", handleRequest(allEventsQueue))
	router.HandleFunc("GET /api/critical", handleRequest(criticalEventsQueue))

	return router
}

// Hilo principal del programa, donde nos conectamos a la DB, creamos el modelo de eventos,
// y lo inyectamos al middleware
func main() {
	dsn := os.Getenv("DATABASE_URL")

	// Esto se hace para que podamos correr el programa tanto con Go + Docker
	// separados, cómo con Go ya Dockerizado
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=security_monitor sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	eventModel := models.NewEventModel(db)

	router := middleware(eventModel)

	log.Println("Servidor escuchando en http://localhost:8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
