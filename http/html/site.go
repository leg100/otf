package html

func (app *Application) siteRoute() string         { return app.link("/site") }
func (app *Application) siteAnchor() anchor        { return anchor{Name: "site", Link: app.siteRoute()} }
func (app *Application) siteBreadcrumbs() []anchor { return []anchor{app.siteAnchor()} }
