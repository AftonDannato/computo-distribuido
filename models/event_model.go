package models

import (
	"database/sql"
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

type EventModel struct {
	DB *sql.DB
}

func NewEventModel(db *sql.DB) *EventModel {
	return &EventModel{
		DB: db,
	}
}

func (m *EventModel) GetEvents() ([]Event, error) {
	rows, err := m.DB.Query(`
		SELECT *
		FROM eventos
		ORDER BY fecha DESC
	`)

	if err != nil {
		return nil, err
	}
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
