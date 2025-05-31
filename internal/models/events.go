package models

import (
	"context"
	"fmt"
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

type Comment struct {
	ID        int64     `json:'id'`
	EventID   int64     `json:'event_id'`
	User      *User     `json:'user'`
	Comment   string    `json:'comment'`
	CreatedAt time.Time `json:'created_at'`
}

type EventDetails struct {
	ID          int64      `json:"id"`
	Incident    int64      `json:"incident"`
	EventTime   time.Time  `json:"event_time"`
	EventType   string     `json:"event_type"`
	Description string     `json:"description"`
	CreatedBy   *User      `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	Endpoint    int64      `json:"endpoint"`
	Comments    []*Comment `json:'comments'`
}

type REvent struct {
	ID          int64     `json:"id"`
	Incident    int64     `json:"incident"`
	EventTime   time.Time `json:"event_time"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	CreatedBy   *User     `json:"created_by"`
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
		//&createdBy,
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

func (m EventModel) Insert(event *Event) (int64, error) {
	query := `
		INSERT INTO events (incident_id, event_time, event_type, description, created_by, endpoint_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING event_id
	`
	var id int32
	err := m.DB.QueryRow(m.CTX, query, event.Incident, event.EventTime,
		event.EventType, event.Description, event.CreatedBy,
		event.Endpoint).Scan(
		&id,
	)

	if err != nil {
		fmt.Printf("Error in event creation: %v", err)
		return -1, fmt.Errorf("failed to create event %v", err)
	}
	fmt.Println("No error in event creation")
	return int64(id), nil
}

func (m EventModel) GetEventsForIncident(incident_id int64) ([]*Event, error) {
	query := `
		SELECT event_id, event_time, event_type, description, created_by, endpoint_id
		FROM events
		WHERE incident_id=$1
		ORDER BY event_time
	`
	rows, err := m.DB.Query(m.CTX, query, incident_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var event Event
		//var createdBy int64
		err := rows.Scan(
			&event.ID,
			&event.EventTime,
			&event.EventType,
			&event.Description,
			&event.CreatedBy,
			&event.Endpoint,
		)
		if err != nil {
			return nil, err
		}

		/*
			user_model := UserModel(m)
			user, err := user_model.GetByID(createdBy)
			if err != nil {
				return nil, err
			}
			event.CreatedBy = user
		*/
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (m EventModel) GetEventDetails(id int64) (*EventDetails, error) {
	/*
			type Comment struct {
			ID        int64     `json:'id'`
			EventID   int64     `json:'event_id'`
			User      *User     `json:'user'`
			Comment   string    `json:'comment'`
			CreatedAt time.Time `json:'created_at'`
			}

			type EventDetails struct {
				ID          int64      `json:"id"`
				Incident    int64      `json:"incident"`
				EventTime   time.Time  `json:"event_time"`
				EventType   string     `json:"event_type"`
				Description string     `json:"description"`
				CreatedBy   *User      `json:"created_by"`
				CreatedAt   time.Time  `json:"created_at"`
				Endpoint    int64      `json:"endpoint"`
				Comments    []*Comment `json:'comments'`
			}
		*

		var event Event
		var Comments *[]Comment

		query := `
		 SELECT events.event_time, events.event_type, ... , event_comments.comment, ... FROM events
		 INNER JOIN event_comments ON events.event_id = event_comments.event_id
		INNER JOIN users ON users.user_id = event_comments.user_id
		where events.event_id=$1
		`
		rows, err := m.DB.Query(m.CTX, query, incident_id)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var events []*Event
		for rows.Next() {
			var event Event
			//var createdBy int64
			err := rows.Scan(
				&event.ID,
				&event.EventTime,
				&event.EventType,
				&event.Description,
				&event.CreatedBy,
				&event.Endpoint,
			)
			if err != nil {
				return nil, err
			}

			/*
				user_model := UserModel(m)
				user, err := user_model.GetByID(createdBy)
				if err != nil {
					return nil, err
				}
				event.CreatedBy = user
			*
			events = append(events, &event)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}
		return events, nil
	*/
	return nil, nil
}

func (m EventModel) AddComment(event_id int64, user_id int64, comment string) error {
	return nil
}
