package behavior

import "utwente.nl/topology-to-dynetkat-coverter/convert"

const HOST_NR = 5

type OutsideHostConn struct{}

func (_ *OutsideHostConn) ModifyNetwork(n *convert.Network) error {
	err := n.AddAndConnectHosts(HOST_NR)
	if err != nil {
		return err
	}
	return nil
}
