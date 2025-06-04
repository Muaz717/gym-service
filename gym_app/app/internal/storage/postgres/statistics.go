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
	// Открываем транзакцию
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("MonthlyStatistics: begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// 1. Получаем по месяцам данные по абонементам
	const subsQuery = `
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

	subRows, err := tx.Query(ctx, subsQuery, from, to)
	if err != nil {
		return nil, fmt.Errorf("MonthlyStatistics (subscriptions query): %w", err)
	}
	defer subRows.Close()

	subsMap := make(map[string]*dto.MonthlyStat)
	for subRows.Next() {
		var stat dto.MonthlyStat
		err := subRows.Scan(&stat.Month, &stat.Income, &stat.NewClients, &stat.SoldSubscriptions)
		if err != nil {
			return nil, fmt.Errorf("MonthlyStatistics subs rows.Scan: %w", err)
		}
		monthKey := stat.Month.Format("2006-01")
		cp := stat
		subsMap[monthKey] = &cp
	}
	if err := subRows.Err(); err != nil {
		return nil, fmt.Errorf("MonthlyStatistics subs rows.Err: %w", err)
	}

	// 2. Получаем данные по разовым посещениям (single_visits)
	const visitsQuery = `
		SELECT
			DATE_TRUNC('month', visit_date) as month,
			COALESCE(SUM(final_price), 0) as single_visits_income,
			COUNT(*) as single_visits_count
		FROM single_visits
		WHERE visit_date >= $1 AND visit_date <= $2
		GROUP BY month
		ORDER BY month
	`

	visitRows, err := tx.Query(ctx, visitsQuery, from, to)
	if err != nil {
		return nil, fmt.Errorf("MonthlyStatistics (single_visits query): %w", err)
	}
	defer visitRows.Close()

	for visitRows.Next() {
		var month time.Time
		var singleVisitsIncome float64
		var singleVisitsCount int
		err := visitRows.Scan(&month, &singleVisitsIncome, &singleVisitsCount)
		if err != nil {
			return nil, fmt.Errorf("MonthlyStatistics visits rows.Scan: %w", err)
		}
		monthKey := month.Format("2006-01")
		stat, ok := subsMap[monthKey]
		if ok {
			stat.SingleVisitsIncome = singleVisitsIncome
			stat.SingleVisitsCount = singleVisitsCount
		} else {
			subsMap[monthKey] = &dto.MonthlyStat{
				Month:              month,
				SingleVisitsIncome: singleVisitsIncome,
				SingleVisitsCount:  singleVisitsCount,
			}
		}
	}
	if err := visitRows.Err(); err != nil {
		return nil, fmt.Errorf("MonthlyStatistics visits rows.Err: %w", err)
	}

	// 3. Генерируем месяцы диапазона от from до to (включительно)
	var months []time.Time
	start := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
	end := time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, to.Location())
	for m := start; !m.After(end); m = m.AddDate(0, 1, 0) {
		months = append(months, m)
	}

	// 4. Собираем финальный слайс, гарантируя, что все месяцы на месте
	stats := make([]dto.MonthlyStat, 0, len(months))
	for _, m := range months {
		monthKey := m.Format("2006-01")
		stat, ok := subsMap[monthKey]
		if ok {
			stat.Month = m
			stats = append(stats, *stat)
		} else {
			stats = append(stats, dto.MonthlyStat{
				Month:              m,
				Income:             0,
				NewClients:         0,
				SoldSubscriptions:  0,
				SingleVisitsIncome: 0,
				SingleVisitsCount:  0,
			})
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("MonthlyStatistics: commit: %w", err)
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

func (s *Storage) TotalSingleVisits(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM single_visits`

	var count int
	err := s.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("TotalSingleVisits: %w", err)
	}
	return count, nil
}

func (s *Storage) SingleVisits(ctx context.Context, from, to time.Time) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM single_visits
		WHERE visit_date >= $1 AND visit_date <= $2
	`

	var count int
	err := s.db.QueryRow(ctx, query, from, to)
	if err != nil {
		return 0, fmt.Errorf("SingleVisits: %w", err)
	}
	return count, nil
}

func (s *Storage) SingleVisitsIncome(ctx context.Context) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(final_price), 0)
		FROM single_visits
	`

	var income float64
	err := s.db.QueryRow(ctx, query).Scan(&income)
	if err != nil {
		return 0, fmt.Errorf("SingleVisitsIncome: %w", err)
	}
	return income, nil
}
