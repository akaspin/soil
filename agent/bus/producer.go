package bus

import "context"

type Producer interface {
	Subscribe(ctx context.Context, consumer Consumer)
}
