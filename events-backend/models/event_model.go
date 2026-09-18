package models

import (
	"database/sql"
	"time"
)

// Estructura que almacena un registro de ciberseguridad
type Event struct {
	ID        int
	Origen    string
	Tipo      string
	Severidad string
	Fecha     time.Time
	Atendido  bool
}

// Modelo dónde se guarda la conexión a la DB (inyección)
type EventModel struct {
	db *sql.DB
}

// Función que se llama para crear el modelo con la conexión inyectada
func NewEventModel(db *sql.DB) *EventModel {
	return &EventModel{
		db: db,
	}
}

// Función encargada de leer  registros obtenidos de una query y convertirlos a estructuras Event
func (model *EventModel) scanEvents(rows *sql.Rows) ([]Event, error) {
	defer rows.Close()

	var eventos []Event

	for rows.Next() {
		var evento Event

		err := rows.Scan(
			&evento.ID,
			&evento.Origen,
			&evento.Tipo,
			&evento.Severidad,
			&evento.Fecha,
			&evento.Atendido,
		)

		if err != nil {
			return nil, err
		}

		eventos = append(eventos, evento)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return eventos, nil
}

// Función encargada de obtener todos los eventos de la DB y devolverlos convertidos ya en estructuras
func (model *EventModel) GetAllEvents() ([]Event, error) {
	rows, err := model.db.Query(`
		SELECT *
		FROM eventos
		ORDER BY fecha DESC
	`)

	if err != nil {
		return nil, err
	}

	return model.scanEvents(rows)
}

// Función encargada de obtener los eventos criticos de la DB y devolverlos convertidos ya en estructuras
func (model *EventModel) GetCriticalEvents() ([]Event, error) {
	rows, err := model.db.Query(`
		SELECT *
		FROM eventos
		WHERE severidad IN ('Alta', 'Crítica')
		ORDER BY fecha DESC
	`)

	if err != nil {
		return nil, err
	}

	return model.scanEvents(rows)
}

func (model *EventModel) CreateEvent(evento Event) error {
	_, err := model.db.Exec(`
		INSERT INTO eventos (origen, tipo, severidad)
		VALUES ($1, $2, $3)
	`, evento.Origen, evento.Tipo, evento.Severidad)
	return err
}
