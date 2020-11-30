package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	. "github.com/shawnsky/ty-network-d3/model"
	"io"
	"log"
	"math"
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
	ran := rand.New(rand.NewSource(time.Now().UnixNano()))
	G.nodes = make([]*Node, 0)
	G.edges = make(map[int][]int)
	// Add nodes
	for i := 1; i < 100; i++ {
		G.nodes = append(G.nodes, &Node{ID: i, Name: "HiggsSocialNode", Subtitle: "", Active: 0, Threshold: ran.Float32(), IsLeader: 0, SpreadWilling: ran.Float32()})
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

// 根据节点id查找邻居节点id列表
func findNeighbors(G *Graph, id int) []int {
	return  G.edges[id]
}

// 权重生成算法，服从正态分布
// 期望中值 0.5 标准差 0.22
func generateWeight() (weight float64) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var devstd, mean float64 = 0.22, 0.5
	weight = r.NormFloat64() * devstd + mean
	if weight < 0 || weight > 1 {
		weight = generateWeight()
	}
	return
}


// 用于初始观点值赋值，取决于源值src
// 如果源值为-1，表示初始节点值生成，随机返回0或1
func generateValue(src float32) (val float32) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if src == -1 {
		if r.Float32() <= 0.5 {
			val = 0
		} else {
			val = 1
		}
		return
	}
	var min, max float32
	if src >=0 && src < 0.5 {
		min, max = 0, 0.5
	} else {
		min, max = 0.5, 1
	}
	val = min + r.Float32() * (max - min)
	return
}

// 尝试激活邻居
func active(G *Graph, starter int, option SimulateOption) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	starterNode := findNodeById(G, starter)
	neighbors := findNeighbors(G, starter)
	p := option.pActive

	for _, nodeId := range neighbors {
		node := findNodeById(G, nodeId)
		if node != nil && node.Active == 0 && r.Float32() < p {
			G.lock.RLock()
			// 设置状态
			node.Active = 1
			node.Value = generateValue(starterNode.Value)
			time.Sleep(time.Millisecond*500)
			G.lock.RUnlock()
			if r.Float32() < node.SpreadWilling {
				go active(G, nodeId, option)
			}
			go evolve(G, nodeId)
		}
	}
}

// 观点值更新算法实现
func updateValue(G *Graph, id int) {
	self := findNodeById(G, id)

	// 计算沟通阈值内节点平均值
	var sum, svg float32 = 0, 0
	cnt := 0
	for _, node := range G.nodes {
		// 如果目标节点已激活，而且其观点值在沟通阈值范围内
		if node.Active == 1 && float32(math.Abs(float64(self.Value-node.Value))) <= self.Threshold {
			cnt += 1
			sum += node.Value
		}
	}
	if cnt == 0 {
		svg = self.Value
	}
	svg = sum / float32(cnt)

	// 权重生成
	weight := float32(generateWeight())
	// 更新观点值
	G.lock.RLock()
	self.Value = weight * self.Value + (1-weight) * svg
	G.lock.RUnlock()
	//if id == 20 {
	//	fmt.Println(self.Value)
	//}

}

// 调用evolve函数进入演化状态，被设置为演化状态的节点值会根据邻居节点值持续改变
func evolve(G *Graph, id int) {
	for {
		updateValue(G, id)
		time.Sleep(time.Second)
	}
}

// 仿真入口点
func simulate(G *Graph, option SimulateOption) {
	// 遍历起始点集合，逐个启动传播
	for _, nodeId := range *option.starter {
		node := findNodeById(G, nodeId)
		node.Active = 1
		node.IsLeader = 1
		node.Value = generateValue(-1)
		go active(G, nodeId, option)
	}
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
