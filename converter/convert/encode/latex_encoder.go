package encode

import (
	"fmt"
	"log"
	"strings"

	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

const (
	NEW_LN               = "\\\\\n"
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

type LatexEncoder struct {
	sym             SymbolEncoding
	proactiveSwitch bool
}

func NewLatexEncoder(proactiveSwitch bool) LatexEncoder {
	return LatexEncoder{
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

func (f *LatexEncoder) SymbolEncodings() SymbolEncoding {
	return f.sym
}

func (f *LatexEncoder) ProactiveSwitch() bool {
	return f.proactiveSwitch
}

func (f *LatexEncoder) Encode(n *convert.Network) (string, error) {
	if n == nil {
		return "", util.NewError(util.ErrNilArgument, "n")
	}

	fmtSwitches, nonEmptySwitches := f.encodeSwitches(n.Switches())
	fmtControllers, usedControllers := f.encodeControllers(n.Controllers())
	arrayBlockStr := fmtSwitches + fmtControllers + f.encodeSDNTerm(
		nonEmptySwitches,
		usedControllers,
	)
	pages := sliceContent(arrayBlockStr, LINES_PER_PAGE, NEW_LN)

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

func (f *LatexEncoder) encodeSwitches(switches []*convert.Switch) (string, []*convert.Switch) {
	nonEmptySwitches := []*convert.Switch{}
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
			nonEmptySwitches = append(nonEmptySwitches, sw)
		}

		if !willReceiveUpdate {
			continue
		}

		// TODO This can be merged in the encodeSwitch function
		newSwName := f.encodeSwitchName(*sw, true)
		updatedSwStrs := f.encodeNetKATPolicies(newFlowTable.ToNetKATPolicies(), newSwName)
		if len(updatedSwStrs) != 0 {
			fmtNewSw := f.joinNonDetThridColumn(updatedSwStrs)
			sb.WriteString(fmt.Sprintf("%s & %s & %s", newSwName, f.sym.DEF, fmtNewSw))
			sb.WriteString(NEW_LN + NEW_LN)
		}
	}

	return sb.String(), nonEmptySwitches
}

func (f *LatexEncoder) encodeSwitch(sw convert.Switch, canBeEmpty bool) string {
	swName := f.encodeSwitchName(sw, false)

	fmtFlowRules := f.encodeNetKATPolicies(sw.FlowTable().ToNetKATPolicies(), swName)

	if len(fmtFlowRules) == 0 {
		if !canBeEmpty {
			return ""
		}
		dropAllStr := fmt.Sprintf("%s%s%s", f.sym.ZERO, f.sym.SEQ, swName)
		fmtFlowRules = append(fmtFlowRules, dropAllStr)
	}

	commStr := f.encodeCommunication(f.encodeSwitchName(sw, true), sw.TopoNode().ID(), false)
	fmtFlowRules = append(fmtFlowRules, commStr)

	fmtSw := f.joinNonDetThridColumn(fmtFlowRules)
	return fmt.Sprintf("%s & %s & %s %s", swName, f.sym.DEF, fmtSw, NEW_LN)
}

func (f *LatexEncoder) encodeNetKATPolicies(
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

func (f *LatexEncoder) encodeCommunication(
	termName string,
	channelId int64,
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

func (f *LatexEncoder) encodeSDNTerm(sws []*convert.Switch, c []*convert.Controller) string {
	var sb strings.Builder

	prefix := ""
	for _, sw := range sws {
		sb.WriteString(prefix + f.encodeSwitchName(*sw, false))
		prefix = f.sym.PAR
	}

	for _, c := range c {
		sb.WriteString(prefix + fmt.Sprintf("%s%d", CONTROLLER_BASE_NAME, c.ID()))
	}

	return fmt.Sprintf("SDN & %s & %s", f.sym.DEF, breakColumn(sb.String()))
}

func (f *LatexEncoder) encodeSwitchName(sw convert.Switch, isNew bool) string {
	name := fmt.Sprintf("%s%d", SW_BASE_NAME, sw.TopoNode().ID())
	if isNew {
		return name + "'"
	}
	return name
}

func (f *LatexEncoder) encodeControllers(
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

func (f *LatexEncoder) encodeController(c *convert.Controller) string {
	fmtCommStrs := []string{}
	cName := fmt.Sprintf("%s%d", CONTROLLER_BASE_NAME, c.ID())

	for key := range c.NewFlowTables() {
		commStr := f.encodeCommunication(cName, key, false)
		fmtCommStrs = append(fmtCommStrs, commStr)
	}

	if len(c.NewFlowTables()) == 0 {
		return ""
	}

	fmtC := f.joinNonDetThridColumn(fmtCommStrs)
	return fmt.Sprintf("%s & %s & %s %s", cName, f.sym.DEF, fmtC, NEW_LN)
}

func (f *LatexEncoder) joinNonDetThridColumn(strs []string) string {
	// '& & ' are for placing the conent in the third column of the array env
	nonDetSep := fmt.Sprintf(" %s %s& & ", f.sym.NONDET, NEW_LN)
	return strings.Join(strs, nonDetSep)
}

/*
Breaks the given string into slices of 'linesPerPage' lines based on the
given separator 'sep'.
*/
func sliceContent(str string, linesPerPage int, sep string) []string {
	contentLines := strings.SplitAfter(str, sep)
	pages := []string{}

	for i := range (len(contentLines) / linesPerPage) + 1 {
		start := i * linesPerPage
		end := min(start+linesPerPage, len(contentLines))
		page := strings.Join(contentLines[start:end], "")
		pages = append(pages, page)
	}

	return pages
}

// assumes the given string is in the third column (considered the last column) of the array environment
func breakColumn(line string) string {
	divisions, err := divideLatexString(line, THIRD_COL_MAX_LEN)
	if err != nil {
		log.Println("Failed to break Latex string. Keeping the string unmodified...")
		return line
	}

	return strings.Join(divisions, NEW_LN+"& & ")
}

/*
Divides the given string into multiple divisions of at most 'divLen'
characters consider Latex symbols escaped with a backslash.
Each division ends on a Latex symbol.

Whitespaces and OS new lines are not counted.
Latex new lines are treated as one character and do NOT reset the char
count of the current division.
*/
func divideLatexString(str string, divLen int) ([]string, error) {
	if divLen < 1 {
		return []string{}, util.NewError(util.ErrZeroOrNegDivisionLength)
	}

	divisions, charCount, divStart, isInsideSymbol, toAdd := []string{}, 0, 0, false, 0

	i := 0
	for i < len(str) {
		toAdd, isInsideSymbol = countSymbol(str[i], isInsideSymbol)
		charCount += toAdd

		if charCount >= divLen && !isInsideSymbol {
			lastSymbol := findLastSymbol(str, i)
			divisions = append(divisions, str[divStart:lastSymbol+1])
			divStart = lastSymbol + 1
			i, charCount = divStart, 0
			continue
		}
		i++
	}

	if divStart < len(str) {
		divisions = append(divisions, str[divStart:])
	}
	return divisions, nil
}

func countSymbol(char byte, isInsideSymbol bool) (int, bool) {
	if char == ' ' || char == '\n' {
		return 0, false
	}

	if char == '\\' && !isInsideSymbol {
		return 1, true
	}

	if !isInsideSymbol {
		return 1, false
	}

	return 0, isInsideSymbol
}

func findLastSymbol(str string, currPos int) int {
	if currPos <= 0 {
		return 0
	}

	lastSpace := currPos
	for i := currPos; i >= 0; i-- {
		if str[i] == ' ' {
			lastSpace = i
			continue
		}

		if str[i] == '\\' {
			return lastSpace
		}
	}

	return currPos
}
