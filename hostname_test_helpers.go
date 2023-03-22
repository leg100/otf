package otf

type FakeHostnameService struct {
	Host string
}

func (s FakeHostnameService) Hostname() string { return s.Host }
