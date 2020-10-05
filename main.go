package main

import "flag"
import "fmt"

import "github.com/golang/glog"
import "math"
import "math/rand"
import "time"

import exprnd "golang.org/x/exp/rand"

import ui "github.com/gizak/termui/v3"

import "github.com/gizak/termui/v3/widgets"
import "gonum.org/v1/gonum/stat/distuv"

var (
	paramNumAgents                  = 100
	paramProb                       = 0.05
	paramConfidence                 = 0.3
	paramAlpha                      = 0.1
	paramNumUpdateNodes             = 1
	paramMaxSteps                   = 6000
	paramMinStationary              = 100
	paramOpinionDiffEpsilon float64 = 0.00000001
)

type Agent struct {
	id               int
	current_opinion  float64
	previous_opinion float64
	new_opinion      float64
	neighbors        map[int]bool
}

type Model struct {
	agents         map[int]*Agent
	confidence     float64
	alpha          float64
	numUpdateNodes int
}

func (m *Model) addAgent(agent *Agent) {
	m.agents[agent.id] = agent
}

func (m *Model) updateOpinions() {
	// N = number of agents, q = numUpdateNodes

	nodes_to_update := make(map[int]bool)
	numAgents := len(m.agents)
	for len(nodes_to_update) < m.numUpdateNodes {
		randId := rand.Intn(numAgents)
		nodes_to_update[randId] = true
	}
	if glog.V(3) {
		glog.Infof("Debug: we have %d nodes to update at this step\n", len(nodes_to_update))
	}

	for agentId, _ := range nodes_to_update {
		agent1 := m.agents[agentId]
		numNeighbors := len(agent1.neighbors)

		if numNeighbors == 0 {
			// Skip this agent, it has ... no friends (!)
			continue
		}

		// Randomly pick a neighbor
		neighborId := rand.Intn(numNeighbors)
		agent2 := m.agents[neighborId]

		opinionDiff := agent1.current_opinion - agent2.current_opinion
		if glog.V(3) {
			glog.Infof("Debug: evaling (%d, %d) who have a diff of %e\n", agentId, neighborId, opinionDiff)
		}

		if math.Abs(opinionDiff) < m.confidence {
			alphaDiff := m.alpha * opinionDiff

			agent2.new_opinion += alphaDiff
			agent1.new_opinion += (-alphaDiff)
		}
	}
}

func NewModel(n int, p float64, c float64, a float64, q int) *Model {
	model := &Model{
		agents:         make(map[int]*Agent),
		confidence:     c,
		alpha:          a,
		numUpdateNodes: q,
	}
	for i := 0; i < n; i++ {
		x := rand.Float64()
		agent := &Agent{
			id:               i,
			current_opinion:  x,
			previous_opinion: x,
			new_opinion:      x,
			neighbors:        make(map[int]bool),
		}

		model.addAgent(agent)
	}

	model.initializeNetwork(n, p)

	return model
}

func (m *Model) initializeNetwork(n int, p float64) {
	numMaxEdges := n * (n - 1) / 2
	binDist := distuv.Binomial{
		N:   float64(numMaxEdges),
		P:   p,
		Src: exprnd.NewSource(uint64(time.Now().UnixNano())),
	}
	numEdges := int(binDist.Rand())

	for edges := 0; edges < numEdges; {
		var id1, id2 int
		for id1 == id2 {
			id1 = rand.Intn(n)
			id2 = rand.Intn(n)
		}

		// Get agents corresponding to these IDs.
		ag1 := m.agents[id1]
		ag2 := m.agents[id2]

		// If link already exists (i.e. if ag2 present in neighbors for ag1) then skip and make a new link.
		if _, ok := ag1.neighbors[id2]; ok {
			continue
		}

		// Create symmetric connection.
		ag1.neighbors[id2] = true
		ag2.neighbors[id1] = true

		if glog.V(3) {
			glog.Infof("Debug: Added links between %d and %d\n", id1, id2)
		}

		edges++
	}
}

func (model *Model) opinionsChanged() bool {
	for id, agent := range model.agents {
		opinionDiff := math.Abs(agent.previous_opinion - agent.new_opinion)
		if opinionDiff > paramOpinionDiffEpsilon {
			if glog.V(3) {
				glog.Infof("Debug: found an opinion diff at %d of %e\n", id, opinionDiff)
			}
			return false
		}
	}

	return true
}

func runner(model *Model, maxSteps int) [][]float64 {
	tstep := 1
	stationary := 0
	var data [][]float64
	for (tstep <= maxSteps) && (stationary < paramMinStationary) {
		if glog.V(2) {
			glog.Infof("Debug: at step %d, stationary=%d\n", tstep, stationary)
		}

		model.updateOpinions()

		var opinions []float64
		for _, agent := range model.agents {
			agent.previous_opinion = agent.current_opinion
			agent.current_opinion = agent.new_opinion

			opinions = append(opinions, agent.current_opinion)
		}
		data = append(data, opinions)

		tstep++
		if model.opinionsChanged() {
			stationary++
		} else {
			stationary = 0
		}
	}

	return data
}

func init() {
	//	flag.Usage = usage
	flag.Parse()
}

func main() {
	fmt.Println("Here we go ...")
	if err := ui.Init(); err != nil {
		glog.Fatalf("Failed to initialize TermUI: %v", err)
	}
	defer ui.Close()

	// Seed rng.
	rand.Seed(time.Now().UnixNano())

	// Create model and run it.
	model := NewModel(paramNumAgents, paramProb, paramConfidence, paramAlpha, paramNumUpdateNodes)
	data := runner(model, paramMaxSteps)

	// "Plot" data.
	plot := widgets.NewPlot()
	plot.Title = "Agent Opinions"

	// Plot the first and last steps (and one somewhere in the middle).
	plot.Data = make([][]float64, 2)
	plot.Data[0] = data[0]
	plot.Data[1] = data[len(data)-1]
	//	plot.Data[2] = data[(len(data)-1)/2]
	plot.AxesColor = ui.ColorWhite
	plot.SetRect(0, 0, 100, 50)

	ui.Render(plot)

	// Wait before quitting ...
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		}
	}

	fmt.Println("Debug: done running")
	// Dump data.
	if glog.V(1) {
		for i, v := range data[0] {
			fmt.Printf("Debug: T0 [%d] -> %f\n", i, v)
		}
		for i, v := range data[len(data)-1] {
			fmt.Printf("Debug: TN [%d] -> %f\n", i, v)
		}
	}
	glog.Flush()
}
