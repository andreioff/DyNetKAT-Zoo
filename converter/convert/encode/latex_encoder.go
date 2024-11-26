package encode

import (
	"strings"

	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

var LATEX_SYMBOLS = SymbolEncoding{
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
}

type CustomFunctions interface {
	encodeSwitch(int, *convert.FlowTable, bool) string
	encodeSwitchNewFT(int, *convert.FlowTable) string
	encodeController(EncodingInfo, int) string
	encodeInformation(EncodingInfo) string
	SymbolEncoding() SymbolEncoding
	ProactiveSwitch() bool
}

type LatexEncoder struct {
	CustomFunctions
}

func NewLatexEncoder(proactiveSwitch bool, cf CustomFunctions) LatexEncoder {
	return LatexEncoder{
		CustomFunctions: cf,
	}
}

func (f LatexEncoder) Encode(ei EncodingInfo) string {
	fmtSwitches := f.encodeSwitches(ei)
	fmtControllers := f.encodeControllers(ei)

	arrayBlockStr := fmtSwitches + fmtControllers + f.encodeInformation(ei)
	return f.splitIntoPages(arrayBlockStr)
}

func (f LatexEncoder) encodeSwitches(ei EncodingInfo) string {
	var sb strings.Builder

	for pair := ei.usedSwitchFTs.Oldest(); pair != nil; pair = pair.Next() {
		swId, ft := pair.Key, pair.Value
		newFT, willReceiveUpdate := ei.FindNewFT(swId)
		swIndex, _ := ei.nodeIdToIndex.Get(swId)

		swStr := f.encodeSwitch(swIndex, ft, willReceiveUpdate)
		sb.WriteString(swStr)
		sb.WriteString(NEW_LN)

		if willReceiveUpdate {
			updateSwStr := f.encodeSwitchNewFT(swIndex, newFT)
			sb.WriteString(updateSwStr)
			sb.WriteString(NEW_LN)
		}

	}

	return sb.String()
}

func (f LatexEncoder) encodeControllers(ei EncodingInfo) string {
	var sb strings.Builder

	for i := range ei.usedContFTs {
		cStr := f.encodeController(ei, i)
		sb.WriteString(cStr)
		sb.WriteString(NEW_LN)
	}

	return sb.String()
}

func (f LatexEncoder) splitIntoPages(arrayBlockStr string) string {
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

	return sb.String()
}
