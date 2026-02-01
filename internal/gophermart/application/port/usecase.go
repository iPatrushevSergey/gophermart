package port

import "context"

// UseCase is the common contract for use cases (input in, output out).
// Each use case implements Execute; one interface instead of many per use case.
type UseCase[In, Out any] interface {
	Execute(ctx context.Context, in In) (Out, error)
}
