package tunnel

import (
	"github.com/lesismal/arpc/log"
)

var (
	tunMethod = "tunnel"
)

func init() {
	log.SetLogger(&logger{})
}
