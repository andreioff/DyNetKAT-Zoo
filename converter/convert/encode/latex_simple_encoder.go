package encode

import (
	"fmt"
	"strings"

	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

const (
	NEW_LN               = "\\\\\n"
	DNEW_LN              = NEW_LN + NEW_LN
	NEW_PAGE             = "\n\\newpage\n"
	BEGIN_EQ_ARRAY       = "\\begin{equation} \\begin{array}{rcl}\n\n"
	END_EQ_ARRAY         = "\n\n\\end{array} \\end{equation}\n"
	THIRD_COL_MAX_LEN    = 40 // nr of chars before the third column of the array env overflows
	LINES_PER_PAGE       = 40
	SW_BASE_NAME         = "SW"
	CONTROLLER_BASE_NAME = "C"
	UP_CHANNEL_NAME      = "Up"
	HELP_CHANNEL_NAME    = "Help"
)

type LatexSimpleEncoder struct {
	sym             SymbolEncoding
	proactiveSwitch bool
}

func NewLatexSimpleEncoder(proactiveSwitch bool) NetworkEncoder {
	return NewLatexEncoder(proactiveSwitch, LatexSimpleEncoder{sym: LATEX_SYMBOLS})
}

func (f LatexSimpleEncoder) SymbolEncoding() SymbolEncoding {
	return f.sym
}

func (f LatexSimpleEncoder) ProactiveSwitch() bool {
	return f.proactiveSwitch
}

func (f LatexSimpleEncoder) encodeInformation(ei EncodingInfo) string {
	return f.encodeSDNTerm(ei)
}

func (f LatexSimpleEncoder) encodeSwitchNewFT(swIndex int, newFT *convert.FlowTable) string {
	newSwName := f.encodeSwitchName(swIndex, true)
	updatedSwStrs := f.encodeNetKATPolicies(newFT.ToNetKATPolicies(), newSwName)
	if len(updatedSwStrs) == 0 {
		return fmt.Sprintf("%s & %s & %s%s", newSwName, f.sym.DEF, f.sym.BOT, NEW_LN)
	}
	fmtNewSw := f.joinNonDetThridColumn(updatedSwStrs)
	return fmt.Sprintf("%s & %s & %s%s", newSwName, f.sym.DEF, fmtNewSw, NEW_LN)
}

func (f LatexSimpleEncoder) encodeSwitch(
	swIndex int,
	ft *convert.FlowTable,
	canBeEmpty bool,
) string {
	swName := f.encodeSwitchName(swIndex, false)

	fmtFlowRules := f.encodeNetKATPolicies(ft.ToNetKATPolicies(), swName)

	if len(fmtFlowRules) == 0 {
		if !canBeEmpty {
			return ""
		}
		dropAllStr := fmt.Sprintf("%s%s%s", f.sym.ZERO, f.sym.SEQ, swName)
		fmtFlowRules = append(fmtFlowRules, dropAllStr)
	}

	commStr := f.encodeCommunication(f.encodeSwitchName(swIndex, true), swIndex, false)
	fmtFlowRules = append(fmtFlowRules, commStr)

	fmtSw := f.joinNonDetThridColumn(fmtFlowRules)
	return fmt.Sprintf("%s & %s & %s %s", swName, f.sym.DEF, fmtSw, NEW_LN)
}

func (f LatexSimpleEncoder) encodeNetKATPolicies(
	policies []*convert.SimpleNetKATPolicy,
	swName string,
) []string {
	fmtFlowRules := []string{}
	for _, policy := range policies {
		fmtFlowRules = append(fmtFlowRules, fmt.Sprintf(
			"(%s) %s %s",
			policy.ToString(f.sym.AND, f.sym.EQ, f.sym.ASSIGN),
			f.sym.SEQ, swName,
		))
	}

	return fmtFlowRules
}

func (f LatexSimpleEncoder) encodeCommunication(
	termName string,
	channelId int,
	fromSwitch bool,
) string {
	upCommSym := f.sym.SEND
	helpCommSym := f.sym.RECV
	if fromSwitch {
		upCommSym = f.sym.RECV
		helpCommSym = f.sym.SEND
	}

	commStr := fmt.Sprintf(
		"%s%d %s %s %s %s",
		UP_CHANNEL_NAME,
		channelId,
		upCommSym,
		f.sym.ONE,
		f.sym.SEQ,
		termName,
	)

	if !f.proactiveSwitch {
		return commStr
	}

	return fmt.Sprintf("%s%d%s%s %s %s",
		HELP_CHANNEL_NAME,
		channelId,
		helpCommSym,
		f.sym.ONE,
		f.sym.SEQ,
		commStr,
	)
}

func (f LatexSimpleEncoder) encodeSDNTerm(ei EncodingInfo) string {
	var sb strings.Builder

	prefix := ""
	for pair := ei.usedSwitchFTs.Oldest(); pair != nil; pair = pair.Next() {
		swIndex, _ := ei.nodeIdToIndex.Get(pair.Key)
		sb.WriteString(prefix + f.encodeSwitchName(swIndex, false))
		prefix = f.sym.PAR
	}

	for i := range ei.usedContFTs {
		sb.WriteString(prefix + fmt.Sprintf("%s%d", CONTROLLER_BASE_NAME, i))
	}
	content := util.BreakColumn(sb.String(), THIRD_COL_MAX_LEN, NEW_LN+"& & ")
	return fmt.Sprintf("SDN & %s & %s", f.sym.DEF, content)
}

func (f LatexSimpleEncoder) encodeSwitchName(swIndex int, isNew bool) string {
	name := fmt.Sprintf("%s%d", SW_BASE_NAME, swIndex)
	if isNew {
		return name + "'"
	}
	return name
}

func (f LatexSimpleEncoder) encodeController(ei EncodingInfo, cIndex int) string {
	fmtCommStrs := []string{}
	cName := fmt.Sprintf("%s%d", CONTROLLER_BASE_NAME, cIndex)

	for pair := ei.usedContFTs[cIndex].Oldest(); pair != nil; pair = pair.Next() {
		swIndex, _ := ei.nodeIdToIndex.Get(pair.Key)
		commStr := f.encodeCommunication(cName, swIndex, false)
		fmtCommStrs = append(fmtCommStrs, commStr)
	}

	fmtC := f.joinNonDetThridColumn(fmtCommStrs)
	return fmt.Sprintf("%s & %s & %s %s", cName, f.sym.DEF, fmtC, NEW_LN)
}

func (f LatexSimpleEncoder) joinNonDetThridColumn(strs []string) string {
	// '& & ' are for placing the conent in the third column of the array env
	nonDetSep := fmt.Sprintf(" %s %s& & ", f.sym.NONDET, NEW_LN)
	return strings.Join(strs, nonDetSep)
}
