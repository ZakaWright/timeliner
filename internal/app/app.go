package app

import (
	"context"
	"timeliner/internal/models"
	"timeliner/internal/services"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	DB     *pgxpool.Pool
	CTX    context.Context
	Models models.Models
	Auth   services.AuthService
}

func NewApp(db *pgxpool.Pool, ctx context.Context, jwtSecret []byte) *App {
	return &App{
		DB:     db,
		CTX:    ctx,
		Models: models.GetModels(db, ctx),
		Auth:   *services.NewAuthService(db, ctx, jwtSecret),
	}
}
