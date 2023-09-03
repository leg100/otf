// Package ghapp provides a github application.
package ghapp

type ghapp struct {
	ID            string
	WebhookSecret string
	Pem           string
	Organization  string
}
