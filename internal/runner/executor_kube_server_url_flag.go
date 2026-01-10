package runner

import (
	"github.com/spf13/pflag"
)

var _ pflag.Value = (*KubeServerURLFlag)(nil)

type KubeServerURLFlag struct {
	url string
}

func (k KubeServerURLFlag) Type() string { return "kubernetes-job-url" }

func (k KubeServerURLFlag) String() string { return k.url }

func (e *KubeServerURLFlag) Set(v string) error {
	e.url = v
	return nil
}
