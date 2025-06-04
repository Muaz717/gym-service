package postgres

import (
	"context"
	"errors"
	"github.com/Muaz717/gym_app/app/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"time"
)

// AddSingleVisit inserts a new single visit into the database.
func (s *Storage) AddSingleVisit(ctx context.Context, singleVis models.SingleVisit) error {
	const query = `
		INSERT INTO single_visits (visit_date, final_price)
		VALUES ($1, $2)
	`
	_, err := s.db.Exec(ctx, query, singleVis.VisitDate, singleVis.FinalPrice)
	return err
}

// GetAllSingleVisits retrieves all single visits from the database.
func (s *Storage) GetAllSingleVisits(ctx context.Context) ([]models.SingleVisit, error) {
	const query = `
		SELECT id, visit_date, final_price
		FROM single_visits
		ORDER BY visit_date DESC, id DESC
	`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var visits []models.SingleVisit
	for rows.Next() {
		var v models.SingleVisit
		err := rows.Scan(&v.Id, &v.VisitDate, &v.FinalPrice)
		if err != nil {
			return nil, err
		}
		visits = append(visits, v)
	}
	return visits, rows.Err()
}

// GetSingleVisitById retrieves a single visit by its ID.
func (s *Storage) GetSingleVisitById(ctx context.Context, id int) (models.SingleVisit, error) {
	const query = `
		SELECT id, visit_date, final_price
		FROM single_visits
		WHERE id = $1
	`
	row := s.db.QueryRow(ctx, query, id)

	var v models.SingleVisit
	err := row.Scan(&v.Id, &v.VisitDate, &v.FinalPrice)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.SingleVisit{}, nil // No visit found with the given ID
		}
		return models.SingleVisit{}, err
	}
	return v, nil
}

// GetSingleVisitsByDay retrieves all single visits for the specified date.
func (s *Storage) GetSingleVisitsByDay(ctx context.Context, date time.Time) ([]models.SingleVisit, error) {
	const query = `
		SELECT id, visit_date, final_price
		FROM single_visits
		WHERE visit_date = $1
		ORDER BY id DESC
	`
	rows, err := s.db.Query(ctx, query, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var visits []models.SingleVisit
	for rows.Next() {
		var v models.SingleVisit
		err := rows.Scan(&v.Id, &v.VisitDate, &v.FinalPrice)
		if err != nil {
			return nil, err
		}
		visits = append(visits, v)
	}
	return visits, rows.Err()
}

// GetSingleVisitsByPeriod retrieves all single visits within the specified period (inclusive).
func (s *Storage) GetSingleVisitsByPeriod(ctx context.Context, from, to time.Time) ([]models.SingleVisit, error) {
	const query = `
		SELECT id, visit_date, final_price
		FROM single_visits
		WHERE visit_date >= $1 AND visit_date <= $2
		ORDER BY visit_date DESC, id DESC
	`
	rows, err := s.db.Query(ctx, query, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var visits []models.SingleVisit
	for rows.Next() {
		var v models.SingleVisit
		err := rows.Scan(&v.Id, &v.VisitDate, &v.FinalPrice)
		if err != nil {
			return nil, err
		}
		visits = append(visits, v)
	}
	return visits, rows.Err()
}

// DeleteSingleVisit deletes a single visit by its ID.
func (s *Storage) DeleteSingleVisit(ctx context.Context, visitID int) error {
	const query = `
		DELETE FROM single_visits
		WHERE id = $1
	`
	_, err := s.db.Exec(ctx, query, visitID)
	if err != nil {
		return err
	}
	return nil
}
