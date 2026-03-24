//go:build ignore
// +build ignore

// NOTE: Este archivo requiere que internal/grpc/proto/product.proto
// haya sido compilado con protoc primero. Elimina las líneas build ignore
// una vez que el pb package esté disponible.

package grpc

import (
	"context"

	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/grpc/pb"
	"github.com/JoX23/go-without-magic/internal/service"
)

// ProductServiceServerImpl implementa pb.ProductServiceServer
// usando la capa de servicio de dominio existente.
type ProductServiceServerImpl struct {
	svc    *service.ProductService
	logger *zap.Logger
}

func NewProductServiceServer(svc *service.ProductService, logger *zap.Logger) *ProductServiceServerImpl {
	return &ProductServiceServerImpl{svc: svc, logger: logger}
}

func (s *ProductServiceServerImpl) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	e, err := s.svc.CreateProduct(ctx, req.Sku, req.Name, req.Price)
	if err != nil {
		return nil, ToGRPCError(err)
	}
	return &pb.CreateProductResponse{
		Product: asProtoProduct(e),
	}, nil
}

func (s *ProductServiceServerImpl) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	e, err := s.svc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, ToGRPCError(err)
	}
	return &pb.GetProductResponse{
		Product: asProtoProduct(e),
	}, nil
}

func (s *ProductServiceServerImpl) ListProducts(ctx context.Context, _ *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	items, err := s.svc.ListProducts(ctx)
	if err != nil {
		return nil, ToGRPCError(err)
	}

	out := make([]*pb.Product, 0, len(items))
	for _, e := range items {
		out = append(out, asProtoProduct(e))
	}

	return &pb.ListProductsResponse{
		Products: out,
	}, nil
}

func asProtoProduct(e *domain.Product) *pb.Product {
	if e == nil {
		return nil
	}
	return &pb.Product{
		Id:          e.ID.String(),
		Sku:         e.Sku,
		Name:        e.Name,
		Price:       e.Price,
		Description: e.Description,
		Status:      string(e.Status),
		CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
