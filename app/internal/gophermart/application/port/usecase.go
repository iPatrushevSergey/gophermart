package port

import "context"

// UseCase is the common contract for request-response use cases.
type UseCase[In, Out any] interface {
	Execute(ctx context.Context, in In) (Out, error)
}

// BackgroundRunner is the contract for background task use cases.
type BackgroundRunner interface {
	Run(ctx context.Context) (int, error)
}
