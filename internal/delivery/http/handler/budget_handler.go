package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/dto"
	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/middleware"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

type BudgetHandler struct {
	budgetService domain.BudgetService
}

func NewBudgetHandler(budgetService domain.BudgetService) *BudgetHandler {
	return &BudgetHandler{
		budgetService: budgetService,
	}
}

// List godoc
//
//	@Summary	List budgets owned by the logged-in user
//	@Tags		budgets
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	response.Success{data=[]dto.BudgetResponseDTO}
//	@Failure	401	{object}	response.Error
//	@Router		/budgets [get]
func (h *BudgetHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	budgets, err := h.budgetService.List(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "budget list", dto.NewBudgetResponseDTOs(budgets))
}

// GetByID godoc
//
//	@Summary	Budget detail
//	@Tags		budgets
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Budget ID (UUID)"
//	@Success	200	{object}	response.Success{data=dto.BudgetResponseDTO}
//	@Failure	401	{object}	response.Error
//	@Failure	404	{object}	response.Error
//	@Router		/budgets/{id} [get]
func (h *BudgetHandler) GetByID(c *gin.Context) {
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

	budget, err := h.budgetService.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "budget found", dto.NewBudgetResponseDTO(budget))
}

// Create godoc
//
//	@Summary	Create a budget
//	@Tags		budgets
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		dto.CreateBudgetRequestDTO	true	"Budget data"
//	@Success	201		{object}	response.Success{data=dto.BudgetResponseDTO}
//	@Failure	400		{object}	response.Error
//	@Failure	401		{object}	response.Error
//	@Failure	409		{object}	response.Error
//	@Failure	422		{object}	response.Error
//	@Router		/budgets [post]
func (h *BudgetHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	var req dto.CreateBudgetRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	budget, err := h.budgetService.Create(c.Request.Context(), userID, req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusCreated, "budget created", dto.NewBudgetResponseDTO(budget))
}

// Update godoc
//
//	@Summary	Update a budget
//	@Tags		budgets
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		string						true	"Budget ID (UUID)"
//	@Param		request	body		dto.UpdateBudgetRequestDTO	true	"Fields to update"
//	@Success	200		{object}	response.Success{data=dto.BudgetResponseDTO}
//	@Failure	400		{object}	response.Error
//	@Failure	401		{object}	response.Error
//	@Failure	404		{object}	response.Error
//	@Failure	409		{object}	response.Error
//	@Router		/budgets/{id} [patch]
func (h *BudgetHandler) Update(c *gin.Context) {
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

	var req dto.UpdateBudgetRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	budget, err := h.budgetService.Update(c.Request.Context(), userID, id, req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "budget updated", dto.NewBudgetResponseDTO(budget))
}

// Delete godoc
//
//	@Summary	Delete a budget
//	@Tags		budgets
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Budget ID (UUID)"
//	@Success	200	{object}	response.Success
//	@Failure	401	{object}	response.Error
//	@Failure	404	{object}	response.Error
//	@Router		/budgets/{id} [delete]
func (h *BudgetHandler) Delete(c *gin.Context) {
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

	if err := h.budgetService.Delete(c.Request.Context(), userID, id); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "budget deleted", nil)
}

// Overview godoc
//
//	@Summary	Budget summary for the current cycle
//	@Tags		budgets
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	response.Success{data=dto.BudgetOverviewResponseDTO}
//	@Failure	401	{object}	response.Error
//	@Router		/budgets/overview [get]
func (h *BudgetHandler) Overview(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	overview, err := h.budgetService.Overview(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "budget overview", dto.NewBudgetOverviewResponseDTO(overview))
}
