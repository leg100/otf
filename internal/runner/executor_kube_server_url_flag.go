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

func (k *KubeServerURLFlag) Set(v string) error {
	k.url = v
	return nil
}
