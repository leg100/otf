package otf

import "context"

type fakeLatestRunSetter struct{}

func (f *fakeLatestRunSetter) SetLatestRun(context.Context, string, string) error { return nil }
