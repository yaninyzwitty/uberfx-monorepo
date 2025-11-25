package controllers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	productsv1 "github.com/yaninyzwitty/go-fx-v1/gen/products/v1"
	"github.com/yaninyzwitty/go-fx-v1/packages/gateway-service/internal/router"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Params contains dependencies for the product controller.
type Params struct {
	fx.In

	Logger        *zap.Logger
	ProductClient productsv1.ProductServiceClient
}

// ProductController handles business logic for products.
type ProductController struct {
	logger *zap.Logger
	client productsv1.ProductServiceClient
}

// NewProductController creates a new product controller.
func NewProductController(p Params) *ProductController {
	return &ProductController{
		logger: p.Logger.Named("product_controller"),
		client: p.ProductClient,
	}
}

// ProductsRouteHandler handles HTTP requests for product endpoints.
type ProductsRouteHandler struct {
	controller *ProductController
}

// routeType distinguishes between collection and item routes.
type routeType int

const (
	routeTypeUnknown routeType = iota
	routeTypeCollection
	routeTypeItem
)

type parsedRoute struct {
	Type routeType
	ID   int64
}

// NewProductsRouteHandler constructs a route handler for product endpoints.
func NewProductsRouteHandler(controller *ProductController) router.RouteHandler {
	return &ProductsRouteHandler{controller: controller}
}

// Pattern returns the base route for products.
func (h *ProductsRouteHandler) Pattern() string {
	return "/api/v1/products"
}

// Patterns returns all supported route patterns.
func (h *ProductsRouteHandler) Patterns() []string {
	return []string{
		"/api/v1/products",  // collection
		"/api/v1/products/", // item prefix
	}
}

// ServeHTTP dispatches requests to the appropriate handler.
func (h *ProductsRouteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := h.parseRoute(r.URL.Path)
	if route == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	switch route.Type {
	case routeTypeCollection:
		switch r.Method {
		case http.MethodGet:
			h.handleListProducts(w, r)
		case http.MethodPost:
			h.handleCreateProduct(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	case routeTypeItem:
		switch r.Method {
		case http.MethodGet:
			h.handleGetProduct(w, r, route.ID)
		case http.MethodDelete:
			h.handleDeleteProduct(w, r, route.ID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// parseRoute determines if the request is for a collection or item route.
func (h *ProductsRouteHandler) parseRoute(path string) *parsedRoute {
	base := "/api/v1/products"
	path = strings.TrimSuffix(path, "/")

	if path == base {
		return &parsedRoute{Type: routeTypeCollection}
	}

	if after, ok := strings.CutPrefix(path, base+"/"); ok {
		idStr := after
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			return &parsedRoute{Type: routeTypeItem, ID: id}
		}
	}

	return nil
}

// handleGetProduct retrieves a single product by ID.
func (h *ProductsRouteHandler) handleGetProduct(w http.ResponseWriter, r *http.Request, id int64) {
	ctx := h.controller.contextWithTelemetry(r.Context())

	resp, err := h.controller.client.GetProduct(ctx, &productsv1.GetProductRequest{Id: id})
	if err != nil {
		h.controller.handleError(w, err, "failed to get product")
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// handleListProducts retrieves a paginated list of products.
func (h *ProductsRouteHandler) handleListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := h.controller.contextWithTelemetry(r.Context())

	// Default pagination
	pageSize := uint32(10)
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.ParseUint(ps, 10, 32); err == nil {
			pageSize = uint32(parsed)
		}
	}

	pageToken := r.URL.Query().Get("page_token")
	var pageTokenUint32 uint32
	if pageToken == "" {
		pageTokenUint32 = 0
	} else {
		parsed, err := strconv.ParseUint(pageToken, 10, 32)
		if err != nil {
			h.controller.handleError(w, err, "invalid page_token")
			return
		}
		pageTokenUint32 = uint32(parsed)
	}

	resp, err := h.controller.client.ListProducts(ctx, &productsv1.ListProductsRequest{
		PageSize:  pageSize,
		PageToken: pageTokenUint32,
	})

	if err != nil {
		h.controller.handleError(w, err, "failed to list products")
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// handleCreateProduct creates a new product.
func (h *ProductsRouteHandler) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := h.controller.contextWithTelemetry(r.Context())

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.controller.logger.Error("failed to read request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.controller.logger.Error("failed to close body", zap.Error(err))
			http.Error(w, "failed to close request body", http.StatusBadRequest)
		}

	}()

	var req productsv1.CreateProductRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.controller.logger.Error("failed to unmarshal request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.controller.client.CreateProduct(ctx, &req)
	if err != nil {
		h.controller.handleError(w, err, "failed to create product")
		return
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

// handleDeleteProduct deletes a product by ID.
func (h *ProductsRouteHandler) handleDeleteProduct(w http.ResponseWriter, r *http.Request, id int64) {
	ctx := h.controller.contextWithTelemetry(r.Context())

	resp, err := h.controller.client.DeleteProduct(ctx, &productsv1.DeleteProductRequest{Id: id})
	if err != nil {
		h.controller.handleError(w, err, "failed to delete product")
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// writeJSON encodes a response as JSON and writes it to the ResponseWriter.
func (h *ProductsRouteHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.controller.logger.Error("failed to encode response", zap.Error(err))
	}
}

// handleError converts gRPC errors to appropriate HTTP responses.
func (c *ProductController) handleError(w http.ResponseWriter, err error, msg string) {
	st, ok := status.FromError(err)
	if !ok {
		c.logger.Error(msg, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	switch st.Code() {
	case codes.InvalidArgument, codes.FailedPrecondition:
		http.Error(w, st.Message(), http.StatusBadRequest)
	case codes.NotFound:
		http.Error(w, st.Message(), http.StatusNotFound)
	case codes.AlreadyExists:
		http.Error(w, st.Message(), http.StatusConflict)
	default:
		c.logger.Error(msg, zap.Error(err), zap.String("code", st.Code().String()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// contextWithTelemetry adds metadata to outgoing gRPC requests.
func (c *ProductController) contextWithTelemetry(ctx context.Context) context.Context {
	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.RFC3339Nano),
		"client-id", "web-api-client-us-east-1",
		"user-id", "some-test-user-id",
	)
	return metadata.NewOutgoingContext(ctx, md)
}

// Module exports the product controller and route handler.
var Module = fx.Module("controllers",
	fx.Provide(
		NewProductController,
		fx.Annotate(
			NewProductsRouteHandler,
			fx.ResultTags(`group:"routes"`),
		),
	),
)
