package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/dto"
	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/middleware"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

type CategoryHandler struct {
	categoryService domain.CategoryService
}

func NewCategoryHandler(categoryService domain.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// List godoc
//
//	@Summary	List categories owned by the logged-in user
//	@Tags		categories
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	response.Success{data=[]dto.CategoryResponseDTO}
//	@Failure	401	{object}	response.Error
//	@Router		/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	categories, err := h.categoryService.List(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "category list", dto.NewCategoryResponseDTOs(categories))
}

// GetByID godoc
//
//	@Summary	Category detail
//	@Tags		categories
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Category ID (UUID)"
//	@Success	200	{object}	response.Success{data=dto.CategoryResponseDTO}
//	@Failure	404	{object}	response.Error
//	@Router		/categories/{id} [get]
func (h *CategoryHandler) GetByID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	id, err := parseUUIDParam(c, "id")
	if err != nil {
		handleError(c, err)
		return
	}

	category, err := h.categoryService.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "category found", dto.NewCategoryResponseDTO(category))
}

// Create godoc
//
//	@Summary	Create a category
//	@Tags		categories
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		dto.CreateCategoryRequestDTO	true	"Category data"
//	@Success	201		{object}	response.Success{data=dto.CategoryResponseDTO}
//	@Failure	400		{object}	response.Error
//	@Failure	409		{object}	response.Error
//	@Failure	422		{object}	response.Error
//	@Router		/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	var req dto.CreateCategoryRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	category, err := h.categoryService.Create(c.Request.Context(), userID, req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusCreated, "category created", dto.NewCategoryResponseDTO(category))
}

// Update godoc
//
//	@Summary	Update a category
//	@Tags		categories
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		string							true	"Category ID (UUID)"
//	@Param		request	body		dto.UpdateCategoryRequestDTO	true	"Fields to update"
//	@Success	200		{object}	response.Success{data=dto.CategoryResponseDTO}
//	@Failure	400		{object}	response.Error
//	@Failure	404		{object}	response.Error
//	@Failure	409		{object}	response.Error
//	@Router		/categories/{id} [patch]
func (h *CategoryHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	id, err := parseUUIDParam(c, "id")
	if err != nil {
		handleError(c, err)
		return
	}

	var req dto.UpdateCategoryRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	category, err := h.categoryService.Update(c.Request.Context(), userID, id, req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "category updated", dto.NewCategoryResponseDTO(category))
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
//	@Failure	409	{object}	response.Error
//	@Router		/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	id, err := parseUUIDParam(c, "id")
	if err != nil {
		handleError(c, err)
		return
	}

	if err := h.categoryService.Delete(c.Request.Context(), userID, id); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "category deleted", nil)
}

// OptionsCategoryType godoc
//
//	@Summary	Category options for one screen
//	@Tags		categories
//	@Produce	json
//	@Security	BearerAuth
//	@Param		slug	query		string	true	"Screen asking: filter takes every type, budget only expense"	Enums(filter, budget)
//	@Success	200		{object}	response.Success{data=[]dto.OptionCategoryResponseDTO}
//	@Failure	400		{object}	response.Error
//	@Failure	401		{object}	response.Error
//	@Router		/categories/options [get]
func (h *CategoryHandler) OptionsCategoryType(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	var query dto.OptionCategoryTypeQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		handleBindError(c, err)
		return
	}

	options, err := h.categoryService.OptionsCategoryType(c.Request.Context(), userID, query.Slug)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "filter category types", dto.NewOptionCategoryResponseDTOs(options))
}
