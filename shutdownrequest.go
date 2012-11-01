package lifecycle

import (
	"sync/atomic"
)

/*
ShutdownRequest solves the problem of quickly and cleanly shutting down a goroutine. When a 
goroutine is reading incoming work items from a channel and executing them, using the quit 
channel approach may not immediately shut down the worker:

	keepLooping := true
	for keepLooping {
		select {
    		workItem := <-workChan:
		    	doWork(workItem)
	    	<-quitChan:
		    	keepLooping = false
		}
	}

In the above example, the worker may continue processing workChan for many iterations until it
happens to read from quitChan. Using ShutdownRequest you can guarantee termination within one
iteration:

	for !shutdownRequest.IsShutdownRequested() {
		select {
		case toAdd := <-onesChan:
			atomic.AddInt32(&sum, int32(toAdd))
		case <-shutdownRequest.GetShutdownRequestChan():
			break
		}
	}
*/
type ShutdownRequest struct {
	shutdownRequestFlag int32
	shutdownRequestChan chan interface{}
}

func NewShutdownRequest() *ShutdownRequest {
	return &ShutdownRequest{
		shutdownRequestChan: make(chan interface{}, 1),
	}
}

// Call this in the server main loop to check whether it should shut down. Use this together
// with the shutdown request channel.
func (this *ShutdownRequest) IsShutdownRequested() bool {
	shouldShutdown := atomic.LoadInt32(&this.shutdownRequestFlag) > 0
	return shouldShutdown
}

// Get the channel that will receive an object when shutdown is requested. Use the returned
// channel in the server's main loop select statement. 
// 
// There is only one such channel, so don't have multiple goroutines receiving from the channel.
//
// Use this together with IsShutdownRequested().
func (this *ShutdownRequest) GetShutdownRequestChan() chan interface{} {
	return this.shutdownRequestChan
}

// Call this 
func (this *ShutdownRequest) RequestShutdown() {
	// The CompareAndSwap check ensures that we only write to the "shutdown requested" channel
	// the first time that shutdown is requested. This prevents channel senders from blocking
	// when this method is called multiple times.
	didSwap := atomic.CompareAndSwapInt32(&this.shutdownRequestFlag, 0, 1)
	if didSwap {
		this.shutdownRequestChan <- true
	}
}
