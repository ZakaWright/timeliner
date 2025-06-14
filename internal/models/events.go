package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgtype"
)

type Event struct {
	ID          int64     			`json:"id"`
	Incident    int64     			`json:"incident"`
	EventTime   pgtype.Timestamptz	`json:"event_time"`
	//EventTime   time.Time `json:"event_time"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	Endpoint    int64     `json:"endpoint"`
}

type Comment struct {
	ID        int64     `json:'id'`
	EventID   int64     `json:'event_id'`
	UserName  string    `json:'user'`
	Comment   string    `json:'comment'`
	CreatedAt time.Time `json:'created_at'`
}

type IOC struct {
	ID             int64     `json:'ioc_id`
	IocType        string    `json:'ioc_type'`
	IocDescription string    `json:'ioc_description`
	Value          string    `json:'ioc_value'`
	AddedAt        time.Time `json:'added_at'`
	AddedBy        string    `json:'added_by'`
	IsMalicious    bool      `json:'is_malicious'`
}

type IOCType struct {
	ID int64
	Name string
	Description string
	
}

type EventDetails struct {
	Event    *Event     `json:'event'`
	Comments []*Comment `json:'comments'`
	IOCs     []*IOC     `json:'iocs'`
}

type REvent struct {
	ID          int64     `json:"id"`
	Incident    int64     `json:"incident"`
	EventTime   pgtype.Timestamptz	`json:"event_time"`
	//EventTime   time.Time `json:"event_time"`
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
		return -1, fmt.Errorf("failed to create event %v", err)
	}
	return int64(id), nil
}

func (m EventModel) GetIOCTypes() ([]* IOCType, error) {
	query := `
		SELECT ioc_type_id, type_name, description
		FROM ioc_types
	`

	rows, err := m.DB.Query(m.CTX, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var iocTypes []*IOCType
	for rows.Next() {
		var iocType IOCType
		//var createdBy int64

		err := rows.Scan(
			&iocType.ID,
			&iocType.Name,
			&iocType.Description,
		)
		if err != nil {
			return nil, err
		}

		iocTypes = append(iocTypes, &iocType)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return iocTypes, nil
}

func (m EventModel) InsertIOC(event_id, added_by int64, ioc_type, value string) (int64, error) {
	// get ioc_type
	var ioc_type_id int32
	query := `
		SELECT ioc_type_id
		FROM ioc_types
		WHERE type_name=$1
	`
	err := m.DB.QueryRow(m.CTX, query, ioc_type).Scan(&ioc_type_id)
	if err != nil {
		return -1, fmt.Errorf("failed to get ioc_type_id %v", err)
	}
	// create the IOC
	var ioc_id int32
	query = `
		INSERT INTO iocs (added_by, ioc_type_id, value)
		VALUES ($1, $2, $3)
		RETURNING ioc_id
	`
	err = m.DB.QueryRow(m.CTX, query, added_by, ioc_type_id, value).Scan(&ioc_id)
	if err != nil {
		//if err.Contains("duplicate key value") {}
		return -1, fmt.Errorf("failed to create ioc %v", err)
	}

	// insert into the cross reference table
	query = `
		INSERT INTO event_iocs (event_id, ioc_id)
		VALUES ($1, $2)
	`
	_, err = m.DB.Exec(m.CTX, query, event_id, ioc_id)
	if err != nil {
		return -1, fmt.Errorf("failed to insert into event-ioc table event_id: %d, %v", event_id, err)
	}
	return int64(ioc_id), nil
}

func (m EventModel) GetEventsForIncident(incident_id int64) ([]*Event, error) {
	/*
	query := `
		SELECT event_id, event_time, event_type, description, created_by, endpoint_id
		FROM events
		WHERE incident_id=$1
		ORDER BY event_time
	`
	*/ 
	//TODO get the mitre tactic as well
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

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (m EventModel) GetEventDetails(id int64) (*EventDetails, error) {
	var eventDetails EventDetails
	//var event Event
	var comments []*Comment
	var iocs []*IOC

	event, err := m.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get event %v", err)
	}
	eventDetails.Event = event

	// get comments
	query := `
			SELECT event_comments.comment_id, event_comments.comment, users.username
			FROM event_comments
			INNER JOIN users ON event_comments.user_id = users.user_id
			WHERE event_comments.event_id=$1
		`
	rows, err := m.DB.Query(m.CTX, query, eventDetails.Event.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var comment Comment
		err := rows.Scan(
			&comment.ID,
			&comment.Comment,
			&comment.UserName,
		)
		if err != nil {
			return nil, err
		}

		comments = append(comments, &comment)
	}

	eventDetails.Comments = comments

	// get IOCs
	query = `
		SELECT event_iocs.ioc_id, ioc_types.type_name, ioc_types.description, iocs.value, users.username
		FROM event_iocs
		INNER JOIN iocs ON event_iocs.ioc_id = iocs.ioc_id
		INNER JOIN ioc_types ON iocs.ioc_type_id = ioc_types.ioc_type_id
		INNER JOIN users ON iocs.added_by = users.user_id
		WHERE event_iocs.event_id=$1
	`

	rows, err = m.DB.Query(m.CTX, query, eventDetails.Event.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var ioc IOC
		err := rows.Scan(
			&ioc.ID,
			&ioc.IocType,
			&ioc.IocDescription,
			&ioc.Value,
			&ioc.AddedBy,
		)
		if err != nil {
			return nil, err
		}

		iocs = append(iocs, &ioc)
	}

	eventDetails.IOCs = iocs

	return &eventDetails, nil
}

func (m EventModel) GetEventDetailsForIncident(incident_id int64) ([]*EventDetails, error) {
	var eventDetails []*EventDetails

	query := `
		SELECT event_id
		FROM events
		WHERE incident_id=$1
	`
	rows, err := m.DB.Query(m.CTX, query, incident_id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var eventID int32
		err := rows.Scan(&eventID)
		if err != nil {
			return nil, err
		}
		eventDetail, err := m.GetEventDetails(int64(eventID))
		if err != nil {
			return nil, err
		}
		eventDetails = append(eventDetails, eventDetail)
	}

	return eventDetails, nil

}

func (m EventModel) AddComment(event_id int64, user_id int64, comment string) error {
	query := `
		INSERT INTO event_comments (event_id, user_id, comment)
		VALUES ($1, $2, $3)
	`

	_, err := m.DB.Exec(m.CTX, query, event_id, user_id, comment)
	if err != nil {
		return err
	}
	return nil
}
