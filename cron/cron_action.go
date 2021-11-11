package cron

import (
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/jasonlvhit/gocron"
)

type ICronAction interface {
	Start()
	Stop()
	IsRunning() bool
}

type cron struct {
	cron        *gocron.Scheduler
	cronActions *map[string]ICronAction
	queue       *goconcurrentqueue.FIFO
}

func NewCron(subCron *gocron.Scheduler, cronActions *map[string]ICronAction, queue *goconcurrentqueue.FIFO) ICronAction {
	return &cron{
		cron:        subCron,
		cronActions: cronActions,
		queue:       queue,
	}
}

func (s *cron) Start() {

	for _, v := range *s.cronActions {
		if !v.IsRunning() {
			v.Start()
		}
	}
}

func (s *cron) Stop() {
	s.cron.Clear()
}

func (s *cron) IsRunning() bool {
	return "startat" == (*gocron.Jobs()[0]).Tags()[0] && "endat" == (*gocron.Jobs()[1]).Tags()[0]
}
