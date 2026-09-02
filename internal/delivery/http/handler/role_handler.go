package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/dto"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/pagination"
	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

type RoleHandler struct {
	roleService domain.RoleService
}

func NewRoleHandler(roleService domain.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// List godoc
//
//	@Summary	List roles
//	@Tags		roles
//	@Produce	json
//	@Security	BearerAuth
//	@Param		search		query		string	false	"Search by name"
//	@Param		sort		query		string	false	"Order: name, is_system, user_count, created_at, updated_at. Prefix - for desc"	default(-is_system,name)
//	@Param		page		query		int		false	"Page"			default(1)
//	@Param		per_page	query		int		false	"Items per page"	default(10)
//	@Success	200			{object}	response.Success{data=[]dto.RoleResponseDTO}
//	@Router		/roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	var query dto.ListRoleQueryDTO
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
	roles, total, err := h.roleService.List(c.Request.Context(), domain.RoleFilter{
		Search:  query.Search,
		Limit:   params.Limit(),
		Offset:  params.Offset(),
		OrderBy: orderBy,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	response.Paginated(c, http.StatusOK, "roles", dto.NewRoleResponseDTOs(roles), response.Meta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalItems: total,
		TotalPages: pagination.TotalPages(total, params.PerPage),
	})
}

// GetByID godoc
//
//	@Summary	Role detail
//	@Tags		roles
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Role ID (UUID)"
//	@Success	200	{object}	response.Success{data=dto.RoleResponseDTO}
//	@Failure	404	{object}	response.Error
//	@Router		/roles/{id} [get]
func (h *RoleHandler) GetByID(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		handleError(c, err)
		return
	}

	role, err := h.roleService.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "role", dto.NewRoleResponseDTO(role))
}

// Create godoc
//
//	@Summary	Create a new role
//	@Tags		roles
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		dto.CreateRoleRequestDTO	true	"Role data"
//	@Success	201		{object}	response.Success{data=dto.RoleResponseDTO}
//	@Failure	400		{object}	response.Error
//	@Failure	422		{object}	response.Error
//	@Failure	409		{object}	response.Error
//	@Router		/roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.CreateRoleRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	role, err := h.roleService.Create(c.Request.Context(), req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusCreated, "role created", dto.NewRoleResponseDTO(role))
}

// Update godoc
//
//	@Summary	Update role data
//	@Tags		roles
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		string					true	"Role ID (UUID)"
//	@Param		request	body		dto.UpdateRoleRequestDTO	true	"Fields to update"
//	@Success	200		{object}	response.Success{data=dto.RoleResponseDTO}
//	@Failure	400		{object}	response.Error
//	@Failure	403		{object}	response.Error
//	@Failure	404		{object}	response.Error
//	@Failure	409		{object}	response.Error
//	@Failure	422		{object}	response.Error
//	@Router		/roles/{id} [patch]
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		handleError(c, err)
		return
	}

	var req dto.UpdateRoleRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	role, err := h.roleService.Update(c.Request.Context(), id, req.ToInput())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "role updated", dto.NewRoleResponseDTO(role))
}

// Delete godoc
//
//	@Summary	Delete a role
//	@Tags		roles
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Role ID (UUID)"
//	@Success	200	{object}	response.Success
//	@Failure	403	{object}	response.Error
//	@Failure	404	{object}	response.Error
//	@Router		/roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		handleError(c, err)
		return
	}

	if err := h.roleService.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "role deleted", nil)
}
