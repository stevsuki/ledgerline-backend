package domain

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AuditKind names the shape of an audit entry's details. Each kind has exactly
// one struct below, and the struct is what decides the kind: pick the type and
// the compiler enforces the fields that go with it.
type AuditKind string

const (
	AuditKindSession          AuditKind = "session"
	AuditKindSessionFailed    AuditKind = "session_failed"
	AuditKindMoneyEntry       AuditKind = "money_entry"
	AuditKindLimitChange      AuditKind = "limit_change"
	AuditKindRecord           AuditKind = "record"
	AuditKindMembership       AuditKind = "membership"
	AuditKindPermissionChange AuditKind = "permission_change"
	AuditKindDataJob          AuditKind = "data_job"
	AuditKindSettingChange    AuditKind = "setting_change"
	AuditKindReportViewed     AuditKind = "report_viewed"
	AuditKindAlertSent        AuditKind = "alert_sent"
)

// AuditDetail is what AuditLog.Details can hold. Build one with its New
// constructor so kind and payload can never disagree.
type AuditDetail interface {
	// Text renders the entry for people to read. Part of the interface so a new
	// kind cannot be added without one.
	Text() string
	AuditKind() AuditKind
}

type SessionMethod string

const (
	SessionMethodPassword  SessionMethod = "password"
	SessionMethodOAuth     SessionMethod = "oauth"
	SessionMethodBiometric SessionMethod = "biometric"
	SessionMethodInvite    SessionMethod = "invite"
)

type RecordType string

const (
	RecordTypeWallet      RecordType = "wallet"
	RecordTypeGoal        RecordType = "goal"
	RecordTypeRole        RecordType = "role"
	RecordTypeBudget      RecordType = "budget"
	RecordTypeInstitution RecordType = "institution"
)

// PermissionAction mirrors the can_* columns on role_menu_permissions.
type PermissionAction string

const (
	PermissionCreate  PermissionAction = "create"
	PermissionRead    PermissionAction = "read"
	PermissionUpdate  PermissionAction = "update"
	PermissionDelete  PermissionAction = "delete"
	PermissionApprove PermissionAction = "approve"
)

type DataJob string

const (
	DataJobExport DataJob = "export"
	DataJobSync   DataJob = "sync"
)

type DataFormat string

const (
	DataFormatCSV DataFormat = "CSV"
	DataFormatPDF DataFormat = "PDF"
)

type SessionDetail struct {
	Kind      AuditKind     `json:"kind"`
	Method    SessionMethod `json:"method"`
	UserAgent string        `json:"user_agent,omitempty"`
}

func NewSessionDetail(method SessionMethod, userAgent string) SessionDetail {
	return SessionDetail{Kind: AuditKindSession, Method: method, UserAgent: userAgent}
}

func (SessionDetail) AuditKind() AuditKind { return AuditKindSession }

// SessionFailedReason separates a rejected password from an attempt that never
// got that far because the account was already locked.
type SessionFailedReason string

const (
	SessionFailedWrongPassword SessionFailedReason = "wrong_password"
	SessionFailedLocked        SessionFailedReason = "locked"
)

type SessionFailedDetail struct {
	Kind        AuditKind           `json:"kind"`
	Reason      SessionFailedReason `json:"reason"`
	Attempts    int                 `json:"attempts"`
	MaxAttempts int                 `json:"max_attempts,omitempty"`
	Email       string              `json:"email"`
}

// NewWrongPasswordDetail: the password was checked and rejected.
func NewWrongPasswordDetail(attempts, maxAttempts int, email string) SessionFailedDetail {
	return SessionFailedDetail{
		Kind:        AuditKindSessionFailed,
		Reason:      SessionFailedWrongPassword,
		Attempts:    attempts,
		MaxAttempts: maxAttempts,
		Email:       email,
	}
}

// NewAccountLockedDetail: the attempt was turned away by the lock itself.
func NewAccountLockedDetail(attempts int, email string) SessionFailedDetail {
	return SessionFailedDetail{
		Kind:     AuditKindSessionFailed,
		Reason:   SessionFailedLocked,
		Attempts: attempts,
		Email:    email,
	}
}

func (SessionFailedDetail) AuditKind() AuditKind { return AuditKindSessionFailed }

type MoneyEntryDetail struct {
	Kind   AuditKind `json:"kind"`
	Amount Money     `json:"amount"`
	Label  string    `json:"label"`
	Wallet string    `json:"wallet,omitempty"`
	Reason string    `json:"reason,omitempty"`
}

func NewMoneyEntryDetail(amount Money, label string) MoneyEntryDetail {
	return MoneyEntryDetail{Kind: AuditKindMoneyEntry, Amount: amount, Label: label}
}

func (MoneyEntryDetail) AuditKind() AuditKind { return AuditKindMoneyEntry }

type LimitChangeDetail struct {
	Kind   AuditKind `json:"kind"`
	Target string    `json:"target"`
	From   Money     `json:"from"`
	To     Money     `json:"to"`
}

func NewLimitChangeDetail(target string, from, to Money) LimitChangeDetail {
	return LimitChangeDetail{Kind: AuditKindLimitChange, Target: target, From: from, To: to}
}

func (LimitChangeDetail) AuditKind() AuditKind { return AuditKindLimitChange }

type RecordDetail struct {
	Kind       AuditKind  `json:"kind"`
	RecordType RecordType `json:"record_type"`
	Name       string     `json:"name"`
	Note       string     `json:"note,omitempty"`
	Amount     *Money     `json:"amount,omitempty"`
}

func NewRecordDetail(recordType RecordType, name string) RecordDetail {
	return RecordDetail{Kind: AuditKindRecord, RecordType: recordType, Name: name}
}

func (RecordDetail) AuditKind() AuditKind { return AuditKindRecord }

type MembershipDetail struct {
	Kind    AuditKind `json:"kind"`
	Subject string    `json:"subject"`
	Email   string    `json:"email,omitempty"`
	Role    string    `json:"role,omitempty"`
	Reason  string    `json:"reason,omitempty"`
}

func NewMembershipDetail(subject string) MembershipDetail {
	return MembershipDetail{Kind: AuditKindMembership, Subject: subject}
}

func (MembershipDetail) AuditKind() AuditKind { return AuditKindMembership }

type PermissionChangeDetail struct {
	Kind       AuditKind        `json:"kind"`
	Role       string           `json:"role"`
	Permission PermissionAction `json:"permission"`
	Module     string           `json:"module"`
	Granted    bool             `json:"granted"`
}

func NewPermissionChangeDetail(role string, permission PermissionAction, module string, granted bool) PermissionChangeDetail {
	return PermissionChangeDetail{
		Kind:       AuditKindPermissionChange,
		Role:       role,
		Permission: permission,
		Module:     module,
		Granted:    granted,
	}
}

func (PermissionChangeDetail) AuditKind() AuditKind { return AuditKindPermissionChange }

type DataJobDetail struct {
	Kind   AuditKind  `json:"kind"`
	Job    DataJob    `json:"job"`
	Format DataFormat `json:"format,omitempty"`
	Source string     `json:"source,omitempty"`
	Rows   int        `json:"rows,omitempty"`
	Period string     `json:"period,omitempty"`
}

func NewDataJobDetail(job DataJob) DataJobDetail {
	return DataJobDetail{Kind: AuditKindDataJob, Job: job}
}

func (DataJobDetail) AuditKind() AuditKind { return AuditKindDataJob }

type SettingChangeDetail struct {
	Kind    AuditKind `json:"kind"`
	Setting string    `json:"setting"`
	Enabled *bool     `json:"enabled,omitempty"`
	From    string    `json:"from,omitempty"`
	To      string    `json:"to,omitempty"`
}

func NewSettingChangeDetail(setting string) SettingChangeDetail {
	return SettingChangeDetail{Kind: AuditKindSettingChange, Setting: setting}
}

func (SettingChangeDetail) AuditKind() AuditKind { return AuditKindSettingChange }

type ReportViewedDetail struct {
	Kind   AuditKind `json:"kind"`
	Report string    `json:"report"`
	Period string    `json:"period"`
}

func NewReportViewedDetail(report, period string) ReportViewedDetail {
	return ReportViewedDetail{Kind: AuditKindReportViewed, Report: report, Period: period}
}

func (ReportViewedDetail) AuditKind() AuditKind { return AuditKindReportViewed }

type AlertSentDetail struct {
	Kind             AuditKind `json:"kind"`
	Subject          string    `json:"subject"`
	ThresholdPercent int       `json:"threshold_percent"`
}

func NewAlertSentDetail(subject string, thresholdPercent int) AlertSentDetail {
	return AlertSentDetail{Kind: AuditKindAlertSent, Subject: subject, ThresholdPercent: thresholdPercent}
}

func (AlertSentDetail) AuditKind() AuditKind { return AuditKindAlertSent }

// RawAuditDetail is a detail read back from storage. Nothing in the backend
// looks inside one, so it is carried and re-emitted as the JSON it was stored
// as, which keeps the read path free of any per-kind decoding.
type RawAuditDetail json.RawMessage

func (r RawAuditDetail) AuditKind() AuditKind {
	var probe struct {
		Kind AuditKind `json:"kind"`
	}
	if json.Unmarshal(r, &probe) != nil {
		return ""
	}
	return probe.Kind
}

func (r RawAuditDetail) MarshalJSON() ([]byte, error) {
	if len(r) == 0 {
		return []byte("null"), nil
	}
	return r, nil
}

// note appends a trailing remark. The web UI separates attributes with "·" and
// a closing note with an em dash: "Auditor — read-only, 2 permissions".
func note(head, tail string) string {
	if tail == "" {
		return head
	}
	if head == "" {
		return tail
	}
	return head + " — " + tail
}

// join glues the parts that are actually filled with a separator.
func join(parts ...string) string {
	kept := parts[:0]
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, " · ")
}

func (d SessionDetail) Text() string {
	method := map[SessionMethod]string{
		SessionMethodPassword:  "Email + password",
		SessionMethodOAuth:     "Google OAuth",
		SessionMethodBiometric: "Biometric",
		SessionMethodInvite:    "Invite link accepted",
	}[d.Method]
	if method == "" {
		method = string(d.Method)
	}
	return join(method, d.UserAgent)
}

func (d SessionFailedDetail) Text() string {
	attempts := fmt.Sprintf("%d attempts", d.Attempts)
	if d.Attempts == 1 {
		attempts = "1 attempt"
	}
	if d.Reason == SessionFailedLocked {
		return join("Account locked", attempts, d.Email)
	}
	return join(attempts, d.Email)
}

func (d MoneyEntryDetail) Text() string {
	return join(d.Amount.String(), d.Label, d.Wallet, d.Reason)
}

func (d LimitChangeDetail) Text() string {
	return fmt.Sprintf("%s limit %s → %s", d.Target, d.From.String(), d.To.String())
}

func (d RecordDetail) Text() string {
	amount := ""
	if d.Amount != nil {
		amount = "target " + d.Amount.String()
	}
	return note(join(d.Name, amount), d.Note)
}

func (d MembershipDetail) Text() string {
	subject := d.Subject
	if d.Email != "" && d.Email != d.Subject {
		subject = fmt.Sprintf("%s (%s)", d.Subject, d.Email)
	}
	if d.Role != "" {
		subject += " as " + d.Role
	}
	return note(subject, d.Reason)
}

// permissionLabels: what the UI calls each action.
var permissionLabels = map[PermissionAction]string{
	PermissionCreate:  "Create",
	PermissionRead:    "Read",
	PermissionUpdate:  "Update",
	PermissionDelete:  "Delete",
	PermissionApprove: "Approval",
}

func (d PermissionChangeDetail) Text() string {
	verb := "revoked"
	if d.Granted {
		verb = "granted"
	}
	permission := permissionLabels[d.Permission]
	if permission == "" {
		permission = string(d.Permission)
	}
	return note(d.Role, fmt.Sprintf("%s %s on %s", verb, permission, d.Module))
}

func (d DataJobDetail) Text() string {
	// A sync reads as a sentence about its source; an export as a list of
	// attributes: "BCA — 12 transactions imported" vs "CSV · 184 rows · August 2026".
	if d.Job == DataJobSync {
		imported := ""
		if d.Rows > 0 {
			imported = fmt.Sprintf("%d transactions imported", d.Rows)
		}
		return join(note(d.Source, imported), d.Period)
	}

	rows := ""
	if d.Rows > 0 {
		rows = fmt.Sprintf("%d rows", d.Rows)
	}
	return join(string(d.Format), d.Source, rows, d.Period)
}

func (d SettingChangeDetail) Text() string {
	switch {
	case d.Enabled != nil && *d.Enabled:
		return d.Setting + " enabled"
	case d.Enabled != nil:
		return d.Setting + " disabled"
	case d.From != "" || d.To != "":
		return fmt.Sprintf("%s %s → %s", d.Setting, d.From, d.To)
	default:
		return d.Setting
	}
}

func (d ReportViewedDetail) Text() string {
	return join(d.Report, d.Period)
}

func (d AlertSentDetail) Text() string {
	return fmt.Sprintf("%s crossed %d%%", d.Subject, d.ThresholdPercent)
}

// Text is empty on the read path: the sentence was rendered when the entry was
// written and is read back from its own column.
func (r RawAuditDetail) Text() string { return "" }
