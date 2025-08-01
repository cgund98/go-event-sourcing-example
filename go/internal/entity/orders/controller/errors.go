package controller

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrOrderNotFound = status.Errorf(codes.NotFound, "order not found")
var ErrInternal = status.Errorf(codes.Internal, "internal server error")
