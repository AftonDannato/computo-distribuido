package models

import (
	"database/sql"
	"time"
)

// Estructura que almacena un registro de ciberseguridad
type Incident struct {
	ID            int
	AlertaID      int
	Descripcion   string
	Estado        string
	FechaCreacion time.Time
	FechaCierre   *time.Time // puede ser nil
}

// Modelo dónde se guarda la conexión a la DB (inyección)
type IncidentModel struct {
	db *sql.DB
}

// Función que se llama para crear el modelo con la conexión inyectada
func NewIncidentModel(db *sql.DB) *IncidentModel {
	return &IncidentModel{
		db: db,
	}
}

// Función encargada de leer  registros obtenidos de una query y convertirlos a estructuras Alert
func (model *IncidentModel) scanIncidents(rows *sql.Rows) ([]Incident, error) {
	defer rows.Close()

	var incidentes []Incident

	for rows.Next() {
		var incidente Incident

		err := rows.Scan(
			&incidente.ID,
			&incidente.AlertaID,
			&incidente.Descripcion,
			&incidente.Estado,
			&incidente.FechaCreacion,
			&incidente.FechaCierre,
		)

		if err != nil {
			return nil, err
		}

		incidentes = append(incidentes, incidente)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return incidentes, nil
}

// Función encargada de obtener todos los eventos de la DB y devolverlos convertidos ya en estructuras
func (model *IncidentModel) GetAllIncidents() ([]Incident, error) {
	rows, err := model.db.Query(`
		SELECT *
		FROM incidentes
		ORDER BY fecha_creacion ASC
	`)

	if err != nil {
		return nil, err
	}

	return model.scanIncidents(rows)
}

// Función encargada de obtener los eventos criticos de la DB y devolverlos convertidos ya en estructuras
func (model *IncidentModel) GetOpenIncidents() ([]Incident, error) {
	rows, err := model.db.Query(`
		SELECT *
		FROM incidentes
		WHERE estado = 'Abierto'
		ORDER BY fecha_creacion ASC
	`)

	if err != nil {
		return nil, err
	}

	return model.scanIncidents(rows)
}

func (model *IncidentModel) CreateIncident(incidente Incident) error {
	_, err := model.db.Exec(`
		INSERT INTO incidentes (alerta_id, descripcion)
		VALUES ($1, $2)
	`, incidente.AlertaID, incidente.Descripcion)
	return err
}
