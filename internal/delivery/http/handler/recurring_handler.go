package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/dto"
	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/middleware"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

type RecurringHandler struct {
	recurringService domain.RecurringService
}

func NewRecurringHandler(recurringService domain.RecurringService) *RecurringHandler {
	return &RecurringHandler{recurringService: recurringService}
}

// List godoc
//
//	@Summary	List recurring items owned by the logged-in user
//	@Tags		recurring
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	response.Success{data=[]dto.RecurringResponseDTO}
//	@Failure	401	{object}	response.Error
//	@Router		/recurring [get]
func (h *RecurringHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	items, err := h.recurringService.List(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "recurring list", dto.NewRecurringResponseDTOs(items))
}

// GetByID godoc
//
//	@Summary	Recurring item detail
//	@Tags		recurring
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Recurring ID (UUID)"
//	@Success	200	{object}	response.Success{data=dto.RecurringResponseDTO}
//	@Failure	401	{object}	response.Error
//	@Failure	404	{object}	response.Error
//	@Router		/recurring/{id} [get]
func (h *RecurringHandler) GetByID(c *gin.Context) {
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

	item, err := h.recurringService.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "recurring found", dto.NewRecurringResponseDTO(item))
}

// Create godoc
//
//	@Summary	Create a recurring item
//	@Tags		recurring
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		dto.CreateRecurringRequestDTO	true	"Recurring data"
//	@Success	201		{object}	response.Success{data=dto.RecurringResponseDTO}
//	@Failure	400		{object}	response.Error
//	@Failure	401		{object}	response.Error
//	@Router		/recurring [post]
func (h *RecurringHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	var req dto.CreateRecurringRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	item, err := h.recurringService.Create(c.Request.Context(), userID, req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusCreated, "recurring created", dto.NewRecurringResponseDTO(item))
}

// Update godoc
//
//	@Summary	Update a recurring item
//	@Tags		recurring
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		string							true	"Recurring ID (UUID)"
//	@Param		request	body		dto.UpdateRecurringRequestDTO	true	"Fields to update"
//	@Success	200		{object}	response.Success{data=dto.RecurringResponseDTO}
//	@Failure	400		{object}	response.Error
//	@Failure	401		{object}	response.Error
//	@Failure	404		{object}	response.Error
//	@Router		/recurring/{id} [patch]
func (h *RecurringHandler) Update(c *gin.Context) {
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

	var req dto.UpdateRecurringRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	item, err := h.recurringService.Update(c.Request.Context(), userID, id, req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "recurring updated", dto.NewRecurringResponseDTO(item))
}

// Delete godoc
//
//	@Summary	Delete a recurring item
//	@Tags		recurring
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Recurring ID (UUID)"
//	@Success	200	{object}	response.Success
//	@Failure	401	{object}	response.Error
//	@Failure	404	{object}	response.Error
//	@Router		/recurring/{id} [delete]
func (h *RecurringHandler) Delete(c *gin.Context) {
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

	if err := h.recurringService.Delete(c.Request.Context(), userID, id); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "recurring deleted", nil)
}

// Overview godoc
//
//	@Summary	Recurring summary: monthly commitment and what is due soon
//	@Tags		recurring
//	@Produce	json
//	@Security	BearerAuth
//	@Param		date	query		string	false	"The caller's today (YYYY-MM-DD)"
//	@Success	200		{object}	response.Success{data=dto.RecurringOverviewResponseDTO}
//	@Failure	401		{object}	response.Error
//	@Router		/recurring/overview [get]
func (h *RecurringHandler) Overview(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	var query dto.OverviewRecurringQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		handleBindError(c, err)
		return
	}

	overview, err := h.recurringService.Overview(c.Request.Context(), userID, query.Today(time.Now()))
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "recurring overview", dto.NewRecurringOverviewResponseDTO(overview))
}

// Confirm godoc
//
//	@Summary	Record this cycle's payment and move the item to its next due date
//	@Tags		recurring
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		string							true	"Recurring ID (UUID)"
//	@Param		request	body		dto.ConfirmRecurringRequestDTO	false	"Overrides for this payment"
//	@Success	200		{object}	response.Success{data=dto.RecurringResponseDTO}
//	@Failure	400		{object}	response.Error
//	@Failure	401		{object}	response.Error
//	@Failure	404		{object}	response.Error
//	@Failure	409		{object}	response.Error
//	@Router		/recurring/{id}/confirm [post]
func (h *RecurringHandler) Confirm(c *gin.Context) {
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

	var req dto.ConfirmRecurringRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	item, err := h.recurringService.Confirm(c.Request.Context(), userID, id, req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "recurring payment recorded", dto.NewRecurringResponseDTO(item))
}
