package otf

type FakePublisher struct{}

func (f *FakePublisher) Publish(Event) {}
