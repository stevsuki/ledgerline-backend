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

type TransactionHandler struct {
	transactionService domain.TransactionService
}

func NewTransactionHandler(transactionService domain.TransactionService) *TransactionHandler {
	return &TransactionHandler{transactionService: transactionService}
}

// List godoc
//
//	@Summary	List transactions owned by the logged-in user
//	@Tags		transactions
//	@Produce	json
//	@Security	BearerAuth
//	@Param		search		query		string	false	"Search by transaction name"
//	@Param		sort		query		string	false	"Order: name, amount, note, type, category, wallet, occurred_at, created_at, updated_at. Prefix - for desc"	default(-occurred_at)
//	@Param		page		query		int		false	"Page"							default(1)
//	@Param		per_page	query		int		false	"Items per page"				default(10)
//	@Param		category		query		string	false	"Filter by category ID (UUID)"
//	@Param		wallet			query		string	false	"Filter by wallet ID (UUID)"
//	@Param		type			query		string	false	"Filter by type: income or expense"	Enums(income, expense)
//	@Param		occurred_from	query		string	false	"Filter occurred_at from (RFC3339)"
//	@Param		occurred_to		query		string	false	"Filter occurred_at to (RFC3339)"
//	@Success	200			{object}	response.Success{data=dto.TransactionListResponseDTO}
//	@Failure	400			{object}	response.Error
//	@Failure	401			{object}	response.Error
//	@Router		/transactions [get]
func (h *TransactionHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	var query dto.ListTransactionsQuery
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
	filter := domain.TransactionFilter{
		Search:       query.Search,
		Limit:        params.Limit(),
		Offset:       params.Offset(),
		OrderBy:      orderBy,
		CategoryID:   query.CategoryID,
		WalletID:     query.WalletID,
		OccurredFrom: query.OccurredFrom,
		OccurredTo:   query.OccurredTo,
	}
	if query.Type != "" {
		t := domain.TransactionType(query.Type)
		filter.Type = &t
	}

	transactions, total, err := h.transactionService.List(c.Request.Context(), userID, filter)
	if err != nil {
		handleError(c, err)
		return
	}

	summary, err := h.transactionService.Summary(c.Request.Context(), userID, filter)
	if err != nil {
		handleError(c, err)
		return
	}

	body := dto.TransactionListResponseDTO{
		Summary: dto.NewTransactionSummaryResponseDTO(summary),
		Groups:  dto.NewTransactionGroupResponseDTOs(transactions),
	}

	response.Paginated(c, http.StatusOK, "success", body, response.Meta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalItems: total,
		TotalPages: pagination.TotalPages(total, params.PerPage),
	})
}

// GetByID godoc
//
//	@Summary	Transaction detail
//	@Tags		transactions
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Transaction ID (UUID)"
//	@Success	200	{object}	response.Success{data=dto.TransactionResponseDTO}
//	@Failure	404	{object}	response.Error
//	@Router		/transactions/{id} [get]
func (h *TransactionHandler) GetByID(c *gin.Context) {
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

	transaction, err := h.transactionService.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "transaction found", dto.NewTransactionResponseDTO(transaction))
}

// Create godoc
//
//	@Summary	Create a transaction
//	@Tags		transactions
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		dto.CreateTransactionRequestDTO	true	"Transaction data"
//	@Success	201		{object}	response.Success{data=dto.TransactionResponseDTO}
//	@Failure	400		{object}	response.Error
//	@Failure	422		{object}	response.Error
//	@Router		/transactions [post]
func (h *TransactionHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	var req dto.CreateTransactionRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	transaction, err := h.transactionService.Create(c.Request.Context(), userID, req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusCreated, "transaction created", dto.NewTransactionResponseDTO(transaction))
}

// Update godoc
//
//	@Summary	Update a transaction
//	@Tags		transactions
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		string							true	"Transaction ID (UUID)"
//	@Param		request	body		dto.UpdateTransactionRequestDTO	true	"Fields to update"
//	@Success	200		{object}	response.Success{data=dto.TransactionResponseDTO}
//	@Failure	400		{object}	response.Error
//	@Failure	404		{object}	response.Error
//	@Failure	422		{object}	response.Error
//	@Router		/transactions/{id} [patch]
func (h *TransactionHandler) Update(c *gin.Context) {
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

	var req dto.UpdateTransactionRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	transaction, err := h.transactionService.Update(c.Request.Context(), userID, id, req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "transaction updated", dto.NewTransactionResponseDTO(transaction))
}

// Delete godoc
//
//	@Summary	Delete a transaction
//	@Tags		transactions
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Transaction ID (UUID)"
//	@Success	200	{object}	response.Success
//	@Failure	404	{object}	response.Error
//	@Router		/transactions/{id} [delete]
func (h *TransactionHandler) Delete(c *gin.Context) {
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

	if err := h.transactionService.Delete(c.Request.Context(), userID, id); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "transaction deleted", nil)
}
