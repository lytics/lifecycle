package lifecycle

import (
	// "github.com/ngmoco/timber"
	"sync/atomic"
	"testing"
	"time"
)

var loggerInited bool = false

func requireLoggerInited() {
	if loggerInited {
		return
	}

	// timber.AddLogger(timber.ConfigLogger{
	// 	LogWriter: new(timber.ConsoleWriter),
	// 	Level:     timber.DEBUG,
	// 	Formatter: timber.NewPatFormatter("[%D %T] [%L] %-10x %M"),
	// })
}

func TestShutdownRequest(t *testing.T) {
	requireLoggerInited()

	shutdownRequest := NewShutdownRequest()

	// Set up a channel that will stream 1's. This will be a workload for our pretend 
	// server below.
	onesChan := make(chan int)
	go func() {
		onesChan <- 1
	}()

	var sum int32 = 0
	lifeCycle := NewLifeCycle()

	// Set up a pretend "main loop" of a server and shut it down using ShutdownRequest
	go func() {
		lifeCycle.Transition(STATE_RUNNING)
		for !shutdownRequest.IsShutdownRequested() {
			select {
			case toAdd := <-onesChan:
				atomic.AddInt32(&sum, int32(toAdd))
			case <-shutdownRequest.GetShutdownRequestChan():
				break
			}
		}
		lifeCycle.Transition(STATE_STOPPED)
	}()

	// Give the server time to get some work done before we shut it down
	lifeCycle.WaitForState(STATE_RUNNING)
	time.Sleep(500 * time.Millisecond)

	// Tricky: after we request shutdown, then the server should execute its main loop at
	// *most* one more time. The shutdown flag should catch it before it loops again.
	// Therefore, we should expect sum to increase by at most 1 after we request shutdown.

	shutdownRequest.RequestShutdown()
	sumAfterShutdownRequest := atomic.LoadInt32(&sum)
	lifeCycle.WaitForState(STATE_STOPPED)
	sumAfterShutdownComplete := atomic.LoadInt32(&sum)

	reqsWhileShuttingDown := sumAfterShutdownComplete - sumAfterShutdownRequest
	if reqsWhileShuttingDown > 1 {
		t.Fail()
	}
}

func TestLifeCycle(t *testing.T) {
	requireLoggerInited()

	lifeCycle := NewLifeCycle()

	triggeredCh := make(chan interface{}, 1)

	// This goroutine waits for the lifeCycle to be RUNNING, then sends to channel
	go func() {
		state := lifeCycle.WaitForState(STATE_RUNNING)
		if state != STATE_RUNNING {
			t.Error("State should be 'running'")
		}
		triggeredCh <- true
	}()

	select {
	case <-triggeredCh:
		t.Error("Channel should not have been ready to read yet")
	default:
	}

	if lifeCycle.GetState() != STATE_NEW {
		t.Error("State should have been 'new'")
	}

	lifeCycle.Transition(STATE_RUNNING)
	if lifeCycle.GetState() != STATE_RUNNING {
		t.Error("State should have been 'running'")
	}
	select {
	case <-triggeredCh:
		break
	case <-time.After(1 * time.Second):
		t.Error("goroutine waiting for lifeCycle startup didn't get triggered")
	}

	// This goroutine waits for the lifeCycle to be STOPPED, then sends to channel
	triggeredCh = make(chan interface{}, 1)
	go func() {
		state := lifeCycle.WaitForState(STATE_STOPPED)
		if state != STATE_STOPPED {
			t.Error("State should have been 'stopped' but was %v", state)
		}
		triggeredCh <- true
	}()
	time.Sleep(100 * time.Millisecond)

	lifeCycle.Transition(STATE_STOPPED)
	if lifeCycle.GetState() != STATE_STOPPED {
		t.Error("State should have been 'stopped'")
	}

	select {
	case <-triggeredCh:
		break
	case <-time.After(1 * time.Second):
		t.Error("goroutine waiting for lifeCycle shutdown didn't get triggered")
	}

	time.Sleep(100 * time.Millisecond) // Let log lines have time to print
}
