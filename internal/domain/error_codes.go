package domain

// Error codes are the public contract: the frontend switches on Code.
const (
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

	CodeInvalidCredentials = "AUTH_INVALID_CREDENTIALS" //nolint:gosec // an error code, not a secret
	CodeAccountLocked      = "AUTH_ACCOUNT_LOCKED"
	CodeTokenMissing       = "AUTH_TOKEN_MISSING" //nolint:gosec // an error code, not a secret
	CodeTokenInvalid       = "AUTH_TOKEN_INVALID" //nolint:gosec // an error code, not a secret
	CodeTokenExpired       = "AUTH_TOKEN_EXPIRED" //nolint:gosec // an error code, not a secret
	CodeInvalidOTP         = "AUTH_INVALID_OTP"
	CodeOTPMaxAttempts     = "AUTH_OTP_MAX_ATTEMPTS"
	CodeResetTooSoon       = "AUTH_RESET_REQUESTED_TOO_SOON"

	CodeUserNotFound    = "USER_NOT_FOUND"
	CodeUserEmailTaken  = "USER_EMAIL_TAKEN"
	CodeUserInvalidRole = "USER_INVALID_ROLE"
	CodeUserInvalidData = "USER_INVALID_DATA"

	CodeCategoryNotFound      = "CATEGORY_NOT_FOUND"
	CodeCategoryNameTaken     = "CATEGORY_NAME_TAKEN"
	CodeCategoryInvalidType   = "CATEGORY_INVALID_TYPE"
	CodeCategoryInvalidMaster = "CATEGORY_INVALID_MASTER"
	CodeCategoryInvalidData   = "CATEGORY_INVALID_DATA"
	CodeCategoryInvalidSlug   = "CATEGORY_INVALID_SLUG"
	CodeCategoryInUse         = "CATEGORY_IN_USE"

	CodeRoleNotFound        = "ROLE_NOT_FOUND"
	CodeRoleNameTaken       = "ROLE_NAME_TAKEN"
	CodeRoleSystemImmutable = "ROLE_SYSTEM_IMMUTABLE"
	CodeRoleInvalidMenu     = "ROLE_INVALID_MENU"
	CodeRoleInvalidData     = "ROLE_INVALID_DATA"

	CodeWalletNotFound        = "WALLET_NOT_FOUND"
	CodeWalletNameTaken       = "WALLET_NAME_TAKEN"
	CodeWalletInvalidData     = "WALLET_INVALID_DATA"
	CodeWalletInvalidType     = "WALLET_INVALID_TYPE"
	CodeWalletInvalidCurrency = "WALLET_INVALID_CURRENCY"
	CodeWalletInvalidCard     = "WALLET_INVALID_CARD"

	CodeMasterCategoryNotFound  = "MASTER_CATEGORY_NOT_FOUND"
	CodeMasterCategoryNameTaken = "MASTER_CATEGORY_NAME_TAKEN"

	CodeBudgetNotFound         = "BUDGET_NOT_FOUND"
	CodeBudgetCategoryTaken    = "BUDGET_CATEGORY_TAKEN"
	CodeBudgetInvalidCategory  = "BUDGET_INVALID_CATEGORY"
	CodeBudgetInvalidCurrency  = "BUDGET_INVALID_CURRENCY"
	CodeBudgetInvalidLimit     = "BUDGET_INVALID_LIMIT"
	CodeBudgetInvalidThreshold = "BUDGET_INVALID_THRESHOLD"
	CodeBudgetInvalidFixed     = "BUDGET_INVALID_FIXED"

	CodeTransactionNotFound        = "TRANSACTION_NOT_FOUND"
	CodeTransactionInvalid         = "TRANSACTION_INVALID"
	CodeTransactionInvalidCategory = "TRANSACTION_INVALID_CATEGORY"
	CodeTransactionInvalidWallet   = "TRANSACTION_INVALID_WALLET"
	CodeTransactionInvalidData     = "TRANSACTION_INVALID_DATA"
	CodeTransactionInvalidType     = "TRANSACTION_INVALID_TYPE"
	CodeTransactionInvalidAmount   = "TRANSACTION_INVALID_AMOUNT"
	CodeTransactionInvalidCurrency = "TRANSACTION_INVALID_CURRENCY"
	// A row in a currency its wallet does not hold; the wallet balance could not absorb it.
	CodeTransactionCurrencyMismatch = "TRANSACTION_CURRENCY_MISMATCH"
	// An expense filed under an income category, or the other way around.
	CodeTransactionCategoryMismatch = "TRANSACTION_CATEGORY_MISMATCH"

	CodeMenuNotFound = "MENU_NOT_FOUND"

	CodeAuditLogNotFound = "AUDIT_LOG_NOT_FOUND"
)
