# Guide: Adding a New API

A real example: the **`categories`** module is already complete in this repo. Follow the 7 steps below, just swap the word `Category` for your own entity (say `Transaction`).

The order is always **inside out**: domain → migration → repository → service → dto → handler → route.

## Naming convention

The suffix tells you which layer a type belongs to, so the same entity never looks ambiguous:

| Layer | Package | Example | Rule |
|---|---|---|---|
| Domain | `internal/domain` | `Role`, `CreateRoleInput` | plain name, no suffix |
| Persistence | `internal/repository/postgres/model` | `RoleModel` | suffix `Model` |
| Transport | `internal/delivery/http/dto` | `CreateRoleRequestDTO`, `RoleResponseDTO` | suffix `DTO` |

Mappers follow the same rule: `RoleFromDomain()` / `(RoleModel).ToDomain()` in the model package, `(CreateRoleRequestDTO).ToInput()` / `NewRoleResponseDTO()` in the dto package.

---

## 1. Migration — create the table

```bash
make migrate-create name=create_categories_table
```

Fill in `db/migrations/000002_create_categories_table.up.sql`:

```sql
CREATE TABLE IF NOT EXISTS categories (
    id         UUID PRIMARY KEY,
    user_id    UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name       VARCHAR(100) NOT NULL,
    type       VARCHAR(20)  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT categories_type_check CHECK (type IN ('income', 'expense'))
);
```

Do not forget the `.down.sql` file (`DROP TABLE ...`). Then run `make migrate-up`.

## 2. Domain — entity + interfaces

File: [internal/domain/category.go](internal/domain/category.go)

It contains 4 things:

```go
type Category struct { ... }                    // 1. entity (no json/db tags)
type CategoryFilter struct { ... }              // 2. filter for the list
type CategoryRepository interface { ... }       // 3. port to the database
type CategoryService interface { ... }          // 4. business logic contract
```

Rule: this file **must not import gin/GORM**. Stdlib + uuid is enough.

## 3. Repository — the GORM implementation

File: [internal/repository/postgres/category_repository.go](internal/repository/postgres/category_repository.go)

```go
type categoryRepository struct{ db *gorm.DB }   // the struct stays private

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository { // return the interface
	return &categoryRepository{db: db}
}
```

Things you must watch out for:

- First build a GORM-tagged persistence model in `model/<entity>_model.go` (`CategoryModel`) plus its `ToDomain()` / `CategoryFromDomain()` mappers.
- Wrap every error with `wrapErr()`: `gorm.ErrRecordNotFound` → `domain.ErrNotFound`, `gorm.ErrDuplicatedKey` → `domain.ErrConflict`.
- Always use `WithContext(ctx)` so request timeouts & cancellation also cancel the query.
- For later reporting/aggregation (SUM, GROUP BY), use `db.Raw(...).Scan(...)` — the ORM does not help there.

## 4. Service — business logic

File: [internal/service/category_service.go](internal/service/category_service.go)

```go
type categoryService struct{ repo domain.CategoryRepository }  // depends on the INTERFACE

func NewCategoryService(repo domain.CategoryRepository) domain.CategoryService {
	return &categoryService{repo: repo}
}
```

Every business rule lives here: value validation, duplicate checks, input normalisation, ownership checks. This layer knows nothing about HTTP — on failure, return a domain error:

```go
return nil, domain.Conflict(domain.CodeCategoryNameTaken, "a category with that name already exists").WithField("name")
```

The constructor picks the status: `Conflict` 409, `NotFound` 404, `InvalidInput` 400, `Forbidden` 403, `Unauthorized` 401, `RateLimited` 429. Add the code to [internal/domain/error_codes.go](internal/domain/error_codes.go) and a row to [ERROR_CODES.md](ERROR_CODES.md); the frontend switches on it, so make it specific.

Modifiers all return a copy: `.WithField("name")` highlights an input, `.WithCause(err)` keeps the technical detail for the log only, `.WithRetryAfter(d)` fills the `Retry-After` header.

When forwarding a repository error, never flatten it into something coarser. That is how a database outage turns into a 401:

```go
if err != nil {
	if errors.Is(err, domain.ErrNotFound) { // an expected outcome
		return nil, domain.ErrInvalidCredentials
	}
	return nil, err // anything else is a real failure, let it be a 500
}
```

## 5. DTO — HTTP request/response

File: [internal/delivery/http/dto/category_dto.go](internal/delivery/http/dto/category_dto.go)

```go
type CreateCategoryRequestDTO struct {
	Name string `json:"name" binding:"required,min=2,max=100" example:"Salary"`
	Type string `json:"type" binding:"required,oneof=income expense" example:"income"`
}

func (r CreateCategoryRequestDTO) ToInput() domain.CreateCategoryInput { ... }   // mapper to the domain
func NewCategoryResponseDTO(c *domain.Category) CategoryResponseDTO { ... }     // mapper to the response
```

The `binding` tag holds the validation rules. The `example` tag is the sample shown in Swagger. For PATCH, use pointers (`*string`) so fields that were not sent do not get updated.

### Sorting on list endpoints

The `sort` query param uses comma + minus-prefix syntax: `?sort=-created_at,name` means newest date first, with name as the tie-breaker. Define the whitelist in the DTO file:

```go
// categorySort: only columns in this map may enter ORDER BY.
var categorySort = pagination.Sortable{
	Allowed:    pagination.Whitelist{"name": "name", "type": "type", "created_at": "created_at"},
	Default:    "-created_at",
	TieBreaker: "id",
}

func (q ListCategoryQuery) OrderBy() (string, error) { return categorySort.OrderBy(q.Sort) }
```

Three things that matter:

- **The whitelist is mandatory.** `.Order()` in GORM is not parameterised the way `.Where()` is, so a user string reaching it means SQL injection. Columns outside the map are rejected with `pagination.ErrInvalidSort` (mapped to 422 in `handleError`).
- **`TieBreaker` must be a unique column.** If the sort uses a column whose values can tie, Postgres does not guarantee the order of tied rows — so rows get duplicated or skipped between pages as `OFFSET` shifts.
- **Map key = public name, value = DB column name.** Columns can be renamed without breaking the API contract.

The handler just translates it before calling the service:

```go
orderBy, err := query.OrderBy()
if err != nil {
	handleError(c, err)
	return
}
```

## 6. Handler — HTTP adapter + Swagger annotations

File: [internal/delivery/http/handler/category_handler.go](internal/delivery/http/handler/category_handler.go)

Every handler always follows the same 4 working lines:

```go
// Create godoc
//
//	@Summary	Create a category
//	@Tags		categories
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		dto.CreateCategoryRequestDTO	true	"Category data"
//	@Success	201		{object}	response.Success{data=dto.CategoryResponse}
//	@Failure	409		{object}	response.Error
//	@Router		/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)              // 1. take the user from the token
	if !ok { handleError(c, domain.ErrUnauthorized); return }

	var req dto.CreateCategoryRequestDTO                  // 2. bind + validate
	if err := c.ShouldBindJSON(&req); err != nil { handleBindError(c, err); return }

	category, err := h.categoryService.Create(c.Request.Context(), userID, req.ToInput())  // 3. call the service
	if err != nil { handleError(c, err); return }

	response.OK(c, http.StatusCreated, "category created", dto.NewCategoryResponseDTO(category)) // 4. respond
}
```

Never write `if errors.Is(...)` in a handler, and never compose a message there. `handleError(c, err)` is enough: status, code, field detail and `request_id` are all decided in [internal/delivery/http/apierr](internal/delivery/http/apierr/apierr.go), which the middleware uses too.

## 7. Route + wiring

Register the route in [internal/delivery/http/router/router.go](internal/delivery/http/router/router.go):

```go
func registerCategoryRoutes(rg *gin.RouterGroup, deps Dependencies) {
	categories := rg.Group("/categories", middleware.Authenticate(deps.TokenManager))
	{
		categories.POST("", deps.Category.Create)
		categories.GET("", deps.Category.List)
		categories.GET("/:id", deps.Category.GetByID)
		categories.PATCH("/:id", deps.Category.Update)
		categories.DELETE("/:id", deps.Category.Delete)
	}
}
```

Then call `registerCategoryRoutes(v1, deps)` inside `New()`, add a `Category *handler.CategoryHandler` field to the `Dependencies` struct, and wire it up in [cmd/api/main.go](cmd/api/main.go):

```go
categoryRepo := postgres.NewCategoryRepository(db)
categoryService := service.NewCategoryService(categoryRepo)
// ...
Category: handler.NewCategoryHandler(categoryService),
```

For admin-only endpoints, add the middleware: `middleware.RequireRoles(domain.RoleIDAdmin)`.

## 8. Tests + Swagger

Add a mock repository under `internal/mocks/`, then test the service the way [category_service_test.go](internal/service/category_service_test.go) does:

```go
repo := new(mocks.CategoryRepository)
repo.On("ExistsByName", mock.Anything, userID, "Salary").Return(false, nil)
repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Category")).Return(nil)

category, err := service.NewCategoryService(repo).Create(context.Background(), userID, input)

require.NoError(t, err)
repo.AssertExpectations(t)
```

Finally:

```bash
make swag   # regenerate the documentation
make test   # make sure it is green
make lint
```

---

## Quick checklist

- [ ] Migration up + down
- [ ] `internal/domain/xxx.go` — entity + repository interface + service interface
- [ ] `internal/repository/postgres/model/xxx_model.go` — `XxxModel` + mappers to/from the domain
- [ ] `internal/repository/postgres/xxx_repository.go` (return the interface; errors through `xxxErrors.wrap()`, registered in `postgres/errors.go`)
- [ ] `internal/service/xxx_service.go` (business logic; errors via `domain.Conflict/NotFound/InvalidInput(...)` with a code from `error_codes.go`)
- [ ] New codes added to `ERROR_CODES.md`
- [ ] `internal/delivery/http/dto/xxx_dto.go` — request/response + mappers + sort whitelist
- [ ] `internal/delivery/http/handler/xxx_handler.go` — swagger annotations + `handleError`
- [ ] Route in `router.go` + wiring in `main.go`
- [ ] Mock in `internal/mocks/` + service unit test
- [ ] `make swag && make test && make lint`
