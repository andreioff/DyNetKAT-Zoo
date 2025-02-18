package encode

import (
	"fmt"
	"strings"

	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

const (
	LATEX_NEW_LN            = "\\\\\n"
	LATEX_DNEW_LN           = LATEX_NEW_LN + LATEX_NEW_LN
	LATEX_NEW_PAGE          = "\n\\newpage\n"
	LATEX_BEGIN_EQ_ARRAY    = "\\begin{equation} \\begin{array}{rcl}\n\n"
	LATEX_END_EQ_ARRAY      = "\n\n\\end{array} \\end{equation}\n"
	LATEX_THIRD_COL_MAX_LEN = 40 // nr of chars before the third column of the array env overflows
	LATEX_LINES_PER_PAGE    = 40
)

type LatexSimpleEncoder struct {
	sym             SymbolEncoding
	proactiveSwitch bool
}

func NewLatexSimpleEncoder(proactiveSwitch bool) NetworkEncoder {
	return NewLatexEncoder(proactiveSwitch, LatexSimpleEncoder{sym: DYNETKAT_LATEX_SYMBOLS})
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
	fmtNewSw := f.encodeSwFlowTable(newFT, newSwName)

	if fmtNewSw == "" {
		return fmt.Sprintf(
			"%s & %s & %s%s%s%s",
			newSwName,
			f.sym.DEF,
			f.sym.ZERO,
			f.sym.SEQ,
			newSwName,
			LATEX_NEW_LN,
		)
	}
	return fmt.Sprintf("%s & %s & %s%s", newSwName, f.sym.DEF, fmtNewSw, LATEX_NEW_LN)
}

func (f LatexSimpleEncoder) encodeSwitch(
	swIndex int,
	ft *convert.FlowTable,
) string {
	swName := f.encodeSwitchName(swIndex, false)

	fmtFlowRules := f.encodeSwFlowTable(ft, swName)
	if fmtFlowRules == "" {
		fmtFlowRules = fmt.Sprintf("%s%s%s", f.sym.ZERO, f.sym.SEQ, swName)
	}
	commStr := f.encodeCommunication(f.encodeSwitchName(swIndex, true), swIndex, false)

	fmtSw := f.joinNonDetThridColumn([]string{fmtFlowRules, commStr})
	return fmt.Sprintf("%s & %s & %s %s", swName, f.sym.DEF, fmtSw, LATEX_NEW_LN)
}

func (f LatexSimpleEncoder) encodeSwFlowTable(
	flowTable *convert.FlowTable,
	swName string,
) string {
	if flowTable == nil || flowTable.Entries().Len() == 0 {
		return ""
	}

	fmtFlowRules := flowTable.ToNetKATStr(f.sym.AND, f.sym.EQ, f.sym.ASSIGN, f.sym.OR)
	return fmt.Sprintf("(%s) %s %s", fmtFlowRules, f.sym.SEQ, swName)
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
		swIndex := ei.GetSwIndex(pair.Key)
		sb.WriteString(prefix + f.encodeSwitchName(swIndex, false))
		prefix = f.sym.PAR
	}

	for i := range ei.usedContUpdates {
		sb.WriteString(prefix + fmt.Sprintf("%s%d", CONTROLLER_BASE_NAME, i))
	}
	content := util.BreakColumn(sb.String(), LATEX_THIRD_COL_MAX_LEN, LATEX_NEW_LN+"& & ")
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

	update := ei.usedContUpdates[cIndex]
	for pair := update.flowTables.Oldest(); pair != nil; pair = pair.Next() {
		swIndex := ei.GetSwIndex(pair.Key)
		commStr := f.encodeCommunication(cName, swIndex, false)
		fmtCommStrs = append(fmtCommStrs, commStr)
	}

	fmtC := f.joinNonDetThridColumn(fmtCommStrs)
	return fmt.Sprintf("%s & %s & %s %s", cName, f.sym.DEF, fmtC, LATEX_NEW_LN)
}

func (f LatexSimpleEncoder) joinNonDetThridColumn(strs []string) string {
	// '& & ' are for placing the conent in the third column of the array env
	nonDetSep := fmt.Sprintf(" %s %s& & ", f.sym.NONDET, LATEX_NEW_LN)
	return strings.Join(strs, nonDetSep)
}
