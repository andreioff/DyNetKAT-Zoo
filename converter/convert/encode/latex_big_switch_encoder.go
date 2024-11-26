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
	PACKET_OUT_CHANNEL   = "po"
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
}

func NewLatexBigSwitchEncoder(proactiveSwitch bool) NetworkEncoder {
	return NewLatexEncoder(proactiveSwitch, LatexBigSwitchEncoder{
		sym:             LATEX_SYMBOLS,
		proactiveSwitch: proactiveSwitch,
	})
}

func (f LatexBigSwitchEncoder) SymbolEncoding() SymbolEncoding {
	return f.sym
}

func (f LatexBigSwitchEncoder) ProactiveSwitch() bool {
	return f.proactiveSwitch
}

func (f LatexBigSwitchEncoder) encodeInformation(ei EncodingInfo) string {
	link := f.encodeLinkTerm(ei)

	fmtBigSwitchTerm := f.encodeBigSwitchTerm(ei)
	fmtSDNTerm := f.encodeSDNTerm(ei)

	return link + fmtBigSwitchTerm + fmtSDNTerm
}

func (f LatexBigSwitchEncoder) encodeSwitch(
	swIndex int,
	ft *convert.FlowTable,
	canBeEmpty bool,
) string {
	swName := f.encodeSwitchName(swIndex, false)

	onlyNonLinkFt := ft.Filter(func(fr convert.FlowRule) bool {
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

func (f LatexBigSwitchEncoder) encodeSwitchNewFT(swIndex int, newFT *convert.FlowTable) string {
	newSwName := f.encodeSwitchName(swIndex, true)
	noLinksFt := newFT.Filter(func(fr convert.FlowRule) bool {
		return !fr.IsLink()
	})
	updatedSwStrs := f.encodeNetKATPolicies(noLinksFt.ToNetKATPolicies())
	if updatedSwStrs == "" {
		updatedSwStrs = f.sym.BOT
	}
	return fmt.Sprintf("%s & %s & %s%s", newSwName, f.sym.DEF, updatedSwStrs, NEW_LN)
}

func (f LatexBigSwitchEncoder) encodeNetKATPolicies(
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

func (f LatexBigSwitchEncoder) encodeLinkTerm(ei EncodingInfo) string {
	linksFt := convert.NewFlowTable()
	isLinkPred := func(fr convert.FlowRule) bool {
		return fr.IsLink()
	}

	for pair := ei.usedSwitchFTs.Oldest(); pair != nil; pair = pair.Next() {
		linksFt.Extend(pair.Value.Filter(isLinkPred))
	}

	for _, c := range ei.usedContFTs {
		for pair := c.Oldest(); pair != nil; pair = pair.Next() {
			linksFt.Extend(pair.Value.Filter(isLinkPred))
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

func (f LatexBigSwitchEncoder) encodeBigSwitchTerm(
	ei EncodingInfo,
) string {
	n := ei.usedSwitchFTs.Len()

	bigSwitchName := f.encodeBigSwitchName(VAR_BASE_NAME, n, -1, "")
	packetProcPolicy := f.encodePacketProcPolicy(n, bigSwitchName)
	fmtBigSw := []string{packetProcPolicy}
	fmtBigSw = append(fmtBigSw, f.encodeSwitchPolicyComm(ei)...)

	if f.proactiveSwitch {
		fmtBigSw = append(fmtBigSw, f.getActivePiPoComm(ei, true, "")...)
	} else {
		fmtBigSw = append(fmtBigSw, f.getPassivePiPoComm(true, bigSwitchName))
	}

	return fmt.Sprintf(
		"%s & %s & %s%s",
		bigSwitchName,
		f.sym.DEF,
		f.joinNonDetThridColumn(fmtBigSw),
		NEW_LN,
	)
}

func (f LatexBigSwitchEncoder) encodeBigSwitchName(
	varName string,
	n, index int,
	termName string,
) string {
	if n < 0 {
		return BIG_SWITCH_BASE_NAME
	}

	if index < 0 || index > n-1 {
		return fmt.Sprintf(
			"%s_{%s}",
			BIG_SWITCH_BASE_NAME,
			f.encodeDottedSequence(0, n, varName),
		)
	}

	commaBefore, commaAfter := COMMA_SYM, COMMA_SYM
	if index == 0 {
		commaBefore = ""
	}
	if index == n-1 {
		commaAfter = ""
	}

	fmtVarSeq := f.encodeDottedSequence(0, index, varName) +
		fmt.Sprintf("%s%s%s", commaBefore, termName, commaAfter) +
		f.encodeDottedSequence(index+1, n, varName)

	return fmt.Sprintf(
		"%s_{%s}",
		BIG_SWITCH_BASE_NAME,
		fmtVarSeq,
	)
}

func (f LatexBigSwitchEncoder) encodeDottedSequence(
	startIndex, endIndex int,
	varName string,
) string {
	n := endIndex - startIndex
	if n < 1 {
		return ""
	}

	dotsStr := ""
	if n > 2 {
		dotsStr = DOTS_SYM + COMMA_SYM
	}

	fmtVars := fmt.Sprintf("%s%d", varName, startIndex)
	if n > 1 {
		fmtVars += fmt.Sprintf("%s %s %s%d", COMMA_SYM, dotsStr, varName, endIndex-1)
	}

	return fmtVars
}

func (f LatexBigSwitchEncoder) encodePacketProcPolicy(n int, bigSwitchName string) string {
	if n < 1 {
		return fmt.Sprintf("%s^{%s} %s %s", LINK_TERM_NAME, f.sym.STAR, f.sym.SEQ, bigSwitchName)
	}

	dotsStr := ""
	if n > 2 {
		dotsStr = DOTS_SYM + f.sym.OR
	}

	concatVarsStr := VAR_BASE_NAME + "0"
	if n > 1 {
		concatVarsStr += fmt.Sprintf("%s %s %s%d", f.sym.OR, dotsStr, VAR_BASE_NAME, n-1)
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

func (f LatexBigSwitchEncoder) encodeSwitchPolicyComm(
	ei EncodingInfo,
) []string {
	commStrs := []string{}
	for pair := ei.usedSwitchFTs.Oldest(); pair != nil; pair = pair.Next() {
		swId := pair.Key

		_, exists := ei.FindNewFT(swId)
		if !exists {
			continue
		}

		swIndex, _ := ei.nodeIdToIndex.Get(swId)
		newSwName := f.encodeSwitchName(swIndex, true)
		commStr := fmt.Sprintf(
			"%s%d %s %s %s %s",
			UP_CHANNEL_NAME,
			swIndex,
			f.sym.RECV,
			newSwName,
			f.sym.SEQ,
			f.encodeBigSwitchName(VAR_BASE_NAME, ei.usedSwitchFTs.Len(), swIndex, newSwName),
		)

		commStrs = append(commStrs, commStr)
	}

	return commStrs
}

func (f LatexBigSwitchEncoder) encodeControllerPolicyComm(
	cName string,
	swIndex int,
) string {
	newSwName := f.encodeSwitchName(swIndex, true)
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

func (f LatexBigSwitchEncoder) encodeSDNTerm(
	ei EncodingInfo,
) string {
	var sb strings.Builder

	sb.WriteString(f.encodeBigSwitchName(SW_BASE_NAME, ei.usedSwitchFTs.Len(), -1, ""))

	for i := range ei.usedContFTs {
		sb.WriteString(fmt.Sprintf("%s %s%d", f.sym.PAR, CONTROLLER_BASE_NAME, i))
	}

	return fmt.Sprintf(
		"SDN & %s & %s",
		f.sym.DEF,
		util.BreakColumn(sb.String(), THIRD_COL_MAX_LEN, NEW_LN+"& & "),
	)
}

func (f LatexBigSwitchEncoder) encodeSwitchName(swIndex int, isNew bool) string {
	name := fmt.Sprintf("%s%d", SW_BASE_NAME, swIndex)
	if isNew {
		return "new" + name
	}
	return name
}

func (f LatexBigSwitchEncoder) encodeController(
	ei EncodingInfo,
	cIndex int,
) string {
	fmtCommStrs := []string{}
	cName := fmt.Sprintf("%s%d", CONTROLLER_BASE_NAME, cIndex)

	for pair := ei.usedContFTs[cIndex].Oldest(); pair != nil; pair = pair.Next() {
		swIndex, _ := ei.nodeIdToIndex.Get(pair.Key)
		commStr := f.encodeControllerPolicyComm(cName, swIndex)
		fmtCommStrs = append(fmtCommStrs, commStr)
	}

	if f.proactiveSwitch {
		fmtCommStrs = append(fmtCommStrs, f.getActivePiPoComm(ei, false, cName)...)
	} else {
		fmtCommStrs = append(fmtCommStrs, f.getPassivePiPoComm(false, cName))
	}

	fmtC := f.joinNonDetThridColumn(fmtCommStrs)
	return fmt.Sprintf("%s & %s & %s%s", cName, f.sym.DEF, fmtC, NEW_LN)
}

func (f LatexBigSwitchEncoder) getPassivePiPoComm(
	forSwitch bool,
	termName string,
) string {
	commSym1 := f.sym.RECV
	commSym2 := f.sym.SEND
	if forSwitch {
		commSym1 = f.sym.SEND
		commSym2 = f.sym.RECV
	}

	return fmt.Sprintf("%s%s%s %s %s%s%s %s %s%s",
		PACKET_IN_CHANNEL, commSym1, f.sym.ONE,
		f.sym.SEQ, PACKET_OUT_CHANNEL, commSym2, f.sym.ONE,
		f.sym.SEQ, termName,
		NEW_LN,
	)
}

func (f LatexBigSwitchEncoder) getActivePiPoComm(
	ei EncodingInfo,
	forSwitch bool,
	termName string,
) []string {
	commSym1 := f.sym.RECV
	commSym2 := f.sym.SEND
	if forSwitch {
		commSym1 = f.sym.SEND
		commSym2 = f.sym.RECV
	}

	commStrs := []string{}
	for pair := ei.usedSwitchFTs.Oldest(); pair != nil; pair = pair.Next() {
		swId := pair.Key

		_, exists := ei.FindNewFT(swId)
		if !exists {
			continue
		}

		swIndex, _ := ei.nodeIdToIndex.Get(swId)
		newSwName := f.encodeSwitchName(swIndex, true)
		if forSwitch {
			termName = f.encodeBigSwitchName(
				VAR_BASE_NAME,
				ei.usedSwitchFTs.Len(),
				swIndex,
				newSwName,
			)
		}

		commStr := fmt.Sprintf(
			"%s%d %s %s %s %s%d %s %s %s %s",
			PACKET_IN_CHANNEL,
			swIndex,
			commSym1,
			f.sym.ONE,
			f.sym.SEQ,
			PACKET_OUT_CHANNEL,
			swIndex,
			commSym2,
			newSwName,
			f.sym.SEQ,
			termName,
		)

		commStrs = append(commStrs, commStr)
	}

	return commStrs
}

func (f LatexBigSwitchEncoder) joinNonDetThridColumn(strs []string) string {
	// '& & ' are for placing the conent in the third column of the array env
	nonDetSep := fmt.Sprintf(" %s %s& & ", f.sym.NONDET, NEW_LN)
	return strings.Join(strs, nonDetSep)
}
