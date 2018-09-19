package event

//accounts--manager--Manager struct
// Feed implements one-to-many subscriptions where the carrier of events is a channel.
// Values sent to a Feed are delivered to all subscribed channels simultaneously.
type Feed struct {
}

// To be finished.
func (f *Feed) Send(value interface{}) (nsent int) {

	return nsent
}

func (f *Feed) Subscribe(channel interface{}) Subscription {

	return nil
}
