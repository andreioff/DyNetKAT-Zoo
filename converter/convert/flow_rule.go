package convert

type FlowRule struct {
	inPort  int64
	outPort int64
	isLink  bool
}

func (fr FlowRule) InPort() int64 {
	return fr.inPort
}

func (fr FlowRule) OutPort() int64 {
	return fr.outPort
}

func (fr FlowRule) IsLink() bool {
	return fr.isLink
}

func (fr1 FlowRule) IsEqual(fr2 FlowRule) bool {
	return fr1.inPort == fr2.inPort &&
		fr1.outPort == fr2.outPort &&
		fr1.isLink == fr2.isLink
}

func NewFlowRule(inPort, outPort int64, isLink bool) FlowRule {
	return FlowRule{
		inPort:  inPort,
		outPort: outPort,
		isLink:  isLink,
	}
}
