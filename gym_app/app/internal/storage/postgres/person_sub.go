package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/models"
	"github.com/Muaz717/gym_app/app/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) AddPersonSub(ctx context.Context, personSub models.PersonSubscription) (string, error) {
	const op = "storage.postgres.AddPersonSub"

	query := `
		INSERT INTO person_subscriptions (number, person_id, subscription_id, start_date, end_date, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING number
	`

	var number string
	err := s.db.QueryRow(ctx, query,
		personSub.Number,
		personSub.PersonID,
		personSub.SubscriptionID,
		personSub.StartDate,
		personSub.EndDate,
		personSub.Status,
	).Scan(&number)

	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return "", fmt.Errorf("%s: %w", op, storage.ErrSubscriptionExists)
			case "23503":
				return "", fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
			}
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return number, nil
}

func (s *Storage) GetPersonSubByNumber(ctx context.Context, number string) (models.PersonSubscription, error) {
	const op = "storage.postgres.FindPersonSubByNumber"

	query := `SELECT * FROM person_subscriptions WHERE number = $1`

	var personSub models.PersonSubscription
	err := s.db.QueryRow(ctx, query, number).Scan(
		&personSub.Number,
		&personSub.PersonID,
		&personSub.SubscriptionID,
		&personSub.StartDate,
		&personSub.EndDate,
		&personSub.Status,
	)

	if err != nil {
		return models.PersonSubscription{}, fmt.Errorf("%s: %w", op, err)
	}

	return personSub, nil
}

func (s *Storage) DeletePersonSub(ctx context.Context, number string) error {
	const op = "storage.postgres.DeletePersonSub"

	query := `DELETE FROM person_subscriptions WHERE number = $1`

	result, err := s.db.Exec(ctx, query, number)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrSubscriptionNotFound)
	}

	return nil
}

func (s *Storage) GetAllPersonSubs(ctx context.Context) ([]models.PersonSubscription, error) {
	const op = "storage.postgres.GetAllPersonSubs"

	query := `SELECT * FROM person_subscriptions`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.PersonSubscription])
}

func (s *Storage) FindPersonSubByPersonName(ctx context.Context, name string) ([]models.PersonSubscription, error) {
	const op = "storage.postgres.FindPersonSubByPersonName"

	query := `
		SELECT ps.* FROM person_subscriptions ps
		JOIN person p ON ps.person_id = p.id
		WHERE p.full_name = $1
	`

	rows, err := s.db.Query(ctx, query, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.PersonSubscription])
}

func (s *Storage) UpdatePersonSubStatus(ctx context.Context, number string, status string) error {
	const op = "storage.postgres.UpdatePersonSubStatus"

	query := `UPDATE person_subscriptions SET status = $1 WHERE number = $2`

	result, err := s.db.Exec(ctx, query, status, number)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrSubscriptionNotFound)
	}

	return nil
}
