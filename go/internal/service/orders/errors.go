package orders

import (
	"github.com/cgund98/go-eventsrc-example/internal/infra/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrInternal() error {
	return status.Errorf(codes.Internal, "Internal error")
}

// Checks if the error is a non-grpc error and logs it.
// Returns an internal server error if the error is not a non-grpc error.
func WrapNonGrpcError[T any](payload T, err error) (T, error) {
	if err == nil {
		return payload, nil
	}

	if status.Code(err) == codes.Unknown {
		logging.Logger.Error("unexpected error when handling request", "error", err)
		return payload, ErrInternal()
	}

	return payload, err
}
