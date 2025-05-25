package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Incident struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CaseNumber  string    `json:"case_number"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   int64     `json:"created_by"`
	ClosedAt    time.Time `json:"closed_at"`
}

type IncidentModel struct {
	DB  *pgxpool.Pool
	CTX context.Context
}

func (m IncidentModel) GetByID(id int64) (*Incident, error) {
	query := `
		SELECT incident_id, name, description, case_number, status, created_at, created_by, closed_at
		FROM incidents
		WHERE incident_id=$1
	`
	var incident Incident
	err := m.DB.QueryRow(m.CTX, query, id).Scan(
		&incident.ID,
		&incident.Name,
		&incident.Description,
		&incident.CaseNumber,
		&incident.Status,
		&incident.CreatedAt,
		&incident.CreatedBy,
		&incident.ClosedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &incident, nil
}
