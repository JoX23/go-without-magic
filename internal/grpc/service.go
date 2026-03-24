package grpc

import (
	"context"

	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/grpc/pb"
	"github.com/JoX23/go-without-magic/internal/service"
)

// UserServiceServerImpl implementa pb.UserServiceServer
// usando la capa de servicio de dominio existente.
type UserServiceServerImpl struct {
	svc    *service.UserService
	logger *zap.Logger
}

func NewUserServiceServer(svc *service.UserService, logger *zap.Logger) *UserServiceServerImpl {
	return &UserServiceServerImpl{svc: svc, logger: logger}
}

func (s *UserServiceServerImpl) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user, err := s.svc.CreateUser(ctx, req.Email, req.Name)
	if err != nil {
		return nil, ToGRPCError(err)
	}

	return &pb.CreateUserResponse{User: asPBUser(user)}, nil
}

func (s *UserServiceServerImpl) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := s.svc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, ToGRPCError(err)
	}

	return &pb.GetUserResponse{User: asPBUser(user)}, nil
}

func (s *UserServiceServerImpl) ListUsers(ctx context.Context, _ *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, err := s.svc.ListUsers(ctx)
	if err != nil {
		return nil, ToGRPCError(err)
	}

	out := make([]*pb.User, 0, len(users))
	for _, u := range users {
		out = append(out, asPBUser(u))
	}

	return &pb.ListUsersResponse{Users: out}, nil
}

func asPBUser(u *domain.User) *pb.User {
	if u == nil {
		return nil
	}
	return &pb.User{
		Id:        u.ID.String(),
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
