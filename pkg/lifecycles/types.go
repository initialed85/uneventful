package lifecycles

type Worker interface {
	GetName() string
	IsStarted() bool
	IsStopped() bool
	Start() error
	Stop() error
	Healthz() error
}

type LifecycleFunc func() error

type ErrorHandlerFunc func(error)
