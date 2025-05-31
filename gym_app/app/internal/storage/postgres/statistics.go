package postgres

import (
	"context"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/domain/dto"
	"time"
)

// Методы статистики реализуются на основной структуре Storage

// MonthlyStatistics возвращает агрегированные данные по месяцам для статистики
func (s *Storage) MonthlyStatistics(ctx context.Context, from, to time.Time) ([]dto.MonthlyStat, error) {
	const query = `
			SELECT
				DATE_TRUNC('month', ps.start_date) as month,
				COALESCE(SUM(s.price - ps.discount), 0) as income,
				COUNT(DISTINCT ps.person_id) as new_clients,
				COUNT(*) as sold_subscriptions
			FROM person_subscriptions ps
			JOIN subscriptions s ON ps.subscription_id = s.id
			WHERE ps.start_date >= $1 AND ps.start_date <= $2
			GROUP BY month
			ORDER BY month
		`

	rows, err := s.db.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("MonthlyStatistics: %w", err)
	}
	defer rows.Close()

	var stats []dto.MonthlyStat
	for rows.Next() {
		var stat dto.MonthlyStat
		err := rows.Scan(&stat.Month, &stat.Income, &stat.NewClients, &stat.SoldSubscriptions)
		if err != nil {
			return nil, fmt.Errorf("MonthlyStatistics rows.Scan: %w", err)
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("MonthlyStatistics rows.Err: %w", err)
	}
	return stats, nil
}

// TotalClients возвращает общее количество клиентов
func (s *Storage) TotalClients(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM person`
	var total int
	err := s.db.QueryRow(ctx, query).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("TotalClients: %w", err)
	}
	return total, nil
}

// NewClients возвращает количество новых клиентов за период (по дате первого абонемента)
func (s *Storage) NewClients(ctx context.Context, from, to time.Time) (int, error) {
	const query = `
		SELECT COUNT(DISTINCT ps.person_id)
		FROM person_subscriptions ps
		WHERE ps.start_date >= $1 AND ps.start_date <= $2
	`
	var count int
	err := s.db.QueryRow(ctx, query, from, to).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("NewClients: %w", err)
	}
	return count, nil
}

// TotalIncome возвращает общий доход с учетом скидок
func (s *Storage) TotalIncome(ctx context.Context) (float64, error) {
	const query = `
			SELECT COALESCE(SUM(s.price - ps.discount), 0)
			FROM person_subscriptions ps
			JOIN subscriptions s ON ps.subscription_id = s.id
		`
	var income float64
	err := s.db.QueryRow(ctx, query).Scan(&income)
	if err != nil {
		return 0, fmt.Errorf("TotalIncome: %w", err)
	}
	return income, nil
}

// Income возвращает доход за период с учетом скидок
func (s *Storage) Income(ctx context.Context, from, to time.Time) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(s.price - ps.discount), 0)
		FROM person_subscriptions ps
		JOIN subscriptions s ON ps.subscription_id = s.id
		WHERE ps.start_date >= $1 AND ps.start_date <= $2
	`
	var income float64
	err := s.db.QueryRow(ctx, query, from, to).Scan(&income)
	if err != nil {
		return 0, fmt.Errorf("Income: %w", err)
	}
	return income, nil
}

func (s *Storage) TotalSoldSubscriptions(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM person_subscriptions`
	var count int
	err := s.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("TotalSoldSubscriptions: %w", err)
	}
	return count, nil
}

// SoldSubscriptions возвращает количество проданных абонементов за период
func (s *Storage) SoldSubscriptions(ctx context.Context, from, to time.Time) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM person_subscriptions
		WHERE start_date >= $1 AND start_date <= $2
	`
	var count int
	err := s.db.QueryRow(ctx, query, from, to).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("SoldSubscriptions: %w", err)
	}
	return count, nil
}
