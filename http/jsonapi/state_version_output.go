package jsonapi

type StateVersionOutput struct {
	ID        string `jsonapi:"primary,state-version-outputs"`
	Name      string `jsonapi:"attr,name"`
	Sensitive bool   `jsonapi:"attr,sensitive"`
	Type      string `jsonapi:"attr,type"`
	Value     string `jsonapi:"attr,value"`
}
