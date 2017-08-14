package watcher

import (
	"testing"
	"time"

	"runtime"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	Run()

	pipe := make(chan jobDoneChannel, 1)
	keyChain.attachChannel <- pipeline{
		key:  CompositeKey{1, "test"},
		pipe: pipe,
	}

	ch := <-pipe
	require.NotNil(t, ch)
	expected := make(jobDoneChannel)
	assert.IsType(t, expected, ch)

	Halt()
}

func TestStart(t *testing.T) {
	Run()

	testChannel := make(chan int, 1)
	val := 1

	isStart := Start(func() {
		testChannel <- val
		val++
	}, time.Second, CompositeKey{1, "test"})

	assert.True(t, isStart)
	receive := <-testChannel
	assert.Equal(t, 1, receive)
	receive = <-testChannel
	assert.Equal(t, 2, receive)

	_, ok := keyChain.storage[CompositeKey{1, "test"}]
	assert.True(t, ok)

	Halt()
}

func TestStartOnlyOneTaskWithSameKey(t *testing.T) {
	Run()

	val := 1
	startFunc := func() bool {
		return Start(func() {
			val++
		}, time.Second, CompositeKey{1, "test"})
	}

	routines := runtime.NumGoroutine()

	isStart := startFunc()
	assert.True(t, isStart)

	for i := 1; i <= 10; i++ {
		if isStart = startFunc(); isStart {
			break
		}
	}
	assert.False(t, isStart)

	routinesAfterStart := runtime.NumGoroutine()
	assert.Equal(t, routines+1, routinesAfterStart)

	assert.Equal(t, 1, len(keyChain.storage))

	Halt()
}

func TestStop(t *testing.T) {
	Run()

	val := 1
	ch := make(chan int)
	routines := runtime.NumGoroutine()

	job := func(val int) func() {
		return func() {
			val++
			ch <- val
		}
	}

	Start(job(val), time.Second, CompositeKey{1, "test"})

	receive := <-ch
	assert.True(t, receive > 1)

	Stop(CompositeKey{1, "test"})

	time.Sleep(time.Second * 2)
	routinesAfterStop := runtime.NumGoroutine()
	assert.Equal(t, routines, routinesAfterStop)

	_, ok := keyChain.storage[CompositeKey{1, "test"}]
	assert.False(t, ok)

	Halt()
}

func TestStopAfterAlreadyStop(t *testing.T) {
	Run()

	val := 1
	Start(func() {
		val++
	}, time.Second, CompositeKey{1, "test"})
	time.Sleep(time.Second * 1)

	stopFunc := func() bool {
		return Stop(CompositeKey{1, "test"})
	}

	isStop := stopFunc()
	assert.True(t, isStop)

	for i := 1; i <= 10; i++ {
		if isStop = stopFunc(); isStop {
			break
		}
	}
	assert.False(t, isStop)

	Halt()
}

func TestStartStopMultipleWorker(t *testing.T) {
	Run()

	val := 0
	num := uint(0)
	ch := make(chan int)

	job := func(val int) func() {
		return func() {
			val++
			ch <- val
		}
	}

	startFunc := func() bool {
		num++
		return Start(job(val), time.Second, CompositeKey{num, "test"})
	}

	routines := runtime.NumGoroutine()
	isStart := false
	for i := 1; i <= 10; i++ {
		if isStart = startFunc(); !isStart {
			break
		}
	}
	assert.True(t, isStart)

	routinesAfterStart := runtime.NumGoroutine()
	assert.Equal(t, routines+10, routinesAfterStart)

	receiver := 0
	for i := 1; i <= 10; i++ {
		receiver += <-ch
	}

	assert.Equal(t, 10, receiver)
	assert.Equal(t, 10, len(keyChain.storage))

	stopFunc := func(num uint) bool {
		return Stop(CompositeKey{num, "test"})
	}

	isStop := false
	for i := 1; i <= 10; i++ {
		if isStop = stopFunc(uint(i)); !isStart {
			break
		}
	}
	assert.True(t, isStop)

	time.Sleep(time.Second * 4)

	routinesAfterStop := runtime.NumGoroutine()
	assert.Equal(t, routines, routinesAfterStop)

	assert.Equal(t, 0, len(keyChain.storage))

	Halt()
}
