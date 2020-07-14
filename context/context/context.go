// Package context defines the Context type, which carries deadlines,
// cancellation signals, and other request-scoped values across API boundaries
// and between processes.

// Incoming requests to a server should create a Context, and outgoing
// calls to servers should accept a Context. The chain of function
// calls between them must propagate the Context, optionally replacing
// it with a derived Context created using WithCancel, WithDeadline,
// WithTimeout, or WithValue. When a Context is canceled, all
// Contexts derived from it are also canceled.

// The WithCancel, WithDeadline, and WithTimeout functions take a
// Context (the parent) and return a derived Context (the child) and a
// CancelFunc. Calling the CancelFunc cancels the child and its
// children, removes the parent's reference to the child, and stops
// any associated timers. Failing to call the CancelFunc leaks the
// child and its children until the parent is canceled or the timer
// fires. The go vet tool checks that CancelFuncs are used on all
// control-flow paths.

// Programs that use Contexts should follow these rules to keep interfaces
// consistent across packages and enable static analysis tools to check context
// propagation.

// Do not store Contexts inside a struct type; instead, pass a Context
// explicitly to each function that needs it. The Context should be the first
// parameter, typically named ctx:

//f unc Dosomething(ctx context.Context, arg Arg) {
//     // ... using ctx ...
// }

// Do not pass a nil context, even if a function permits it. Pass context.TODO
// if you are unsure about which Context to use.

// Use context Values only for request-scoped data that transits processed and
// APIs, not for passing optional parameters to functions.

// The same Context may be passed to functions running in different goroutines;
// Contexts are same for simultaneous use by multiple goroutines.

package context

import (
    "./internal/reflectlite"
    "errors"
    "sync"
    "sync/atomic"
    "time"
)

// A Context carries a deadline, a cancellation signal, and other values across
// API boundaries.
//
// Context's methods may be called by multiple goroutines simultaneously.
type Context interface {
    // Deadline returns the time when work done on behalf of this context
    // should be canceled. Deadline returns ok==false when no deadline is set.
    // Successive calls to Deadline return the same results.
    Deadline() (deadline time.Time, ok bool)

    // Done returns a channel that's closed when work done on behalf of this
    // context should be canceled. Done may return nil if this context can
    // never be canceled. Successive calls to Done return the same value.
    // The close of this Done channel may happen asynchronously,
    // after the cancel function returns.
    //
    // WithCancel arranges for Done to be closed when cancel is called;
    // WithDeadline arranges for Done to be closed when the deadline
    // expires; WithTimeout arranges for Done to be closed when the timeout
    // elapses.
    //
    // Done is provided for use in select statements:
    // a Done channel for cancellation
    Done() <- chan struct {}

    // If Done is not yet closed, Err returns nil.
    // if Done is closed, Err returns a non-nil error explaining why;
    // Canceled if the the context canceled
    // or DeadlineExceeded if the context's deadline passed.
    // After Err returns a non-nil error, successive calls to Err return the same error
    Err() error

    Value(key interface{}) interface{}
}

// Canceled is the error returned by Context.Err when the context is canceled.
var Canceled = errors.New("context canceled")

// DeadlineExceed is the error returned by Context.Err when the context's
// deadline passes
var DeadlineExceed error = deadlineExceededError{}

type deadlineExceededError struct {}

func (deadlineExceededError) Error() string { return "context deadline exceeded" }
func (deadlineExceededError) Timeout() bool { return true }
func (deadlineExceededError) Temporary() bool { return true }

// An emptyCtx is never canceled, has no values, and has no deadline. It is not
// struct{}, since vars of this type must have distinct addresses.
type emptyCtx int

func (*emptyCtx) Deadline() (deadline time.Time, ok bool) {
    return
}

func (*emptyCtx) Done() <- chan struct{} {
    return nil
}

func (*emptyCtx) Err() error {
    return nil
}

func (*emptyCtx) Value(key interface{}) interface{} {
    return nil
}

func (e *emptyCtx) String() string {
    switch e {
    case background:
        return "context.Background"
    case todo:
        return "context.TODO"
    }
    return "unknown empty Context"
}

// Background returns a non-nil, empty Context. It is never canceled, has no
// values, and has no deadline. It is typically used by the main function,
// initialization, and tests, and as the top-level Context for incoming
// requests.
var (
    background = new(emptyCtx)
    todo = new(emptyCtx)
)

func Background() Context {
    return background
}

func TODO() Context {
    return todo
}


// A CancelFunc tells an operation to abandon its work.
// A CancelFunc does not wait for the work to stop.
// A CancelFunc may be called by multiple goroutines simultaneously.
// After the first call, subsequent calls to a CancelFunc do nothing.
type CancelFunc func()

// WithCancel returns a copy of parent with a new Done channel. The returned
// context's Done channel is closed when the returned cancel function is called
// or when the parent context's Done channel is closed, whichever happens first.
//
// Canceling this context releases resources associated with it, so code should
// call cancel as soon as the operations running in this Context complete.

func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
    c := newCancelCtx(parent)
    propagateCancel(parent, &c)
    return &c, func(){ c.cancel(true, Canceled) }
}

func newCancelCtx(parent Context) cancelCtx {
    return cancelCtx{Context: parent}
}

// goroutines counts the number of goroutines ever created; for testing
var goroutines int32

// propagateCancel arranges for child to be canceled when parent is.
func propagateCancel(parent Context, child canceler) {
    done := parent.Done()
    if done == nil {
        return // parent is never canceled
    }

    select {
    case <-done:
        // parent is already canceled
        child.cancel(false, parent.Err())
        return
    default:
    }

    if p, ok := parentCancelCtx(parent); ok {
        p.mu.Lock()
        if p.err != nil {
            child.cancel(false, p.err)
        } else {
            if p.children == nil {
                p.children = make(map[canceler]struct{})
            }
            p.children[child] = struct{}{}
        }
        p.mu.Unlock()
    } else {
        atomic.AddInt32(&goroutines, +1)
        go func() {
            select {
            case <-parent.Done():
                child.cancel(false, parent.Err())
            case <-child.Done():
            }
        }()
    }
}



// parentCancelCtx returns the underlying *contextCtx for parent.
// It does this by looking up parent.Value(&cancelContextKey) to find
// the innermost enclosing *cancelCtx and then checking whether
// parent.Done() matches that *cancelCtx. (If not, the *cancelCtx
// has been wrapped in a custom implementation providing a
// different done channel, in which case wh should not bypass it.
func parentCancelCtx(parent Context) (*cancelCtx, bool) {
    done := parent.Done()
    if done == closedchan || done == nil {
        return nil, false
    }
    p, ok := parent.Value(&cancelCtxKey).(*cancelCtx)
    if !ok {
        return nil, false
    }

    p.mu.Lock()
    ok = p.done == done
    p.mu.Unlock()
    if !ok {
        return nil, false
    }
    return p, true
}

// removeChild removes a context from its parent.
func removeChild(parent Context, child canceler) {
    p, ok := parentCancelCtx(parent)
    if !ok {
        return
    }
    p.mu.Lock()
    if p.children != nil {
        delete(p.children, child)
    }
    p.mu.Unlock()
}

type canceler interface {
    cancel(removeFromParent bool, err error)
    Done() <-chan struct{}
}

// closedchan is a reusable closed channel
var closedchan = make(chan struct{})

func init() {
    close(closedchan)
}

type cancelCtx struct {
    Context

    mu       sync.Mutex
    done     chan struct{}
    children map[canceler]struct{}
    err      error
}

func (c *cancelCtx) Value(key interface{}) interface{} {
    if key == cancelCtxKey {
        return c
    }
    return c.Context.Value(key)
}

func (c *cancelCtx) Done() <-chan struct{} {
    c.mu.Lock()
    if c.done == nil {
        c.done = make(chan struct{})
    }
    d := c.done
    c.mu.Unlock()
    return d
}

func (c *cancelCtx) Err() error {
    c.mu.Lock()
    err := c.err
    c.mu.Unlock()
    return err
}

type stringer interface {
    String() string
}

func contextName(c Context) string {
    if s, ok :=c.(stringer); ok {
        return s.String()
    }
    return reflectlite.TypeOf(c).String()
}

func (c *cancelCtx) String() string {
    return contextName(c.Context) + ".WithCancel"
}

// cancel closes c.done, cancels each of c's children, and, if
// removeFromParent is true, removes c from its parent's children.
func (c *cancelCtx) cancel(removeFromParent bool, err error) {
    if err == nil {
        panic("context: internal error: missing cancel error")
    }
    c.mu.Lock()
    if c.err != nil {
        c.mu.Unlock()
        return  // already canceled
    }

    c.err = err
    if c.done == nil {
        c.done = closedchan
    } else {
        close(c.done)
    }
    for child := range c.children {
        child.cancel(false, err)
    }
    c.children = nil
    c.mu.Unlock()

    if removeFromParent {
        removeChild(c.Context, c)
    }

}

var cancelCtxKey int

