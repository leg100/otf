package state

type output struct {
	id             string
	name           string
	typ            string
	value          []byte
	sensitive      bool
	stateVersionID string
}

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (out *output) ToJSONAPI() any {
	return &jsonapiVersionOutput{
		ID:        out.id,
		Name:      out.name,
		Sensitive: out.sensitive,
		Type:      out.typ,
		Value:     out.value,
	}
}

type outputList []*output

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (out outputList) ToJSONAPI() any {
	var to jsonapiVersionOutputList
	for _, from := range out {
		to.Items = append(to.Items, from.ToJSONAPI().(*jsonapiVersionOutput))
	}
	return &to
}
