package domain

// OTPGenerator: port for one-time password generation.
type OTPGenerator interface {
	Generate() (string, error)
}
