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
//	@Param			request	body		dto.RegisterRequest	true	"Registration data"
//	@Success		201		{object}	response.Success{data=dto.UserResponse}
//	@Failure		409		{object}	response.Error
//	@Failure		422		{object}	response.Error
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusCreated, "registration successful", dto.NewUserResponse(user))
}

// Login godoc
//
//	@Summary		Login
//	@Description	Exchanges email and password for an access + refresh token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequest	true	"Login credentials"
//	@Success		200		{object}	response.Success{data=dto.TokenResponse}
//	@Failure		401		{object}	response.Error
//	@Failure		422		{object}	response.Error
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
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
	response.OK(c, http.StatusOK, "login successful", dto.NewTokenResponse(tokens))
}

// Refresh godoc
//
//	@Summary		Refresh token
//	@Description	Issues a new access token using the refresh token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RefreshTokenRequest	true	"Refresh token"
//	@Success		200		{object}	response.Success{data=dto.TokenResponse}
//	@Failure		401		{object}	response.Error
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	tokens, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "token refreshed", dto.NewTokenResponse(tokens))
}

// Me godoc
//
//	@Summary	Profile of the currently logged-in user
//	@Tags		auth
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	response.Success{data=dto.ProfileResponse}
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
	response.OK(c, http.StatusOK, "success", dto.NewProfileResponse(profile))
}

// ForgotPassword godoc
//
//	@Summary		Request a password reset OTP
//	@Description	Sends an OTP code to the email address if it is registered
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.ForgotPasswordRequest	true	"Destination email"
//	@Success		200		{object}	response.Success
//	@Failure		422		{object}	response.Error
//	@Failure		429		{object}	response.Error
//	@Router			/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
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
//	@Param			request	body		dto.VerifyOTPResetPasswordRequest	true	"Email + OTP code"
//	@Success		200		{object}	response.Success{data=dto.ResetTokenResponse}
//	@Failure		400		{object}	response.Error
//	@Failure		422		{object}	response.Error
//	@Failure		429		{object}	response.Error
//	@Router			/auth/verify-otp [post]
func (h *AuthHandler) VerifyOTPResetPassword(c *gin.Context) {
	var req dto.VerifyOTPResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	resetToken, err := h.authService.VerifyOTPResetPassword(c.Request.Context(), req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "OTP verified", dto.NewResetTokenResponse(resetToken))
}

// ResetPassword godoc
//
//	@Summary		Change password with a reset token
//	@Description	Changes the password using the reset token from the OTP verification step
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.ResetPasswordRequest	true	"Reset token + new password"
//	@Success		200		{object}	response.Success
//	@Failure		400		{object}	response.Error
//	@Failure		422		{object}	response.Error
//	@Router			/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
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
