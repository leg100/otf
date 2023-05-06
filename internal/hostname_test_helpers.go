package internal

type FakeHostnameService struct {
	Host string

	HostnameService
}

func (s FakeHostnameService) Hostname() string { return s.Host }
