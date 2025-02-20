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
	Switches    []DNKSwitch
	Links       string `json:",omitempty"`
	Controllers map[string]string
}

type DNKDirectUpdate struct {
	Channel string
	Policy  string
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

func newDNKDirectUpdate(channel, policy string) DNKDirectUpdate {
	return DNKDirectUpdate{channel, policy}
}

func newDNKRequestedUpdate(reqCh, reqPol, respCh, respPol string) DNKRequestedUpdate {
	return DNKRequestedUpdate{reqCh, reqPol, respCh, respPol}
}

// Guarantees that the result is not nil
func getOrSetDNKSwitch(dnkSwMap *om.OrderedMap[int64, *DNKSwitch], swId int64) *DNKSwitch {
	dnkSw, exists := dnkSwMap.Get(swId)
	if !exists {
		dnkSw = newEmptyDNKSwitch()
		dnkSwMap.Set(swId, dnkSw)
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

func (je JsonEncoder) encodeSwitches(ei EncodingInfo) []DNKSwitch {
	dnkSwMap := om.New[int64, *DNKSwitch]()
	for _, u := range ei.usedContUpdates {
		je.collectSwFlowRuleUpdates(dnkSwMap, u.frSequences, ei)
		je.collectSwFlowTableUpdates(dnkSwMap, u.flowTables, ei)
	}

	dnkSwitches := []DNKSwitch{}
	for pair := ei.usedSwitchFTs.Oldest(); pair != nil; pair = pair.Next() {
		swId, initialFt := pair.Key, pair.Value.ToNetKATStr(
			je.sym.AND,
			je.sym.EQ,
			je.sym.ASSIGN,
			je.sym.OR,
		)
		dnkSwitch := getOrSetDNKSwitch(dnkSwMap, swId)
		dnkSwitch.InitialFlowTable = initialFt
		dnkSwitches = append(dnkSwitches, *dnkSwitch)
	}

	return dnkSwitches
}

func (je JsonEncoder) collectSwFlowRuleUpdates(
	dnkSwMap *om.OrderedMap[int64, *DNKSwitch],
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
			)
			dnkSw.DirectUpdates = append(dnkSw.DirectUpdates, swDirUpdate)
		}
	}
}

func (je JsonEncoder) collectPiPoSwitchComm(
	dnkSwMap *om.OrderedMap[int64, *DNKSwitch],
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
	dnkSwMap *om.OrderedMap[int64, *DNKSwitch],
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
		swDirUpdate := newDNKDirectUpdate(UP_CHANNEL_NAME+swIndxStr, newFT)
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

	var sb strings.Builder
	sb.WriteString(je.encodePiPoComm(seq.Entries()[0], true, ei))
	for _, e := range seq.Entries()[1:] {
		swIndxStr := ei.GetSwIndexStr(e.SwitchId)
		fmComm := je.encodeRecvComm(FLOW_MOD_CHANNEL, swIndxStr, e.ToNetKATPolicy())
		sb.WriteString(je.sym.SEQ + fmComm)
	}
	sb.WriteString(je.sym.SEQ + cName)
	return sb.String()
}

func (je JsonEncoder) encodePiPoComm(
	entry convert.FRSequenceEntry,
	recvFirst bool,
	ei EncodingInfo,
) string {
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

	return PACKET_IN_CHANNEL + swIndxStr + commOp1 + pkt0Test +
		je.sym.SEQ +
		PACKET_OUT_CHANNEL + swIndxStr + commOp2 + pkt0Policy
}

func (je JsonEncoder) encodeRecvComm(
	chBaseName string,
	swIndxStr string,
	policy convert.SimpleNetKATPolicy,
) string {
	policyStr := ASCII_QUOTE + policy.ToString(je.sym.AND, je.sym.EQ, je.sym.ASSIGN) + ASCII_QUOTE
	return chBaseName + swIndxStr + je.sym.RECV + policyStr
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
		sb.WriteString(commStr)
		sb.WriteString(je.sym.SEQ + cName)
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
	return UP_CHANNEL_NAME + swIndexStr + commSym + policy
}
