package behavior

import (
	"utwente.nl/topology-to-dynetkat-coverter/convert"
)

const (
	HOSTS_NR         = 2
	OUTSIDE_HOSTS_NR = 2
	CONTROLLERS_NR   = 1
)

type OutsideHostConn struct{}

func (_ *OutsideHostConn) ModifyNetwork(n *convert.Network) error {
	err := n.AddAndConnectHosts(HOSTS_NR)
	if err != nil {
		return err
	}

	err = n.AddControllers(CONTROLLERS_NR)
	if err != nil {
		return err
	}

	_, err = n.CreateHosts(OUTSIDE_HOSTS_NR)
	if err != nil {
		return err
	}

	// TODO the controllers will modify the network to
	// allow each new host to communicate to all hosts
	// already in the network (communication between the new hosts is not necessary at this point)

	return nil
}
