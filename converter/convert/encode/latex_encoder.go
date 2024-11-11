package encode

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"utwente.nl/topology-to-dynetkat-coverter/convert"
)

const (
	NEW_LN            = "\\\\\n"
	NEW_PAGE          = "\n\\newpage\n"
	BEGIN_EQ_ARRAY    = "\\begin{equation} \\begin{array}{rcl}\n\n"
	END_EQ_ARRAY      = "\n\n\\end{array} \\end{equation}\n"
	THIRD_COL_MAX_LEN = 40 // nr of chars before the third column of the array env overflows
	LINES_PER_PAGE    = 40
)

type LatexEncoder struct {
	sym SymbolEncoding
}

func NewLatexEncoder() LatexEncoder {
	return LatexEncoder{
		SymbolEncoding{
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
			RECV:   "?",
			SEND:   "!",
			PAR:    "\\, \\|\\, ",
			DEF:    "\\triangleq",
			NONDET: "\\, \\oplus\\,",
		},
	}
}

func (f *LatexEncoder) GetSymbolEncodings() SymbolEncoding {
	return f.sym
}

func (f *LatexEncoder) Encode(n *convert.Network) (string, error) {
	if n == nil {
		return "", errors.New("Received nil network!")
	}

	fmtSwitches, nonEmptySwitches := f.encodeSwitches(n.Switches())
	arrayBlockStr := fmtSwitches + f.encodeSDNTerm(nonEmptySwitches)
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

func (f *LatexEncoder) encodeSwitches(switches []convert.Switch) (string, []convert.Switch) {
	nonEmptySwitches := []convert.Switch{}
	var sb strings.Builder

	for _, sw := range switches {
		swStr := f.encodeSwitch(sw)
		if swStr != "" {
			sb.WriteString(swStr)
			sb.WriteString(NEW_LN)
			nonEmptySwitches = append(nonEmptySwitches, sw)
		}
	}

	return sb.String(), nonEmptySwitches
}

func (f *LatexEncoder) encodeSwitch(sw convert.Switch) string {
	fmtDstTableEntries := []string{}
	swName := f.encodeSwitchName(sw)

	prefix := ""
	for dstHostId, inPortToOutPort := range sw.DestTable() {
		for inPort, outPort := range inPortToOutPort {
			fmtDstTableEntries = append(fmtDstTableEntries, fmt.Sprintf(
				"%s((dst %s %d) %s (port %s %d) %s (port %s %d)) %s %s",
				prefix, f.sym.EQ,
				dstHostId, f.sym.AND,
				f.sym.EQ, inPort,
				f.sym.AND, f.sym.ASSIGN,
				outPort, f.sym.SEQ,
				swName,
			))
			prefix = "& & "
		}
	}

	if len(fmtDstTableEntries) == 0 {
		return ""
	}

	nonDetSep := fmt.Sprintf(" %s %s", f.sym.NONDET, NEW_LN)
	fmtSw := strings.Join(fmtDstTableEntries, nonDetSep)
	return fmt.Sprintf("%s & %s & %s %s", swName, f.sym.DEF, fmtSw, NEW_LN)
}

func (f *LatexEncoder) encodeSDNTerm(sws []convert.Switch) string {
	var sb strings.Builder

	prefix := ""
	for _, sw := range sws {
		sb.WriteString(prefix + f.encodeSwitchName(sw))
		prefix = f.sym.PAR
	}

	return fmt.Sprintf("SDN & %s & %s", f.sym.DEF, breakColumn(sb.String()))
}

func (f *LatexEncoder) encodeSwitchName(sw convert.Switch) string {
	return fmt.Sprintf("SW%d", sw.TopoNode().ID())
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
		return []string{}, errors.New("Zero or negative division length!")
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
