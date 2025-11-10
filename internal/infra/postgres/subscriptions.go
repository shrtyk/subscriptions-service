package postgres

import (
	"database/sql"

	"github.com/shrtyk/subscriptions-service/internal/config"
)

const (
	opCreate  = "subsRepo.Create"
	opGetByID = "subsRepo.GetByID"
	opUpdate  = "subsRepo.Update"
	opDelete  = "subsRepo.Delete"
	opList    = "subsRepo.List"
)

type subsRepo struct {
	db  *sql.DB
	cfg *config.RepoConfig
}

func NewSubsRepo(db *sql.DB, cfg *config.RepoConfig) *subsRepo {
	return &subsRepo{
		db:  db,
		cfg: cfg,
	}
}
