package convert

import (
	"errors"
	"fmt"
	"math/rand"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

// TODO Check if it is better to put these into an init function
var (
	SEED    int64     = 3
	randGen rand.Rand = *rand.New(rand.NewSource(SEED))
)

type Network struct {
	topology      util.Graph
	shortestPaths map[int64]map[int64][]*Switch

	switches []Switch
	hosts    []Host
	portNr   int64
	hostId   int64
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
		portNr:        portNr,
		hostId:        0,
		hosts:         []Host{},
	}, nil
}

func (n *Network) PortNr() int64 {
	return n.portNr
}

func makeSwitchesFromTopology(topo util.Graph, portNr *int64) ([]Switch, error) {
	if portNr == nil {
		return []Switch{}, errors.New("Nil portNr argument!")
	}

	switches := []Switch{}

	iter := topo.Nodes()
	for iter.Next() {
		links, err := makeLinks(topo, iter.Node(), portNr)
		if err != nil {
			return []Switch{}, err
		}

		newSw := *NewSwitch(iter.Node(), links)
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

func mapNodeToSwitch(switches []Switch) map[int64]*Switch {
	nodeIdToSwitch := make(map[int64]*Switch, len(switches))

	for _, sw := range switches {
		nodeIdToSwitch[sw.topoNode.ID()] = &sw
	}

	return nodeIdToSwitch
}

func nodePathToSwitchPath(path []graph.Node, nodeToSwitch map[int64]*Switch) []*Switch {
	switchPath := []*Switch{}
	for _, node := range path {
		switchPath = append(switchPath, nodeToSwitch[node.ID()])
	}
	return switchPath
}

func computeShortestPaths(topo util.Graph, switches []Switch) map[int64]map[int64][]*Switch {
	nodePaths := path.DijkstraAllPaths(&topo)
	nodeToSwitch := mapNodeToSwitch(switches)
	switchPaths := make(map[int64]map[int64][]*Switch)

	for i := range len(switches) {
		sw1Id := switches[i].topoNode.ID()
		switchPaths[sw1Id] = make(map[int64][]*Switch)

		for j := range len(switches) {
			if i == j {
				continue
			}

			sw2Id := switches[j].topoNode.ID()
			nodePath, _, _ := nodePaths.Between(sw1Id, sw2Id)
			switchPaths[sw1Id][sw2Id] = nodePathToSwitchPath(nodePath, nodeToSwitch)
		}
	}
	return switchPaths
}

func (n *Network) assignHosts(hostsNr int64) error {
	switchesLen := len(n.switches)
	hosts := []Host{}

	for range hostsNr {
		randSw := &n.switches[randGen.Intn(switchesLen)]

		newHost, err := NewHost(n.hostId, n.portNr, randSw)
		if err != nil {
			return err
		}
		hosts = append(hosts, newHost)
		randSw.AddHost(&newHost)

		n.hostId++
		n.portNr++
	}
	n.hosts = hosts

	return nil
}

func (n *Network) populateDestinationTables(h1, h2 *Host) error {
	if h1 == nil || h2 == nil {
		return errors.New("Null arguments!")
	}

	path := n.shortestPaths[h1.sw.topoNode.ID()][h2.sw.topoNode.ID()]

	receivingPort := h1.SwitchPort()
	for i := range len(path) - 1 {
		currSw := path[i]
		nextSwId := path[i+1].topoNode.ID()
		link := currSw.FindLink(nextSwId)

		currSw.destTable[h2.ID()] = make(map[int64]int64)
		currSw.destTable[h2.ID()][receivingPort] = link.FromPort()
		currSw.destTable[h2.ID()][link.FromPort()] = link.ToPort()
		receivingPort = link.ToPort()
	}

	// TODO Not a huge fan of this. To be refactored...
	h2.sw.destTable[h2.ID()] = make(map[int64]int64)
	h2.sw.destTable[h2.ID()][receivingPort] = h2.switchPort
	return nil
}

func (n *Network) AddAndConnectHosts(hostsNr int64) error {
	if hostsNr < 2 {
		return errors.New("Number of hosts must be at least 2!")
	}

	err := n.assignHosts(hostsNr)
	if err != nil {
		return err
	}

	for i := range len(n.hosts) {
		for j := i + 1; j < len(n.hosts); j++ {
			err := n.populateDestinationTables(&n.hosts[i], &n.hosts[j])
			if err != nil {
				return err
			}

			err = n.populateDestinationTables(&n.hosts[j], &n.hosts[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *Network) String() string {
	str := ""

	for _, h := range n.hosts {
		str += fmt.Sprintf("H%3d: Sw: %3d, Port: %3d\n", h.ID(), h.sw.topoNode.ID(), h.SwitchPort())
	}

	str += "\n\n"

	tab := "       "
	for _, sw := range n.switches {
		str += fmt.Sprintf("SW%3d: \n", sw.topoNode.ID())
		for dstHostId, inPortToOutPort := range sw.destTable {
			for inPort, outPort := range inPortToOutPort {
				str += fmt.Sprintf(
					"%s(dst = %d) . (port = %d) . (port <- %d) + \n",
					tab,
					dstHostId,
					inPort,
					outPort,
				)
			}
		}
	}

	return str
}
