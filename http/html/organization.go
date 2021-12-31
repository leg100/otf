package html

func (app *Application) organizationShowRoute(organizationID string) string {
	return app.link("organizations", organizationID)
}
func (app *Application) organizationShowAnchor(organizationID string) anchor {
	return anchor{Name: organizationID, Link: app.organizationShowRoute(organizationID)}
}
func (app *Application) organizationShowBreadcrumbs(organizationID string) []anchor {
	return append(app.organizationListBreadcrumbs(), app.organizationShowAnchor(organizationID))
}

func (app *Application) organizationListRoute() string { return app.link("organizations") }
func (app *Application) organizationListAnchor() anchor {
	return anchor{Name: "organizations", Link: app.organizationListRoute()}
}
func (app *Application) organizationListBreadcrumbs() []anchor {
	return append(app.siteBreadcrumbs(), app.organizationListAnchor())
}
