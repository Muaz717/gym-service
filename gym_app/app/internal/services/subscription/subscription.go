package subscriptionService

import (
	"context"
	"errors"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/domain/models"
	"github.com/Muaz717/gym_app/app/internal/lib/logger/sl"
	"github.com/Muaz717/gym_app/app/internal/storage"

	"log/slog"
)

type SubscriptionService struct {
	log                 *slog.Logger
	subscriptionStorage SubscriptionStorage
}

type SubscriptionStorage interface {
	SaveSubscription(ctx context.Context, subscription models.Subscription) (int, error)
	FindAllSubscriptions(ctx context.Context) ([]models.Subscription, error)
	UpdateSubscription(ctx context.Context, subscription models.Subscription, subID int) (int, error)
	DeleteSubscription(ctx context.Context, subID int) error
}

var (
	ErrSubExists   = errors.New("subscription with that number already exists")
	ErrSubNotFound = errors.New("subscription not found")
)

func New(
	log *slog.Logger,
	subscriptionStorage SubscriptionStorage,
) *SubscriptionService {
	return &SubscriptionService{
		log:                 log,
		subscriptionStorage: subscriptionStorage,
	}
}

func (m *SubscriptionService) AddSubscription(ctx context.Context, subscription models.Subscription) (int, error) {
	const op = "services.subscription.AddSubscription"

	log := m.log.With(
		slog.String("op", op),
	)

	log.Info("Adding new membership")

	subId, err := m.subscriptionStorage.SaveSubscription(ctx, subscription)
	if err != nil {

		if errors.Is(err, storage.ErrSubscriptionExists) {
			log.Warn("subscription already exists", sl.Error(err))
			return 0, fmt.Errorf("%s: %w", op, ErrSubExists)
		}
		return 0, fmt.Errorf("%s: %w", op, sl.Error(err))
	}

	log.Info("subscription registered", "mid", subId)

	return subId, nil
}

func (m *SubscriptionService) UpdateSubscription(ctx context.Context, subscription models.Subscription, subID int) (int, error) {
	const op = "services.subscription.UpdateSubscription"

	log := m.log.With(
		slog.String("op", op),
	)

	log.Info("Updating subscription")

	subId, err := m.subscriptionStorage.UpdateSubscription(ctx, subscription, subID)
	if err != nil {
		if errors.Is(err, storage.ErrSubscriptionNotFound) {
			log.Warn("subscription not found", sl.Error(err))
			return 0, fmt.Errorf("%s: %w", op, ErrSubNotFound)
		}

		return 0, fmt.Errorf("%s: %w", op, sl.Error(err))
	}

	log.Info("subscription updated", "mid", subId)

	return subId, nil
}

func (m *SubscriptionService) DeleteSubscription(ctx context.Context, subID int) error {
	const op = "services.subscription.DeleteSubscription"

	log := m.log.With(
		slog.String("op", op),
	)

	log.Info("Deleting subscription")

	err := m.subscriptionStorage.DeleteSubscription(ctx, subID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, sl.Error(err))
	}

	log.Info("subscription deleted")

	return nil
}

func (m *SubscriptionService) FindAllSubscriptions(ctx context.Context) ([]models.Subscription, error) {
	const op = "services.subscription.FindAllSubscriptions"

	log := m.log.With(
		slog.String("op", op),
	)

	subscriptions, err := m.subscriptionStorage.FindAllSubscriptions(ctx)
	if err != nil {
		log.Warn("error", sl.Error(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Subscriptions are found")

	return subscriptions, nil
}
