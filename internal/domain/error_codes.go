package domain

// Error codes are the public contract: the frontend switches on Code, never on Message.
// A code never changes meaning once released; add a new one instead.
const (
	// Generic — usable by any resource.
	CodeInternal         = "INTERNAL_ERROR"
	CodeValidation       = "VALIDATION_ERROR"
	CodeBadRequest       = "BAD_REQUEST"
	CodeInvalidParam     = "INVALID_PARAM"
	CodeInvalidInput     = "INVALID_INPUT"
	CodeNotFound         = "NOT_FOUND"
	CodeConflict         = "CONFLICT"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeTooManyRequests  = "TOO_MANY_REQUESTS"
	CodeTimeout          = "REQUEST_TIMEOUT"
	CodeRouteNotFound    = "ROUTE_NOT_FOUND"
	CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
	CodeDBUnavailable    = "DB_UNAVAILABLE"

	// Auth.
	CodeInvalidCredentials = "AUTH_INVALID_CREDENTIALS" //nolint:gosec // an error code, not a secret
	CodeAccountLocked      = "AUTH_ACCOUNT_LOCKED"
	CodeTokenMissing       = "AUTH_TOKEN_MISSING" //nolint:gosec // an error code, not a secret
	CodeTokenInvalid       = "AUTH_TOKEN_INVALID" //nolint:gosec // an error code, not a secret
	CodeTokenExpired       = "AUTH_TOKEN_EXPIRED" //nolint:gosec // an error code, not a secret
	CodeInvalidOTP         = "AUTH_INVALID_OTP"
	CodeOTPMaxAttempts     = "AUTH_OTP_MAX_ATTEMPTS"
	CodeResetTooSoon       = "AUTH_RESET_REQUESTED_TOO_SOON"

	// User.
	CodeUserNotFound    = "USER_NOT_FOUND"
	CodeUserEmailTaken  = "USER_EMAIL_TAKEN"
	CodeUserInvalidRole = "USER_INVALID_ROLE"
	CodeUserInvalidData = "USER_INVALID_DATA"

	// Category.
	CodeCategoryNotFound    = "CATEGORY_NOT_FOUND"
	CodeCategoryNameTaken   = "CATEGORY_NAME_TAKEN"
	CodeCategoryInvalidType = "CATEGORY_INVALID_TYPE"
	CodeCategoryInvalidData = "CATEGORY_INVALID_DATA"

	// Role.
	CodeRoleNotFound        = "ROLE_NOT_FOUND"
	CodeRoleNameTaken       = "ROLE_NAME_TAKEN"
	CodeRoleSystemImmutable = "ROLE_SYSTEM_IMMUTABLE"
	CodeRoleInvalidMenu     = "ROLE_INVALID_MENU"
	CodeRoleInvalidData     = "ROLE_INVALID_DATA"

	// Wallet.
	CodeWalletNotFound        = "WALLET_NOT_FOUND"
	CodeWalletNameTaken       = "WALLET_NAME_TAKEN"
	CodeWalletInvalidData     = "WALLET_INVALID_DATA"
	CodeWalletInvalidType     = "WALLET_INVALID_TYPE"
	CodeWalletInvalidCurrency = "WALLET_INVALID_CURRENCY"
	CodeWalletInvalidCard     = "WALLET_INVALID_CARD"

	// Master category.
	CodeMasterCategoryNotFound  = "MASTER_CATEGORY_NOT_FOUND"
	CodeMasterCategoryNameTaken = "MASTER_CATEGORY_NAME_TAKEN"

	// Menu.
	CodeMenuNotFound = "MENU_NOT_FOUND"

	// Audit log.
	CodeAuditLogNotFound = "AUDIT_LOG_NOT_FOUND"
)
