package grpc

import (
	"errors"
	"net/http"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToGRPCError mapea errores de dominio/app a errores gRPC gradables.
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := status.FromError(err); ok {
		return err
	}

	var appErr *validator.AppError
	if errors.As(err, &appErr) {
		code := codeFromHTTPStatus(appErr.Status)
		return status.Error(code, appErr.Error())
	}

	var parseErr *validator.ParseError
	if errors.As(err, &parseErr) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	code := grpcCodeFromDomain(err)
	if code == codes.Unknown {
		return status.Error(codes.Internal, err.Error())
	}

	return status.Error(code, err.Error())
}

func grpcCodeFromDomain(err error) codes.Code {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return codes.NotFound
	case errors.Is(err, domain.ErrUserDuplicated):
		return codes.AlreadyExists
	case errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrInvalidName):
		return codes.InvalidArgument
	default:
		return codes.Unknown
	}
}

func codeFromHTTPStatus(statusCode int) codes.Code {
	switch statusCode {
	case http.StatusOK:
		return codes.OK
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusServiceUnavailable:
		return codes.Unavailable
	case http.StatusInternalServerError:
		return codes.Internal
	default:
		return codes.Unknown
	}
}
