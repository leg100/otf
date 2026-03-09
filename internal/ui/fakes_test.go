package ui

type fakeHostnamesClient struct {
	HostnameClient
	hostname string
}

func (f *fakeHostnamesClient) URL(path string) string        { return path }
func (f *fakeHostnamesClient) WebhookURL(path string) string { return path }
func (f *fakeHostnamesClient) Hostname() string              { return f.hostname }
