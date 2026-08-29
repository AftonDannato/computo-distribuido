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
