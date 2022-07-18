package otf

import "context"

type FakeLatestRunSetter struct{}

func (f *FakeLatestRunSetter) SetLatestRun(context.Context, string, string) error { return nil }
