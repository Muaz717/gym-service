package personSubService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/domain/dto"
	"github.com/Muaz717/gym_app/app/internal/domain/models"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	"github.com/Muaz717/gym_app/app/internal/services/cache"
	"github.com/Muaz717/gym_app/app/internal/storage"
	"log/slog"
	"time"
)

const (
	activeStatus  = "active"
	frozenStatus  = "frozen"
	expiredStatus = "expired"
	closedStatus  = "closed"
)

type PersonSubStorage interface {
	AddPersonSub(ctx context.Context, personSub models.PersonSubscription) (string, error)
	GetPersonSubByNumber(ctx context.Context, number string) (dto.PersonSubResponse, error)
	GetAllPersonSubs(ctx context.Context) ([]dto.PersonSubResponse, error)
	DeletePersonSub(ctx context.Context, number string) error
	FindPersonSubByPersonName(ctx context.Context, name string) ([]dto.PersonSubResponse, error)
	UpdatePersonSubStatus(ctx context.Context, number string, status string) error
	FindPersonSubByPersonId(ctx context.Context, personID int) ([]dto.PersonSubResponse, error)
}

type PersonFinder interface {
	FindPersonById(ctx context.Context, id int) (models.Person, error)
}

type PersonSubCache interface {
	cache.Cache
}

type StatCache interface {
	cache.Cache
}

type PersonSubService struct {
	log              *slog.Logger
	personSubStorage PersonSubStorage
	personSubCache   PersonSubCache
	personFinder     PersonFinder
	statCache        StatCache
}

func New(
	log *slog.Logger,
	personSubStorage PersonSubStorage,
	personSubCache PersonSubCache,
	personFinder PersonFinder,
	statCache StatCache,
) *PersonSubService {
	return &PersonSubService{
		log:              log,
		personSubStorage: personSubStorage,
		personSubCache:   personSubCache,
		personFinder:     personFinder,
		statCache:        statCache,
	}
}

var (
	ErrSubExists      = errors.New("subscription with that number already exists")
	ErrSubNotFound    = errors.New("subscription not found")
	ErrPersonNotFound = errors.New("person not found")
)

// Инвалидация статистического кэша с поддержкой DelByPrefix для Redis
func (p *PersonSubService) invalidateStatisticsCache(ctx context.Context) {
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

func (p *PersonSubService) AddPersonSub(ctx context.Context, input dto.PersonSubInput) (string, error) {
	const op = "services.personSub.AddPersonSub"

	log := p.log.With(
		slog.String("op", op),
	)

	log.Info("Adding new person subscription")

	// Преобразование строковых дат в time.Time с учетом локальной временной зоны
	layout := "2006-01-02"
	loc := time.Local

	var startDate, endDate time.Time
	var err error

	if input.StartDate != "" {
		startDate, err = time.ParseInLocation(layout, input.StartDate, loc)
		if err != nil {
			log.Error("failed to parse start_date", sl.Error(err))
			return "", fmt.Errorf("%s: invalid start_date: %w", op, err)
		}
	}

	if input.EndDate != "" {
		endDate, err = time.ParseInLocation(layout, input.EndDate, loc)
		if err != nil {
			log.Error("failed to parse end_date", sl.Error(err))
			return "", fmt.Errorf("%s: invalid end_date: %w", op, err)
		}
	}

	// Собираем структуру для сохранения в базу
	personSub := models.PersonSubscription{
		Number:            input.Number,
		PersonID:          input.PersonID,
		SubscriptionID:    input.SubscriptionID,
		SubscriptionPrice: input.SubscriptionPrice,
		StartDate:         startDate,
		EndDate:           endDate,
		Status:            input.Status,
		Discount:          input.Discount,
		FinalPrice:        input.FinalPrice,
	}

	personSubNumber, err := p.personSubStorage.AddPersonSub(ctx, personSub)
	if err != nil {
		if errors.Is(err, storage.ErrSubscriptionExists) {
			log.Warn("subscription already exists", slog.String("number", personSub.Number), sl.Error(err))
			return "", fmt.Errorf("%s: %w", op, ErrSubExists)
		} else if errors.Is(err, storage.ErrPersonNotFound) {
			log.Warn("person not found", slog.String("number", personSub.Number), sl.Error(err))
			return "", fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to add person subscription", sl.Error(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Инвалидируем кэш подписок
	if err := p.personSubCache.Delete(ctx, "person_subs:all"); err != nil {
		log.Warn("failed to invalidate cache", sl.Error(err))
	}

	// Инвалидируем кэш по имени пользователя
	person, err := p.personFinder.FindPersonById(ctx, personSub.PersonID)
	if err != nil {
		log.Warn("failed to get person name for cache invalidation", slog.Int("personID", personSub.PersonID), sl.Error(err))
	} else {
		cacheKey := fmt.Sprintf("person_sub:person:%s", person.Name)
		if err := p.personSubCache.Delete(ctx, cacheKey); err != nil {
			log.Warn("failed to invalidate cache", sl.Error(err))
		}
	}

	// Инвалидируем статистику!
	p.invalidateStatisticsCache(ctx)

	log.Info("person subscription added", "number", personSubNumber)

	return personSubNumber, nil
}

func (p *PersonSubService) DeletePersonSub(ctx context.Context, number string) error {
	const op = "services.personSub.DeletePersonSub"

	log := p.log.With(
		slog.String("op", op),
	)

	log.Info("Deleting person subscription")

	// Получаем подписку для получения PersonID перед удалением
	personSub, err := p.personSubStorage.GetPersonSubByNumber(ctx, number)
	if err != nil && !errors.Is(err, storage.ErrSubscriptionNotFound) {
		log.Error("failed to get person subscription before deletion", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	err = p.personSubStorage.DeletePersonSub(ctx, number)
	if err != nil {

		if errors.Is(err, storage.ErrSubscriptionNotFound) {
			log.Warn("subscription not found", slog.String("number", number), sl.Error(err))

			return fmt.Errorf("%s: %w", op, ErrSubNotFound)
		}

		log.Error("failed to delete person subscription", sl.Error(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	// Инвалидируем кэш подписок
	cacheKey := fmt.Sprintf("person_sub:number:%s", number)
	if err := p.personSubCache.Delete(ctx, cacheKey); err != nil {
		log.Warn("failed to invalidate cache", sl.Error(err))
	}
	if err := p.personSubCache.Delete(ctx, "person_subs:all"); err != nil {
		log.Warn("failed to invalidate cache", sl.Error(err))
	}
	// Инвалидируем кэш по имени пользователя, если PersonID существует
	if personSub.PersonID != 0 {
		person, err := p.personFinder.FindPersonById(ctx, personSub.PersonID)
		if err != nil {
			log.Warn("failed to get person name for cache invalidation", slog.Int("personID", personSub.PersonID), sl.Error(err))
		} else {
			personCacheKey := fmt.Sprintf("person_sub:person:%s", person.Name)
			if err := p.personSubCache.Delete(ctx, personCacheKey); err != nil {
				log.Warn("failed to invalidate cache", sl.Error(err))
			}
		}
	}

	// Инвалидируем статистику!
	p.invalidateStatisticsCache(ctx)

	log.Info("person subscription deleted", "number", number)

	return nil
}

func (p *PersonSubService) GetPersonSubByNumber(ctx context.Context, number string) (dto.PersonSubResponse, error) {
	const op = "services.personSub.FindPersonSubByNumber"

	log := p.log.With(
		slog.String("op", op),
	)

	log.Info("Getting person subscription by number")

	// Проверяем кэш
	cacheKey := fmt.Sprintf("person_sub:number:%s", number)
	if cached, err := p.personSubCache.Get(ctx, cacheKey); err == nil {
		var personSub dto.PersonSubResponse
		if err := json.Unmarshal([]byte(cached), &personSub); err == nil {
			log.Info("cache hit", slog.String("key", cacheKey))
			return personSub, nil
		}
		log.Warn("failed to unmarshal cached data", sl.Error(err))
	}

	personSub, err := p.personSubStorage.GetPersonSubByNumber(ctx, number)
	if err != nil {
		return dto.PersonSubResponse{}, err
	}

	// Сохраняем в кэш
	if data, err := json.Marshal(personSub); err == nil {
		if err := p.personSubCache.Set(ctx, cacheKey, data, 10*time.Minute); err != nil {
			log.Warn("failed to set cache", sl.Error(err))
		}
	}

	log.Info("person subscription found", "number", number)

	return personSub, nil
}

func (p *PersonSubService) GetAllPersonSubs(ctx context.Context) ([]dto.PersonSubResponse, error) {
	const op = "services.personSub.GetAllPersonSubs"

	log := p.log.With(
		slog.String("op", op),
	)

	log.Info("Getting all person subscriptions")

	// Проверяем кэш
	cacheKey := "person_subs:all"
	if cached, err := p.personSubCache.Get(ctx, cacheKey); err == nil {
		var personSubs []dto.PersonSubResponse
		if err := json.Unmarshal([]byte(cached), &personSubs); err == nil {
			log.Info("cache hit", slog.String("key", cacheKey))
			return personSubs, nil
		}
		log.Warn("failed to unmarshal cached data", sl.Error(err))
	}

	personSubs, err := p.personSubStorage.GetAllPersonSubs(ctx)
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш
	if data, err := json.Marshal(personSubs); err == nil {
		if err := p.personSubCache.Set(ctx, cacheKey, data, 30*time.Minute); err != nil {
			log.Warn("failed to set cache", sl.Error(err))
		}
	}

	log.Info("all person subscriptions found")

	return personSubs, nil
}

func (p *PersonSubService) FindPersonSubByPersonName(ctx context.Context, name string) ([]dto.PersonSubResponse, error) {
	const op = "services.personSub.FindPersonSubByPersonName"

	log := p.log.With(
		slog.String("op", op),
	)

	log.Info("Finding person subscription by person name")

	// Проверяем кэш
	cacheKey := fmt.Sprintf("person_sub:person:%s", name)
	if cached, err := p.personSubCache.Get(ctx, cacheKey); err == nil {
		var personSubs []dto.PersonSubResponse
		if err := json.Unmarshal([]byte(cached), &personSubs); err == nil {
			log.Info("cache hit", slog.String("key", cacheKey))
			return personSubs, nil
		}
		log.Warn("failed to unmarshal cached data", sl.Error(err))
	}

	personSubs, err := p.personSubStorage.FindPersonSubByPersonName(ctx, name)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Warn("person not found", slog.String("name", name), sl.Error(err))
			return nil, ErrPersonNotFound
		}

		if errors.Is(err, storage.ErrSubscriptionNotFound) {
			log.Warn("no subscriptions found for person", slog.String("name", name), sl.Error(err))
			return nil, ErrSubNotFound
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Сохраняем в кэш
	if data, err := json.Marshal(personSubs); err == nil {
		if err := p.personSubCache.Set(ctx, cacheKey, data, 10*time.Minute); err != nil {
			log.Warn("failed to set cache", sl.Error(err))
		}
	}

	log.Info("person subscriptions found")
	return personSubs, nil
}

func (p *PersonSubService) FindPersonSubByPersonId(ctx context.Context, personID int) ([]dto.PersonSubResponse, error) {
	const op = "services.personSub.FindPersonSubByPersonId"

	log := p.log.With(
		slog.String("op", op),
	)
	log.Info("Finding person subscription by person ID", slog.Int("personID", personID))

	// Проверяем кэш
	cacheKey := fmt.Sprintf("person_sub:person:%d", personID)
	if cached, err := p.personSubCache.Get(ctx, cacheKey); err == nil {
		var personSubs []dto.PersonSubResponse
		if err := json.Unmarshal([]byte(cached), &personSubs); err == nil {
			log.Info("cache hit", slog.String("key", cacheKey))
			return personSubs, nil
		}
		log.Warn("failed to unmarshal cached data", sl.Error(err))
	}

	// Получаем подписки по PersonID
	personSubs, err := p.personSubStorage.FindPersonSubByPersonId(ctx, personID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Warn("person not found", slog.Int("personID", personID), sl.Error(err))
			return nil, ErrPersonNotFound
		}
		if errors.Is(err, storage.ErrSubscriptionNotFound) {
			log.Warn("no subscriptions found for person", slog.Int("personID", personID), sl.Error(err))
			return nil, ErrSubNotFound
		}
		log.Error("failed to find person subscription by person ID", sl.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Сохраняем в кэш
	if data, err := json.Marshal(personSubs); err == nil {
		if err := p.personSubCache.Set(ctx, cacheKey, data, 10*time.Minute); err != nil {
			log.Warn("failed to set cache", sl.Error(err))
		}
	} else {
		log.Warn("failed to marshal person subscriptions for cache", sl.Error(err))
	}

	log.Info("person subscriptions found by person ID", slog.Int("personID", personID))
	return personSubs, nil
}

func (p *PersonSubService) UpdateStatuses(ctx context.Context) error {
	const op = "services.personSub.UpdateStatuses"

	log := p.log.With(
		slog.String("op", op),
	)

	log.Info("Updating person subscription statuses")

	subs, err := p.personSubStorage.GetAllPersonSubs(ctx)
	if err != nil {
		log.Error("failed to get all person subscriptions", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	today := time.Now().Truncate(24 * time.Hour)

	for _, sub := range subs {
		newStatus := ""

		if sub.StartDate.After(today) {
			newStatus = frozenStatus
		} else if sub.EndDate.Before(today) {
			newStatus = expiredStatus
		} else {
			newStatus = activeStatus
		}

		if sub.Status != newStatus {
			err := p.personSubStorage.UpdatePersonSubStatus(ctx, sub.Number, newStatus)
			if err != nil {
				log.Error("failed to update person subscription status", sl.Error(err))
				return fmt.Errorf("%s: %w", op, err)
			}

			// Инвалидируем кэш для подписки
			cacheKey := fmt.Sprintf("person_sub:number:%s", sub.Number)
			if err := p.personSubCache.Delete(ctx, cacheKey); err != nil {
				log.Warn("failed to invalidate cache", sl.Error(err))
			}

			// Инвалидируем кэш для пользователя
			person, err := p.personFinder.FindPersonById(ctx, sub.PersonID)
			if err != nil {
				log.Warn("failed to get person name for cache invalidation", slog.Int("personID", sub.PersonID), sl.Error(err))
			} else {
				personCacheKey := fmt.Sprintf("person_sub:person:%s", person.Name)
				if err := p.personSubCache.Delete(ctx, personCacheKey); err != nil {
					log.Warn("failed to invalidate cache", sl.Error(err))
				}
			}
		}
	}

	// Инвалидируем кэш всех подписок
	if err := p.personSubCache.Delete(ctx, "person_subs:all"); err != nil {
		log.Warn("failed to invalidate cache", sl.Error(err))
	}

	// И статистику!
	p.invalidateStatisticsCache(ctx)

	log.Info("person subscription statuses updated")
	return nil
}
