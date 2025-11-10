package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	productsv1 "github.com/yaninyzwitty/go-fx-v1/gen/products/v1"
	"github.com/yaninyzwitty/go-fx-v1/packages/gateway-service/internal/router"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Params contains dependencies for product controller
type Params struct {
	fx.In

	Logger        *zap.Logger
	ProductClient productsv1.ProductServiceClient
}

// ProductController handles HTTP requests for product operations
type ProductController struct {
	logger *zap.Logger
	client productsv1.ProductServiceClient
}

// NewProductController creates a new product controller
func NewProductController(p Params) *ProductController {
	return &ProductController{
		logger: p.Logger.Named("product_controller"),
		client: p.ProductClient,
	}
}

// ProductsRouteHandler handles all product routes: GET/POST /api/v1/products and GET/DELETE /api/v1/products/{id}
type ProductsRouteHandler struct {
	controller *ProductController
}

func (h *ProductsRouteHandler) Pattern() string {
	return "/api/v1/products"
}

// Patterns returns both the collection and item patterns
// This allows the router to register both "/api/v1/products" and "/api/v1/products/"
func (h *ProductsRouteHandler) Patterns() []string {
	return []string{
		"/api/v1/products",  // Exact match for collection
		"/api/v1/products/", // Prefix match for items with IDs
	}
}

func (h *ProductsRouteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	basePath := "/api/v1/products" // base path without trailing slash

	// Handle exact match for collection endpoint: /api/v1/products
	if path == basePath {
		switch r.Method {
		case http.MethodGet:
			h.handleListProducts(w, r)
		case http.MethodPost:
			h.handleCreateProduct(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// Handle item endpoints: /api/v1/products/{id}
	if len(path) > len(basePath)+1 && path[:len(basePath)+1] == basePath+"/" {
		// Extract ID from path: /api/v1/products/{id}
		idStr := path[len(basePath)+1:]
		// Check if there are more path segments (e.g., /api/v1/products/123/something)
		// If so, this is not a valid product ID endpoint
		for i := 0; i < len(idStr); i++ {
			if idStr[i] == '/' {
				http.Error(w, "Invalid product path", http.StatusBadRequest)
				return
			}
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.handleGetProduct(w, r, id)
		case http.MethodDelete:
			h.handleDeleteProduct(w, r, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// If we get here, the path doesn't match our expected patterns
	http.Error(w, "Not found", http.StatusNotFound)
}

func (h *ProductsRouteHandler) handleGetProduct(w http.ResponseWriter, r *http.Request, id int64) {
	req := &productsv1.GetProductRequest{Id: id}
	resp, err := h.controller.client.GetProduct(r.Context(), req)
	if err != nil {
		h.controller.handleError(w, err, "failed to get product")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.controller.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *ProductsRouteHandler) handleListProducts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	pageSize := uint32(10) // default
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.ParseUint(ps, 10, 32); err == nil {
			pageSize = uint32(parsed)
		}
	}

	pageToken := uint32(0)
	if pt := r.URL.Query().Get("page_token"); pt != "" {
		if parsed, err := strconv.ParseUint(pt, 10, 32); err == nil {
			pageToken = uint32(parsed)
		}
	}

	req := &productsv1.ListProductsRequest{
		PageSize:  pageSize,
		PageToken: pageToken,
	}

	resp, err := h.controller.client.ListProducts(r.Context(), req)
	if err != nil {
		h.controller.handleError(w, err, "failed to list products")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.controller.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *ProductsRouteHandler) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.controller.logger.Error("failed to read request body", zap.Error(err))
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req productsv1.CreateProductRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.controller.logger.Error("failed to unmarshal request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.controller.client.CreateProduct(r.Context(), &req)
	if err != nil {
		h.controller.handleError(w, err, "failed to create product")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.controller.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *ProductsRouteHandler) handleDeleteProduct(w http.ResponseWriter, r *http.Request, id int64) {
	req := &productsv1.DeleteProductRequest{Id: id}
	resp, err := h.controller.client.DeleteProduct(r.Context(), req)
	if err != nil {
		h.controller.handleError(w, err, "failed to delete product")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.controller.logger.Error("failed to encode response", zap.Error(err))
	}
}

// NewProductsRouteHandler creates a route handler for all product operations
func NewProductsRouteHandler(controller *ProductController) router.RouteHandler {
	return &ProductsRouteHandler{controller: controller}
}

// handleError converts gRPC errors to HTTP responses
func (c *ProductController) handleError(w http.ResponseWriter, err error, message string) {
	st, ok := status.FromError(err)
	if !ok {
		c.logger.Error(message, zap.Error(err))
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
		c.logger.Error(message, zap.Error(err), zap.String("code", st.Code().String()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Module exports the product controller module
// Provides the product controller and route handler
var Module = fx.Module("controllers",
	fx.Provide(
		NewProductController,
		fx.Annotate(
			NewProductsRouteHandler,
			fx.ResultTags(`group:"routes"`),
		),
	),
)
