package encode

import (
	"strings"

	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

var DYNETKAT_LATEX_SYMBOLS = SymbolEncoding{
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
	encodeSwitch(int, *convert.FlowTable) string
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

func (f LatexEncoder) Encode(ei EncodingInfo) (string, error) {
	fmtSwitches := f.encodeSwitches(ei)
	fmtControllers := f.encodeControllers(ei)

	arrayBlockStr := fmtSwitches + fmtControllers + f.encodeInformation(ei)
	return f.splitIntoPages(arrayBlockStr), nil
}

func (f LatexEncoder) encodeSwitches(ei EncodingInfo) string {
	var sb strings.Builder

	for pair := ei.usedSwitchFTs.Oldest(); pair != nil; pair = pair.Next() {
		swId, ft := pair.Key, pair.Value
		newFT, newFTExists := ei.FindNewFT(swId)
		swIndex := ei.GetSwIndex(swId)

		swStr := f.encodeSwitch(swIndex, ft)
		sb.WriteString(swStr)
		sb.WriteString(LATEX_NEW_LN)

		if newFTExists {
			updateSwStr := f.encodeSwitchNewFT(swIndex, newFT)
			sb.WriteString(updateSwStr)
			sb.WriteString(LATEX_NEW_LN)
		}

	}

	return sb.String()
}

func (f LatexEncoder) encodeControllers(ei EncodingInfo) string {
	var sb strings.Builder

	for i := range ei.usedContUpdates {
		cStr := f.encodeController(ei, i)
		sb.WriteString(cStr)
		sb.WriteString(LATEX_NEW_LN)
	}

	return sb.String()
}

func (f LatexEncoder) splitIntoPages(arrayBlockStr string) string {
	pages := util.SliceContent(arrayBlockStr, LATEX_LINES_PER_PAGE, LATEX_NEW_LN)

	var sb strings.Builder
	sep := ""
	for _, page := range pages {
		sb.WriteString(sep)
		sb.WriteString(LATEX_BEGIN_EQ_ARRAY)
		sb.WriteString(page)
		sb.WriteString(LATEX_END_EQ_ARRAY)
		sep = LATEX_NEW_PAGE
	}

	return sb.String()
}
