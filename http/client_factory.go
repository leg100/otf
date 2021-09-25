package http

// ClientFactory implementations of capable of constructing new OTF clients
type ClientFactory interface {
	NewClient() (Client, error)
}
