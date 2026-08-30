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

type Job struct {
	Response chan<- Result
}

type Result struct {
	Events []models.Event
	Err    error
}

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

func main() {
	dsn := os.Getenv("DATABASE_URL")

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
