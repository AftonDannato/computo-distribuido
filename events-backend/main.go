package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/AftonDannato/computo-distribuido/events-backend/controllers"
	"github.com/AftonDannato/computo-distribuido/events-backend/models"
)

// Función que responde si el backend está vivo
func heartbeatHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("backend alive"))
}

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
	eventController := controllers.NewEventController(eventModel)

	router := http.NewServeMux()
	router.HandleFunc("GET /events/all", eventController.GetAllEvents)
	router.HandleFunc("GET /events/critical", eventController.GetCriticalEvents)
	router.HandleFunc("GET /heartbeat", heartbeatHandler)

	log.Println("Servidor escuchando en http://localhost:8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
