package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/handler"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/mocks"
	"github.com/stevensuki/ledgerline-backend/pkg/validator"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	validator.RegisterGinValidator()
	m.Run()
}

// setup builds a minimal router holding the single route under test.
func setup(method, path string, h gin.HandlerFunc) *gin.Engine {
	engine := gin.New()
	engine.Handle(method, path, h)
	return engine
}

func TestUserHandler_GetByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	user := &domain.User{ID: id, Email: "budi@example.com", FullName: "Budi", RoleID: domain.RoleIDUser}

	t.Run("200 when the user is found", func(t *testing.T) {
		t.Parallel()

		svc := new(mocks.UserService)
		svc.On("GetByID", mock.Anything, id).Return(user, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users/"+id.String(), nil)
		setup(http.MethodGet, "/users/:id", handler.NewUserHandler(svc).GetByID).ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var body struct {
			Success bool `json:"success"`
			Data    struct {
				Email string `json:"email"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.True(t, body.Success)
		assert.Equal(t, "budi@example.com", body.Data.Email)
	})

	t.Run("404 when the user does not exist", func(t *testing.T) {
		t.Parallel()

		svc := new(mocks.UserService)
		svc.On("GetByID", mock.Anything, id).Return(nil, domain.ErrNotFound)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users/"+id.String(), nil)
		setup(http.MethodGet, "/users/:id", handler.NewUserHandler(svc).GetByID).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("400 when the id is not a UUID", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users/not-a-uuid", nil)
		setup(http.MethodGet, "/users/:id", handler.NewUserHandler(new(mocks.UserService)).GetByID).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestUserHandler_Create_Validation(t *testing.T) {
	t.Parallel()

	body := `{"email":"not-an-email","full_name":"ab","password":"123"}`

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	setup(http.MethodPost, "/users", handler.NewUserHandler(new(mocks.UserService)).Create).ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var resp struct {
		Code   string `json:"code"`
		Errors []struct {
			Field string `json:"field"`
		} `json:"errors"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Len(t, resp.Errors, 3)
}
