package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Event struct {
	ID        int
	Origen    string
	Tipo      string
	Severidad string
	Fecha     time.Time
	Atendido  bool
}

type Alert struct {
	ID        int
	EventoID  int
	Regla     string
	Severidad string
	Fecha     time.Time
}

type Incident struct {
	ID            int
	AlertaID      int
	Descripcion   string
	Estado        string
	FechaCreacion time.Time
	FechaCierre   *time.Time
}

type DashboardData struct {
	Eventos    []Event
	Alertas    []Alert
	Incidentes []Incident
}

var gatewayURL string

func heartbeatHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("frontend alive"))
}

func getJSON(path string, target any) error {
	response, err := http.Get(gatewayURL + path)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("gateway respondió con %s", response.Status)
	}

	return json.NewDecoder(response.Body).Decode(target)
}

func makeGetter[T any](path string) func() ([]T, error) {
	return func() ([]T, error) {
		var data []T

		if err := getJSON(path, &data); err != nil {
			return nil, err
		}

		return data, nil
	}
}

var getAllEvents = makeGetter[Event]("/events")
var getCriticalEvents = makeGetter[Event]("/events/critical")

var getAllAlerts = makeGetter[Alert]("/detection")
var getCriticalAlerts = makeGetter[Alert]("/detection/critical")

var getAllIncidents = makeGetter[Incident]("/incidents")
var getOpenIncidents = makeGetter[Incident]("/incidents/open")

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var eventos []Event
	var alertas []Alert
	var incidentes []Incident
	var err error

	switch r.URL.Query().Get("events") {
	case "critical":
		eventos, err = getCriticalEvents()
	default:
		eventos, err = getAllEvents()
	}

	if err != nil {
		http.Error(w, "Error al obtener eventos", http.StatusInternalServerError)
		return
	}

	switch r.URL.Query().Get("alerts") {
	case "critical":
		alertas, err = getCriticalAlerts()
	default:
		alertas, err = getAllAlerts()
	}

	if err != nil {
		http.Error(w, "Error al obtener alertas", http.StatusInternalServerError)
		return
	}

	switch r.URL.Query().Get("incidents") {
	case "open":
		incidentes, err = getOpenIncidents()
	default:
		incidentes, err = getAllIncidents()
	}

	if err != nil {
		http.Error(w, "Error al obtener incidentes", http.StatusInternalServerError)
		return
	}

	data := DashboardData{
		Eventos:    eventos,
		Alertas:    alertas,
		Incidentes: incidentes,
	}

	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "Error al cargar el dashboard", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Error al generar el dashboard", http.StatusInternalServerError)
	}
}

func postJSON(path string, data any) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	response, err := http.Post(
		gatewayURL+path,
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("gateway respondió con %s", response.Status)
	}

	return nil
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	switch r.FormValue("accion") {
	case "evento":
		data := map[string]any{
			"origen":    r.FormValue("origen"),
			"tipo":      r.FormValue("tipo"),
			"severidad": r.FormValue("severidad"),
		}

		if err := postJSON("/events", data); err != nil {
			http.Error(w, "Error al crear el evento", http.StatusInternalServerError)
			return
		}
	case "alerta":
		eventoID, err := strconv.Atoi(r.FormValue("evento_id"))
		if err != nil {
			http.Error(w, "Evento ID inválido", http.StatusBadRequest)
			return
		}

		data := map[string]any{
			"evento_id": eventoID,
			"regla":     r.FormValue("regla"),
		}

		if err := postJSON("/detection", data); err != nil {
			http.Error(w, "Error al crear la alerta", http.StatusInternalServerError)
			return
		}

	case "incidente":
		alertaID, err := strconv.Atoi(r.FormValue("alerta_id"))
		if err != nil {
			http.Error(w, "Alerta ID inválido", http.StatusBadRequest)
			return
		}

		data := map[string]any{
			"alerta_id":   alertaID,
			"descripcion": r.FormValue("descripcion"),
		}

		if err := postJSON("/incidents", data); err != nil {
			http.Error(w, "Error al crear el incidente", http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "Acción inválida", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	gatewayURL = os.Getenv("GATEWAY_URL")

	if gatewayURL == "" {
		gatewayURL = "http://gateway:8080"
	}

	router := http.NewServeMux()

	router.HandleFunc("GET /heartbeat", heartbeatHandler)
	router.HandleFunc("GET /{$}", indexHandler)
	router.HandleFunc("POST /{$}", postHandler)

	log.Println("Frontend escuchando en http://localhost:8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
