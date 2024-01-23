package action

import "context"

type Action interface {
	Do(ctx context.Context) error
}
