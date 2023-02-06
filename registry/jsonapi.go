package registry

// jsonapiSession is an unmarshalled JSONAPI representation of an otf registry session
type jsonapiSession struct {
	Token            string `jsonapi:"primary,registry_sessions"`
	OrganizationName string `jsonapi:"attr,organization_name"`
}

func (s *jsonapiSession) toSession() *Session {
	return &Session{
		token:        s.Token,
		organization: s.OrganizationName,
	}
}

// RegistrySessionCreateOptions represents the options for creating a new
// JSONAPI registry session.
type jsonapiCreateOptions struct {
	// Type is a public field utilized by JSON:API to set the resource type via
	// the field tag.  It is not a user-defined value and does not need to be
	// set.  https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,registry_sessions"`

	// OrganizationName is the name of the organization in which to create the
	// session.
	OrganizationName string `jsonapi:"attr,organization_name"`
}
