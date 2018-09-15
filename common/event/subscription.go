package event

//accounts -- accounts -- backend interface
// Subscription represents a stream of events. The carrier of the events is typically a
// channel, but isn't part of the interface.
type Subscription interface {
	Err() <-chan error // returns the error channel
	Unsubscribe()      // cancels sending of events, closing the error channel
}
