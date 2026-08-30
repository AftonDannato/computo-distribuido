package controllers

import (
	"encoding/json"
	"html/template"
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

func (c *EventController) Dashboard(w http.ResponseWriter, r *http.Request) {
	eventos, err := c.Model.GetEvents()

	if err != nil {
		http.Error(w, "Error al obtener los eventos", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")

	if err != nil {
		http.Error(w, "Error al cargar la vista", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, eventos)

	if err != nil {
		http.Error(w, "Error al generar la vista", http.StatusInternalServerError)
	}
}
