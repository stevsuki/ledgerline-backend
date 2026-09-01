package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/dto"
	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/middleware"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/pagination"
	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

type CategoryHandler struct {
	categoryService domain.CategoryService
}

func NewCategoryHandler(categoryService domain.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// Create godoc
//
//	@Summary	Create a category
//	@Tags		categories
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		dto.CreateCategoryRequest	true	"Category data"
//	@Success	201		{object}	response.Success{data=dto.CategoryResponse}
//	@Failure	409		{object}	response.Error
//	@Failure	422		{object}	response.Error
//	@Router		/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrUnauthorized)
		return
	}

	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	category, err := h.categoryService.Create(c.Request.Context(), userID, req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusCreated, "category created", dto.NewCategoryResponse(category))
}

// List godoc
//
//	@Summary	List categories owned by the logged-in user
//	@Tags		categories
//	@Produce	json
//	@Security	BearerAuth
//	@Param		search		query		string	false	"Search by category name"
//	@Param		type		query		string	false	"Filter by type"		Enums(income, expense)
//	@Param		sort		query		string	false	"Order: name, type, created_at, updated_at. Prefix - for desc"	default(-created_at)
//	@Param		page		query		int		false	"Page"			default(1)
//	@Param		per_page	query		int		false	"Items per page"	default(10)
//	@Success	200			{object}	response.Success{data=[]dto.CategoryResponse}
//	@Router		/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrUnauthorized)
		return
	}

	var query dto.ListCategoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		handleBindError(c, err)
		return
	}

	orderBy, err := query.OrderBy()
	if err != nil {
		handleError(c, err)
		return
	}

	params := pagination.Params{Page: query.Page, PerPage: query.PerPage}.Normalize()
	categories, total, err := h.categoryService.List(c.Request.Context(), domain.CategoryFilter{
		UserID:  userID,
		Search:  query.Search,
		Type:    domain.CategoryType(query.Type),
		Limit:   params.Limit(),
		Offset:  params.Offset(),
		OrderBy: orderBy,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	response.Paginated(c, http.StatusOK, "success", dto.NewCategoryResponses(categories), response.Meta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalItems: total,
		TotalPages: pagination.TotalPages(total, params.PerPage),
	})
}

// GetByID godoc
//
//	@Summary	Category detail
//	@Tags		categories
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Category ID (UUID)"
//	@Success	200	{object}	response.Success{data=dto.CategoryResponse}
//	@Failure	404	{object}	response.Error
//	@Router		/categories/{id} [get]
func (h *CategoryHandler) GetByID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrUnauthorized)
		return
	}

	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}

	category, err := h.categoryService.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "success", dto.NewCategoryResponse(category))
}

// Update godoc
//
//	@Summary	Update a category
//	@Tags		categories
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		string						true	"Category ID (UUID)"
//	@Param		request	body		dto.UpdateCategoryRequest	true	"Fields to update"
//	@Success	200		{object}	response.Success{data=dto.CategoryResponse}
//	@Failure	404		{object}	response.Error
//	@Router		/categories/{id} [patch]
func (h *CategoryHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrUnauthorized)
		return
	}

	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	category, err := h.categoryService.Update(c.Request.Context(), userID, id, req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "category updated", dto.NewCategoryResponse(category))
}

// Delete godoc
//
//	@Summary	Delete a category
//	@Tags		categories
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Category ID (UUID)"
//	@Success	200	{object}	response.Success
//	@Failure	404	{object}	response.Error
//	@Router		/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrUnauthorized)
		return
	}

	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}

	if err := h.categoryService.Delete(c.Request.Context(), userID, id); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "category deleted", nil)
}
