package integration

type hostnameService struct {
	hostname string
}

func (s hostnameService) Hostname() string { return s.hostname }
