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

func (s *Storage) SavePerson(
	ctx context.Context,
	person models.Person,
) (int, error) {
	const op = "postgres.savePerson"

	query := `INSERT INTO person(full_name, phone) VALUES($1, $2) RETURNING id`
	row := s.db.QueryRow(ctx, query, person.Name, person.Phone)

	var personId int
	if err := row.Scan(&personId); err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
			}
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return personId, nil
}

func (s *Storage) UpdatePerson(
	ctx context.Context,
	person models.Person,
	pID int,
) (int, error) {
	const op = "postgres.updatePerson"

	query := `UPDATE person SET full_name = $1, phone = $2 WHERE id = $3 RETURNING id`
	row := s.db.QueryRow(ctx, query, person.Name, person.Phone, pID)

	var personId int
	if err := row.Scan(&personId); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
			}
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return personId, nil
}

func (s *Storage) DeletePerson(ctx context.Context, pID int) error {
	const op = "postgres.deletePerson"

	query := `DELETE FROM person WHERE id = $1`
	result, err := s.db.Exec(ctx, query, pID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
	}

	return nil
}

func (s *Storage) FindPersonByName(ctx context.Context, name string) (models.Person, error) {
	const op = "postgres.findPersonByName"

	query := `SELECT * FROM person WHERE full_name = $1`
	var person models.Person
	err := s.db.QueryRow(ctx, query, name).Scan(
		&person.Id,
		&person.Name,
		&person.Phone,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Person{}, fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
		}
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	return person, nil
}

func (s *Storage) FindAllPeople(ctx context.Context) ([]models.Person, error) {
	const op = "postgres.findAllPeople"

	query := `SELECT * FROM person`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Person])
}

func (s *Storage) FindPersonById(ctx context.Context, id int64) (models.Person, error) {
	const op = "postgres.findPersonById"

	query := `SELECT * FROM person WHERE id = $1`

	var person models.Person
	err := s.db.QueryRow(ctx, query, id).Scan(
		&person.Id,
		&person.Name,
		&person.Phone,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Person{}, fmt.Errorf("%s: %w", op, storage.ErrPersonNotFound)
		}
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	return person, nil
}
