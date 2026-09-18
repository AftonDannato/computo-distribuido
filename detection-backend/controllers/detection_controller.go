package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/AftonDannato/computo-distribuido/detection-backend/models"
)

// Controlador de eventos que almacena el modelo de eventos
type DetectionController struct {
	model *models.DetectionModel
}

// Función para crear el controlador de eventos
func NewDetectionController(model *models.DetectionModel) *DetectionController {
	return &DetectionController{
		model: model,
	}
}

func (controller *DetectionController) handleRequest(request func() ([]models.Alert, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		alertas, err := request()

		if err != nil {
			http.Error(w, "Error al obtener las alertas", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(alertas)

		if err != nil {
			http.Error(w, "Error al generar la respuesta", http.StatusInternalServerError)
		}
	}
}

func (controller *DetectionController) GetAllAlerts(w http.ResponseWriter, r *http.Request) {
	controller.handleRequest(controller.model.GetAllAlerts)(w, r)
}

func (controller *DetectionController) GetCriticalAlerts(w http.ResponseWriter, r *http.Request) {
	controller.handleRequest(controller.model.GetCriticalAlerts)(w, r)
}

func (controller *DetectionController) PostAlert(w http.ResponseWriter, r *http.Request) {
	var request struct {
		EventoID int    `json:"evento_id"`
		Regla    string `json:"regla"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	var severidad string
	switch request.Regla {
	case "Evento de severidad crítica":
		severidad = "Crítica"
	case "Prueba":
		severidad = "Media"
	default:
		http.Error(w, "Regla inválida", http.StatusBadRequest)
		return
	}

	alerta := models.Alert{
		EventoID:  request.EventoID,
		Regla:     request.Regla,
		Severidad: severidad,
	}

	err = controller.model.CreateAlert(alerta)

	if err != nil {
		http.Error(w, "Error al crear la alerta", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(map[string]string{"message": "Alerta creada correctamente"})
	if err != nil {
		http.Error(w, "Error al generar la respuesta", http.StatusInternalServerError)
	}
}
