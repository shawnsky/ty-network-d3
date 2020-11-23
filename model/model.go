package model


type Node struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Subtitle string  `json:"subtitle"`
	Value    float32 `json:"value"`
}

type Edge struct {
	Src int `json:"src"`
	Dst int `json:"dst"`
	Weight int `json:"weight"`
}

// PushMessage defines message struct send by client to produce to ws client
type PushMessage struct {
	Nodes []Node   `json:"nodes"`
	Edges []Edge   `json:"edges"`
	Appendix string `json:"appendix"`
}
