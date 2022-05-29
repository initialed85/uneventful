package workers

import "fmt"

type LazyWorker struct {
	*Lifecycle
}

func NewLazyWorker(
	name string,
	setupFunc LifecycleFunc,
	teardownFunc LifecycleFunc,
) Worker {
	w := LazyWorker{
		NewLifecycle(
			fmt.Sprintf("LazyWorker_%v", name),
			setupFunc,
			nil,
			nil,
			nil,
			teardownFunc,
		),
	}

	return &w
}
