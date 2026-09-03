package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/dto"
	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/middleware"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

type WalletHandler struct {
	walletService domain.WalletService
}

func NewWalletHandler(walletService domain.WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

// List godoc
//
//	@Summary	List wallets owned by the logged-in user
//	@Tags		wallets
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	response.Success{data=[]dto.WalletResponseDTO}
//	@Failure	401	{object}	response.Error
//	@Router		/wallets [get]
func (h *WalletHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	wallets, err := h.walletService.List(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "wallet list", dto.NewWalletResponseDTOs(wallets))
}

// Create godoc
//
//	@Summary	Create a wallet
//	@Tags		wallets
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		dto.CreateWalletRequestDTO	true	"Wallet data"
//	@Success	201		{object}	response.Success{data=dto.WalletResponseDTO}
//	@Failure	400		{object}	response.Error
//	@Failure	409		{object}	response.Error
//	@Failure	422		{object}	response.Error
//	@Router		/wallets [post]
func (h *WalletHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	var req dto.CreateWalletRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	wallet, err := h.walletService.Create(c.Request.Context(), userID, req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusCreated, "wallet created", dto.NewWalletResponseDTO(wallet))
}

// GetByID godoc
//
//	@Summary	Wallet detail
//	@Tags		wallets
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Wallet ID (UUID)"
//	@Success	200	{object}	response.Success{data=dto.WalletResponseDTO}
//	@Failure	404	{object}	response.Error
//	@Router		/wallets/{id} [get]
func (h *WalletHandler) GetByID(c *gin.Context) {
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

	wallet, err := h.walletService.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "wallet found", dto.NewWalletResponseDTO(wallet))
}

// Update godoc
//
//	@Summary	Update a wallet
//	@Tags		wallets
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		string						true	"Wallet ID (UUID)"
//	@Param		request	body		dto.UpdateWalletRequestDTO	true	"Fields to update"
//	@Success	200		{object}	response.Success{data=dto.WalletResponseDTO}
//	@Failure	400		{object}	response.Error
//	@Failure	404		{object}	response.Error
//	@Failure	409		{object}	response.Error
//	@Router		/wallets/{id} [patch]
func (h *WalletHandler) Update(c *gin.Context) {
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

	var req dto.UpdateWalletRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	wallet, err := h.walletService.Update(c.Request.Context(), userID, id, req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "wallet updated", dto.NewWalletResponseDTO(wallet))
}

// Delete godoc
//
//	@Summary	Delete a wallet
//	@Tags		wallets
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Wallet ID (UUID)"
//	@Success	200	{object}	response.Success
//	@Failure	404	{object}	response.Error
//	@Router		/wallets/{id} [delete]
func (h *WalletHandler) Delete(c *gin.Context) {
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

	if err := h.walletService.Delete(c.Request.Context(), userID, id); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "wallet deleted", nil)
}

// Overview godoc
//
//	@Summary		Balance summary across the logged-in user's wallets
//	@Description	Only base-currency wallets reach total_held; card debt and other currencies are reported separately, as no exchange rate is available.
//	@Tags			wallets
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Success{data=dto.WalletOverviewResponseDTO}
//	@Failure		401	{object}	response.Error
//	@Router			/wallets/overview [get]
func (h *WalletHandler) Overview(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrAuthRequired)
		return
	}

	overview, err := h.walletService.Overview(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "wallet overview", dto.NewWalletOverviewResponseDTO(overview))
}
