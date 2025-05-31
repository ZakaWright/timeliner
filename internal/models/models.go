package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Models struct {
	Users     UserModel
	Events    EventModel
	Incident  IncidentModel
	Endpoints EndpointModel
}

func GetModels(db *pgxpool.Pool, ctx context.Context) Models {
	return Models{
		Users:     UserModel{DB: db, CTX: ctx},
		Events:    EventModel{DB: db, CTX: ctx},
		Incident:  IncidentModel{DB: db, CTX: ctx},
		Endpoints: EndpointModel{DB: db, CTX: ctx},
	}
}
