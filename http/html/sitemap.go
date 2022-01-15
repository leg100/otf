package html

var (
	siteMap = map[string]interface{}{
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

	parentLookupTable map[string]string
)

func init() {
	parentLookupTable = make(map[string]string)
	buildParentLookupTable("", siteMap)
}

func buildParentLookupTable(parent string, m map[string]interface{}) {
	for k, v := range m {
		if parent != "" {
			parentLookupTable[k] = parent
		}

		children, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		buildParentLookupTable(k, children)
	}
}
