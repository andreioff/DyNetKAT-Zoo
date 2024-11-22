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
	}
}

func (f *LatexBigSwitchEncoder) SymbolEncodings() SymbolEncoding {
	return f.sym
}

func (f *LatexBigSwitchEncoder) ProactiveSwitch() bool {
	return f.proactiveSwitch
}

func (f *LatexBigSwitchEncoder) Encode(ei EncodingInfo) (string, error) {
	fmtSwitches := f.encodeSwitches(ei)
	fmtControllers := f.encodeControllers(ei)
	link := f.encodeLinkTerm(ei)

	fmtBigSwitchTerm := f.encodeBigSwitchTerm(ei)
	fmtSDNTerm := f.encodeSDNTerm(ei)

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
	ei EncodingInfo,
) string {
	var sb strings.Builder

	for swId, ft := range ei.usedSwitchFTs {
		newFt, willReceiveUpdate := ei.FindNewFT(swId)

		swIndex := ei.nodeIdToIndex[swId]
		swStr := f.encodeSwitch(swIndex, ft, willReceiveUpdate)
		sb.WriteString(swStr)
		sb.WriteString(NEW_LN)

		if !willReceiveUpdate {
			continue
		}

		newSwName := f.encodeSwitchName(swIndex, true)
		noLinksFt := newFt.Filter(func(fr convert.FlowRule) bool {
			return !fr.IsLink()
		})
		fmtNewSw := f.encodeNetKATPolicies(noLinksFt.ToNetKATPolicies())
		if fmtNewSw != "" {
			sb.WriteString(fmt.Sprintf("%s & %s & %s%s", newSwName, f.sym.DEF, fmtNewSw, DNEW_LN))
		}
	}

	return sb.String()
}

func (f *LatexBigSwitchEncoder) encodeSwitch(
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

func (f *LatexBigSwitchEncoder) encodeLinkTerm(ei EncodingInfo) string {
	linksFt := convert.NewFlowTable()
	isLinkPred := func(fr convert.FlowRule) bool {
		return fr.IsLink()
	}

	for _, ft := range ei.usedSwitchFTs {
		linksFt.Extend(ft.Filter(isLinkPred))
	}

	for _, c := range ei.usedContFTs {
		for _, ft := range c {
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
	ei EncodingInfo,
) string {
	n := len(ei.usedSwitchFTs)

	bigSwitchName := f.encodeBigSwitchName(VAR_BASE_NAME, n, -1, "")
	packetProcPolicy := f.encodePacketProcPolicy(n, bigSwitchName)
	fmtBigSw := []string{packetProcPolicy}
	fmtBigSw = append(fmtBigSw, f.encodeSwitchPolicyComm(ei)...)
	fmtBigSw = append(fmtBigSw, f.getPassivePiPoComm(true, bigSwitchName))

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
	ei EncodingInfo,
) []string {
	commStrs := []string{}
	for swId := range ei.usedSwitchFTs {
		_, exists := ei.FindNewFT(swId)
		if !exists {
			continue
		}

		swIndex := ei.nodeIdToIndex[swId]
		newSwName := f.encodeSwitchName(swIndex, true)
		commStr := fmt.Sprintf(
			"%s%d %s %s %s %s",
			UP_CHANNEL_NAME,
			swIndex,
			f.sym.RECV,
			newSwName,
			f.sym.SEQ,
			f.encodeBigSwitchName(VAR_BASE_NAME, len(ei.usedSwitchFTs), swIndex, newSwName),
		)

		commStrs = append(commStrs, commStr)
	}

	return commStrs
}

func (f *LatexBigSwitchEncoder) encodeControllerPolicyComm(
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

func (f *LatexBigSwitchEncoder) encodeSDNTerm(
	ei EncodingInfo,
) string {
	var sb strings.Builder

	sb.WriteString(f.encodeBigSwitchName(SW_BASE_NAME, len(ei.usedSwitchFTs), -1, ""))

	for i := range ei.usedContFTs {
		sb.WriteString(fmt.Sprintf("%s %s%d", f.sym.PAR, CONTROLLER_BASE_NAME, i))
	}

	return fmt.Sprintf(
		"SDN & %s & %s",
		f.sym.DEF,
		util.BreakColumn(sb.String(), THIRD_COL_MAX_LEN, NEW_LN+"& & "),
	)
}

func (f *LatexBigSwitchEncoder) encodeSwitchName(swIndex int, isNew bool) string {
	name := fmt.Sprintf("%s%d", SW_BASE_NAME, swIndex)
	if isNew {
		return "new" + name
	}
	return name
}

func (f *LatexBigSwitchEncoder) encodeControllers(ei EncodingInfo) string {
	var sb strings.Builder

	for i := range ei.usedContFTs {
		cStr := f.encodeController(ei, i)
		sb.WriteString(cStr)
		sb.WriteString(NEW_LN)
	}

	return sb.String()
}

func (f *LatexBigSwitchEncoder) encodeController(
	ei EncodingInfo,
	cIndex int,
) string {
	fmtCommStrs := []string{}
	cName := fmt.Sprintf("%s%d", CONTROLLER_BASE_NAME, cIndex)

	for swId := range ei.usedContFTs[cIndex] {
		commStr := f.encodeControllerPolicyComm(cName, ei.nodeIdToIndex[swId])
		fmtCommStrs = append(fmtCommStrs, commStr)
	}

	fmtCommStrs = append(fmtCommStrs, f.getPassivePiPoComm(false, cName))

	fmtC := f.joinNonDetThridColumn(fmtCommStrs)
	return fmt.Sprintf("%s & %s & %s%s", cName, f.sym.DEF, fmtC, NEW_LN)
}

func (f *LatexBigSwitchEncoder) getPassivePiPoComm(
	forSwitch bool,
	termName string,
) string {
	// TODO This should also be encoded separately considering the f.proactiveSwitch flag
	commSym := f.sym.RECV
	if forSwitch {
		commSym = f.sym.SEND
	}

	return fmt.Sprintf("%s%s%s %s %s%s%s %s %s%s",
		PACKET_IN_CHANNEL, commSym, f.sym.ONE,
		f.sym.SEQ, PACKET_OUT_CHANNEL, f.sym.RECV, f.sym.ONE,
		f.sym.SEQ, termName,
		NEW_LN,
	)
}

func (f *LatexBigSwitchEncoder) joinNonDetThridColumn(strs []string) string {
	// '& & ' are for placing the conent in the third column of the array env
	nonDetSep := fmt.Sprintf(" %s %s& & ", f.sym.NONDET, NEW_LN)
	return strings.Join(strs, nonDetSep)
}
