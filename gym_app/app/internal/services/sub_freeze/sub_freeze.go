package subFreezeService

import (
	"context"
	"encoding/json"
	"github.com/Muaz717/gym_app/app/internal/domain/models"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	"github.com/Muaz717/gym_app/app/internal/services/cache"
	"log/slog"
	"time"
)

type SubFreezeStorage interface {
	FreezeSubscription(ctx context.Context, subscriptionNumber string, freezeStart time.Time) error
	UnfreezeSubscription(ctx context.Context, subscriptionNumber string, unfreezeDate time.Time) error
	GetAllActiveFreeze(ctx context.Context) ([]models.SubscriptionFreeze, error)
}

type SubFreezeCache interface {
	cache.Cache
}

type SubFreezeService struct {
	log              *slog.Logger
	subFreezeStorage SubFreezeStorage
	subFreezeCache   SubFreezeCache
}

func New(
	log *slog.Logger,
	subFreezeStorage SubFreezeStorage,
	subFreezeCache SubFreezeCache,
) *SubFreezeService {
	return &SubFreezeService{
		log:              log,
		subFreezeStorage: subFreezeStorage,
		subFreezeCache:   subFreezeCache,
	}
}

func (s *SubFreezeService) FreezeSubscription(ctx context.Context, subscriptionNumber string, freezeStart time.Time) error {
	const op = "services.sub_freeze.FreezeSubscription"
	log := s.log.With(slog.String("op", op))

	log.Info("freezing subscription", slog.String("subscriptionNumber", subscriptionNumber), slog.Time("freezeStart", freezeStart))

	if err := s.subFreezeStorage.FreezeSubscription(ctx, subscriptionNumber, freezeStart); err != nil {
		log.Error("failed to freeze subscription", slog.String("subscriptionNumber", subscriptionNumber), sl.Error(err))
		return err
	}

	// Инвалидация кэша (не прерываем бизнес-логику, если кэш не удалился)
	cacheKey := "sub_freezed:all"
	if err := s.subFreezeCache.DelByPrefix(ctx, cacheKey); err != nil {
		log.Error("failed to invalidate cache", slog.String("cacheKey", cacheKey), sl.Error(err))
	}

	// Инвалидация кеша абонементов клиентов
	cacheKeySubs := "person_subs:all"
	if err := s.subFreezeCache.Delete(ctx, cacheKeySubs); err != nil {
		log.Error("failed to invalidate cache", slog.String("cacheKey", cacheKey), sl.Error(err))
	}

	log.Info("subscription frozen successfully")
	return nil
}

func (s *SubFreezeService) UnfreezeSubscription(ctx context.Context, subscriptionNumber string, unfreezeDate time.Time) error {
	const op = "services.sub_freeze.UnfreezeSubscription"
	log := s.log.With(slog.String("op", op))

	log.Info("unfreezing subscription", slog.String("subscriptionNumber", subscriptionNumber), slog.Time("unfreezeDate", unfreezeDate))

	if err := s.subFreezeStorage.UnfreezeSubscription(ctx, subscriptionNumber, unfreezeDate); err != nil {
		log.Error("failed to unfreeze subscription", slog.String("subscriptionNumber", subscriptionNumber), sl.Error(err))
		return err
	}

	// Инвалидация кэша (не прерываем бизнес-логику)
	cacheKey := "sub_freezed:all"
	if err := s.subFreezeCache.DelByPrefix(ctx, cacheKey); err != nil {
		log.Error("failed to invalidate cache", slog.String("cacheKey", cacheKey), sl.Error(err))
	}

	// Инвалидация кеша абонементов клиентов
	cacheKeySubs := "person_subs:all"
	if err := s.subFreezeCache.Delete(ctx, cacheKeySubs); err != nil {
		log.Error("failed to invalidate cache", slog.String("cacheKey", cacheKey), sl.Error(err))
	}

	log.Info("subscription unfrozen successfully")
	return nil
}

func (s *SubFreezeService) GetAllActiveFreeze(ctx context.Context) ([]models.SubscriptionFreeze, error) {
	const op = "services.sub_freeze.GetAllActiveFreeze"
	log := s.log.With(slog.String("op", op))

	log.Info("getting all active freezed subscriptions")

	cacheKey := "sub_freezed:all"
	if cached, err := s.subFreezeCache.Get(ctx, cacheKey); err == nil {
		var freezes []models.SubscriptionFreeze
		if err := json.Unmarshal([]byte(cached), &freezes); err == nil {
			log.Info("cache hit", slog.String("cacheKey", cacheKey))
			return freezes, nil
		}
		log.Warn("failed to unmarshal cached data", sl.Error(err))
	} else {
		log.Info("cache miss", slog.String("cacheKey", cacheKey))
	}

	freezes, err := s.subFreezeStorage.GetAllActiveFreeze(ctx)
	if err != nil {
		log.Error("failed to get active freezes", sl.Error(err))
		return nil, err
	}

	if data, err := json.Marshal(freezes); err == nil {
		_ = s.subFreezeCache.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return freezes, nil
}
