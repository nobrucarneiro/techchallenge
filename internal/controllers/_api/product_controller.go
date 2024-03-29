package _api

import (
	"errors"
	application "g37-lanchonete/internal/controllers/application"
	"g37-lanchonete/internal/core/usecases"
	"g37-lanchonete/internal/core/usecases/dto"
	"g37-lanchonete/internal/infra/drivers/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	productUsecase usecases.ProductUsecase
}

func NewProductController(productUsecase usecases.ProductUsecase) ProductController {
	return ProductController{
		productUsecase: productUsecase,
	}
}

func (c ProductController) GetProducts(ctx *gin.Context) {
	category := ctx.Query("category")
	pageParams, err := application.GetPageParams(ctx)
	if err != nil {
		application.HandleBadRequestResponse(ctx, "invalid query parameters", err)
	}

	if category != "" {
		c.getProductsByCategory(ctx, pageParams, category)
		return
	}

	c.getAllProducts(ctx, pageParams)
}

func (c ProductController) CreateProducts(ctx *gin.Context) {
	var product dto.ProductDTO
	err := ctx.ShouldBindJSON(&product)
	if err != nil {
		application.HandleBadRequestResponse(ctx, "failed to bind product payload", err)
		return
	}

	valid, err := product.ValidateProduct()
	if !valid {
		application.HandleBadRequestResponse(ctx, "invalid product payload", err)
		return
	}

	err = c.productUsecase.CreateProduct(product)
	if err != nil {
		application.HandleInternalServerResponse(ctx, "failed to create product", err)
		return
	}

	ctx.Status(http.StatusOK)
}

func (c ProductController) UpdateProduct(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		application.HandleBadRequestResponse(ctx, "id path param is required", errors.New("id path parameter is missing"))
		return
	}

	var product dto.ProductDTO
	err := ctx.ShouldBindJSON(&product)
	if err != nil {
		application.HandleBadRequestResponse(ctx, "failed to bind product payload", err)
		return
	}

	valid, err := product.ValidateProduct()
	if !valid {
		application.HandleBadRequestResponse(ctx, "invalid product payload", err)
		return
	}

	err = c.productUsecase.UpdateProduct(id, product)
	if err != nil {
		if errors.Is(err, sql.ErrNotFound) {
			application.HandleNotFoundResponse(ctx, "product not found", err)
			return
		}
		application.HandleInternalServerResponse(ctx, "failed to create product", err)
		return
	}

	ctx.Status(http.StatusOK)
}

func (c ProductController) DeleteProduct(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		application.HandleBadRequestResponse(ctx, "id path param is required", errors.New("id path parameter is missing"))
		return
	}

	err := c.productUsecase.DeleteProduct(id)
	if err != nil {
		if errors.Is(err, sql.ErrNotFound) {
			application.HandleNotFoundResponse(ctx, "product not found", err)
			return
		}
		application.HandleInternalServerResponse(ctx, "failed to delete product", err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c ProductController) getAllProducts(ctx *gin.Context, pageParameters dto.PageParams) {
	products, err := c.productUsecase.GetAllProducts(pageParameters)
	if err != nil {
		application.HandleInternalServerResponse(ctx, "failed to get all products", err)
		return
	}
	ctx.JSON(http.StatusOK, products)
}

func (c ProductController) getProductsByCategory(ctx *gin.Context, pageParameters dto.PageParams, category string) {
	products, err := c.productUsecase.GetProductsByCategory(pageParameters, category)
	if err != nil {
		application.HandleInternalServerResponse(ctx, "failed to get products by category", err)
		return
	}
	ctx.JSON(http.StatusOK, products)
}
