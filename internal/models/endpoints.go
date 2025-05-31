package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Endpoint struct {
	ID         int64      `json:"id"`
	Name       string     `json:'endpoint_name'`
	OS         string     `json:'os'`
	OSVersion  string     `json:'os_version'`
	IP         string     `json:'ip'`
	Mac        string     `json:'mac'`
	Last_Seen  *time.Time `json:'last_seen'`
	AddedAt    time.Time  `json:'added_at`
	IncidentID int64      `json:'incident_id'`
}

type EndpointModel struct {
	DB  *pgxpool.Pool
	CTX context.Context
}

func (m EndpointModel) GetByID(id int64) (*Endpoint, error) {
	query := `
		SELECT endpoint_id, device_name, os, os_version, ip_address, mac_address, last_seen 
		FROM endpoints
		WHERE event_id=$1
	`
	var endpoint Endpoint
	err := m.DB.QueryRow(m.CTX, query, id).Scan(
		&endpoint.ID,
		&endpoint.Name,
		&endpoint.OS,
		&endpoint.OSVersion,
		&endpoint.IP,
		&endpoint.Mac,
		&endpoint.Last_Seen,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &endpoint, nil
}

func (m EndpointModel) Insert(endpoint *Endpoint) (int64, error) {
	query := `
		INSERT INTO endpoints (device_name, os, os_version, ip_address, mac_address, last_seen, incident_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING endpoint_id
	`
	var endpoint_id int64
	err := m.DB.QueryRow(m.CTX, query, endpoint.Name, endpoint.OS, endpoint.OSVersion,
		endpoint.IP, endpoint.Mac, endpoint.Last_Seen, endpoint.IncidentID).Scan(
		&endpoint_id,
	)

	if err != nil {
		fmt.Printf("Error: %v", err)
		return -1, fmt.Errorf("failed to create new endpoint %v", err)
	}
	return endpoint_id, nil
}

func (m EndpointModel) GetByIncidentID(incident_id int64) ([]*Endpoint, error) {
	query := `
		SELECT endpoint_id, device_name, ip_address, os
		FROM endpoints
		WHERE incident_id=$1
		ORDER BY device_name
	`
	rows, err := m.DB.Query(m.CTX, query, incident_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endpoints []*Endpoint
	for rows.Next() {
		var endpoint Endpoint
		err := rows.Scan(
			&endpoint.ID,
			&endpoint.Name,
			&endpoint.IP,
			&endpoint.OS,
		)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, &endpoint)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return endpoints, nil
}

func (m EndpointModel) GetNamesByIncidentID(incident_id int64) ([]*Endpoint, error) {
	query := `
		SELECT endpoint_id, device_name, ip_address
		FROM endpoints
		WHERE incident_id=$1
		ORDER BY device_name
	`
	rows, err := m.DB.Query(m.CTX, query, incident_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endpoints []*Endpoint
	for rows.Next() {
		var endpoint Endpoint
		err := rows.Scan(
			&endpoint.ID,
			&endpoint.Name,
			&endpoint.IP,
		)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, &endpoint)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return endpoints, nil
}

// need to modify to get endpoint ID from events somehow
/*
func (m EndpointModel) GetEventsForIncident(incident_id int64) ([]*Endpoint, error) {
	query := `
		SELECT endpoint_id, device_name, os, os_version, ip_address, mac_address, last_seen
		FROM endpoints
		WHERE incident_id=$1
	`
	rows, err := m.DB.Query(m.CTX, query, incident_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*REvent
	for rows.Next() {
		var event REvent
		var createdBy int64
		err := rows.Scan(
			&event.ID,
			&event.EventTime,
			&event.EventType,
			&event.Description,
			createdBy,
			//&event.CreatedBy,
			&event.Endpoint,
		)
		if err != nil {
			return nil, err
		}

		user_model := UserModel(m)
		user, err := user_model.GetByID(createdBy)
		if err != nil {
			return nil, err
		}
		event.CreatedBy = user
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}
*/
