package forgejo

import "github.com/leg100/otf/internal/vcs"

type Service struct{}

func (s *Service) NewClient() (vcs.Client, error) {
}
