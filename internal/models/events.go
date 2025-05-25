package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Event struct {
	ID          int64     `json:"id"`
	Incident    int64     `json:"incident"`
	EventTime   time.Time `json:"event_time"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	Endpoint    int64     `json:"endpoint"`
}

type EventModel struct {
	DB  *pgxpool.Pool
	CTX context.Context
}

func (m EventModel) GetByID(id int64) (*Event, error) {
	query := `
		SELECT event_id, incident_id, event_time, event_type, description, created_by, created_at, endpoint_id 
		FROM events
		WHERE event_id=$1
	`
	var event Event
	err := m.DB.QueryRow(m.CTX, query, id).Scan(
		&event.ID,
		&event.Incident,
		&event.EventTime,
		&event.EventType,
		&event.Description,
		&event.CreatedBy,
		&event.CreatedAt,
		&event.Endpoint,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &event, nil
}

func (m EventModel) Insert(event *Event) error {
	query := `
		INSERT INTO event (event_id, incident_id, event_time, event_type, description, created_by, created_at, endpoint_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING event_id
	`

	return m.DB.QueryRow(m.CTX, query, event.ID, event.Incident, event.EventTime,
		event.EventType, event.Description, event.CreatedBy,
		event.CreatedAt, event.Endpoint).Scan(
		&event.ID,
		&event.Incident,
		&event.EventTime,
		&event.EventType,
		&event.Description,
		&event.CreatedBy,
		&event.CreatedAt,
		&event.Endpoint,
	)
}
