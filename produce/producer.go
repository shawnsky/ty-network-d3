package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/shawnsky/ty-network-d3/model"
	"net/http"
	"time"
)

func main() {
	pushURL := "http://127.0.0.1:7341/produce"
	contentType := "application/json"
	pm := model.PushMessage{
	}
	pm.Nodes = make([]model.Node, 0)
	pm.Edges = make([]model.Edge, 0)

	for {
		pm.Appendix = fmt.Sprintf("Data in %s", time.Now().Format("2006-01-02 15:04:05.000"))
		pm.Nodes = append(pm.Nodes, model.Node{ID: 0, Name: "hi", Subtitle: "hi"})
		pm.Edges = append(pm.Edges, model.Edge{Src: 1, Dst: 2})
		b, _ := json.Marshal(pm)

		fmt.Println(pm)

		_, err := http.DefaultClient.Post(pushURL, contentType, bytes.NewReader(b))
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second)
	}
}
