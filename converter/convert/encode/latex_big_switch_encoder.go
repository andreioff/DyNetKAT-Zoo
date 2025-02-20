package encode

import (
	"fmt"
	"strings"

	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

const (
	LATEX_FLOW_MOD_SET_NAME    = "FM"
	LATEX_BIG_SWITCH_BASE_NAME = "SDN"
	LATEX_VAR_BASE_NAME        = "X"
	LATEX_VAR_I                = "i"
	LATEX_DOTS_SYM             = "\\, \\ldots \\, "
	LATEX_COMMA_SYM            = ",\\, "
	LATEX_FT_SET_NAME          = "FT"
	LATEX_OPEN_CURLY           = "\\{"
	LATEX_CLOSED_CURLY         = "\\}"
)

type LatexBigSwitchEncoder struct {
	sym             SymbolEncoding
	proactiveSwitch bool
}

func NewLatexBigSwitchEncoder(proactiveSwitch bool) NetworkEncoder {
	return NewLatexEncoder(proactiveSwitch, LatexBigSwitchEncoder{
		sym:             DYNETKAT_LATEX_SYMBOLS,
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
) string {
	if ft == nil {
		return f.sym.ZERO
	}

	swName := f.encodeSwitchName(swIndex, false)

	fmtFlowRules := f.encodeFlowTable(*ft)
	if fmtFlowRules == "" {
		fmtFlowRules = f.sym.ZERO
	}

	return fmt.Sprintf("%s & %s & %s %s", swName, f.sym.DEF, fmtFlowRules, LATEX_NEW_LN)
}

func (f LatexBigSwitchEncoder) encodeSwitchNewFT(swIndex int, newFT *convert.FlowTable) string {
	if newFT == nil {
		return f.sym.ZERO
	}

	newSwName := f.encodeSwitchName(swIndex, true)
	updatedSwStrs := f.encodeFlowTable(*newFT)

	if updatedSwStrs == "" {
		updatedSwStrs = f.sym.ZERO
	}
	return fmt.Sprintf("%s & %s & %s%s", newSwName, f.sym.DEF, updatedSwStrs, LATEX_NEW_LN)
}

func (f LatexBigSwitchEncoder) encodeFlowTable(
	ft convert.FlowTable,
) string {
	orSep := fmt.Sprintf(" %s %s& & ", f.sym.OR, LATEX_NEW_LN)
	return ft.ToNetKATStr(f.sym.AND, f.sym.EQ, f.sym.ASSIGN, orSep)
}

func (f LatexBigSwitchEncoder) encodeLinkTerm(ei EncodingInfo) string {
	var sb strings.Builder

	prefix := ""
	for _, link := range ei.links {
		sb.WriteString(prefix)
		sb.WriteString(link.ToString(f.sym.AND, f.sym.EQ, f.sym.ASSIGN))
		prefix = fmt.Sprintf(" %s %s& & ", f.sym.OR, LATEX_NEW_LN)
	}

	return fmt.Sprintf(
		"%s & %s & %s %s",
		LINK_TERM_NAME,
		f.sym.DEF,
		sb.String(),
		LATEX_DNEW_LN,
	)
}

func (f LatexBigSwitchEncoder) encodeBigSwitchTerm(
	ei EncodingInfo,
) string {
	n := ei.usedSwitchFTs.Len()

	bigSwitchName := f.encodeBigSwitchName(LATEX_VAR_BASE_NAME, n, -1, "")
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
		LATEX_NEW_LN,
	)
}

func (f LatexBigSwitchEncoder) encodeBigSwitchName(
	varName string,
	n, index int,
	termName string,
) string {
	if n < 0 {
		return LATEX_BIG_SWITCH_BASE_NAME
	}

	if index < 0 || index > n-1 {
		return fmt.Sprintf(
			"%s_{%s}",
			LATEX_BIG_SWITCH_BASE_NAME,
			f.encodeDottedSequence(0, n, varName),
		)
	}

	commaBefore, commaAfter := LATEX_COMMA_SYM, LATEX_COMMA_SYM
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
		LATEX_BIG_SWITCH_BASE_NAME,
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
		dotsStr = LATEX_DOTS_SYM + LATEX_COMMA_SYM
	}

	fmtVars := fmt.Sprintf("%s%d", varName, startIndex)
	if n > 1 {
		fmtVars += fmt.Sprintf("%s %s %s%d", LATEX_COMMA_SYM, dotsStr, varName, endIndex-1)
	}

	return fmtVars
}

func (f LatexBigSwitchEncoder) encodePacketProcPolicy(n int, bigSwitchName string) string {
	if n < 1 {
		return fmt.Sprintf("%s^{%s} %s %s", LINK_TERM_NAME, f.sym.STAR, f.sym.SEQ, bigSwitchName)
	}

	dotsStr := ""
	if n > 2 {
		dotsStr = LATEX_DOTS_SYM + f.sym.OR
	}

	concatVarsStr := LATEX_VAR_BASE_NAME + "0"
	if n > 1 {
		concatVarsStr += fmt.Sprintf("%s %s %s%d", f.sym.OR, dotsStr, LATEX_VAR_BASE_NAME, n-1)
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

		swIndex := ei.GetSwIndex(swId)
		newSwName := f.encodeSwitchName(swIndex, true)
		commStr := fmt.Sprintf(
			"%s%d %s %s %s %s",
			UP_CHANNEL_NAME,
			swIndex,
			f.sym.RECV,
			newSwName,
			f.sym.SEQ,
			f.encodeBigSwitchName(LATEX_VAR_BASE_NAME, ei.usedSwitchFTs.Len(), swIndex, newSwName),
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

	for i := range ei.usedContUpdates {
		sb.WriteString(fmt.Sprintf("%s %s%d", f.sym.PAR, CONTROLLER_BASE_NAME, i))
	}

	return fmt.Sprintf(
		"SDN & %s & %s",
		f.sym.DEF,
		util.BreakColumn(sb.String(), LATEX_THIRD_COL_MAX_LEN, LATEX_NEW_LN+"& & "),
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

	update := ei.usedContUpdates[cIndex]
	for pair := update.flowTables.Oldest(); pair != nil; pair = pair.Next() {
		swIndex := ei.GetSwIndex(pair.Key)
		commStr := f.encodeControllerPolicyComm(cName, swIndex)
		fmtCommStrs = append(fmtCommStrs, commStr)
	}

	if f.proactiveSwitch {
		fmtCommStrs = append(fmtCommStrs, f.getActivePiPoComm(ei, false, cName)...)
	} else {
		fmtCommStrs = append(fmtCommStrs, f.getPassivePiPoComm(false, cName))
	}

	fmtC := f.joinNonDetThridColumn(fmtCommStrs)
	return fmt.Sprintf("%s & %s & %s%s", cName, f.sym.DEF, fmtC, LATEX_NEW_LN)
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
		LATEX_NEW_LN,
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

		swIndex := ei.GetSwIndex(swId)
		newSwName := f.encodeSwitchName(swIndex, true)
		if forSwitch {
			termName = f.encodeBigSwitchName(
				LATEX_VAR_BASE_NAME,
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
	// '& & ' are for placing the content in the third column of the array env
	nonDetSep := fmt.Sprintf(" %s %s& & ", f.sym.NONDET, LATEX_NEW_LN)
	return strings.Join(strs, nonDetSep)
}
