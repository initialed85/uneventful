package workers

import "fmt"

type BlockedWorker struct {
	*Lifecycle
}

func NewBlockedWorker(
	name string,
	setupFunc LifecycleFunc,
	workFunc LifecycleFunc,
	workErrorHandlerFunc ErrorHandlerFunc,
	teardownFunc LifecycleFunc,
) Worker {
	w := BlockedWorker{
		NewLifecycle(
			fmt.Sprintf("BlockedWorker_%v", name),
			setupFunc,
			nil,
			workFunc,
			workErrorHandlerFunc,
			teardownFunc,
		),
	}

	return &w
}
