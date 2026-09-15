package controllers

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/AftonDannato/computo-distribuido/models"
)

// Controlador de eventos que almacena el modelo de eventos
type EventController struct {
	Model *models.EventModel
}

// Función para crear el controlador de eventos
func NewEventController(model *models.EventModel) *EventController {
	return &EventController{
		Model: model,
	}
}

// Función que, en el MVC original, devolvía todos los eventos.
// Deprecado, las consultas ahora se hacen desde los workers
func (c *EventController) GetEvents(w http.ResponseWriter, r *http.Request) {
	eventos, err := c.Model.GetAllEvents()

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

// Función que, en el MVC original, mostraba un dashboard con los eventos
// Deprecado, actualmente no implementado en la arquitectura actual
func (c *EventController) Dashboard(w http.ResponseWriter, r *http.Request) {
	eventos, err := c.Model.GetAllEvents()

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
