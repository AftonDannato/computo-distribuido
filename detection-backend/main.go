package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/AftonDannato/computo-distribuido/detection-backend/controllers"
	"github.com/AftonDannato/computo-distribuido/detection-backend/models"
)

// Función que responde si el backend está vivo
func heartbeatHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("backend alive"))
}

var instanceID = "unknown"

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func registerWithConsul(consulAddr string) error {
	payload := map[string]any{
		"ID":      instanceID,
		"Name":    "detection",
		"Address": instanceID, // resolvable via Docker's embedded DNS
		"Port":    8080,
		"Check": map[string]any{
			"HTTP":                           fmt.Sprintf("http://%s:8080/heartbeat", instanceID),
			"Interval":                       "5s",
			"Timeout":                        "2s",
			"DeregisterCriticalServiceAfter": "1m",
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 1; attempt <= 10; attempt++ {
		req, err := http.NewRequest(http.MethodPut, consulAddr+"/v1/agent/service/register", bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				log.Printf("registered %s with Consul at %s", instanceID, consulAddr)
				return nil
			}
			lastErr = fmt.Errorf("consul returned status %d", resp.StatusCode)
		} else {
			lastErr = err
		}
		time.Sleep(time.Duration(attempt) * time.Second)
	}
	return fmt.Errorf("giving up registering with consul: %w", lastErr)
}

// deregisterFromConsul is called on shutdown so Consul (and therefore
// every gateway watching it) stops routing traffic here immediately,
// instead of waiting for DeregisterCriticalServiceAfter to time it out.
func deregisterFromConsul(consulAddr string) {
	req, err := http.NewRequest(http.MethodPut, consulAddr+"/v1/agent/service/deregister/"+instanceID, nil)
	if err != nil {
		log.Printf("failed to build deregister request: %v", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("failed to deregister %s from consul: %v", instanceID, err)
		return
	}
	resp.Body.Close()
	log.Printf("deregistered %s from consul", instanceID)
}

func main() {
	dsn := os.Getenv("DATABASE_URL")

	// Esto se hace para que podamos correr el programa tanto con Go + Docker
	// separados, cómo con Go ya Dockerizado
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=security_monitor sslmode=disable"
	}

	if v := os.Getenv("INSTANCE_ID"); v != "" {
		instanceID = v
	}
	consulAddr := getEnv("CONSUL_HTTP_ADDR", "http://consul:8500")

	if err := registerWithConsul(consulAddr); err != nil {
		log.Printf("WARNING: %v", err)
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

	srv := &http.Server{Addr: ":8080", Handler: router}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Printf("shutting down %s", instanceID)
		deregisterFromConsul(consulAddr)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	log.Println("Servidor escuchando en http://localhost:8080")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
