package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	. "github.com/shawnsky/ty-network-d3/model"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type Producer struct {
	conn *Conn
	running bool
	chActive chan int
	chEvolve chan int
	chCheckEnd chan int
}

type SimulateOption struct {
	starter *[]int
	pActive float32
}


var instance *Producer

func GetInstance() *Producer {
	if instance == nil {
		pd := Producer{
			chActive: make(chan int),
			chEvolve: make(chan int),
			chCheckEnd: make(chan int),
		}
		instance = &pd
	}
	return instance
}

func (p *Producer) SetConn(c *Conn) {
	p.conn = c
}

func (p *Producer) buildHiggsSocialNetwork(G *Graph) {
	ran := rand.New(rand.NewSource(time.Now().UnixNano()))
	G.Nodes = make([]*Node, 0)
	G.Edges = make(map[int][]int)
	// Add nodes
	for i := 1; i < 100; i++ {
		G.Nodes = append(G.Nodes, &Node{
			ID: i,
			Name: "HiggsSocialNode",
			Subtitle: "",
			Active: 0,
			Threshold: ran.Float64(),
			IsLeader: 0,
			SpreadWilling: ran.Float64(),
			Evolved: 0,
			Spread: 0,
		})
	}

	// Add edges
	pwd, _ := os.Getwd()
	csvFile, err := os.Open(pwd + "/server/higgs-social_network-1000.csv")
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
		_, ok := G.Edges[src]
		if !ok {
			G.Edges[src] = make([]int, 0)
		}
		G.Edges[src] = append(G.Edges[src], dst)

	}
}

// 根据节点id查找节点对象指针
func (p *Producer) findNodeById(G *Graph, id int) *Node {
	nodeSet := G.Nodes
	for _, node := range nodeSet {
		if node.ID == id {
			return node
		}
	}
	return nil
}

// 根据节点id查找邻居节点id列表
func (p *Producer) findNeighbors(G *Graph, id int) []int {
	return  G.Edges[id]
}

// 权重生成算法，服从正态分布
// 期望中值 0.5 标准差 0.22
func (p *Producer) generateWeight() (weight float64) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var devstd, mean float64 = 0.22, 0.5
	weight = r.NormFloat64() * devstd + mean
	if weight < 0 || weight > 1 {
		weight = p.generateWeight()
	}
	return
}


// 用于初始观点值赋值，取决于源值src
// 如果源值为-1，表示初始节点值生成，随机返回0或1
func (p *Producer) generateValue(src float64) (val float64) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if src == -1 {
		if r.Float64() <= 0.5 {
			val = 0
		} else {
			val = 1
		}
		return
	}
	var min, max float64
	if src >=0 && src < 0.5 {
		min, max = 0, 0.5
	} else {
		min, max = 0.5, 1
	}
	val = min + r.Float64() * (max - min)
	val = math.Round(val * 100) / 100  // 保留两位小数
	return
}

// 尝试激活邻居
func (p *Producer) active(G *Graph, starter int, option SimulateOption) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	starterNode := p.findNodeById(G, starter)
	neighbors := p.findNeighbors(G, starter)

	for _, nodeId := range neighbors {
		node := p.findNodeById(G, nodeId)
		if node != nil && node.Active == 0 && r.Float32() < option.pActive {
			G.Lock.RLock()
			// 设置状态
			node.Active = 1
			node.Value = p.generateValue(starterNode.Value)
			time.Sleep(time.Millisecond*500)
			G.Lock.RUnlock()
			if r.Float64() < node.SpreadWilling {
				// 愿意传播
				node.Spread = 1
				go p.active(G, nodeId, option)
			}
			go p.evolve(G, nodeId)
		}
	}
}

// 观点值更新算法实现
func (p *Producer) updateValue(G *Graph, id int) {
	self := p.findNodeById(G, id)
	// 设置为演化状态
	self.Evolved = 1

	// 计算沟通阈值内节点平均值
	var sum, svg float64 = 0, 0
	cnt := 0
	for _, node := range G.Nodes {
		// 如果目标节点已激活，而且其观点值在沟通阈值范围内
		if node.Active == 1 && math.Abs(self.Value-node.Value) <= self.Threshold {
			cnt += 1
			sum += node.Value
		}
	}
	if cnt == 0 {
		svg = self.Value
	}
	svg = sum / float64(cnt)

	// 权重生成
	weight := p.generateWeight()
	// 更新观点值
	G.Lock.RLock()
	val := weight * self.Value + (1-weight) * svg
	// 如果观点值还在变化，演化没有结束
	if val == self.Value {
		p.chCheckEnd <- 1
	}
	self.Value = math.Round(val * 100) / 100
	G.Lock.RUnlock()
	//if id == 20 {
	//	fmt.Println(self.Value)
	//}

}

// 调用evolve函数进入演化状态，被设置为演化状态的节点值会根据邻居节点值持续改变
func (p *Producer) evolve(G *Graph, id int) {
	for {
		p.updateValue(G, id)
		time.Sleep(time.Second)
	}
}

// 仿真入口点
func (p *Producer) simulate(G *Graph, option SimulateOption) {
	// 遍历起始点集合，逐个启动传播
	for _, nodeId := range *option.starter {
		node := p.findNodeById(G, nodeId)
		node.Active = 1
		node.IsLeader = 1
		node.Value = p.generateValue(-1)
		go p.active(G, nodeId, option)
	}
}

// 生成PushMessage，传送整个网络结构
func (p *Producer) initGraphFrame(G *Graph) (pm PushMessage) {
	pm = PushMessage{}
	pm.Edges = make([]Edge, 0)
	pm.Nodes = make([]Node, 0)

	for i := 0; i < len(G.Nodes); i++ {
		nodeId := G.Nodes[i].ID
		pm.Nodes = append(pm.Nodes, *G.Nodes[i])
		singleNodeEdgePairs := make([]Edge, 0)
		nodeEdges := G.Edges[nodeId]
		for k := 0; k < len(nodeEdges); k++ {
			singleNodeEdgePairs = append(singleNodeEdgePairs, Edge{Src: nodeId, Dst: nodeEdges[k]})
		}
		pm.Edges = append(pm.Edges, singleNodeEdgePairs...)
	}

	return
}

// 推送
func (p *Producer) push(G *Graph, conn *Conn) {
	pm := p.initGraphFrame(G)

	pm.Appendix = fmt.Sprintf("Data in %s", time.Now().Format("2006-01-02 15:04:05.000"))

	jsonData,_ := json.Marshal(pm)
	_, err := conn.Write(jsonData)
	if err != nil {
		panic(err)
	}
}



func (p *Producer) pushForever(G *Graph, conn *Conn) {
	for {
		p.push(G, conn)
		time.Sleep(time.Millisecond * 500)
	}
}

// 持续输出网络状态，分别是四种状态节点数量
func (p *Producer) makeStatics(G *Graph) {
	for {
		all := len(G.Nodes)
		var activeCount, evlovedCount, spreadCount int = 0, 0, 0
		for _, node := range G.Nodes {
			if node.Active == 1 {
				activeCount += 1
			}
			if node.Evolved == 1 {
				evlovedCount += 1
			}
			if node.Spread == 1 {
				spreadCount += 1
			}
		}
		fmt.Printf("[*] 网络状态 激活：%d，未激活：%d，传播：%d，演化：%d\n", activeCount, all-activeCount, spreadCount, evlovedCount)
		time.Sleep(time.Millisecond *500)
	}
}


func (p *Producer) Start() {
	var G = Graph{}
	p.buildHiggsSocialNetwork(&G)
	option := SimulateOption{
		pActive: 0.6,
		starter: &[]int{1, 24},
	}
	p.simulate(&G, option)
	go p.pushForever(&G, p.conn)
	go p.makeStatics(&G)


}

func (p *Producer) Pause() {
	p.chActive <- 1
	p.chEvolve <- 1
}

