package encode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	om "github.com/wk8/go-ordered-map/v2"
	"utwente.nl/topology-to-dynetkat-coverter/convert"
)

const (
	ASCII_QUOTE = "\""
)

var DYNETKAT_ASCII_SYMBOLS = SymbolEncoding{
	ONE:    " one ",
	ZERO:   " zero ",
	EQ:     "=",
	OR:     " + ",
	AND:    " . ",
	NEG:    " NEGATE ", // not needed, but just in case
	STAR:   " * ",
	ASSIGN: "<-",

	BOT:    " bot ",
	SEQ:    " ; ",
	RECV:   " ? ",
	SEND:   " ! ",
	PAR:    " || ",
	DEF:    " DEFINE ", // not needed, but just in case
	NONDET: " o+ ",
}

// JSON structs
type DNKNetwork struct {
	Switches    map[string]DNKSwitch
	Links       string `json:",omitempty"`
	Controllers map[string]string
}

type DNKDirectUpdate struct {
	Channel string
	Policy  string
	Append  bool
}

type DNKRequestedUpdate struct {
	RequestChannel  string
	RequestPolicy   string
	ResponseChannel string
	ResponsePolicy  string
}

type DNKSwitch struct {
	InitialFlowTable string `json:",omitempty"`
	DirectUpdates    []DNKDirectUpdate
	RequestedUpdates []DNKRequestedUpdate
}

func newEmptyDNKSwitch() *DNKSwitch {
	return &DNKSwitch{"", []DNKDirectUpdate{}, []DNKRequestedUpdate{}}
}

func newDNKDirectUpdate(channel, policy string, appen bool) DNKDirectUpdate {
	return DNKDirectUpdate{channel, policy, appen}
}

func newDNKRequestedUpdate(reqCh, reqPol, respCh, respPol string) DNKRequestedUpdate {
	return DNKRequestedUpdate{reqCh, reqPol, respCh, respPol}
}

// Guarantees that the result is not nil
func getOrSetDNKSwitch(dnkSwMap map[int64]*DNKSwitch, swId int64) *DNKSwitch {
	dnkSw, exists := dnkSwMap[swId]
	if !exists {
		dnkSw = newEmptyDNKSwitch()
		dnkSwMap[swId] = dnkSw
	}
	return dnkSw
}

type JsonEncoder struct {
	sym SymbolEncoding
}

func NewJsonEncoder() NetworkEncoder {
	return JsonEncoder{
		sym: DYNETKAT_ASCII_SYMBOLS,
	}
}

func (je JsonEncoder) SymbolEncoding() SymbolEncoding {
	return je.sym
}

func (je JsonEncoder) ProactiveSwitch() bool {
	return false
}

func (je JsonEncoder) Encode(ei EncodingInfo) (string, error) {
	fmtSwitches := je.encodeSwitches(ei)
	fmtControllers := je.encodeControllers(ei)
	links := je.encodeLinks(ei)

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(DNKNetwork{
		Switches:    fmtSwitches,
		Controllers: fmtControllers,
		Links:       links,
	})
	return buffer.String(), err
}

func (je JsonEncoder) buildSeqExpr(exprs []string) string {
	if len(exprs) < 3 {
		return strings.Join(exprs, je.sym.SEQ)
	}

	var sb, closedPars strings.Builder

	sb.WriteString(exprs[0])
	for _, e := range exprs[1 : len(exprs)-1] {
		sb.WriteString(je.sym.SEQ)
		sb.WriteString("(")
		sb.WriteString(e)
		closedPars.WriteString(")")
	}
	sb.WriteString(je.sym.SEQ)
	sb.WriteString(exprs[len(exprs)-1])
	sb.WriteString(closedPars.String())
	return sb.String()
}

func (je JsonEncoder) encodeLinks(ei EncodingInfo) string {
	var sb strings.Builder
	prefix := ""
	for _, link := range ei.links {
		sb.WriteString(prefix)
		sb.WriteString(link.ToString(
			je.sym.AND,
			je.sym.EQ,
			je.sym.ASSIGN,
		))
		prefix = je.sym.OR
	}
	return sb.String()
}

func (je JsonEncoder) encodeSwitches(ei EncodingInfo) map[string]DNKSwitch {
	dnkSwMap := make(map[int64]*DNKSwitch)
	for _, u := range ei.usedContUpdates {
		je.collectSwFlowRuleUpdates(dnkSwMap, u.frSequences, ei)
		je.collectSwFlowTableUpdates(dnkSwMap, u.flowTables, ei)
	}

	fmtDnkSwitches := make(map[string]DNKSwitch)
	for pair := ei.usedSwitchFTs.Oldest(); pair != nil; pair = pair.Next() {
		swId, initialFt := pair.Key, pair.Value.ToNetKATStr(
			je.sym.AND,
			je.sym.EQ,
			je.sym.ASSIGN,
			je.sym.OR,
		)
		dnkSwitch := getOrSetDNKSwitch(dnkSwMap, swId)
		dnkSwitch.InitialFlowTable = initialFt
		fmtDnkSwitches[SW_BASE_NAME+ei.GetSwIndexStr(swId)] = *dnkSwitch
	}

	return fmtDnkSwitches
}

func (je JsonEncoder) collectSwFlowRuleUpdates(
	dnkSwMap map[int64]*DNKSwitch,
	frSeqs []*convert.FlowRuleSequence,
	ei EncodingInfo,
) {
	for _, seq := range frSeqs {
		firstEntry := seq.Entries()[0]
		je.collectPiPoSwitchComm(dnkSwMap, firstEntry.SwitchId, firstEntry.ToNetKATPolicy(), ei)

		for _, entry := range seq.Entries()[1:] {
			dnkSw := getOrSetDNKSwitch(dnkSwMap, entry.SwitchId)
			swIndxStr := ei.GetSwIndexStr(entry.SwitchId)
			swDirUpdate := newDNKDirectUpdate(
				FLOW_MOD_CHANNEL+swIndxStr,
				entry.ToNetKATPolicy().ToString(je.sym.AND, je.sym.EQ, je.sym.ASSIGN),
				true,
			)
			dnkSw.DirectUpdates = append(dnkSw.DirectUpdates, swDirUpdate)
		}
	}
}

func (je JsonEncoder) collectPiPoSwitchComm(
	dnkSwMap map[int64]*DNKSwitch,
	swId int64,
	policy convert.SimpleNetKATPolicy,
	ei EncodingInfo,
) {
	dnkSw := getOrSetDNKSwitch(dnkSwMap, swId)
	swIndxStr := ei.GetSwIndexStr(swId)
	swReqUpdate := newDNKRequestedUpdate(
		PACKET_IN_CHANNEL+swIndxStr,
		policy.TestToString(je.sym.AND, je.sym.EQ),
		PACKET_OUT_CHANNEL+swIndxStr,
		policy.ToString(je.sym.AND, je.sym.EQ, je.sym.ASSIGN),
	)
	dnkSw.RequestedUpdates = append(dnkSw.RequestedUpdates, swReqUpdate)
}

func (je JsonEncoder) collectSwFlowTableUpdates(
	dnkSwMap map[int64]*DNKSwitch,
	ftUpdates om.OrderedMap[int64, *convert.FlowTable],
	ei EncodingInfo,
) {
	for pair := ftUpdates.Oldest(); pair != nil; pair = pair.Next() {
		swId, newFT := pair.Key, pair.Value.ToNetKATStr(
			je.sym.AND,
			je.sym.EQ,
			je.sym.ASSIGN,
			je.sym.OR,
		)
		dnkSw := getOrSetDNKSwitch(dnkSwMap, swId)

		swIndxStr := ei.GetSwIndexStr(swId)
		swDirUpdate := newDNKDirectUpdate(UP_CHANNEL_NAME+swIndxStr, newFT, false)
		dnkSw.DirectUpdates = append(dnkSw.DirectUpdates, swDirUpdate)
	}
}

func (je JsonEncoder) encodeControllers(ei EncodingInfo) map[string]string {
	fmtConts := make(map[string]string)
	for i, update := range ei.usedContUpdates {
		cName := CONTROLLER_BASE_NAME + fmt.Sprintf("%d", i)
		seqEnc := je.encodeContFRSequences(cName, update.frSequences, ei)
		ftsEnc := je.encodeContFlowTables(cName, update.flowTables, ei)

		sep := je.sym.NONDET
		if seqEnc == "" || ftsEnc == "" {
			sep = ""
		}

		fmtConts[cName] = seqEnc + sep + ftsEnc
	}
	return fmtConts
}

func (je JsonEncoder) encodeContFRSequences(
	cName string,
	seqs []*convert.FlowRuleSequence,
	ei EncodingInfo,
) string {
	var sb strings.Builder
	prefix := ""
	for _, seq := range seqs {
		if len(seq.Entries()) == 0 {
			continue
		}

		sb.WriteString(prefix)
		sb.WriteString(je.encodeContFRSequence(cName, seq, ei))
		prefix = je.sym.NONDET
	}
	return sb.String()
}

func (je JsonEncoder) encodeContFRSequence(
	cName string,
	seq *convert.FlowRuleSequence,
	ei EncodingInfo,
) string {
	if seq == nil || len(seq.Entries()) == 0 {
		return ""
	}

	fmtSeqTerms := je.encodePiPoComm(seq.Entries()[0], true, ei)
	for _, e := range seq.Entries()[1:] {
		swIndxStr := ei.GetSwIndexStr(e.SwitchId)
		fmComm := je.encodeSendComm(FLOW_MOD_CHANNEL, swIndxStr, e.ToNetKATPolicy())
		fmtSeqTerms = append(fmtSeqTerms, fmComm)
	}
	fmtSeqTerms = append(fmtSeqTerms, cName)
	return je.buildSeqExpr(fmtSeqTerms)
}

func (je JsonEncoder) encodePiPoComm(
	entry convert.FRSequenceEntry,
	recvFirst bool,
	ei EncodingInfo,
) []string {
	commOp1 := je.sym.SEND
	commOp2 := je.sym.RECV
	if recvFirst {
		commOp1 = je.sym.RECV
		commOp2 = je.sym.SEND
	}

	swIndxStr := ei.GetSwIndexStr(entry.SwitchId)

	nkEntry := entry.ToNetKATPolicy()
	pkt0Test := ASCII_QUOTE + nkEntry.TestToString(je.sym.AND, je.sym.EQ) + ASCII_QUOTE
	pkt0Policy := ASCII_QUOTE + nkEntry.ToString(je.sym.AND, je.sym.EQ, je.sym.ASSIGN) + ASCII_QUOTE

	return []string{
		"(" + PACKET_IN_CHANNEL + swIndxStr + commOp1 + pkt0Test + ")",
		"(" + PACKET_OUT_CHANNEL + swIndxStr + commOp2 + pkt0Policy + ")",
	}
}

func (je JsonEncoder) encodeSendComm(
	chBaseName string,
	swIndxStr string,
	policy convert.SimpleNetKATPolicy,
) string {
	policyStr := ASCII_QUOTE + policy.ToString(je.sym.AND, je.sym.EQ, je.sym.ASSIGN) + ASCII_QUOTE
	return "(" + chBaseName + swIndxStr + je.sym.SEND + policyStr + ")"
}

func (je JsonEncoder) encodeContFlowTables(
	cName string,
	flowTables om.OrderedMap[int64, *convert.FlowTable],
	ei EncodingInfo,
) string {
	var sb strings.Builder

	prefix := ""
	for pair := flowTables.Oldest(); pair != nil; pair = pair.Next() {
		swIndexStr := ei.GetSwIndexStr(pair.Key)
		commStr := je.encodeFlowTableComm(swIndexStr, *pair.Value, je.sym.SEND)
		if commStr == "" {
			continue
		}

		sb.WriteString(prefix)
		sb.WriteString(je.buildSeqExpr([]string{commStr, cName}))
		prefix = je.sym.NONDET
	}

	return sb.String()
}

func (je JsonEncoder) encodeFlowTableComm(
	swIndexStr string,
	flowTable convert.FlowTable,
	commSym string,
) string {
	policy := flowTable.ToNetKATStr(je.sym.AND, je.sym.EQ, je.sym.ASSIGN, je.sym.OR)
	policy = ASCII_QUOTE + policy + ASCII_QUOTE
	return "(" + UP_CHANNEL_NAME + swIndexStr + commSym + policy + ")"
}
