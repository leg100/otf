package resource

// SiteID identifies the "site", a top-level abstraction identifying the OTF
// system as a whole.
var SiteID site

type site struct{}

func (s site) String() string { return "site" }
func (s site) Kind() Kind     { return SiteKind }
