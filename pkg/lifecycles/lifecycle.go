package lifecycles

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Lifecycle struct {
	mu                        sync.RWMutex
	started, stopped          bool
	stopRequest, stopResponse chan bool
	name                      string
	startFunc                 LifecycleFunc
	workTriggerAndStopSignal  chan bool
	workFunc                  LifecycleFunc
	workErrorHandlerFunc      ErrorHandlerFunc
	stopFunc                  LifecycleFunc
}

func NewLifecycle(name string, startFunc LifecycleFunc, workTriggerAndStopSignal chan bool, // true = work, false = stop
	workFunc LifecycleFunc, workErrorHandlerFunc ErrorHandlerFunc, stopFunc LifecycleFunc) *Lifecycle {
	l := Lifecycle{started: false, stopped: true, stopRequest: make(chan bool), stopResponse: make(chan bool), name: name, startFunc: startFunc, workFunc: workFunc, workErrorHandlerFunc: workErrorHandlerFunc, stopFunc: stopFunc, workTriggerAndStopSignal: workTriggerAndStopSignal}

	return &l
}

func (l *Lifecycle) run() {
	log.Printf("%v - run() entered", l.name)
	defer log.Printf("%v - run() exited", l.name)

	var err error
	var before, after time.Time

exit:
	for {
		if l.workTriggerAndStopSignal != nil { // this path blocks on either of these
			select {
			case <-l.stopRequest:
				break exit
			case workOrStop := <-l.workTriggerAndStopSignal:
				if !workOrStop {
					break exit
				}
			}
		} else { // this path checks but doesn't block
			select {
			case <-l.stopRequest:
				break exit
			default:
			}
		}

		before = time.Now()
		err = l.workFunc()
		if err != nil {
			log.Printf("%v - warning: workFunc=%#+v returned error: %v", l.name, l.workFunc, err)

			if l.workErrorHandlerFunc != nil {
				l.workErrorHandlerFunc(err)
			}
		}
		after = time.Now()

		log.Printf("%v - {workFunc: %#+v}=%v", l.name, l.workFunc, after.Sub(before))
	}

	l.stopResponse <- true
}

func (l *Lifecycle) GetName() string {
	return l.name
}

func (l *Lifecycle) IsStarted() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.started
}

func (l *Lifecycle) IsStopped() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.stopped
}

func (l *Lifecycle) Start() (err error) {
	log.Printf("%v - Start() entered", l.name)
	defer log.Printf("%v - Start() exited", l.name)

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.started {
		return fmt.Errorf("cannot start, already started")
	}

	if l.startFunc != nil {
		err = l.startFunc()
		if err != nil {
			return err
		}
	}

	l.started = true
	l.stopped = false

	if l.workFunc != nil {
		go l.run()
	}

	return nil
}

func (l *Lifecycle) Stop() (err error) {
	log.Printf("%v - Stop() entered", l.name)
	defer log.Printf("%v - Stop() exited", l.name)

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.stopped {
		return fmt.Errorf("cannot stop, already stopped")
	}

	if l.workFunc != nil {
		l.stopRequest <- true
	}

	if l.stopFunc != nil {
		err = l.stopFunc()
		if err != nil {
			return err
		}
	}

	if l.workFunc != nil {
		<-l.stopResponse
	}

	l.started = false
	l.stopped = true

	return nil
}

func (l *Lifecycle) Healthz() (err error) {
	if !l.IsStarted() {
		return fmt.Errorf("%v - IsStarted() returned false to Healthz()", l.name)
	}

	return nil
}
