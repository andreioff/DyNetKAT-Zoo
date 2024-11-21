package encode

import (
	"fmt"
	"strings"

	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

const (
	LINK_TERM_NAME       = "L"
	PACKET_IN_CHANNEL    = "pi"
	PACKET_OUT_CHANNEL   = "pi"
	FLOW_MOD_SET_NAME    = "FM"
	BIG_SWITCH_BASE_NAME = "SDN"
	VAR_BASE_NAME        = "X"
	VAR_I                = "i"
	DOTS_SYM             = "\\, \\ldots \\, "
	COMMA_SYM            = ",\\, "
	FT_SET_NAME          = "FT"
	OPEN_CURLY           = "\\{"
	CLOSED_CURLY         = "\\}"
)

type LatexBigSwitchEncoder struct {
	sym             SymbolEncoding
	proactiveSwitch bool
	swIdToIndex     map[int64]int
}

func NewLatexBigSwitchEncoder(proactiveSwitch bool) LatexBigSwitchEncoder {
	return LatexBigSwitchEncoder{
		sym: SymbolEncoding{
			ONE:    "1",
			ZERO:   "0",
			EQ:     "=",
			OR:     "+",
			AND:    "\\cdot",
			NEG:    "\\neg",
			STAR:   "*",
			ASSIGN: "\\leftarrow",

			BOT:    "\\bot",
			SEQ:    "\\, ;\\, ",
			RECV:   "\\, ?\\, ",
			SEND:   "\\, !\\, ",
			PAR:    "\\, \\|\\, ",
			DEF:    "\\triangleq",
			NONDET: "\\, \\oplus\\,",
		},
		proactiveSwitch: proactiveSwitch,
		swIdToIndex:     make(map[int64]int), // TODO Not ideal
	}
}

func (f *LatexBigSwitchEncoder) SymbolEncodings() SymbolEncoding {
	return f.sym
}

func (f *LatexBigSwitchEncoder) ProactiveSwitch() bool {
	return f.proactiveSwitch
}

func (f *LatexBigSwitchEncoder) setSwIdToIndexMap(usedSwitches []*convert.Switch) {
	newMap := make(map[int64]int)
	for i, sw := range usedSwitches {
		newMap[sw.TopoNode().ID()] = i + 1
	}
	f.swIdToIndex = newMap
}

func getUsedSwitches(switches []*convert.Switch) []*convert.Switch {
	usedSwitches := []*convert.Switch{}

	for _, sw := range switches {
		c := sw.Controller()
		willReceiveUpdate := false
		if c != nil {
			_, willReceiveUpdate = c.NewFlowTables()[sw.TopoNode().ID()]
		}

		if len(sw.FlowTable().Entries()) > 0 || willReceiveUpdate {
			usedSwitches = append(usedSwitches, sw)
		}
	}

	return usedSwitches
}

func (f *LatexBigSwitchEncoder) Encode(n *convert.Network) (string, error) {
	if n == nil {
		return "", util.NewError(util.ErrNilArgument, "n")
	}

	usedSwitches := getUsedSwitches(n.Switches())
	f.setSwIdToIndexMap(usedSwitches)

	fmtSwitches := f.encodeSwitches(usedSwitches)
	fmtControllers, usedControllers := f.encodeControllers(n.Controllers())
	link := f.encodeLinkTerm(usedSwitches, usedControllers)

	fmtBigSwitchTerm := f.encodeBigSwitchTerm(usedSwitches, n.GetSwitchesWithUpdates())
	fmtSDNTerm := f.encodeSDNTerm(usedSwitches, usedControllers)

	arrayBlockStr := link + fmtSwitches + fmtBigSwitchTerm + fmtControllers + fmtSDNTerm
	pages := util.SliceContent(arrayBlockStr, LINES_PER_PAGE, NEW_LN)

	var sb strings.Builder
	sep := ""
	for _, page := range pages {
		sb.WriteString(sep)
		sb.WriteString(BEGIN_EQ_ARRAY)
		sb.WriteString(page)
		sb.WriteString(END_EQ_ARRAY)
		sep = NEW_PAGE
	}

	return sb.String(), nil
}

func (f *LatexBigSwitchEncoder) encodeSwitches(
	switches []*convert.Switch,
) string {
	usedSwitches := []*convert.Switch{}
	var sb strings.Builder

	for _, sw := range switches {
		c := sw.Controller()
		newFlowTable, willReceiveUpdate := convert.NewFlowTable(), false
		if c != nil {
			newFlowTable, willReceiveUpdate = c.NewFlowTables()[sw.TopoNode().ID()]
		}

		swStr := f.encodeSwitch(*sw, willReceiveUpdate)
		if swStr != "" {
			sb.WriteString(swStr)
			sb.WriteString(NEW_LN)
			usedSwitches = append(usedSwitches, sw)
		}

		if !willReceiveUpdate {
			continue
		}

		newSwName := f.encodeSwitchName(*sw, true)
		noLinksFt := newFlowTable.Filter(func(fr convert.FlowRule) bool {
			return !fr.IsLink()
		})
		fmtNewSw := f.encodeNetKATPolicies(noLinksFt.ToNetKATPolicies())
		if fmtNewSw != "" {
			sb.WriteString(fmt.Sprintf("%s & %s & %s%s", newSwName, f.sym.DEF, fmtNewSw, DNEW_LN))
		}
	}

	return sb.String()
}

func (f *LatexBigSwitchEncoder) encodeSwitch(sw convert.Switch, canBeEmpty bool) string {
	swName := f.encodeSwitchName(sw, false)

	onlyNonLinkFt := sw.FlowTable().Filter(func(fr convert.FlowRule) bool {
		return !fr.IsLink()
	})
	fmtFlowRules := f.encodeNetKATPolicies(onlyNonLinkFt.ToNetKATPolicies())

	if fmtFlowRules == "" {
		if !canBeEmpty {
			return ""
		}
		fmtFlowRules = fmt.Sprintf("%s", f.sym.ZERO)
	}

	return fmt.Sprintf("%s & %s & %s %s", swName, f.sym.DEF, fmtFlowRules, NEW_LN)
}

func (f *LatexBigSwitchEncoder) encodeNetKATPolicies(
	policies []*convert.SimpleNetKATPolicy,
) string {
	strs := []string{}
	for _, policy := range policies {
		policyStr := policy.ToString(f.sym.AND, f.sym.EQ, f.sym.ASSIGN)
		strs = append(strs, fmt.Sprintf("(%s)", policyStr))
	}

	orSep := fmt.Sprintf(" %s %s& & ", f.sym.OR, NEW_LN)
	return strings.Join(strs, orSep)
}

func (f *LatexBigSwitchEncoder) encodeLinkTerm(
	switches []*convert.Switch,
	controllers []*convert.Controller,
) string {
	linksFt := convert.NewFlowTable()
	isLinkPred := func(fr convert.FlowRule) bool {
		return fr.IsLink()
	}

	for _, sw := range switches {
		linksFt.Extend(sw.FlowTable().Filter(isLinkPred))
	}

	for _, c := range controllers {
		for _, ft := range c.NewFlowTables() {
			linksFt.Extend(ft.Filter(isLinkPred))
		}
	}

	fmtLinks := f.encodeNetKATPolicies(linksFt.ToNetKATPolicies())
	return fmt.Sprintf(
		"%s & %s & %s %s%s",
		LINK_TERM_NAME,
		f.sym.DEF,
		fmtLinks,
		NEW_LN,
		NEW_LN,
	)
}

func (f *LatexBigSwitchEncoder) encodeBigSwitchTerm(
	switches []*convert.Switch,
	switchesWithUpdates []*convert.Switch,
) string {
	n := len(switches)
	if n == 0 {
		return ""
	}

	bigSwitchName := f.encodeBigSwitchName(VAR_BASE_NAME, n, -1, "")
	packetProcPolicy := f.encodePacketProcPolicy(n, bigSwitchName)
	fmtBigSw := []string{packetProcPolicy}
	fmtBigSw = append(fmtBigSw, f.encodeSwitchPolicyComm(n, switchesWithUpdates)...)
	// TODO This should be encoded separately considering the f.proactiveSwitch flag
	fmtBigSw = append(fmtBigSw, fmt.Sprintf("%s%s%s %s %s%s%s %s %s%s",
		PACKET_IN_CHANNEL, f.sym.SEND, f.sym.ONE,
		f.sym.SEQ, PACKET_OUT_CHANNEL, f.sym.RECV, f.sym.ONE,
		f.sym.SEQ, f.encodeBigSwitchName(VAR_BASE_NAME, n, -1, ""),
		NEW_LN,
	))

	return fmt.Sprintf(
		"%s & %s & %s%s",
		bigSwitchName,
		f.sym.DEF,
		f.joinNonDetThridColumn(fmtBigSw),
		NEW_LN,
	)
}

func (f *LatexBigSwitchEncoder) encodeBigSwitchName(
	varName string,
	n, index int,
	termName string,
) string {
	if n < 1 {
		return BIG_SWITCH_BASE_NAME
	}

	if index < 1 || index > n {
		return fmt.Sprintf(
			"%s_{%s}",
			BIG_SWITCH_BASE_NAME,
			f.encodeDottedSequence(1, n, varName),
		)
	}

	commaBefore, commaAfter := COMMA_SYM, COMMA_SYM
	if index == 1 {
		commaBefore = ""
	}
	if index == n {
		commaAfter = ""
	}

	fmtVarSeq := f.encodeDottedSequence(1, index-1, varName) +
		fmt.Sprintf("%s%s%s", commaBefore, termName, commaAfter) +
		f.encodeDottedSequence(index+1, n, varName)

	return fmt.Sprintf(
		"%s_{%s}",
		BIG_SWITCH_BASE_NAME,
		fmtVarSeq,
	)
}

func (f *LatexBigSwitchEncoder) encodeDottedSequence(
	startIndex, endIndex int,
	varName string,
) string {
	n := endIndex - startIndex + 1
	if n < 1 {
		return ""
	}

	dotsStr := ""
	if n > 2 {
		dotsStr = DOTS_SYM + COMMA_SYM
	}

	fmtVars := fmt.Sprintf("%s%d", varName, startIndex)
	if n > 1 {
		fmtVars += fmt.Sprintf("%s %s %s%d", COMMA_SYM, dotsStr, varName, endIndex)
	}

	return fmtVars
}

func (f *LatexBigSwitchEncoder) encodePacketProcPolicy(n int, bigSwitchName string) string {
	if n < 1 {
		return fmt.Sprintf("%s^{%s} %s %s", LINK_TERM_NAME, f.sym.STAR, f.sym.SEQ, bigSwitchName)
	}

	dotsStr := ""
	if n > 2 {
		dotsStr = DOTS_SYM + f.sym.OR
	}

	concatVarsStr := VAR_BASE_NAME + "1"
	if n > 1 {
		concatVarsStr += fmt.Sprintf("%s %s %s%d", f.sym.OR, dotsStr, VAR_BASE_NAME, n)
	}

	return fmt.Sprintf(
		"((%s) %s %s)^{%s} %s %s",
		concatVarsStr,
		f.sym.AND,
		LINK_TERM_NAME,
		f.sym.STAR,
		f.sym.SEQ,
		bigSwitchName,
	)
}

func (f *LatexBigSwitchEncoder) encodeSwitchPolicyComm(
	n int,
	switches []*convert.Switch,
) []string {
	commStrs := []string{}
	for _, sw := range switches {
		swIndex, exists := f.swIdToIndex[sw.TopoNode().ID()]
		if !exists {
			panic("This should not happen")
		}

		newSwName := f.encodeSwitchName(*sw, true)
		commStr := fmt.Sprintf(
			"%s%d %s %s %s %s",
			UP_CHANNEL_NAME,
			swIndex,
			f.sym.RECV,
			newSwName,
			f.sym.SEQ,
			f.encodeBigSwitchName(VAR_BASE_NAME, n, swIndex, newSwName),
		)

		commStrs = append(commStrs, commStr)
	}

	return commStrs
}

func (f *LatexBigSwitchEncoder) encodeControllerPolicyComm(
	cName string,
	sw *convert.Switch,
) string {
	swIndex, exists := f.swIdToIndex[sw.TopoNode().ID()]
	if !exists {
		panic("This should not happen")
	}

	newSwName := f.encodeSwitchName(*sw, true)
	return fmt.Sprintf(
		"%s%d %s %s %s %s",
		UP_CHANNEL_NAME,
		swIndex,
		f.sym.SEND,
		newSwName,
		f.sym.SEQ,
		cName,
	)
}

func (f *LatexBigSwitchEncoder) encodeSDNTerm(
	sws []*convert.Switch,
	c []*convert.Controller,
) string {
	var sb strings.Builder

	sb.WriteString(f.encodeBigSwitchName(SW_BASE_NAME, len(sws), -1, ""))

	for _, c := range c {
		sb.WriteString(fmt.Sprintf("%s %s%d", f.sym.PAR, CONTROLLER_BASE_NAME, c.ID()))
	}

	return fmt.Sprintf(
		"SDN & %s & %s",
		f.sym.DEF,
		util.BreakColumn(sb.String(), THIRD_COL_MAX_LEN, NEW_LN+"& & "),
	)
}

func (f *LatexBigSwitchEncoder) encodeSwitchName(sw convert.Switch, isNew bool) string {
	swIndex, exists := f.swIdToIndex[sw.TopoNode().ID()]
	if !exists {
		panic("This should not happen!")
	}

	name := fmt.Sprintf("%s%d", SW_BASE_NAME, swIndex)
	if isNew {
		return "new" + name
	}
	return name
}

func (f *LatexBigSwitchEncoder) encodeControllers(
	controllers []*convert.Controller,
) (string, []*convert.Controller) {
	usedControllers := []*convert.Controller{}
	var sb strings.Builder

	for _, c := range controllers {
		cStr := f.encodeController(c)
		if cStr != "" {
			sb.WriteString(cStr)
			sb.WriteString(NEW_LN)
			usedControllers = append(usedControllers, c)
		}
	}

	return sb.String(), usedControllers
}

func (f *LatexBigSwitchEncoder) encodeController(c *convert.Controller) string {
	fmtCommStrs := []string{}
	cName := fmt.Sprintf("%s%d", CONTROLLER_BASE_NAME, c.ID())

	for swId := range c.NewFlowTables() {
		sw := c.FindSwitch(swId)

		commStr := f.encodeControllerPolicyComm(cName, sw)
		fmtCommStrs = append(fmtCommStrs, commStr)
	}

	if len(c.NewFlowTables()) == 0 {
		return ""
	}

	// TODO Can be merged with the one from the switch
	fmtCommStrs = append(fmtCommStrs, fmt.Sprintf("%s%s%s %s %s%s%s %s %s",
		PACKET_IN_CHANNEL, f.sym.RECV, f.sym.ONE,
		f.sym.SEQ, PACKET_OUT_CHANNEL, f.sym.SEND, f.sym.ONE,
		f.sym.SEQ, cName,
	))

	fmtC := f.joinNonDetThridColumn(fmtCommStrs)
	return fmt.Sprintf("%s & %s & %s%s", cName, f.sym.DEF, fmtC, NEW_LN)
}

func (f *LatexBigSwitchEncoder) joinNonDetThridColumn(strs []string) string {
	// '& & ' are for placing the conent in the third column of the array env
	nonDetSep := fmt.Sprintf(" %s %s& & ", f.sym.NONDET, NEW_LN)
	return strings.Join(strs, nonDetSep)
}
