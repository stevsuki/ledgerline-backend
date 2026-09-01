package dto

import (
	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"budi@example.com"`
	FullName string `json:"full_name" binding:"required,min=3,max=100" example:"Budi Santoso"`
	Password string `json:"password" binding:"required,min=8,max=72" example:"Rahasia123!"`
}

func (r RegisterRequest) ToInput() domain.RegisterInput {
	return domain.RegisterInput{Email: r.Email, FullName: r.FullName, Password: r.Password}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"budi@example.com"`
	Password string `json:"password" binding:"required,min=8,max=72" example:"Rahasia123!"`
}

func (r LoginRequest) ToInput() domain.LoginInput {
	return domain.LoginInput{Email: r.Email, Password: r.Password}
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"budi@example.com"`
}

func (r ForgotPasswordRequest) ToInput() domain.ForgotPasswordInput {
	return domain.ForgotPasswordInput{Email: r.Email}
}

type VerifyOTPResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"budi@example.com"`
	OTP   string `json:"otp" binding:"required" example:"123456"`
}

func (r VerifyOTPResetPasswordRequest) ToInput() domain.VerifyOTPResetPasswordInput {
	return domain.VerifyOTPResetPasswordInput{
		Email: r.Email,
		OTP:   r.OTP,
	}
}

// ResetTokenResponse: the grant returned once an OTP is verified.
type ResetTokenResponse struct {
	ResetToken string `json:"reset_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	ExpiresIn  int64  `json:"expires_in" example:"600"`
}

func NewResetTokenResponse(t *domain.ResetToken) ResetTokenResponse {
	return ResetTokenResponse{ResetToken: t.Token, ExpiresIn: t.ExpiresIn}
}

type ResetPasswordRequest struct {
	ResetToken         string `json:"reset_token" binding:"required" example:"abc123"`
	NewPassword        string `json:"new_password" binding:"required,min=8,max=72" example:"RahasiaBaru123!"`
	ConfirmNewPassword string `json:"confirm_new_password" binding:"required,eqfield=NewPassword" example:"RahasiaBaru123!"`
}

func (r ResetPasswordRequest) ToInput() domain.ResetPasswordInput {
	return domain.ResetPasswordInput{
		ResetToken:         r.ResetToken,
		NewPassword:        r.NewPassword,
		ConfirmNewPassword: r.ConfirmNewPassword,
	}
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIs..."`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int64  `json:"expires_in" example:"900"`
}

func NewTokenResponse(t *domain.TokenPair) TokenResponse {
	return TokenResponse{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    t.ExpiresIn,
	}
}

// ProfileResponse: /auth/me -- the account plus the sidebar it may see.
type ProfileResponse struct {
	User  UserResponse   `json:"user"`
	Menus []MenuResponse `json:"menus"`
}

// MenuResponse: a group carries children, a page carries access.
type MenuResponse struct {
	ID       uuid.UUID           `json:"id"`
	Code     string              `json:"code" example:"transactions"`
	Name     string              `json:"name" example:"Transactions"`
	Path     string              `json:"path,omitempty"`
	Icon     string              `json:"icon,omitempty" example:"swap"`
	Access   *MenuAccessResponse `json:"access,omitempty"`
	Children []MenuResponse      `json:"children,omitempty"`
}

type MenuAccessResponse struct {
	CanCreate  bool `json:"can_create" example:"true"`
	CanRead    bool `json:"can_read" example:"true"`
	CanUpdate  bool `json:"can_update" example:"true"`
	CanDelete  bool `json:"can_delete" example:"true"`
	CanApprove bool `json:"can_approve" example:"false"`
}

func NewProfileResponse(p *domain.Profile) ProfileResponse {
	return ProfileResponse{
		User:  NewUserResponse(&p.User),
		Menus: NewMenuResponses(p.Menus),
	}
}

func NewMenuResponses(menus []domain.Menu) []MenuResponse {
	out := make([]MenuResponse, 0, len(menus))
	for i := range menus {
		m := menus[i]
		item := MenuResponse{
			ID:   m.ID,
			Code: m.Code,
			Name: m.Name,
			Path: m.Path,
			Icon: m.Icon,
		}
		if len(m.Children) > 0 {
			item.Children = NewMenuResponses(m.Children)
		} else {
			item.Access = &MenuAccessResponse{
				CanCreate:  m.Access.CanCreate,
				CanRead:    m.Access.CanRead,
				CanUpdate:  m.Access.CanUpdate,
				CanDelete:  m.Access.CanDelete,
				CanApprove: m.Access.CanApprove,
			}
		}
		out = append(out, item)
	}
	return out
}
