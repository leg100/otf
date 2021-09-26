package mock

import (
	"github.com/leg100/otf"
)

type EventService struct {
	PublishFn   func(event otf.Event)
	SubscribeFn func(id string) otf.Subscription
}

func (s EventService) Publish(event otf.Event) {
	s.PublishFn(event)
}

func (s EventService) Subscribe(id string) (otf.Subscription, error) {
	return s.SubscribeFn(id), nil
}
