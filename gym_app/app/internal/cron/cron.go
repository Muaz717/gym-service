package cron

import (
	"context"
	personSubService "github.com/Muaz717/gym_app/app/internal/services/person_sub"
	"github.com/robfig/cron/v3"
)

type CronJobs struct {
	cronScheduler    *cron.Cron
	personSubService *personSubService.PersonSubService
}

func New(personSubService *personSubService.PersonSubService) *CronJobs {
	return &CronJobs{
		cronScheduler:    cron.New(),
		personSubService: personSubService,
	}
}

func (c *CronJobs) Start(ctx context.Context) {

	c.cronScheduler.AddFunc("@daily", func() {
		err := c.personSubService.UpdateStatuses(ctx)
		if err != nil {
			panic(err)
		}
	})

	c.cronScheduler.Start()
}

func (c *CronJobs) Stop() {
	c.cronScheduler.Stop()
}
