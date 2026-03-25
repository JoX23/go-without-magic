package grpc

import (
	"context"
	"errors"
	"net"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/grpc/pb"
	"github.com/JoX23/go-without-magic/internal/repository/memory"
	"github.com/JoX23/go-without-magic/internal/service"
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

func TestGRPCService_Integration(t *testing.T) {
	repo := memory.NewUserRepository()
	svc := service.NewUserService(repo, zap.NewNop())

	grpcServer := grpcpkg.NewServer(grpcpkg.UnaryInterceptor(UnaryServerInterceptor(zap.NewNop())))
	pb.RegisterUserServiceServer(grpcServer, NewUserServiceServer(svc, zap.NewNop()))

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	go func() { _ = grpcServer.Serve(lis) }()
	defer grpcServer.Stop()

	conn, err := grpcpkg.NewClient(lis.Addr().String(), grpcpkg.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial grpc server: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	createResp, err := client.CreateUser(context.Background(), &pb.CreateUserRequest{Email: "x@example.com", Name: "X"})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if createResp.User.Email != "x@example.com" {
		t.Fatalf("unexpected email: %s", createResp.User.Email)
	}

	getResp, err := client.GetUser(context.Background(), &pb.GetUserRequest{Id: createResp.User.Id})
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if getResp.User.Name != "X" {
		t.Fatalf("unexpected name: %s", getResp.User.Name)
	}

	listResp, err := client.ListUsers(context.Background(), &pb.ListUsersRequest{})
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}

	if len(listResp.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(listResp.Users))
	}
}
