package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/AftonDannato/computo-distribuido/events-backend/models"
)

// Controlador de eventos que almacena el modelo de eventos
type EventController struct {
	model *models.EventModel
}

// Función para crear el controlador de eventos
func NewEventController(model *models.EventModel) *EventController {
	return &EventController{
		model: model,
	}
}

func (controller *EventController) handleRequest(request func() ([]models.Event, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventos, err := request()

		if err != nil {
			http.Error(w, "Error al obtener los eventos", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(eventos)

		if err != nil {
			http.Error(w, "Error al generar la respuesta", http.StatusInternalServerError)
		}
	}
}

func (controller *EventController) GetAllEvents(w http.ResponseWriter, r *http.Request) {
	controller.handleRequest(controller.model.GetAllEvents)(w, r)
}

func (controller *EventController) GetCriticalEvents(w http.ResponseWriter, r *http.Request) {
	controller.handleRequest(controller.model.GetCriticalEvents)(w, r)
}

func (controller *EventController) PostEvent(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Origen    string `json:"origen"`
		Tipo      string `json:"tipo"`
		Severidad string `json:"severidad"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if request.Severidad != "Media" && request.Severidad != "Alta" && request.Severidad != "Crítica" {
		http.Error(w, "Severidad inválida", http.StatusBadRequest)
		return
	}

	evento := models.Event{
		Origen:    request.Origen,
		Tipo:      request.Tipo,
		Severidad: request.Severidad,
	}

	err = controller.model.CreateEvent(evento)

	if err != nil {
		http.Error(w, "Error al crear el evento", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(map[string]string{"message": "Evento creado correctamente"})
	if err != nil {
		http.Error(w, "Error al generar la respuesta", http.StatusInternalServerError)
	}
}
