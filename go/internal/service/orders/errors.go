package orders

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrInternal() error {
	return status.Errorf(codes.Internal, "Internal error")
}
