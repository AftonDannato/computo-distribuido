package models

import (
	"database/sql"
	"time"
)

// Estructura que almacena un registro de ciberseguridad
type Alert struct {
	ID        int
	EventoID  int
	Regla     string
	Severidad string
	Fecha     time.Time
}

// Modelo dónde se guarda la conexión a la DB (inyección)
type DetectionModel struct {
	db *sql.DB
}

// Función que se llama para crear el modelo con la conexión inyectada
func NewDetectionModel(db *sql.DB) *DetectionModel {
	return &DetectionModel{
		db: db,
	}
}

// Función encargada de leer  registros obtenidos de una query y convertirlos a estructuras Alert
func (model *DetectionModel) scanAlerts(rows *sql.Rows) ([]Alert, error) {
	defer rows.Close()

	var alertas []Alert

	for rows.Next() {
		var alerta Alert

		err := rows.Scan(
			&alerta.ID,
			&alerta.EventoID,
			&alerta.Regla,
			&alerta.Severidad,
			&alerta.Fecha,
		)

		if err != nil {
			return nil, err
		}

		alertas = append(alertas, alerta)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return alertas, nil
}

// Función encargada de obtener todos los eventos de la DB y devolverlos convertidos ya en estructuras
func (model *DetectionModel) GetAllAlerts() ([]Alert, error) {
	rows, err := model.db.Query(`
		SELECT *
		FROM alertas
		ORDER BY fecha DESC
	`)

	if err != nil {
		return nil, err
	}

	return model.scanAlerts(rows)
}

// Función encargada de obtener los eventos criticos de la DB y devolverlos convertidos ya en estructuras
func (model *DetectionModel) GetCriticalAlerts() ([]Alert, error) {
	rows, err := model.db.Query(`
		SELECT *
		FROM alertas
		WHERE severidad IN ('Crítica')
		ORDER BY fecha DESC
	`)

	if err != nil {
		return nil, err
	}

	return model.scanAlerts(rows)
}

func (model *DetectionModel) CreateAlert(alerta Alert) error {
	_, err := model.db.Exec(`
		INSERT INTO alertas (evento_id, regla, severidad)
		VALUES ($1, $2, $3)
	`, alerta.EventoID, alerta.Regla, alerta.Severidad)
	return err
}
