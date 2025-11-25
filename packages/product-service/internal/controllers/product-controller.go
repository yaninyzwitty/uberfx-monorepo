// Package controllers implements gRPC service handlers for product operations.
package controllers

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	productsv1 "github.com/yaninyzwitty/go-fx-v1/gen/products/v1"
	metrics "github.com/yaninyzwitty/go-fx-v1/packages/product-service/internal/grpc-metrics"
	"github.com/yaninyzwitty/go-fx-v1/packages/shared/repository"
	"github.com/yaninyzwitty/go-fx-v1/packages/shared/sonyflake"
	"go.opentelemetry.io/otel/trace"
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
	Tracer      trace.Tracer
	AppMetrics  *metrics.AppMetrics
}

type ProductServiceHandler struct {
	productsv1.UnimplementedProductServiceServer
	log     *zap.Logger
	queries *repository.Queries
	ids     sonyflake.Generator
	tracer  trace.Tracer
	metrics *metrics.AppMetrics
}

var Module = fx.Module("controllers",
	fx.Provide(
		fx.Annotate(
			NewProductServiceHandler,
			fx.As(new(productsv1.ProductServiceServer)),
		),
	),
)

const (
	defaultPageSize  = uint32(10)
	defaultPageToken = uint32(0)
	dbBackend        = "postgres"
)

func NewProductServiceHandler(p Params) *ProductServiceHandler {
	return &ProductServiceHandler{
		log:     p.Logger.Named("product_controller"),
		queries: p.Queries,
		ids:     p.IDGenerator,
		tracer:  p.Tracer,
		metrics: p.AppMetrics,
	}
}

func (c *ProductServiceHandler) startSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return c.tracer.Start(ctx, name)
}

func (c *ProductServiceHandler) CreateProduct(ctx context.Context, req *productsv1.CreateProductRequest) (*productsv1.CreateProductResponse, error) {
	ctx, span := c.startSpan(ctx, "CreateProduct.Handler")
	defer span.End()

	const op = "create_product"
	timerStart := time.Now()

	defer func() {
		c.metrics.Duration.
			WithLabelValues(op, dbBackend).
			Observe(time.Since(timerStart).Seconds())
	}()

	if req.GetName() == "" {
		c.metrics.Errors.WithLabelValues(op, dbBackend).Inc()
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	if req.GetCurrency() == "" {
		c.metrics.Errors.WithLabelValues(op, dbBackend).Inc()
		return nil, status.Errorf(codes.InvalidArgument, "currency is required")
	}
	if req.GetPrice() <= 0 {
		c.metrics.Errors.WithLabelValues(op, dbBackend).Inc()
		return nil, status.Errorf(codes.InvalidArgument, "price must be greater than 0")
	}

	id, err := c.ids.NextID()
	if err != nil {
		c.metrics.Errors.WithLabelValues(op, dbBackend).Inc()
		return nil, status.Errorf(codes.Internal, "failed to generate product ID: %v", err)
	}

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
		c.metrics.Errors.WithLabelValues(op, dbBackend).Inc()
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	// Optional workflow stage update
	c.metrics.Stage.Set(2)

	return &productsv1.CreateProductResponse{
		Product: mapDBToProto(product),
	}, nil
}

func (c *ProductServiceHandler) ListProducts(ctx context.Context, req *productsv1.ListProductsRequest) (*productsv1.ListProductsResponse, error) {
	ctx, span := c.startSpan(ctx, "ListProducts.Handler")
	defer span.End()

	const op = "list_products"
	timerStart := time.Now()

	defer func() {
		c.metrics.Duration.
			WithLabelValues(op, dbBackend).
			Observe(time.Since(timerStart).Seconds())
	}()

	pageSize := req.GetPageSize()
	pageToken := req.GetPageToken()

	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	if pageToken == 0 {
		pageToken = defaultPageToken
	}

	products, err := c.queries.ListProducts(ctx, repository.ListProductsParams{
		Limit:  int32(pageSize),
		Offset: int32(pageToken),
	})
	if err != nil {
		c.metrics.Errors.WithLabelValues(op, dbBackend).Inc()
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}

	resp := &productsv1.ListProductsResponse{
		Products: make([]*productsv1.Product, 0, len(products)),
	}
	for _, p := range products {
		resp.Products = append(resp.Products, mapDBToProto(p))
	}

	return resp, nil
}

func (c *ProductServiceHandler) DeleteProduct(ctx context.Context, req *productsv1.DeleteProductRequest) (*productsv1.DeleteProductResponse, error) {
	ctx, span := c.startSpan(ctx, "DeleteProduct.Handler")
	defer span.End()

	const op = "delete_product"
	timerStart := time.Now()

	defer func() {
		c.metrics.Duration.
			WithLabelValues(op, dbBackend).
			Observe(time.Since(timerStart).Seconds())
	}()

	if req.GetId() == 0 {
		c.metrics.Errors.WithLabelValues(op, dbBackend).Inc()
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}

	err := c.queries.DeleteProduct(ctx, req.GetId())
	if err != nil {
		c.metrics.Errors.WithLabelValues(op, dbBackend).Inc()
		return nil, status.Errorf(codes.Internal, "failed to delete product: %v", err)
	}

	return &productsv1.DeleteProductResponse{
		Success: true,
	}, nil
}

func mapDBToProto(p repository.Product) *productsv1.Product {
	return &productsv1.Product{
		Id:            uint64(p.ID),
		Name:          p.Name,
		Description:   p.Description.String,
		Price:         p.Price,
		Currency:      p.Currency,
		StockQuantity: uint32(p.StockQuantity),
		CreatedAt:     timestamppb.New(p.CreatedAt),
		UpdatedAt:     timestamppb.New(p.UpdatedAt),
	}
}
