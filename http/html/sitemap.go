package html

import "github.com/gorilla/mux"

var siteTree = map[string]interface{}{
	"listOrganization": map[string]interface{}{
		"newOrganization": nil,
		"getOrganization": map[string]interface{}{
			"getOrganizationOverview": nil,
			"editOrganization":        nil,
			"listWorkspace": map[string]interface{}{
				"newWorkspace": nil,
				"getWorkspace": map[string]interface{}{
					"getWorkspaceOverview": nil,
					"editWorkspace":        nil,
					"editWorkspaceLock":    nil,
					"listRun": map[string]interface{}{
						"getRun": map[string]interface{}{
							"getRunOverview": nil,
						},
					},
				},
			},
		},
	},
}

type SiteMap struct {
	m map[string]string

	router *mux.Router
}

func (sm *SiteMap) Breadcrumbs(name string)
