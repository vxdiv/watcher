// Package watcher provides the ability to run background tasks that are performed
// at the regular intervals. All tasks are marked by the specific composite key
// in order to enable the process stopping. You must run the watcher before tasks
// starting or stopping. Only single task can be run with the same composite key.
package watcher

import (
	"time"
)

type (
	CompositeKey struct {
		Index uint
		Name  string
	}

	JobFunc func()
)

var keyChain *hub

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

func Halt() {
	close(keyChain.haltChannel)
	<-keyChain.done
	keyChain = nil
}

func Start(job JobFunc, duration time.Duration, key CompositeKey) bool {
	return keyChain.start(job, duration, key)
}

func Stop(key CompositeKey) bool {
	return keyChain.stop(key)
}
