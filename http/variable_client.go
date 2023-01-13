package http

//func (c *client) ListVariables(ctx context.Context, workspaceID string) ([]*otf.Variable, error) {
//	u := fmt.Sprintf("workspaces/%s/variables", workspaceID)
//	req, err := c.newRequest("GET", u, nil)
//	if err != nil {
//		return nil, err
//	}
//
//	list := &dto.VariableList{}
//	err = c.do(ctx, req, list)
//	if err != nil {
//		return nil, err
//	}
//
//	return otf.UnmarshalVariableListJSONAPI(list), nil
//}
