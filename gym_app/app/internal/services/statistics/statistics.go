package statistics

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	"github.com/Muaz717/gym_app/app/internal/services/cache"
	"log/slog"
	"time"
)

type StatStorage interface {
	TotalClients(ctx context.Context) (int, error)
	NewClients(ctx context.Context, from, to time.Time) (int, error)
	TotalIncome(ctx context.Context) (float64, error)
	Income(ctx context.Context, from, to time.Time) (float64, error)
	SoldSubscriptions(ctx context.Context, from, to time.Time) (int, error)
	TotalSoldSubscriptions(ctx context.Context) (int, error)
}

type StatCache interface {
	cache.Cache
}

type StatService struct {
	log         *slog.Logger
	statStorage StatStorage
	statCache   StatCache
}

func New(
	log *slog.Logger,
	statStorage StatStorage,
	statCache StatCache,
) *StatService {
	return &StatService{
		log:         log,
		statStorage: statStorage,
		statCache:   statCache,
	}
}

func (s *StatService) TotalClients(ctx context.Context) (int, error) {
	const op = "services.statistics.totalClients"
	log := s.log.With(slog.String("op", op))

	cacheKey := "stat:total_clients"
	if cached, err := s.statCache.Get(ctx, cacheKey); err == nil {
		var total int
		if err := json.Unmarshal([]byte(cached), &total); err == nil {
			log.Info("cache hit", slog.String("key", cacheKey))
			return total, nil
		}
		log.Warn("failed to unmarshal cached data", sl.Error(err))
	}

	total, err := s.statStorage.TotalClients(ctx)
	if err != nil {
		log.Error("failed to get total clients", sl.Error(err))
		return 0, err
	}

	if data, err := json.Marshal(total); err == nil {
		_ = s.statCache.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return total, nil
}

func (s *StatService) NewClients(ctx context.Context, from, to time.Time) (int, error) {
	const op = "services.statistics.newClients"
	log := s.log.With(slog.String("op", op))

	cacheKey := fmt.Sprintf("stat:new_clients:%s:%s", from.Format("2006-01-02"), to.Format("2006-01-02"))
	if cached, err := s.statCache.Get(ctx, cacheKey); err == nil {
		var newClients int
		if err := json.Unmarshal([]byte(cached), &newClients); err == nil {
			log.Info("cache hit", slog.String("key", cacheKey))
			return newClients, nil
		}
		log.Warn("failed to unmarshal cached data", sl.Error(err))
	}

	newClients, err := s.statStorage.NewClients(ctx, from, to)
	if err != nil {
		log.Error("failed to get new clients", sl.Error(err))
		return 0, err
	}

	if data, err := json.Marshal(newClients); err == nil {
		_ = s.statCache.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return newClients, nil
}

func (s *StatService) TotalIncome(ctx context.Context) (float64, error) {
	const op = "services.statistics.totalIncome"
	log := s.log.With(slog.String("op", op))

	cacheKey := "stat:total_income"
	if cached, err := s.statCache.Get(ctx, cacheKey); err == nil {
		var totalIncome float64
		if err := json.Unmarshal([]byte(cached), &totalIncome); err == nil {
			log.Info("cache hit", slog.String("key", cacheKey))
			return totalIncome, nil
		}
		log.Warn("failed to unmarshal cached data", sl.Error(err))
	}

	totalIncome, err := s.statStorage.TotalIncome(ctx)
	if err != nil {
		log.Error("failed to get total income", sl.Error(err))
		return 0, err
	}

	if data, err := json.Marshal(totalIncome); err == nil {
		_ = s.statCache.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return totalIncome, nil
}

func (s *StatService) Income(ctx context.Context, from, to time.Time) (float64, error) {
	const op = "services.statistics.income"
	log := s.log.With(slog.String("op", op))

	cacheKey := fmt.Sprintf("stat:income:%s:%s", from.Format("2006-01-02"), to.Format("2006-01-02"))
	if cached, err := s.statCache.Get(ctx, cacheKey); err == nil {
		var income float64
		if err := json.Unmarshal([]byte(cached), &income); err == nil {
			log.Info("cache hit", slog.String("key", cacheKey))
			return income, nil
		}
		log.Warn("failed to unmarshal cached data", sl.Error(err))
	}

	income, err := s.statStorage.Income(ctx, from, to)
	if err != nil {
		log.Error("failed to get income", sl.Error(err))
		return 0, err
	}

	if data, err := json.Marshal(income); err == nil {
		_ = s.statCache.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return income, nil
}

func (s *StatService) TotalSoldSubscriptions(ctx context.Context) (int, error) {
	const op = "services.statistics.totalSoldSubscriptions"
	log := s.log.With(slog.String("op", op))

	cacheKey := "stat:total_sold_subscriptions"
	if cached, err := s.statCache.Get(ctx, cacheKey); err == nil {
		var totalSold int
		if err := json.Unmarshal([]byte(cached), &totalSold); err == nil {
			log.Info("cache hit", slog.String("key", cacheKey))
			return totalSold, nil
		}
		log.Warn("failed to unmarshal cached data", sl.Error(err))
	}

	totalSold, err := s.statStorage.TotalSoldSubscriptions(ctx)
	if err != nil {
		log.Error("failed to get total sold subscriptions", sl.Error(err))
		return 0, err
	}

	if data, err := json.Marshal(totalSold); err == nil {
		_ = s.statCache.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return totalSold, nil
}

func (s *StatService) SoldSubscriptions(ctx context.Context, from, to time.Time) (int, error) {
	const op = "services.statistics.soldSubscriptions"
	log := s.log.With(slog.String("op", op))

	cacheKey := fmt.Sprintf("stat:sold_subs:%s:%s", from.Format("2006-01-02"), to.Format("2006-01-02"))
	if cached, err := s.statCache.Get(ctx, cacheKey); err == nil {
		var soldSubs int
		if err := json.Unmarshal([]byte(cached), &soldSubs); err == nil {
			log.Info("cache hit", slog.String("key", cacheKey))
			return soldSubs, nil
		}
		log.Warn("failed to unmarshal cached data", sl.Error(err))
	}

	soldSubs, err := s.statStorage.SoldSubscriptions(ctx, from, to)
	if err != nil {
		log.Error("failed to get sold subscriptions", sl.Error(err))
		return 0, err
	}

	if data, err := json.Marshal(soldSubs); err == nil {
		_ = s.statCache.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return soldSubs, nil
}
