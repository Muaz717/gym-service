package personService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/domain/models"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	"github.com/Muaz717/gym_app/app/internal/services/cache"
	"github.com/Muaz717/gym_app/app/internal/storage"

	"log/slog"
	"time"
)

type PersonCache interface {
	cache.Cache
}

type StatCache interface {
	cache.Cache
	DelByPrefix(ctx context.Context, prefix string) error
}

type PersonStorage interface {
	SavePerson(ctx context.Context, person models.Person) (int, error)
	FindAllPeople(ctx context.Context) ([]models.Person, error)
	UpdatePerson(ctx context.Context, person models.Person, pID int) (int, error)
	DeletePerson(ctx context.Context, pID int) error
	FindPersonByName(ctx context.Context, name string) ([]models.Person, error)
	FindPersonById(ctx context.Context, id int) (models.Person, error)
}

type PersonService struct {
	log           *slog.Logger
	personStorage PersonStorage
	personCache   PersonCache
	statCache     StatCache
}

func New(
	log *slog.Logger,
	personStorage PersonStorage,
	personCache PersonCache,
	statCache StatCache,
) *PersonService {
	return &PersonService{
		log:           log,
		personStorage: personStorage,
		personCache:   personCache,
		statCache:     statCache,
	}
}

var (
	ErrPersonExists   = errors.New("person already exists")
	ErrPersonNotFound = errors.New("person not found")
)

// Инвалидация кэша статистики (DelByPrefix для Redis)
func (p *PersonService) invalidateStatisticsCache(ctx context.Context) {
	_ = p.statCache.DelByPrefix(ctx, "stat:income:")
	_ = p.statCache.DelByPrefix(ctx, "stat:sold_subs:")
	_ = p.statCache.DelByPrefix(ctx, "stat:new_clients:")
	_ = p.statCache.DelByPrefix(ctx, "stat:monthly_stats:")
	_ = p.statCache.Delete(ctx, "stat:monthly_stats")
	_ = p.statCache.Delete(ctx, "stat:income")
	_ = p.statCache.Delete(ctx, "stat:total_clients")
	_ = p.statCache.Delete(ctx, "stat:sold_subs")
	_ = p.statCache.Delete(ctx, "stat:new_clients")
	_ = p.statCache.Delete(ctx, "stat:total_sold_subscriptions")
	_ = p.statCache.Delete(ctx, "stat:total_income")
}

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

	// Инвалидируем статистику
	p.invalidateStatisticsCache(ctx)

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

	// Инвалидируем статистику (если ФИО влияет на статистику новых клиентов)
	p.invalidateStatisticsCache(ctx)

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

	// Инвалидируем статистику (кол-во клиентов уменьшилось)
	p.invalidateStatisticsCache(ctx)

	log.Info("person deleted")
	return nil
}

func (p *PersonService) FindPersonByName(ctx context.Context, name string) ([]models.Person, error) {
	const op = "service.PersonService.FindPersonByName"
	log := p.log.With(slog.String("op", op))

	log.Info("Finding people by name", slog.String("name", name))

	cacheKey := fmt.Sprintf("person:name:%s", name)
	if cached, err := p.personCache.Get(ctx, cacheKey); err == nil {
		var people []models.Person
		if err := json.Unmarshal([]byte(cached), &people); err == nil {
			log.Info("cache hit", slog.String("key", cacheKey))
			return people, nil
		}
		log.Warn("failed to unmarshal cached data", sl.Error(err))
	}

	people, err := p.personStorage.FindPersonByName(ctx, name)
	if err != nil {
		log.Error("storage error", sl.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Не найдено — всегда возвращаем пустой слайс, не nil
	if people == nil {
		people = []models.Person{}
	}

	if data, err := json.Marshal(people); err == nil {
		if err := p.personCache.Set(ctx, cacheKey, data, 10*time.Minute); err != nil {
			log.Warn("failed to set cache", sl.Error(err))
		}
	}

	log.Info("people found", slog.Int("count", len(people)))
	return people, nil
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

func (p *PersonService) FindPersonById(ctx context.Context, id int) (models.Person, error) {
	const op = "services.person.FindPersonById"

	log := p.log.With(
		slog.String("op", op),
	)

	log.Info("Finding person by ID", slog.Int("id", id))

	cacheKey := fmt.Sprintf("person:id:%d", id)
	if cached, err := p.personCache.Get(ctx, cacheKey); err == nil {
		var person models.Person
		if err := json.Unmarshal([]byte(cached), &person); err == nil {
			log.Info("cache hit", slog.String("key", cacheKey))
			return person, nil
		}
		log.Warn("failed to unmarshal cached data", sl.Error(err))
	}

	person, err := p.personStorage.FindPersonById(ctx, id)
	if err != nil {
		log.Error("storage error", sl.Error(err))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	if data, err := json.Marshal(person); err == nil {
		if err := p.personCache.Set(ctx, cacheKey, data, 10*time.Minute); err != nil {
			log.Warn("failed to set cache", sl.Error(err))
		}
	}

	log.Info("person found")
	return person, nil
}
