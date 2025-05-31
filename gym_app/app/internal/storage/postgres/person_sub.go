package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/domain/dto"
	"github.com/Muaz717/gym_app/app/internal/domain/models"
	"github.com/Muaz717/gym_app/app/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Добавляем поля subscription_price, final_price, freeze_days и used_freeze_days в запросы/запись

func (s *Storage) AddPersonSub(ctx context.Context, personSub models.PersonSubscription) (string, error) {
	const op = "storage.postgres.AddPersonSub"

	query := `
		INSERT INTO person_subscriptions (
			number, person_id, subscription_id, subscription_price, start_date, end_date, status, discount, final_price
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING number
	`

	var number string
	err := s.db.QueryRow(ctx, query,
		personSub.Number,
		personSub.PersonID,
		personSub.SubscriptionID,
		personSub.SubscriptionPrice,
		personSub.StartDate,
		personSub.EndDate,
		personSub.Status,
		personSub.Discount,
		personSub.FinalPrice,
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

func (s *Storage) GetPersonSubByNumber(ctx context.Context, number string) (dto.PersonSubResponse, error) {
	const op = "storage.postgres.GetPersonSubByNumber"

	query := `
		SELECT 	ps.number,
		        ps.person_id,
		        ps.subscription_id,
		        s.title AS subscription_title,
				ps.subscription_price,
		        ps.start_date,
		        ps.end_date,
		        ps.status,
       			p.full_name AS person_name,
				ps.discount,
				ps.final_price,
				s.freeze_days,
				COALESCE((
					SELECT SUM(EXTRACT(DAY FROM (COALESCE(freeze_end, NOW()) - freeze_start)))
					FROM subscription_freeze
					WHERE subscription_number = ps.number
				), 0) as used_freeze_days
		FROM person_subscriptions ps
		JOIN person p ON ps.person_id = p.id
		JOIN subscriptions s ON ps.subscription_id = s.id
		WHERE ps.number = $1
	`

	var personSub dto.PersonSubResponse
	err := s.db.QueryRow(ctx, query, number).Scan(
		&personSub.Number,
		&personSub.PersonID,
		&personSub.SubscriptionID,
		&personSub.SubscriptionTitle,
		&personSub.SubscriptionPrice,
		&personSub.StartDate,
		&personSub.EndDate,
		&personSub.Status,
		&personSub.PersonName,
		&personSub.Discount,
		&personSub.FinalPrice,
		&personSub.FreezeDays,
		&personSub.UsedFreezeDays,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.PersonSubResponse{}, fmt.Errorf("%s: %w", op, storage.ErrSubscriptionNotFound)
		}
		return dto.PersonSubResponse{}, fmt.Errorf("%s: %w", op, err)
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

func (s *Storage) GetAllPersonSubs(ctx context.Context) ([]dto.PersonSubResponse, error) {
	const op = "storage.postgres.GetAllPersonSubs"

	query := `
	SELECT
    	ps.number,
    	ps.person_id,
		ps.subscription_id,
		s.title AS subscription_title,
		ps.subscription_price,
		ps.start_date,
		ps.end_date,
		ps.status,
		p.full_name AS person_name,
		ps.discount,
		ps.final_price,
		s.freeze_days,
    	COALESCE((
			SELECT SUM(EXTRACT(DAY FROM (COALESCE(freeze_end, NOW()) - freeze_start)))
			FROM subscription_freeze
			WHERE subscription_number = ps.number
    	), 0) as used_freeze_days
	FROM person_subscriptions ps
	JOIN person p ON ps.person_id = p.id
	JOIN subscriptions s ON ps.subscription_id = s.id
	ORDER BY
    	ps.number ~ '[^0-9]',
    	CASE WHEN ps.number ~ '^[0-9]+$' THEN CAST(ps.number AS INTEGER) END DESC
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var result []dto.PersonSubResponse
	for rows.Next() {
		var sub dto.PersonSubResponse
		err := rows.Scan(
			&sub.Number,
			&sub.PersonID,
			&sub.SubscriptionID,
			&sub.SubscriptionTitle,
			&sub.SubscriptionPrice,
			&sub.StartDate,
			&sub.EndDate,
			&sub.Status,
			&sub.PersonName,
			&sub.Discount,
			&sub.FinalPrice,
			&sub.FreezeDays,
			&sub.UsedFreezeDays,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		result = append(result, sub)
	}
	return result, nil
}

func (s *Storage) FindPersonSubByPersonName(ctx context.Context, name string) ([]dto.PersonSubResponse, error) {
	const op = "storage.postgres.FindPersonSubByPersonName"

	// Шаг 1: Проверка существования пользователя
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM person WHERE full_name = $1)", name).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("%s: check person existence: %w", op, err)
	}
	if !exists {
		return nil, storage.ErrPersonNotFound
	}

	// Шаг 2: Запрос на получение абонементов
	query := `
		SELECT ps.number,
		       ps.person_id,
		       ps.subscription_id,
		       s.title AS subscription_title,
			   ps.subscription_price,
		       ps.start_date,
		       ps.end_date,
		       ps.status,
		       p.full_name AS person_name,
			   ps.discount,
			   ps.final_price,
			   s.freeze_days,
			   COALESCE((
					SELECT SUM(EXTRACT(DAY FROM (COALESCE(freeze_end, NOW()) - freeze_start)))
					FROM subscription_freeze
					WHERE subscription_number = ps.number
			   ), 0) as used_freeze_days
		FROM person_subscriptions ps
		JOIN person p ON ps.person_id = p.id
		JOIN subscriptions s ON ps.subscription_id = s.id
		WHERE p.full_name = $1
	`

	rows, err := s.db.Query(ctx, query, name)
	if err != nil {
		return nil, fmt.Errorf("%s: query subscriptions: %w", op, err)
	}
	defer rows.Close()

	var result []dto.PersonSubResponse
	for rows.Next() {
		var sub dto.PersonSubResponse
		err := rows.Scan(
			&sub.Number,
			&sub.PersonID,
			&sub.SubscriptionID,
			&sub.SubscriptionTitle,
			&sub.SubscriptionPrice,
			&sub.StartDate,
			&sub.EndDate,
			&sub.Status,
			&sub.PersonName,
			&sub.Discount,
			&sub.FinalPrice,
			&sub.FreezeDays,
			&sub.UsedFreezeDays,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		result = append(result, sub)
	}

	if len(result) == 0 {
		return nil, storage.ErrSubscriptionNotFound
	}

	return result, nil
}

func (s *Storage) FindPersonSubByPersonId(ctx context.Context, personId int) ([]dto.PersonSubResponse, error) {
	const op = "storage.postgres.FindPersonSubByPersonId"

	// 1. Проверка: существует ли пользователь с таким ID
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM person WHERE id = $1)", personId).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("%s: check person existence: %w", op, err)
	}
	if !exists {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
	}

	// 2. Запрос на получение абонементов
	query := `
		SELECT ps.number,
		       ps.person_id,
		       ps.subscription_id,
		       s.title AS subscription_title,
			   ps.subscription_price,
		       ps.start_date,
		       ps.end_date,
		       ps.status,
		       p.full_name AS person_name,
			   ps.discount,
			   ps.final_price,
			   s.freeze_days,
			   COALESCE((
					SELECT SUM(EXTRACT(DAY FROM (COALESCE(freeze_end, NOW()) - freeze_start)))
					FROM subscription_freeze
					WHERE subscription_number = ps.number
			   ), 0) as used_freeze_days
		FROM person_subscriptions ps
		JOIN person p ON ps.person_id = p.id
		JOIN subscriptions s ON ps.subscription_id = s.id
		WHERE ps.person_id = $1
	`

	rows, err := s.db.Query(ctx, query, personId)
	if err != nil {
		return nil, fmt.Errorf("%s: query subscriptions: %w", op, err)
	}
	defer rows.Close()

	var result []dto.PersonSubResponse
	for rows.Next() {
		var sub dto.PersonSubResponse
		err := rows.Scan(
			&sub.Number,
			&sub.PersonID,
			&sub.SubscriptionID,
			&sub.SubscriptionTitle,
			&sub.SubscriptionPrice,
			&sub.StartDate,
			&sub.EndDate,
			&sub.Status,
			&sub.PersonName,
			&sub.Discount,
			&sub.FinalPrice,
			&sub.FreezeDays,
			&sub.UsedFreezeDays,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		result = append(result, sub)
	}

	// 3. Если абонементы не найдены
	if len(result) == 0 {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrSubscriptionNotFound)
	}

	return result, nil
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
