package watcher

import (
	"sync"
	"time"
)

type jobDoneChannel chan struct{}

type storage map[CompositeKey]jobDoneChannel

func (s storage) attach(key CompositeKey) (jobDoneChannel, bool) {
	channel, ok := s[key]
	if !ok {
		channel = make(jobDoneChannel, 1)
		s[key] = channel
	}

	return channel, !ok
}

func (s storage) detach(key CompositeKey) (jobDoneChannel, bool) {
	channel, ok := s[key]
	if ok {
		delete(s, key)
	}

	return channel, ok
}

type pipeline struct {
	key  CompositeKey
	pipe chan jobDoneChannel
}

type hub struct {
	storage       storage
	attachChannel chan pipeline
	detachChannel chan pipeline
	haltChannel   chan struct{}
	done          chan bool
	wg            sync.WaitGroup
}

func (h *hub) attach(key CompositeKey) (channel jobDoneChannel, ok bool) {
	pipe := make(chan jobDoneChannel, 1)
	pipeline := pipeline{
		key:  key,
		pipe: pipe,
	}

	h.attachChannel <- pipeline
	channel, ok = <-pipe

	return
}

func (h *hub) detach(key CompositeKey) (channel jobDoneChannel, ok bool) {
	pipe := make(chan jobDoneChannel)
	pipeline := pipeline{
		key:  key,
		pipe: pipe,
	}

	h.detachChannel <- pipeline
	channel, ok = <-pipe

	return
}

func (h *hub) start(job JobFunc, duration time.Duration, key CompositeKey) (ok bool) {
	doneChannel, ok := h.attach(key)
	if !ok {
		return
	}

	h.wg.Add(1)
	go func() {
		ticker := time.NewTicker(duration)
		defer func() {
			ticker.Stop()
			h.wg.Done()
		}()

		for {
			select {
			case <-doneChannel:
				return

			case <-ticker.C:
				job()
			}
		}
	}()

	return
}

func (h *hub) stop(key CompositeKey) (ok bool) {
	doneChannel, ok := h.detach(key)
	if ok {
		close(doneChannel)
	}

	return
}

func (h *hub) run() {
	defer func() {
		h.wg.Wait()
		close(h.done)
	}()

	for {
		select {
		case pipeline := <-h.attachChannel:
			if channel, ok := h.storage.attach(pipeline.key); ok {
				pipeline.pipe <- channel
			} else {
				close(pipeline.pipe)
			}

		case pipeline := <-h.detachChannel:
			if channel, ok := h.storage.detach(pipeline.key); ok {
				pipeline.pipe <- channel
			} else {
				close(pipeline.pipe)
			}
		case <-h.haltChannel:
			for _, channel := range h.storage {
				close(channel)
			}

			return
		}
	}
}
