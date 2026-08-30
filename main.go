package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/AftonDannato/computo-distribuido/controllers"
	"github.com/AftonDannato/computo-distribuido/models"
)

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
	eventController := controllers.NewEventController(eventModel)

	router := http.NewServeMux()

	router.HandleFunc("GET /", eventController.Dashboard)
	router.HandleFunc("GET /events", eventController.GetEvents)

	log.Println("Servidor escuchando en http://localhost:8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
