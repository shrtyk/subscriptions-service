package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shrtyk/subscriptions-service/internal/config"
	"github.com/shrtyk/subscriptions-service/internal/core/domain"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/repos"
)

const (
	opCreate  = "subsRepo.Create"
	opGetByID = "subsRepo.GetByID"
	opUpdate  = "subsRepo.Update"
	opDelete  = "subsRepo.Delete"
	opList    = "subsRepo.List"
	opFindAll = "subsRepo.FindAll"
)

type subsRepo struct {
	db  DBTX
	cfg *config.RepoConfig
}

func NewSubsRepo(db DBTX, cfg *config.RepoConfig) *subsRepo {
	return &subsRepo{
		db:  db,
		cfg: cfg,
	}
}

func (r *subsRepo) Create(ctx context.Context, sub *domain.Subscription) error {
	err := r.db.QueryRowContext(
		ctx, createQuery, sub.ServiceName,
		sub.MonthlyCost, sub.UserID, sub.StartDate, sub.EndDate).
		Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repos.WrapErr(opCreate, repos.KindDuplicate, err)
		}
		return repos.WrapErr(opCreate, repos.KindUnknown, err)
	}

	return nil
}

func (r *subsRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	sub := &domain.Subscription{}
	err := r.db.QueryRowContext(ctx, getByIDQuery, id).Scan(
		&sub.ID, &sub.ServiceName, &sub.MonthlyCost, &sub.UserID,
		&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repos.WrapErr(opGetByID, repos.KindNotFound, err)
		}
		return nil, repos.WrapErr(opGetByID, repos.KindUnknown, err)
	}

	return sub, nil
}

func (r *subsRepo) Update(ctx context.Context, sub *domain.Subscription) error {
	res, err := r.db.ExecContext(ctx, updateQuery, sub.ServiceName, sub.MonthlyCost, sub.UserID, sub.StartDate, sub.EndDate, sub.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repos.WrapErr(opUpdate, repos.KindDuplicate, err)
		}
		return repos.WrapErr(opUpdate, repos.KindUnknown, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return repos.WrapErr(opUpdate, repos.KindUnknown, err)
	}
	if rowsAffected == 0 {
		return repos.NewErr(opUpdate, repos.KindNotFound)
	}

	return nil
}

func (r *subsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, deleteQuery, id)
	if err != nil {
		return repos.WrapErr(opDelete, repos.KindUnknown, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return repos.WrapErr(opDelete, repos.KindUnknown, err)
	}
	if rowsAffected == 0 {
		return repos.NewErr(opDelete, repos.KindNotFound)
	}

	return nil
}

func (r *subsRepo) List(
	ctx context.Context,
	filter domain.SubscriptionFilter,
) ([]domain.Subscription, error) {
	return r.listSubs(ctx, filter, r.buildListQuery)
}

func (r *subsRepo) ListAll(
	ctx context.Context,
	filter domain.SubscriptionFilter,
) ([]domain.Subscription, error) {
	return r.listSubs(ctx, filter, r.buildListAllQuery)
}

func (r *subsRepo) listSubs(
	ctx context.Context,
	filter domain.SubscriptionFilter,
	queryBuilder func(domain.SubscriptionFilter) (string, []any, error),
) ([]domain.Subscription, error) {
	query, args, err := queryBuilder(filter)
	if err != nil {
		return nil, repos.WrapErr(opFindAll, repos.KindUnknown, err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, repos.WrapErr(opFindAll, repos.KindUnknown, err)
	}
	defer rows.Close()

	subs := make([]domain.Subscription, 0)
	for rows.Next() {
		var sub domain.Subscription
		if err := rows.Scan(
			&sub.ID, &sub.ServiceName, &sub.MonthlyCost, &sub.UserID,
			&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
		); err != nil {
			return nil, repos.WrapErr(opFindAll, repos.KindUnknown, err)
		}
		subs = append(subs, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, repos.WrapErr(opFindAll, repos.KindUnknown, err)
	}

	return subs, nil
}
