package models

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgtype"
)

type Incident struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CaseNumber  string     `json:"case_number"`
	Status      string     `json:"status"`
	CreatedAt   pgtype.Timestamp	`json:"created_at"`
	//CreatedAt   time.Time  `json:"created_at"`
	CreatedBy   int64      `json:"created_by"`
	ClosedAt    *pgtype.Timestamp `json:"closed_at"` // pointer is to account for null values (unclosed incidents)
	//ClosedAt    *time.Time `json:"closed_at"` // pointer is to account for null values (unclosed incidents)
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

func (m IncidentModel) Insert(name, description, caseNumber, status string, userID int64) (int64, error) {
	// validate user
	userQuery := `
		SELECT is_active
		FROM users
		WHERE user_id=$1
	`
	var isValid bool
	err := m.DB.QueryRow(m.CTX, userQuery, userID).Scan(
		&isValid,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return -1, errors.New("no user found")
		}
		return -1, err
	}
	if !isValid {
		return -1, errors.New("user is inactive")
	}
	// insert new incident
	query := `
	INSERT INTO incidents (name, description, case_number, status, created_by)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING incident_id
	`
	var response int64
	err = m.DB.QueryRow(m.CTX, query, name, description, caseNumber, status, userID).Scan(
		&response,
	)
	if err != nil {
		return -1, fmt.Errorf("failed to create incident %v", err)
	}
	return response, nil
}

func (m IncidentModel) GetOpenIncidents() ([]*Incident, error) {
	query := `
	SELECT incident_id, name, case_number
	FROM incidents
	WHERE status='open'
	`

	rows, err := m.DB.Query(m.CTX, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*Incident
	for rows.Next() {
		var incident Incident
		err := rows.Scan(
			&incident.ID,
			&incident.Name,
			&incident.CaseNumber,
		)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, &incident)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return incidents, nil
}

func (m IncidentModel) GetClosedIncidents() ([]*Incident, error) {
	query := `
	SELECT incident_id, name, case_number
	FROM incidents
	WHERE status='closed'
	`

	rows, err := m.DB.Query(m.CTX, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*Incident
	for rows.Next() {
		var incident Incident
		err := rows.Scan(
			&incident.ID,
			&incident.Name,
			&incident.CaseNumber,
		)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, &incident)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return incidents, nil
}

func (m IncidentModel) GetAllIncidents() ([]*Incident, error) {
	query := `
	SELECT incident_id, name, status, case_number
	FROM incidents
	`

	rows, err := m.DB.Query(m.CTX, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*Incident
	for rows.Next() {
		var incident Incident
		err := rows.Scan(
			&incident.ID,
			&incident.Name,
			&incident.Status,
			&incident.CaseNumber,
		)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, &incident)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return incidents, nil
}

func (m IncidentModel) Close(id int64) error {
	query := `
		UPDATE incidents
		SET status = 'closed', closed_at = NOW()
		WHERE incident_id=$1
	`
	_, err := m.DB.Exec(m.CTX, query, id)
	return err
}

func (m IncidentModel) Reopen(id int64) error {
	query := `
		UPDATE incidents
		SET status = 'open'
		WHERE incident_id=$1
	`
	_, err := m.DB.Exec(m.CTX, query, id)
	return err
}
