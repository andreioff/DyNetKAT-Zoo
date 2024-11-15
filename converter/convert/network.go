package convert

import (
	"errors"
	"maps"
	"slices"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

type Network struct {
	topology      util.Graph
	shortestPaths map[util.I64Tup][]*Switch // maps a tuple of topology node ids to a switch path

	switches   []*Switch
	nodeIdToSw map[int64]*Switch

	controllers []*Controller
	hosts       []*Host
	portNr      int64
	hostId      int64
}

func NewNetwork(topo util.Graph) (*Network, error) {
	var portNr int64 = 0

	switches, err := makeSwitchesFromTopology(topo, &portNr)
	if err != nil {
		return &Network{}, err
	}

	return &Network{
		topology:      topo,
		shortestPaths: computeShortestPaths(topo, switches),
		switches:      switches,
		nodeIdToSw:    mapNodeToSwitch(switches),
		portNr:        portNr,
		hostId:        0,
		hosts:         []*Host{},
	}, nil
}

func mapNodeToSwitch(switches []*Switch) map[int64]*Switch {
	nodeIdToSwitch := make(map[int64]*Switch, len(switches))

	for _, sw := range switches {
		nodeIdToSwitch[sw.topoNode.ID()] = sw
	}

	return nodeIdToSwitch
}

func (n *Network) PortNr() int64 {
	return n.portNr
}

func (n *Network) SetPortNr(newPortNr int64) {
	n.portNr = newPortNr
}

func (n *Network) Switches() []*Switch {
	return n.switches
}

func (n *Network) NodeIdToSw() map[int64]*Switch {
	return n.nodeIdToSw
}

func (n *Network) Hosts() []*Host {
	return n.hosts
}

func (n *Network) Controllers() []*Controller {
	return n.controllers
}

func makeSwitchesFromTopology(topo util.Graph, portNr *int64) ([]*Switch, error) {
	if portNr == nil {
		return []*Switch{}, errors.New("Nil portNr argument!")
	}

	switches := []*Switch{}

	iter := topo.Nodes()
	for iter.Next() {
		links, err := makeLinks(topo, iter.Node(), portNr)
		if err != nil {
			return []*Switch{}, err
		}

		newSw, err := NewSwitch(iter.Node(), links)
		if err != nil {
			return []*Switch{}, err
		}

		switches = append(switches, newSw)
	}

	return switches, nil
}

func makeLinks(topo util.Graph, node graph.Node, portNr *int64) ([]Link, error) {
	if portNr == nil {
		return []Link{}, errors.New("Nil portNr argument!")
	}

	incidentEdges, err := util.GetIncidentEdges(topo, node)
	if err != nil {
		return []Link{}, err
	}

	links := []Link{}
	for _, edge := range incidentEdges {
		newLink := *NewLink(edge, *portNr, *portNr+1)
		links = append(links, newLink)
		*portNr += 2
	}
	return links, nil
}

func nodePathToSwitchPath(path []graph.Node, nodeToSwitch map[int64]*Switch) []*Switch {
	switchPath := []*Switch{}
	for _, node := range path {
		switchPath = append(switchPath, nodeToSwitch[node.ID()])
	}
	return switchPath
}

func computeShortestPaths(topo util.Graph, switches []*Switch) map[util.I64Tup][]*Switch {
	nodePaths := path.DijkstraAllPaths(&topo)
	nodeToSwitch := mapNodeToSwitch(switches)
	switchPaths := make(map[util.I64Tup][]*Switch)

	for i := range len(switches) {
		sw1Id := switches[i].topoNode.ID()

		for j := range len(switches) {
			sw2Id := switches[j].topoNode.ID()
			nodePath, _, _ := nodePaths.Between(sw1Id, sw2Id)
			switchPaths[util.NewI64Tup(sw1Id, sw2Id)] = nodePathToSwitchPath(
				nodePath,
				nodeToSwitch,
			)
		}
	}

	return switchPaths
}

func (n *Network) assignHosts(hostsNr uint) error {
	hosts, err := n.CreateHosts(hostsNr)
	if err != nil {
		return err
	}

	for _, h := range hosts {
		h.sw.AddHost(h)
	}
	n.hosts = hosts

	return nil
}

func (n *Network) CreateHosts(hostsNr uint) ([]*Host, error) {
	hosts := []*Host{}

	randSws, err := n.pickRandomSwitches(hostsNr)
	if err != nil {
		return []*Host{}, err
	}

	for _, randSw := range randSws {
		newHost, err := NewHost(n.hostId, n.portNr, randSw)
		if err != nil {
			return []*Host{}, err
		}
		hosts = append(hosts, &newHost)

		n.hostId++
		n.portNr++
	}

	return hosts, nil
}

// Turns out that the switches order in the array is not static,
// so we must pick them by ID
func (n *Network) pickRandomSwitches(picksNr uint) ([]*Switch, error) {
	if len(n.switches) == 0 {
		return []*Switch{}, errors.New("Network has no switches!")
	}
	nodeIds := slices.Collect(maps.Keys(n.nodeIdToSw))

	randIdPicks := util.RandomFromArrayWithReplc(nodeIds, picksNr)

	randSws := []*Switch{}
	for _, nodeId := range randIdPicks {
		randSws = append(randSws, n.nodeIdToSw[nodeId])
	}

	return randSws, nil
}

func (n *Network) populateFlowTables(h1, h2 *Host) error {
	if h1 == nil || h2 == nil {
		return errors.New("Null arguments!")
	}

	entries, err := n.GetFlowRulesForSwitchPath(h1.sw, h2.sw, h1.SwitchPort(), h2.SwitchPort())
	if err != nil {
		return err
	}

	for nodeId, portTuples := range entries {
		for _, fromPortToPort := range portTuples {
			n.nodeIdToSw[nodeId].FlowTable().
				AddEntry(h2.ID(), fromPortToPort.Fst, fromPortToPort.Snd)
		}
	}

	return nil
}

// Maps the switches on the path between 'srcSw' and 'destSw' to their
// corresponding flow rules for forwarding packets
// from 'srcSw', port 'inPortSrcSw' to 'destSw', port 'outPortDestSw'
func (n *Network) GetFlowRulesForSwitchPath(
	srcSw *Switch,
	destSw *Switch,
	inPortSrcSw int64,
	outPortDestSw int64,
) (map[int64][]util.I64Tup, error) {
	if srcSw == nil || destSw == nil {
		return make(map[int64][]util.I64Tup), errors.New("Null arguments!")
	}

	path, exists := n.shortestPaths[util.NewI64Tup(srcSw.topoNode.ID(), destSw.topoNode.ID())]
	if !exists {
		return make(
				map[int64][]util.I64Tup,
			), errors.New(
				"Path between given switches could not be found!",
			)
	}

	entries := make(map[int64][]util.I64Tup)
	receivingPort := inPortSrcSw

	for i := range len(path) - 1 {
		currSw := path[i]
		nextSwId := path[i+1].topoNode.ID()
		link := currSw.FindLink(nextSwId)

		inPortOutPort := util.I64Tup{Fst: receivingPort, Snd: link.FromPort()}
		outPortNextSwInPort := util.I64Tup{Fst: link.FromPort(), Snd: link.ToPort()}

		entries[currSw.topoNode.ID()] = []util.I64Tup{inPortOutPort, outPortNextSwInPort}

		receivingPort = link.ToPort()
	}

	inPortOutPort := util.I64Tup{Fst: receivingPort, Snd: outPortDestSw}
	entries[destSw.topoNode.ID()] = []util.I64Tup{inPortOutPort}

	return entries, nil
}

func (n *Network) AddAndConnectHosts(hostsNr uint) error {
	if hostsNr < 2 {
		return errors.New("Number of hosts must be at least 2!")
	}

	err := n.assignHosts(hostsNr)
	if err != nil {
		return err
	}

	for i := range len(n.hosts) {
		for j := i + 1; j < len(n.hosts); j++ {
			err := n.populateFlowTables(n.hosts[i], n.hosts[j])
			if err != nil {
				return err
			}

			err = n.populateFlowTables(n.hosts[j], n.hosts[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/*
Adds 'controllersNr' controllers to the network, randomly assigning them to the network' switches.
All switches in the network are equally divided between these controllers. If the number of switches
cannot be equally divided, the remainder is uniformly distributed again, 1 switch per controller
starting from the first controller.
*/
func (n *Network) AddControllers(controllersNr uint) error {
	if controllersNr == 0 {
		return errors.New("Number of controllers to be added must be at least 1")
	}

	if int(controllersNr) > len(n.switches) {
		return errors.New("Cannot have more controllers than switches")
	}

	nodeIds := slices.Collect(maps.Keys(n.nodeIdToSw))
	randOrder, err := util.RandomFromArray(nodeIds, uint(len(nodeIds)))
	if err != nil {
		return err
	}

	slices := util.SplitArray(randOrder, controllersNr)
	for _, slice := range slices {
		switches := []*Switch{}
		for _, nodeId := range slice {
			switches = append(switches, n.nodeIdToSw[nodeId])
		}
		n.controllers = append(n.controllers, NewController(switches))
	}

	return nil
}
