package jsonapi

// RegistrySession is an unmarshalled JSONAPI representation of an otf registry session
type RegistrySession struct {
	Token            string `jsonapi:"primary,registry_sessions"`
	OrganizationName string `jsonapi:"attr,organization_name"`
}

type RegistrySessionCreateOptions struct {
	// Type is a public field utilized by JSON:API to set the resource type via
	// the field tag.  It is not a user-defined value and does not need to be
	// set.  https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,registry_sessions"`

	// Organization is the name of the organization in which to create the
	// session.
	Organization *string `jsonapi:"attr,organization_name"`
}
