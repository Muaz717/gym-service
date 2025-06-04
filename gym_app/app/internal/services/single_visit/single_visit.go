package singleVisitService

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/domain/dto"
	"github.com/Muaz717/gym_app/app/internal/domain/models"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	"github.com/Muaz717/gym_app/app/internal/services/cache"
	"log/slog"
	"time"
)

type SingleVisitStorage interface {
	AddSingleVisit(ctx context.Context, singleVis models.SingleVisit) error
	GetAllSingleVisits(ctx context.Context) ([]models.SingleVisit, error)
	GetSingleVisitById(ctx context.Context, id int) (models.SingleVisit, error)
	GetSingleVisitsByDay(ctx context.Context, date time.Time) ([]models.SingleVisit, error)
	GetSingleVisitsByPeriod(ctx context.Context, from, to time.Time) ([]models.SingleVisit, error)
	DeleteSingleVisit(ctx context.Context, id int) error
}

type SingleVisitCache interface {
	cache.Cache
}

type SingleVisitService struct {
	log                *slog.Logger
	singleVisitStorage SingleVisitStorage
	singleVisitCache   SingleVisitCache
}

func New(
	log *slog.Logger,
	singleVisitStorage SingleVisitStorage,
	singleVisitCache SingleVisitCache,
) *SingleVisitService {
	return &SingleVisitService{
		log:                log,
		singleVisitStorage: singleVisitStorage,
		singleVisitCache:   singleVisitCache,
	}
}

func (p *SingleVisitService) invalidateStatisticsCache(ctx context.Context) {
	_ = p.singleVisitCache.DelByPrefix(ctx, "stat:income:")
	_ = p.singleVisitCache.DelByPrefix(ctx, "stat:sold_subs:")
	_ = p.singleVisitCache.DelByPrefix(ctx, "stat:new_clients:")
	_ = p.singleVisitCache.DelByPrefix(ctx, "stat:monthly_stats:")
	_ = p.singleVisitCache.Delete(ctx, "stat:monthly_stats")
	_ = p.singleVisitCache.Delete(ctx, "stat:income")
	_ = p.singleVisitCache.Delete(ctx, "stat:total_clients")
	_ = p.singleVisitCache.Delete(ctx, "stat:sold_subs")
	_ = p.singleVisitCache.Delete(ctx, "stat:new_clients")
	_ = p.singleVisitCache.Delete(ctx, "stat:total_sold_subscriptions")
	_ = p.singleVisitCache.Delete(ctx, "stat:total_income")
}

func (s *SingleVisitService) AddSingleVisit(ctx context.Context, singleVisStrDate dto.SingleVisitInput) error {
	const op = "services.single_visit.AddSingleVisit"
	log := s.log.With(slog.String("op", op))

	log.Info("adding single visit", slog.Any("singleVisit", singleVisStrDate))

	layout := "2006-01-02"
	loc := time.Local

	if singleVisStrDate.VisitDate == "" {
		log.Error("visit date is empty")
		return fmt.Errorf("visit date is required")
	}

	// Парсим дату: сначала как YYYY-MM-DD, потом как RFC3339
	visitDate, err := time.ParseInLocation(layout, singleVisStrDate.VisitDate, loc)
	if err != nil {
		visitDate, err = time.Parse(time.RFC3339, singleVisStrDate.VisitDate)
		if err != nil {
			log.Error("failed to parse visit date", slog.String("visitDate", singleVisStrDate.VisitDate), slog.Any("error", err))
			return fmt.Errorf("invalid visit date format: %w", err)
		}
	}

	singleVisit := models.SingleVisit{
		VisitDate:  visitDate,
		FinalPrice: singleVisStrDate.FinalPrice,
	}

	if err := s.singleVisitStorage.AddSingleVisit(ctx, singleVisit); err != nil {
		log.Error("failed to add single visit", slog.Any("error", err))
		return err
	}

	s.invalidateStatisticsCache(ctx)

	cachePrefix := "single_visits:"
	if err := s.singleVisitCache.DelByPrefix(ctx, cachePrefix); err != nil {
		log.Error("failed to invalidate cache", slog.String("cacheKey", cachePrefix), slog.Any("error", err))
	}

	return nil
}

func (s *SingleVisitService) GetAllSingleVisits(ctx context.Context) ([]models.SingleVisit, error) {
	const op = "services.single_visit.GetAllSingleVisits"
	log := s.log.With(slog.String("op", op))

	cacheKey := "single_visits:all"
	if cached, err := s.singleVisitCache.Get(ctx, cacheKey); err == nil && len(cached) > 0 {
		var singleVisits []models.SingleVisit
		if err := json.Unmarshal([]byte(cached), &singleVisits); err == nil {
			log.Info("cache hit for all single visits", slog.String("cacheKey", cacheKey))
			return singleVisits, nil
		}
		log.Error("failed to unmarshal cached single visits", slog.Any("error", err), slog.String("cacheKey", cacheKey))
	}

	log.Info("cache miss for all single visits, fetching from storage")
	singleVisits, err := s.singleVisitStorage.GetAllSingleVisits(ctx)
	if err != nil {
		log.Error("failed to get all single visits", slog.Any("error", err))
		return nil, err
	}

	// Сохраняем в кэш
	if data, err := json.Marshal(singleVisits); err == nil {
		if err := s.singleVisitCache.Set(ctx, cacheKey, data, 30*time.Minute); err != nil {
			log.Warn("failed to set cache", sl.Error(err))
		}
	}

	return singleVisits, nil
}

func (s *SingleVisitService) GetSingleVisitById(ctx context.Context, id int) (models.SingleVisit, error) {
	const op = "services.single_visit.GetSingleVisitById"
	log := s.log.With(slog.String("op", op))

	cacheKey := fmt.Sprintf("single_visit:%d", id)
	if cached, err := s.singleVisitCache.Get(ctx, cacheKey); err == nil && len(cached) > 0 {
		var singleVisit models.SingleVisit
		if err := json.Unmarshal([]byte(cached), &singleVisit); err == nil {
			log.Info("cache hit for single visit by ID", slog.String("cacheKey", cacheKey))
			return singleVisit, nil
		}
		log.Error("failed to unmarshal cached single visit by ID", slog.Any("error", err), slog.String("cacheKey", cacheKey))
	}

	log.Info("cache miss for single visit by ID, fetching from storage")
	singleVisit, err := s.singleVisitStorage.GetSingleVisitById(ctx, id)
	if err != nil {
		log.Error("failed to get single visit by ID", slog.Int("id", id), slog.Any("error", err))
		return models.SingleVisit{}, err
	}

	if data, err := json.Marshal(singleVisit); err == nil {
		if err := s.singleVisitCache.Set(ctx, cacheKey, data, 30*time.Minute); err != nil {
			log.Warn("failed to set cache for single visit by ID", sl.Error(err))
		}
	}

	return singleVisit, nil
}

// Теперь GetSingleVisitsByDay принимает dateStr string, парсит дату внутри и вызывает storage с time.Time
func (s *SingleVisitService) GetSingleVisitsByDay(ctx context.Context, dateStr string) ([]models.SingleVisit, error) {
	const op = "services.single_visit.GetSingleVisitsByDay"
	log := s.log.With(slog.String("op", op))

	layout := "2006-01-02"
	date, err := time.Parse(layout, dateStr)
	if err != nil {
		date, err = time.Parse(time.RFC3339, dateStr)
		if err != nil {
			log.Error("failed to parse date for GetSingleVisitsByDay", slog.String("date", dateStr), slog.Any("error", err))
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
	}

	cacheKey := fmt.Sprintf("single_visits:day:%s", date.Format(layout))
	if cached, err := s.singleVisitCache.Get(ctx, cacheKey); err == nil && len(cached) > 0 {
		var singleVisits []models.SingleVisit
		if err := json.Unmarshal([]byte(cached), &singleVisits); err == nil {
			log.Info("cache hit for single visits by day", slog.String("cacheKey", cacheKey))
			return singleVisits, nil
		}
		log.Error("failed to unmarshal cached single visits by day", slog.Any("error", err), slog.String("cacheKey", cacheKey))
	}

	log.Info("cache miss for single visits by day, fetching from storage")
	singleVisits, err := s.singleVisitStorage.GetSingleVisitsByDay(ctx, date)
	if err != nil {
		log.Error("failed to get single visits by day", slog.Any("error", err))
		return nil, err
	}

	if data, err := json.Marshal(singleVisits); err == nil {
		if err := s.singleVisitCache.Set(ctx, cacheKey, data, 30*time.Minute); err != nil {
			log.Warn("failed to set cache for single visits by day", sl.Error(err))
		}
	}

	return singleVisits, nil
}

// GetSingleVisitsByPeriod теперь принимает fromStr, toStr string, парсит даты внутри
func (s *SingleVisitService) GetSingleVisitsByPeriod(ctx context.Context, fromStr, toStr string) ([]models.SingleVisit, error) {
	const op = "services.single_visit.GetSingleVisitsByPeriod"
	log := s.log.With(slog.String("op", op))

	layout := "2006-01-02T15:04:05Z"
	layoutShort := "2006-01-02"

	from, err := time.Parse(layout, fromStr)
	if err != nil {
		from, err = time.Parse(layoutShort, fromStr)
		if err != nil {
			log.Error("failed to parse 'from' date", slog.String("from", fromStr), slog.Any("error", err))
			return nil, fmt.Errorf("invalid 'from' date format: %w", err)
		}
	}

	to, err := time.Parse(layout, toStr)
	if err != nil {
		to, err = time.Parse(layoutShort, toStr)
		if err != nil {
			log.Error("failed to parse 'to' date", slog.String("to", toStr), slog.Any("error", err))
			return nil, fmt.Errorf("invalid 'to' date format: %w", err)
		}
	}

	if from.After(to) {
		return nil, fmt.Errorf("invalid period: 'from' date is after 'to' date")
	}

	cacheKey := fmt.Sprintf("single_visits:period:%s:%s", from.Format(layoutShort), to.Format(layoutShort))
	if cached, err := s.singleVisitCache.Get(ctx, cacheKey); err == nil && len(cached) > 0 {
		var singleVisits []models.SingleVisit
		if err := json.Unmarshal([]byte(cached), &singleVisits); err == nil {
			log.Info("cache hit for single visits by period", slog.String("cacheKey", cacheKey))
			return singleVisits, nil
		}
		log.Error("failed to unmarshal cached single visits by period", slog.Any("error", err), slog.String("cacheKey", cacheKey))
	}

	log.Info("cache miss for single visits by period, fetching from storage")
	singleVisits, err := s.singleVisitStorage.GetSingleVisitsByPeriod(ctx, from, to)
	if err != nil {
		log.Error("failed to get single visits by period", slog.Any("error", err))
		return nil, err
	}

	if data, err := json.Marshal(singleVisits); err == nil {
		if err := s.singleVisitCache.Set(ctx, cacheKey, data, 30*time.Minute); err != nil {
			log.Warn("failed to set cache for single visits by period", sl.Error(err))
		}
	}

	return singleVisits, nil
}

func (s *SingleVisitService) DeleteSingleVisit(ctx context.Context, id int) error {
	const op = "services.single_visit.DeleteSingleVisit"
	log := s.log.With(slog.String("op", op))

	log.Info("deleting single visit", slog.Int("id", id))

	if err := s.singleVisitStorage.DeleteSingleVisit(ctx, id); err != nil {
		log.Error("failed to delete single visit", slog.Int("id", id), slog.Any("error", err))
		return err
	}

	s.invalidateStatisticsCache(ctx)

	cachePrefix := "single_visits:"
	if err := s.singleVisitCache.DelByPrefix(ctx, cachePrefix); err != nil {
		log.Error("failed to invalidate cache after delete", slog.String("cacheKey", cachePrefix), slog.Any("error", err))
	}

	return nil
}
