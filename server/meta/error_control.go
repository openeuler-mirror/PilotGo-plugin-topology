package meta

import (
	"context"
)

type Topoerror struct {
	Err    error
	Cancel context.CancelFunc
}
