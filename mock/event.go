package mock

import (
	"github.com/leg100/ots"
)

type EventService struct {
	PublishFn   func(event ots.Event)
	SubscribeFn func(id string) ots.Subscription
}

func (s EventService) Publish(event ots.Event) {
	s.PublishFn(event)
}

func (s EventService) Subscribe(id string) ots.Subscription {
	return s.SubscribeFn(id)
}
