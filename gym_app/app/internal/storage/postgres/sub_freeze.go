package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"time"
)

// getFreezeDaysBySubscriptionID возвращает максимальное количество freeze_days для тарифа
func (s *Storage) getFreezeDaysBySubscriptionID(ctx context.Context, subscriptionID int) (int, error) {
	const query = `SELECT freeze_days FROM subscriptions WHERE id = $1`
	var freezeDays int
	err := s.db.QueryRow(ctx, query, subscriptionID).Scan(&freezeDays)
	if err != nil {
		return 0, err
	}
	return freezeDays, nil
}

// getUsedFreezeDays возвращает сумму дней заморозки по абонементу
func (s *Storage) getUsedFreezeDays(ctx context.Context, subscriptionNumber string) (int, error) {
	const query = `SELECT COALESCE(SUM(EXTRACT(DAY FROM (COALESCE(freeze_end, NOW()) - freeze_start))), 0)
					FROM subscription_freeze WHERE subscription_number = $1`
	var used int
	err := s.db.QueryRow(ctx, query, subscriptionNumber).Scan(&used)
	if err != nil {
		return 0, err
	}
	return used, nil
}

// getSubscriptionIDByNumber возвращает subscription_id по номеру абонемента
func (s *Storage) getSubscriptionIDByNumber(ctx context.Context, subscriptionNumber string) (int, error) {
	const query = `SELECT subscription_id FROM person_subscriptions WHERE number = $1`
	var id int
	err := s.db.QueryRow(ctx, query, subscriptionNumber).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// FreezeSubscription Заморозка абонемента с учетом лимита freeze_days
func (s *Storage) FreezeSubscription(ctx context.Context, subscriptionNumber string, freezeStart time.Time) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Получаем subscription_id
	subscriptionID, err := s.getSubscriptionIDByNumber(ctx, subscriptionNumber)
	if err != nil {
		return fmt.Errorf("не найден subscription_id: %w", err)
	}

	// Получаем лимит freeze_days
	maxFreezeDays, err := s.getFreezeDaysBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("ошибка получения freeze_days: %w", err)
	}
	if maxFreezeDays <= 0 {
		return errors.New("для этого тарифа нельзя замораживать абонемент")
	}

	// В новой логике freeze_end заранее не задаётся, days_used = 0
	// days_used будет вычисляться при разморозке.
	newFreezeDays := 0 // При постановке на заморозку пока 0

	// Вставляем запись о заморозке (freeze_end = NULL)
	const insertFreeze = `
		INSERT INTO subscription_freeze (subscription_number, freeze_start, freeze_end, days_used, created_at)
		VALUES ($1, $2, NULL, $3, NOW())
	`
	_, err = tx.Exec(ctx, insertFreeze, subscriptionNumber, freezeStart, newFreezeDays)
	if err != nil {
		return err
	}

	// Обновляем статус абонемента
	const updateStatus = `
		UPDATE person_subscriptions
		SET status = 'frozen'
		WHERE number = $1
	`
	_, err = tx.Exec(ctx, updateStatus, subscriptionNumber)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// UnfreezeSubscription Разморозка абонемента: обновляем freeze и меняем статус PersonSubscription на "active"
func (s *Storage) UnfreezeSubscription(ctx context.Context, subscriptionNumber string, unfreezeDate time.Time) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Обновляем запись о заморозке (делаем все с freeze_end IS NULL завершенными)
	const updateFreeze = `
		UPDATE subscription_freeze
		SET freeze_end = $2
		WHERE subscription_number = $1
	`
	ct, err := tx.Exec(ctx, updateFreeze, subscriptionNumber, unfreezeDate)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	// Обновляем статус абонемента
	const updateStatus = `
		UPDATE person_subscriptions
		SET status = 'active'
		WHERE number = $1
	`
	_, err = tx.Exec(ctx, updateStatus, subscriptionNumber)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetAllActiveFreeze Получение всех замороженных абонементов (по статусу person_subscriptions)
func (s *Storage) GetAllActiveFreeze(ctx context.Context) ([]models.SubscriptionFreeze, error) {
	const query = `
		SELECT sf.id, sf.subscription_number, sf.freeze_start, sf.freeze_end, sf.days_used, sf.created_at
		FROM subscription_freeze sf
		JOIN person_subscriptions ps ON ps.number = sf.subscription_number
		WHERE ps.status = 'frozen'
		ORDER BY sf.created_at DESC
	`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var freezes []models.SubscriptionFreeze
	for rows.Next() {
		var f models.SubscriptionFreeze
		var freezeEnd *time.Time
		err := rows.Scan(&f.ID, &f.SubscriptionNumber, &f.FreezeStart, &freezeEnd, &f.DaysUsed, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		if freezeEnd != nil {
			f.FreezeEnd = *freezeEnd
		}
		freezes = append(freezes, f)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return freezes, nil
}
