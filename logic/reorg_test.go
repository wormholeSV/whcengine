package engine

import (
	"testing"

	"github.com/copernet/whccommon/log"
)

func TestReorgWork(t *testing.T) {
	MaybeReorg(log.NewContext())
}
