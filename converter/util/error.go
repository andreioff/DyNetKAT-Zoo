package util

import (
	"errors"
	"fmt"
)

func NewError(errorStr string, args ...any) error {
	fmtErrorStr := fmt.Sprintf(errorStr, args...)
	return errors.New(fmtErrorStr)
}

const (
	ErrNilArgument                 = "Argument '%s' is nil!"
	ErrEmptyGraph                  = "Empty graph!"
	ErrDisconnGraph                = "Disconnected graph with %d components!"
	ErrNodeNotInGraph              = "Node is not part of the graph!"
	ErrNoSwitchWithNodeId          = "No switch matches the node id '%d'"
	ErrNoLinkBetweenSwitches       = "Could not find link between switches!"
	ErrMorePicksThanUniqueElements = "No. of random picks is greater than the no. of unique elements in the array."
	ErrGraphMLExactly1Graph        = "GraphML instance must contain exactly 1 graph!"
	ErrSwitchHasNilController      = "Switch has nil controller!"
	ErrEdgeNotMappedToLink         = "Edge is not mapped to a link!"
	ErrNetworkHasNoSwitches        = "Network has no switches!"
	ErrNoPathBetweenSwitches       = "Could not find path between switches!"
	ErrHostsNrAtLeast2             = "Number of hosts must be at least 2!"
	ErrControllersNrAtLeast1       = "Number of controllers to be added must be at least 1!"
	ErrMoreContsThanSwitches       = "Cannot add more controllers than switches to the network!"
	ErrZeroOrNegDivisionLength     = "Zero or negative division length!"
	ErrOnlyIncidentLinksForSwitch  = "Switch must receive only links that are connected to it!"
	ErrNilInArray                  = "Found nil elements in array '%s'"
	ErrEmptyStringVar              = "Empty string variable '%s'!"
)
