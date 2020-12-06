package model

import "sync"

type Node struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Subtitle string  `json:"subtitle"`
	Value    float32 `json:"value"`
	Active   int     `json:"active"`
	Threshold float32 `json:"threshold"`
	IsLeader int     `json:"is_leader"`
	SpreadWilling float32 `json:"spread_willing"`
}

type Edge struct {
	Src int `json:"sid"`
	Dst int `json:"tid"`
	Weight int `json:"weight"`
}

type Graph struct {
	Nodes []*Node          // 节点集
	Edges map[int][]int // 邻接表表示的无向图
	Lock  sync.RWMutex
}

// PushMessage defines message struct send by client to produce to ws client
type PushMessage struct {
	Nodes []Node   `json:"nodes"`
	Edges []Edge   `json:"edges"`
	Appendix string `json:"appendix"`
}
