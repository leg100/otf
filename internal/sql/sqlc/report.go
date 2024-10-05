package sqlc

type Report struct {
	Additions    int `json:"additions"`
	Changes      int `json:"changes"`
	Destructions int `json:"destructions"`
}
