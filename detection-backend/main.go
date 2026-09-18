package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/AftonDannato/computo-distribuido/detection-backend/controllers"
	"github.com/AftonDannato/computo-distribuido/detection-backend/models"
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

	detectionModel := models.NewDetectionModel(db)
	detectionController := controllers.NewDetectionController(detectionModel)

	router := http.NewServeMux()
	router.HandleFunc("GET /detection", detectionController.GetAllAlerts)
	router.HandleFunc("GET /detection/critical", detectionController.GetCriticalAlerts)
	router.HandleFunc("POST /detection", detectionController.PostAlert)
	router.HandleFunc("GET /heartbeat", heartbeatHandler)

	log.Println("Servidor escuchando en http://localhost:8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
