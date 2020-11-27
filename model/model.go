package model


type Node struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Subtitle string  `json:"subtitle"`
	Value    float32 `json:"value"`
	Active   int     `json:"active"`
}

type Edge struct {
	Src int `json:"sid"`
	Dst int `json:"tid"`
	Weight int `json:"weight"`
}

// PushMessage defines message struct send by client to produce to ws client
type PushMessage struct {
	Nodes []Node   `json:"nodes"`
	Edges []Edge   `json:"edges"`
	Appendix string `json:"appendix"`
}
