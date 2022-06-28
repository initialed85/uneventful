package lifecycles

import (
	"fmt"
	"time"
)

type ScheduledWorker struct {
	ticker                   *time.Ticker
	workTriggerAndStopSignal chan bool
	innerLifeCycle           *Lifecycle
	*Lifecycle
	setupFunc    LifecycleFunc
	teardownFunc LifecycleFunc
	duration     time.Duration
}

func NewScheduledWorker(name string, setupFunc LifecycleFunc, workFunc LifecycleFunc, workErrorHandlerFunc ErrorHandlerFunc, teardownFunc LifecycleFunc, duration time.Duration) Worker {
	s := ScheduledWorker{workTriggerAndStopSignal: make(chan bool), setupFunc: setupFunc, teardownFunc: teardownFunc, duration: duration}

	name = fmt.Sprintf("ScheduledWorker_%.3f_%v", duration.Seconds(), name)

	s.innerLifeCycle = NewLifecycle(name, s.innerStart, nil, s.innerWork, nil, s.innerStop)

	s.Lifecycle = NewLifecycle(name, s.wrappedStart, s.workTriggerAndStopSignal, workFunc, workErrorHandlerFunc, s.wrappedStop)

	return &s
}

func (s *ScheduledWorker) innerStart() error {
	s.ticker = time.NewTicker(s.duration)
	return nil
}

func (s *ScheduledWorker) wrappedStart() (err error) {
	err = s.innerLifeCycle.Start()
	if err != nil {
		return err
	}

	if s.setupFunc == nil {
		return nil
	}

	return s.setupFunc()
}

func (s *ScheduledWorker) innerWork() error {
	select {
	case s.workTriggerAndStopSignal <- true:
	default:
	}

	<-s.ticker.C

	return nil
}

func (s *ScheduledWorker) innerStop() error {
	s.ticker.Stop()
	select {
	case s.workTriggerAndStopSignal <- false:
	default:
	}
	return nil
}

func (s *ScheduledWorker) wrappedStop() (err error) {
	err = s.innerLifeCycle.Stop()
	if err != nil {
		return err
	}

	if s.teardownFunc == nil {
		return nil
	}

	return s.teardownFunc()
}
