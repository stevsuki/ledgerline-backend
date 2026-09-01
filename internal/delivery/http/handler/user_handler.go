package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/dto"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/pagination"
	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

type UserHandler struct {
	userService domain.UserService
}

func NewUserHandler(userService domain.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Create godoc
//
//	@Summary	Create a new user (admin)
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		dto.CreateUserRequest	true	"User data"
//	@Success	201		{object}	response.Success{data=dto.UserResponse}
//	@Failure	403		{object}	response.Error
//	@Failure	409		{object}	response.Error
//	@Router		/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	user, err := h.userService.Create(c.Request.Context(), req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusCreated, "user created", dto.NewUserResponse(user))
}

// List godoc
//
//	@Summary	List users
//	@Tags		users
//	@Produce	json
//	@Security	BearerAuth
//	@Param		search		query		string	false	"Search by email or name"
//	@Param		sort		query		string	false	"Order: email, full_name, role, created_at, updated_at. Prefix - for desc"	default(-created_at)
//	@Param		page		query		int		false	"Page"			default(1)
//	@Param		per_page	query		int		false	"Items per page"	default(10)
//	@Success	200			{object}	response.Success{data=[]dto.UserResponse}
//	@Router		/users [get]
func (h *UserHandler) List(c *gin.Context) {
	var query dto.ListUserQuery
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
	users, total, err := h.userService.List(c.Request.Context(), domain.UserFilter{
		Search:  query.Search,
		Limit:   params.Limit(),
		Offset:  params.Offset(),
		OrderBy: orderBy,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	response.Paginated(c, http.StatusOK, "success", dto.NewUserResponses(users), response.Meta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalItems: total,
		TotalPages: pagination.TotalPages(total, params.PerPage),
	})
}

// GetByID godoc
//
//	@Summary	User detail
//	@Tags		users
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"User ID (UUID)"
//	@Success	200	{object}	response.Success{data=dto.UserResponse}
//	@Failure	404	{object}	response.Error
//	@Router		/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "success", dto.NewUserResponse(user))
}

// Update godoc
//
//	@Summary	Update user data
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		string					true	"User ID (UUID)"
//	@Param		request	body		dto.UpdateUserRequest	true	"Fields to update"
//	@Success	200		{object}	response.Success{data=dto.UserResponse}
//	@Failure	404		{object}	response.Error
//	@Router		/users/{id} [patch]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	user, err := h.userService.Update(c.Request.Context(), id, req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "user updated", dto.NewUserResponse(user))
}

// Delete godoc
//
//	@Summary	Delete a user
//	@Tags		users
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"User ID (UUID)"
//	@Success	200	{object}	response.Success
//	@Failure	404	{object}	response.Error
//	@Router		/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}

	if err := h.userService.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "user deleted", nil)
}

// parseUUIDParam writes its own error response when the param is invalid.
func parseUUIDParam(c *gin.Context, key string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(key))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PARAM", key+" must be a valid UUID", nil)
		return uuid.Nil, err
	}
	return id, nil
}
