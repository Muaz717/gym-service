package personService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	"github.com/Muaz717/gym_app/app/internal/models"
	"github.com/Muaz717/gym_app/app/internal/services/cache"
	"github.com/Muaz717/gym_app/app/internal/storage"

	"log/slog"
	"time"
)

type PersonCache interface {
	cache.Cache
}

type PersonStorage interface {
	SavePerson(ctx context.Context, person models.Person) (int, error)
	FindAllPeople(ctx context.Context) ([]models.Person, error)
	UpdatePerson(ctx context.Context, person models.Person, pID int) (int, error)
	DeletePerson(ctx context.Context, pID int) error
	FindPersonByName(ctx context.Context, name string) (models.Person, error)
}

type PersonService struct {
	log           *slog.Logger
	personStorage PersonStorage
	personCache   PersonCache
}

func New(
	log *slog.Logger,
	personStorage PersonStorage,
	personCache PersonCache,
) *PersonService {
	return &PersonService{
		log:           log,
		personStorage: personStorage,
		personCache:   personCache,
	}
}

var (
	ErrPersonExists   = errors.New("person already exists")
	ErrPersonNotFound = errors.New("person not found")
)

func (p *PersonService) AddPerson(ctx context.Context, person models.Person) (int, error) {

	const op = "services.person.addPerson"

	log := p.log.With(
		slog.String("op", op),
	)

	log.Info("Registering new user")

	personId, err := p.personStorage.SavePerson(ctx, person)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", sl.Error(err))

			return 0, fmt.Errorf("%s: %w", op, ErrPersonExists)
		}

		return 0, err
	}

	// Инвалидируем кэш списка всех пользователей
	if err := p.personCache.Delete(ctx, "people:all"); err != nil {
		log.Warn("failed to invalidate cache", sl.Error(err))
	}

	log.Info("person registered", "pid", personId)

	return personId, nil
}

func (p *PersonService) UpdatePerson(ctx context.Context, person models.Person, pID int) (int, error) {
	const op = "services.person.UpdatePerson"

	log := p.log.With(
		slog.String("op", op),
	)

	log.Info("Updating user")

	personId, err := p.personStorage.UpdatePerson(ctx, person, pID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Warn("user not found", sl.Error(err))

			return 0, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}

		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", sl.Error(err))

			return 0, fmt.Errorf("%s: %w", op, ErrPersonExists)
		}

		return 0, err
	}

	cacheKey := fmt.Sprintf("person:name:%s", person.Name)
	if err := p.personCache.Delete(ctx, cacheKey); err != nil {
		log.Warn("failed to invalidate cache", sl.Error(err))
	}
	if err := p.personCache.Delete(ctx, "people:all"); err != nil {
		log.Warn("failed to invalid cache", sl.Error(err))
	}

	log.Info("person updated", "pid", personId)

	return personId, nil
}

func (p *PersonService) DeletePerson(ctx context.Context, pID int) error {
	const op = "services.person.DeletePerson"

	log := p.log.With(
		slog.String("op", op),
	)

	log.Info("Deleting user")

	err := p.personStorage.DeletePerson(ctx, pID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Warn("user not found", sl.Error(err))

			return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
	}

	if err := p.personCache.Delete(ctx, "people:all"); err != nil {
		log.Warn("failed to invalid cache", sl.Error(err))
	}

	log.Info("person deleted")
	return nil
}

func (p *PersonService) FindPersonByName(ctx context.Context, name string) (models.Person, error) {
	const op = "services.person.FindPersonByName"

	log := p.log.With(
		slog.String("op", op),
	)

	log.Info("Finding user by name")

	cacheKey := fmt.Sprintf("person:name:%s", name)
	if cached, err := p.personCache.Get(ctx, cacheKey); err == nil {
		var person models.Person
		if err := json.Unmarshal([]byte(cached), &person); err == nil {
			log.Info("cache hit", slog.String("key", cacheKey))
			return person, nil
		}
		log.Warn("failed to unmarshal cached data", sl.Error(err))
	}

	person, err := p.personStorage.FindPersonByName(ctx, name)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Warn("user not found", sl.Error(err))

			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
	}

	if data, err := json.Marshal(person); err == nil {
		if err := p.personCache.Set(ctx, cacheKey, data, 10*time.Minute); err != nil {
			log.Warn("failed to set cache", sl.Error(err))
		}
	}

	log.Info("person found")
	return person, nil
}

func (p *PersonService) FindAllPeople(ctx context.Context) ([]models.Person, error) {
	const op = "services.person.FindAllPeople"

	log := p.log.With(
		slog.String("op", op),
	)

	log.Info("Starting to find people")

	cacheKey := fmt.Sprintf("people:all")
	if cached, err := p.personCache.Get(ctx, cacheKey); err == nil {
		var allPeople []models.Person
		if err := json.Unmarshal([]byte(cached), &allPeople); err == nil {
			log.Info("cache hit", slog.String("key", cacheKey))
			return allPeople, nil
		}
	}

	allPeople, err := p.personStorage.FindAllPeople(ctx)
	if err != nil {
		log.Warn("error", sl.Error(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if data, err := json.Marshal(allPeople); err == nil {
		if err := p.personCache.Set(ctx, cacheKey, data, 30*time.Minute); err != nil {
			log.Warn("failed to set cache", sl.Error(err))
		}
	}

	log.Info("People are found")
	return allPeople, nil
}
