lifecycle
=========

Lifecycle is a go package that helps you manage goroutines that provide services to other goroutines.

Things you can do:
 - Have one goroutine wait for another goroutine to initalize itself and be ready to work. The server goroutine calls ```go LifeCycle.Transition(STATE_RUNNING)``` and the client calls ```LifeCycle.WaitForState(STATE_RUNNING)```
 - Wait for a goroutine service to finish shutting down. This is handy for testing graceful shutdown, and it's friendlier than using WaitGroups manually.

Separately, a ShutdownRequest is a good way to quickly shut down the main loop of a server without an unbounded of iterations that can occur with the "quit channel" approach:

```go
for !shutdownRequest.IsShutdownRequested() {
    select {
        case toAdd := <-onesChan:
            atomic.AddInt32(&sum, int32(toAdd))
        case <-shutdownRequest.GetShutdownRequestChan():
            break
        }
}
```
