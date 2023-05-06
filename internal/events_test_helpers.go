package internal

type FakePublisher struct{}

func (f *FakePublisher) Publish(Event) {}
