package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/domain/models"
	"github.com/Muaz717/gym_app/app/internal/storage"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) SaveSubscription(
	ctx context.Context,
	subscription models.Subscription,
) (int, error) {
	const op = "postgres.addSubscription"

	query := `INSERT INTO subscriptions(title, price, duration_days, freeze_days) VALUES($1, $2, $3, $4) RETURNING id`

	row := s.db.QueryRow(ctx, query, subscription.Title, subscription.Price, subscription.DurationDays, subscription.FreezeDays)

	var subId int
	if err := row.Scan(&subId); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return subId, nil
}

func (s *Storage) UpdateSubscription(
	ctx context.Context,
	subscription models.Subscription,
	subID int,
) (int, error) {
	const op = "postgres.updateSubscription"

	query := `UPDATE subscriptions SET title = $1, price = $2, duration_days = $3, freeze_days = $4 WHERE id = $5 RETURNING id`

	row := s.db.QueryRow(ctx, query, subscription.Title, subscription.Price, subscription.DurationDays, subscription.FreezeDays, subID)

	var subId int
	if err := row.Scan(&subId); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return subId, nil
}

func (s *Storage) DeleteSubscription(
	ctx context.Context,
	subID int,
) error {
	const op = "postgres.deleteSubscription"

	query := `DELETE FROM subscriptions WHERE id = $1`

	_, err := s.db.Exec(ctx, query, subID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) FindAllSubscriptions(ctx context.Context) ([]models.Subscription, error) {
	const op = "postgres.FindAllSubscriptions"

	query := `SELECT * FROM subscriptions`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var subs []models.Subscription
	for rows.Next() {
		sub := models.Subscription{}
		err := rows.Scan(&sub.ID, &sub.Title, &sub.Price, &sub.DurationDays, &sub.FreezeDays)

		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}
		subs = append(subs, sub)
	}

	return subs, nil
}
