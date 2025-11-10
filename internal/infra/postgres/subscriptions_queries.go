package postgres

import (
	"github.com/Masterminds/squirrel"
	"github.com/shrtyk/subscriptions-service/internal/core/domain"
)

const (
	createQuery = `
		INSERT INTO subscriptions (service_name, monthly_cost, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at;
	`

	getByIDQuery = `
		SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1;
	`

	updateQuery = `
		UPDATE subscriptions
		SET service_name = $1, monthly_cost = $2, user_id = $3, start_date = $4, end_date = $5, updated_at = NOW()
		WHERE id = $6;
	`

	deleteQuery = `
		DELETE FROM subscriptions WHERE id = $1;
	`
)

func (r *subsRepo) buildListQuery(filter domain.SubscriptionFilter) (string, []any, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	queryBuilder := psql.Select(
		"id", "service_name", "monthly_cost", "user_id",
		"start_date", "end_date", "created_at", "updated_at",
	).From("subscriptions")

	if filter.UserID != nil {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"user_id": *filter.UserID})
	}
	if filter.ServiceName != nil {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"service_name": *filter.ServiceName})
	}

	pageSize := uint64(r.cfg.DefaultPageSize)
	if filter.PageSize != nil && *filter.PageSize > 0 {
		if *filter.PageSize > r.cfg.MaxPageSize {
			pageSize = uint64(r.cfg.MaxPageSize)
		} else {
			pageSize = uint64(*filter.PageSize)
		}
	}
	queryBuilder = queryBuilder.Limit(pageSize)

	page := 1
	if filter.Page != nil && *filter.Page > 0 {
		page = *filter.Page
	}
	offset := uint64((page - 1) * int(pageSize))
	queryBuilder = queryBuilder.Offset(offset)

	return queryBuilder.ToSql()
}
