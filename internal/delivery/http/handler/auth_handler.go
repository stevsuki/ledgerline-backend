package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/dto"
	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/middleware"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

type AuthHandler struct {
	authService domain.AuthService
}

func NewAuthHandler(authService domain.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Creates a new account with the default "user" role
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterRequestDTO	true	"Registration data"
//	@Success		201		{object}	response.Success{data=dto.UserResponseDTO}
//	@Failure		409		{object}	response.Error
//	@Failure		422		{object}	response.Error
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusCreated, "registration successful", dto.NewUserResponseDTO(user))
}

// Login godoc
//
//	@Summary		Login
//	@Description	Exchanges email and password for an access + refresh token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequestDTO	true	"Login credentials"
//	@Success		200		{object}	response.Success{data=dto.TokenResponseDTO}
//	@Failure		401		{object}	response.Error
//	@Failure		422		{object}	response.Error
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	meta := domain.NewRequestMeta(c.ClientIP(), c.GetHeader("User-Agent"))

	tokens, err := h.authService.Login(c.Request.Context(), req.ToInput(), meta)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "login successful", dto.NewTokenResponseDTO(tokens))
}

// Refresh godoc
//
//	@Summary		Refresh token
//	@Description	Issues a new access token using the refresh token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RefreshTokenRequestDTO	true	"Refresh token"
//	@Success		200		{object}	response.Success{data=dto.TokenResponseDTO}
//	@Failure		401		{object}	response.Error
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	tokens, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "token refreshed", dto.NewTokenResponseDTO(tokens))
}

// Me godoc
//
//	@Summary	Profile of the currently logged-in user
//	@Tags		auth
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	response.Success{data=dto.ProfileResponseDTO}
//	@Failure	401	{object}	response.Error
//	@Router		/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handleError(c, domain.ErrUnauthorized)
		return
	}

	profile, err := h.authService.Me(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "success", dto.NewProfileResponseDTO(profile))
}

// ForgotPassword godoc
//
//	@Summary		Request a password reset OTP
//	@Description	Sends an OTP code to the email address if it is registered
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.ForgotPasswordRequestDTO	true	"Destination email"
//	@Success		200		{object}	response.Success
//	@Failure		422		{object}	response.Error
//	@Failure		429		{object}	response.Error
//	@Router			/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	err := h.authService.ForgotPassword(c.Request.Context(), req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "password reset email sent", nil)
}

// VerifyOTPResetPassword godoc
//
//	@Summary		Verify the password reset OTP
//	@Description	Checks the OTP code before the user may change their password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.VerifyOTPResetPasswordRequestDTO	true	"Email + OTP code"
//	@Success		200		{object}	response.Success{data=dto.ResetTokenResponseDTO}
//	@Failure		400		{object}	response.Error
//	@Failure		422		{object}	response.Error
//	@Failure		429		{object}	response.Error
//	@Router			/auth/verify-otp [post]
func (h *AuthHandler) VerifyOTPResetPassword(c *gin.Context) {
	var req dto.VerifyOTPResetPasswordRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	resetToken, err := h.authService.VerifyOTPResetPassword(c.Request.Context(), req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "OTP verified", dto.NewResetTokenResponseDTO(resetToken))
}

// ResetPassword godoc
//
//	@Summary		Change password with a reset token
//	@Description	Changes the password using the reset token from the OTP verification step
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.ResetPasswordRequestDTO	true	"Reset token + new password"
//	@Success		200		{object}	response.Success
//	@Failure		400		{object}	response.Error
//	@Failure		422		{object}	response.Error
//	@Router			/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	err := h.authService.ResetPassword(c.Request.Context(), req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "password reset successfully", nil)
}
