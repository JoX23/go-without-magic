package grpc

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/validator"
	grpcpkg "google.golang.org/grpc"
)

func TestToGRPCError_DomainMapping(t *testing.T) {
	err := ToGRPCError(domain.ErrUserNotFound)
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %s", status.Code(err))
	}

	err = ToGRPCError(domain.ErrUserDuplicated)
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists, got %s", status.Code(err))
	}

	err = ToGRPCError(domain.ErrInvalidEmail)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %s", status.Code(err))
	}
}

func TestToGRPCError_ValidatorMapping(t *testing.T) {
	parseErr := &validator.ParseError{Cause: errors.New("invalid json")}
	err := ToGRPCError(parseErr)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %s", status.Code(err))
	}

	valErrs := validator.ValidationErrors{{Field: "email", Tag: "required", Msg: "required"}}
	err = ToGRPCError(valErrs)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %s", status.Code(err))
	}
}

func TestUnaryServerInterceptor(t *testing.T) {
	interceptor := UnaryServerInterceptor(zap.NewNop())

	info := &grpcpkg.UnaryServerInfo{FullMethod: "/user.UserService/GetUser"}
	_, err := interceptor(context.Background(), nil, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, domain.ErrUserNotFound
	})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %s", status.Code(err))
	}

	_, err = interceptor(context.Background(), nil, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("boom")
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal on panic, got %s", status.Code(err))
	}
}
