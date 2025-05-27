package forgejo

import "github.com/spf13/pflag"

var hostname string

func init() {
	pflag.StringVar(&hostname, "forgejo-hostname", "next.forgejo.org", "forgejo hostname")
}
