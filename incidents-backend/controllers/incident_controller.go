package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/AftonDannato/computo-distribuido/incidents-backend/models"
)

// Controlador de eventos que almacena el modelo de eventos
type IncidentController struct {
	model *models.IncidentModel
}

// Función para crear el controlador de eventos
func NewIncidentController(model *models.IncidentModel) *IncidentController {
	return &IncidentController{
		model: model,
	}
}

func (controller *IncidentController) handleRequest(request func() ([]models.Incident, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		incidentes, err := request()

		if err != nil {
			http.Error(w, "Error al obtener los incidentes", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(incidentes)

		if err != nil {
			http.Error(w, "Error al generar la respuesta", http.StatusInternalServerError)
		}
	}
}

func (controller *IncidentController) GetAllIncidents(w http.ResponseWriter, r *http.Request) {
	controller.handleRequest(controller.model.GetAllIncidents)(w, r)
}

func (controller *IncidentController) GetOpenIncidents(w http.ResponseWriter, r *http.Request) {
	controller.handleRequest(controller.model.GetOpenIncidents)(w, r)
}

func (controller *IncidentController) PostIncident(w http.ResponseWriter, r *http.Request) {
	var request struct {
		AlertaID    int    `json:"alerta_id"`
		Descripcion string `json:"descripcion"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	incidente := models.Incident{
		AlertaID:    request.AlertaID,
		Descripcion: request.Descripcion,
	}

	err = controller.model.CreateIncident(incidente)

	if err != nil {
		http.Error(w, "Error al crear el incidente", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(map[string]string{"message": "Incidente creado correctamente"})
	if err != nil {
		http.Error(w, "Error al generar la respuesta", http.StatusInternalServerError)
	}
}
