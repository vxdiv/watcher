// Package watcher provides the ability to run background tasks that are performed
// at the regular intervals. All tasks are marked by the specific composite key
// in order to enable the process stopping. You must run the watcher before tasks
// starting or stopping. Only single task can be run with the same composite key.
package watcher

import (
	"time"
)

type (
	// CompositeKey is used to mark tasks
	CompositeKey struct {
		Index uint
		Name  string
	}

	//JobFunc there is a task that performs an action in the background
	JobFunc func()
)

var keyChain *hub

// Run starts watcher in background
func Run() {
	keyChain = &hub{
		storage:       make(storage),
		attachChannel: make(chan pipeline, 1),
		detachChannel: make(chan pipeline, 1),
		haltChannel:   make(chan struct{}),
		done:          make(chan bool),
	}
	go keyChain.run()
}

// Halt stops the watcher
func Halt() {
	close(keyChain.haltChannel)
	<-keyChain.done
	keyChain = nil
}

// Start registers a task that will be performed at the regular intervals
func Start(job JobFunc, duration time.Duration, key CompositeKey) bool {
	return keyChain.start(job, duration, key)
}

// Stop stops task by its CompositeKey
func Stop(key CompositeKey) bool {
	return keyChain.stop(key)
}
