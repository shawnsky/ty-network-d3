package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	. "github.com/shawnsky/ty-network-d3/model"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Graph struct {
	nodes []*Node          // 节点集
	edges map[int][]int // 邻接表表示的无向图
	lock  sync.RWMutex

}

type SimulateOption struct {
	starter *[]int
	pActive float32
}

func buildHiggsSocialNetwork(G *Graph) {
	G.nodes = make([]*Node, 0)
	G.edges = make(map[int][]int)
	// Add nodes
	for i := 1; i < 100; i++ {
		G.nodes = append(G.nodes, &Node{ID: i, Name: "HiggsSocialNode", Subtitle: "", Active: 0})
	}

	// Add edges
	csvFile, err := os.Open("/Users/amalthea/Projects/golang/gopath/src/github.com/shawnsky/ty-network-d3/produce/higgs-social_network-1000.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	// Parse the file
	r := csv.NewReader(csvFile)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		src, _ := strconv.Atoi(record[0])
		dst, _ := strconv.Atoi(record[1])
		// 需要判断src节点的边集合是否初始化完成
		_, ok := G.edges[src]
		if !ok {
			G.edges[src] = make([]int, 0)
		}
		G.edges[src] = append(G.edges[src], dst)

	}
}

// 根据节点id查找节点对象指针
func findNodeById(G *Graph, id int) *Node {
	nodeSet := G.nodes
	for _, node := range nodeSet {
		if node.ID == id {
			return node
		}
	}
	return nil
}

// 尝试激活邻居
func active(G *Graph, starter int, option SimulateOption) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	neighbors := G.edges[starter]
	p := option.pActive


	for _, nodeId := range neighbors {
		node := findNodeById(G, nodeId)
		if node != nil && node.Active == 0 && r.Float32() < p {
			G.lock.RLock()
			// 设置状态
			node.Active = 1
			// TODO: 设置观点值
			time.Sleep(time.Millisecond*500)
			G.lock.RUnlock()
			go active(G, nodeId, option)
		}
	}
}

// 仿真入口点
func simulate(G *Graph, option SimulateOption) {
	// 遍历起始点集合，逐个启动传播
	for _, nodeId := range *option.starter {
		node := findNodeById(G, nodeId)
		node.Active = 1
		// 起始点观点值随机：0或1
		if rand.Float32() <= 0.5 {
			node.Value = 0
		} else {
			node.Value = 1
		}
		go active(G, nodeId, option)
	}
	// 观点更新
	// 规则：起始点必须是0/1？
	//      被激活的人的观点值赋值依据？比如传播者观点值为 0.6，那么激活者的观点值为？
	//      观点值更新公式？
}

// 生成PushMessage，传送整个网络结构
func initGraphFrame(G *Graph) (pm PushMessage) {
	pm = PushMessage{}
	pm.Edges = make([]Edge, 0)
	pm.Nodes = make([]Node, 0)

	for i := 0; i < len(G.nodes); i++ {
		nodeId := G.nodes[i].ID
		pm.Nodes = append(pm.Nodes, *G.nodes[i])
		singleNodeEdgePairs := make([]Edge, 0)
		nodeEdges := G.edges[nodeId]
		for k := 0; k < len(nodeEdges); k++ {
			singleNodeEdgePairs = append(singleNodeEdgePairs, Edge{Src: nodeId, Dst: nodeEdges[k]})
		}
		pm.Edges = append(pm.Edges, singleNodeEdgePairs...)
	}

	return
}


func push(G *Graph) {
	pm := initGraphFrame(G)
	pushURL := "http://127.0.0.1:7341/produce"
	contentType := "application/json"
	pm.Appendix = fmt.Sprintf("Data in %s", time.Now().Format("2006-01-02 15:04:05.000"))
	b, _ := json.Marshal(pm)
	_, err := http.DefaultClient.Post(pushURL, contentType, bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
}

func pushForever(G *Graph) {
	for {
		push(G)
		time.Sleep(time.Millisecond * 500)
	}
}

func main() {
	var G = Graph{}
	buildHiggsSocialNetwork(&G)
	option := SimulateOption{
		pActive: 0.6,
		starter: &[]int{1, 24},
	}
	simulate(&G, option)
	pushForever(&G)


}
