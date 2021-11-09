package cron

type ICronAction interface {
	Start()
	Stop()
}
