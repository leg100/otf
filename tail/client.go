package tail

import (
	"github.com/leg100/otf"
)

// client is a channel of logs for a specific run phase
type client struct {
	phase otf.PhaseSpec
	ch    chan []byte
}
