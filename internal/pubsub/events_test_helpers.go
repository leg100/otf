package pubsub

type FakePublisher struct{}

func (f *FakePublisher) Publish(Event) {}
