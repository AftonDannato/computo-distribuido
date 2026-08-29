package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/AftonDannato/computo-distribuido/models"
)

type EventController struct {
	Model *models.EventModel
}

func NewEventController(model *models.EventModel) *EventController {
	return &EventController{
		Model: model,
	}
}

func (c *EventController) GetEvents(w http.ResponseWriter, r *http.Request) {
	eventos, err := c.Model.GetEvents()

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