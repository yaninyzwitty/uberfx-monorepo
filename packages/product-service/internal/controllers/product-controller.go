package controllers

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	productsv1 "github.com/yaninyzwitty/go-fx-v1/gen/products/v1"
	"github.com/yaninyzwitty/go-fx-v1/packages/shared/repository"
	"github.com/yaninyzwitty/go-fx-v1/packages/shared/sonyflake"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Params struct {
	fx.In

	Logger      *zap.Logger
	Queries     *repository.Queries
	IDGenerator sonyflake.Generator
}

type ProductServiceHandler struct {
	productsv1.UnimplementedProductServiceServer
	log     *zap.Logger
	queries *repository.Queries
	ids     sonyflake.Generator
}

// Module exports the product service controller
// Provides the gRPC service handler as a ProductServiceServer interface
var Module = fx.Module("controllers",
	fx.Provide(
		fx.Annotate(
			NewProductServiceHandler,
			fx.As(new(productsv1.ProductServiceServer)),
		),
	),
)

func NewProductServiceHandler(p Params) *ProductServiceHandler {
	return &ProductServiceHandler{
		log:     p.Logger.Named("product_controller"),
		queries: p.Queries,
		ids:     p.IDGenerator,
	}
}

func (c *ProductServiceHandler) GetProduct(ctx context.Context, req *productsv1.GetProductRequest) (*productsv1.GetProductResponse, error) {
	if req.GetId() == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}

	product, err := c.queries.GetProductByID(ctx, req.GetId())
	if err != nil {
		c.log.Error("failed to get product", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get product: %v", err)
	}

	return &productsv1.GetProductResponse{
		Product: mapDBToProto(product),
	}, nil
}

func (c *ProductServiceHandler) ListProducts(ctx context.Context, req *productsv1.ListProductsRequest) (*productsv1.ListProductsResponse, error) {
	products, err := c.queries.ListProducts(ctx, repository.ListProductsParams{
		Limit:  int32(req.GetPageSize()),
		Offset: int32(req.GetPageToken()),
	})
	if err != nil {
		c.log.Error("failed to list products", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}

	resp := &productsv1.ListProductsResponse{}
	for _, p := range products {
		resp.Products = append(resp.Products, mapDBToProto(p))
	}
	return resp, nil
}

func (c *ProductServiceHandler) CreateProduct(ctx context.Context, req *productsv1.CreateProductRequest) (*productsv1.CreateProductResponse, error) {
	// Validate required fields
	if req.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	if req.GetCurrency() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "currency is required")
	}
	if req.GetPrice() <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "price must be greater than 0")
	}

	id, err := c.ids.NextID()
	if err != nil {
		c.log.Error("failed to generate product ID", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to generate product ID: %v", err)
	}

	// Handle description - check if it's empty or not
	description := pgtype.Text{String: req.GetDescription(), Valid: req.GetDescription() != ""}

	product, err := c.queries.CreateProduct(ctx, repository.CreateProductParams{
		ID:            int64(id),
		Name:          req.GetName(),
		Description:   description,
		Price:         req.GetPrice(),
		Currency:      req.GetCurrency(),
		StockQuantity: int32(req.GetStockQuantity()),
	})
	if err != nil {
		c.log.Error("failed to create product", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	return &productsv1.CreateProductResponse{
		Product: mapDBToProto(product),
	}, nil
}

func (c *ProductServiceHandler) DeleteProduct(ctx context.Context, req *productsv1.DeleteProductRequest) (*productsv1.DeleteProductResponse, error) {
	if req.GetId() == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}

	err := c.queries.DeleteProduct(ctx, req.GetId())
	if err != nil {
		c.log.Error("failed to delete product", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete product: %v", err)
	}

	deleted := err == nil
	return &productsv1.DeleteProductResponse{
		Success: deleted,
	}, nil
}

func mapDBToProto(p repository.Product) *productsv1.Product {
	return &productsv1.Product{
		Id:            (uint64(p.ID)),
		Name:          (p.Name),
		Description:   (p.Description.String),
		Price:         (p.Price),
		Currency:      (p.Currency),
		StockQuantity: (uint32(p.StockQuantity)),
		CreatedAt:     timestamppb.New(p.CreatedAt),
		UpdatedAt:     timestamppb.New(p.UpdatedAt),
	}
}
