package pubsub

import (
	"context"
)

type (
	fakePool struct {
		pool
	}
	fakeGetter struct {
		fake *fakeType
	}
	fakeType struct {
		ID    string `json:"id"`
		Stuff []byte `json:"stuff"`
	}
)

func (f *fakeGetter) GetByID(context.Context, string, DBAction) (any, error) {
	return f.fake, nil
}
