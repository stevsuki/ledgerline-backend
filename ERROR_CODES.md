# Error codes

Every failed request answers with the same envelope:

```json
{
  "success": false,
  "message": "email is already registered",
  "code": "USER_EMAIL_TAKEN",
  "errors": [{ "field": "email", "message": "email is already registered" }],
  "request_id": "9b2c1f2e-6f0a-4a1e-9c7d-2f8b0a1c3d4e"
}
```

- **`code` is the contract.** Switch on it. It never changes meaning once released; a new
  situation gets a new code.
- **`message` is for humans**, not for logic. Never parse it, never branch on it. Translate by
  `code` on the client instead.
- **`errors`** is present when the failure belongs to specific request fields.
- **`request_id`** matches the `X-Request-ID` response header and the server log line.

The source of truth is [internal/domain/error_codes.go](internal/domain/error_codes.go).

## Generic

| Code | Status | When |
|---|---|---|
| `VALIDATION_ERROR` | 422 | The body or query failed validation; `errors` lists the fields |
| `BAD_REQUEST` | 400 | The body could not be parsed at all (malformed JSON) |
| `INVALID_PARAM` | 400 | A path parameter has the wrong shape, e.g. a non-UUID id |
| `INVALID_INPUT` | 400 | A rejected value with no more specific code |
| `NOT_FOUND` | 404 | A resource is missing and has no more specific code |
| `CONFLICT` | 409 | A duplicate with no more specific code |
| `UNAUTHORIZED` | 401 | Authentication is required |
| `FORBIDDEN` | 403 | Authenticated, but not allowed |
| `TOO_MANY_REQUESTS` | 429 | Rate limited; see the `Retry-After` header |
| `REQUEST_TIMEOUT` | 504 | The request exceeded `HTTP_REQUEST_TIMEOUT` |
| `ROUTE_NOT_FOUND` | 404 | No such endpoint |
| `METHOD_NOT_ALLOWED` | 405 | The endpoint exists, the method does not |
| `DB_UNAVAILABLE` | 503 | Readiness probe: the database is unreachable |
| `INTERNAL_ERROR` | 500 | Unexpected; details are in the log under `request_id` |

## Auth

| Code | Status | When |
|---|---|---|
| `AUTH_INVALID_CREDENTIALS` | 401 | Wrong email or password. Deliberately identical for both |
| `AUTH_ACCOUNT_LOCKED` | 429 | Too many failed sign-ins; `Retry-After` holds the remaining lock |
| `AUTH_TOKEN_MISSING` | 401 | No `Authorization: Bearer <token>` header |
| `AUTH_TOKEN_INVALID` | 401 | The token is malformed, revoked, or its user is gone |
| `AUTH_TOKEN_EXPIRED` | 401 | The token expired — the client should refresh |
| `AUTH_INVALID_OTP` | 400 | Wrong, expired, or already used reset code |
| `AUTH_OTP_MAX_ATTEMPTS` | 429 | Too many wrong codes; the client must request a new one |
| `AUTH_RESET_REQUESTED_TOO_SOON` | 429 | Reset requested inside the cooldown; `Retry-After` holds the wait |

`POST /auth/forgot-password` always answers 200, whether or not the address exists.

## User

| Code | Status | When |
|---|---|---|
| `USER_NOT_FOUND` | 404 | No user with that id |
| `USER_EMAIL_TAKEN` | 409 | The email is already registered (`field: email`) |
| `USER_INVALID_ROLE` | 400 | `role_id` is empty or points at no role (`field: role_id`) |
| `USER_INVALID_DATA` | 400 | Another user field was rejected |

## Category

| Code | Status | When |
|---|---|---|
| `CATEGORY_NOT_FOUND` | 404 | No category with that id for this user |
| `CATEGORY_NAME_TAKEN` | 409 | The user already has a category with that name (`field: name`) |
| `CATEGORY_INVALID_TYPE` | 400 | `type` is neither `income` nor `expense` (`field: type`) |
| `CATEGORY_INVALID_DATA` | 400 | Another category field was rejected |

## Role

| Code | Status | When |
|---|---|---|
| `ROLE_NOT_FOUND` | 404 | No role with that id |
| `ROLE_NAME_TAKEN` | 409 | A role with that name exists (`field: name`) |
| `ROLE_SYSTEM_IMMUTABLE` | 403 | A built-in role cannot be renamed or deleted |
| `ROLE_INVALID_MENU` | 400 | A menu id in `permissions` is empty, duplicated, or unknown |
| `ROLE_INVALID_DATA` | 400 | Another role field was rejected |

## Menu & audit log

| Code | Status | When |
|---|---|---|
| `MENU_NOT_FOUND` | 404 | No menu with that id |
| `AUDIT_LOG_NOT_FOUND` | 404 | No audit log entry with that id |

## Adding a code

1. Add the constant to [internal/domain/error_codes.go](internal/domain/error_codes.go), grouped by resource.
2. Return it from the service: `domain.Conflict(domain.CodeXxx, "...").WithField("...")`.
3. Add the row to the table above.

Never reuse a code for a different meaning — the frontend has already branched on it.
