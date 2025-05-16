package json_encoder_whole_network_updates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	om "github.com/wk8/go-ordered-map/v2"
	"utwente.nl/topology-to-dynetkat-coverter/convert"
	en "utwente.nl/topology-to-dynetkat-coverter/convert/encode"
)

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

type JsonEncoder struct {
	sym en.SymbolEncoding
}

/*
Ignores flow rule senquences installed by controllers.
*/
func NewJsonEncoder() en.NetworkEncoder {
	return JsonEncoder{
		sym: en.DYNETKAT_ASCII_SYMBOLS,
	}
}

func (je JsonEncoder) SymbolEncoding() en.SymbolEncoding {
	return je.sym
}

func (je JsonEncoder) ProactiveSwitch() bool {
	return false
}

func (je JsonEncoder) Encode(ei en.EncodingInfo) (string, error) {
	fmtBigSwitch := je.encodeBigSwitch(ei)
	fmtControllers := je.encodeControllers(ei)

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(DNKNetwork{
		Switches:    map[string]DNKSwitch{"BigSW": fmtBigSwitch},
		Controllers: fmtControllers,
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

func (je JsonEncoder) encodeLinks(ei en.EncodingInfo) string {
	var sb strings.Builder
	prefix := ""
	for _, link := range ei.Links {
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

func (je JsonEncoder) encodeBigSwitch(ei en.EncodingInfo) DNKSwitch {
	dnkSw := newEmptyDNKSwitch()
	link := je.encodeLinks(ei)
	dnkSw.InitialFlowTable = je.buildNetworkEncoding(&ei.UsedSwitchFTs, link)
	for _, update := range ei.UsedContUpdates {
		wholeNetworkUpdates := je.buildWholeNetworkUpdates(
			link,
			update.FlowTables,
			ei,
		)
		je.collectSwFlowTableUpdates(dnkSw, wholeNetworkUpdates, ei)
	}

	return *dnkSw
}

func (je JsonEncoder) buildNetworkEncoding(
	swFlowTables *om.OrderedMap[int64, *convert.FlowTable],
	link string,
) string {
	initialFT := convert.NewFlowTable()
	for pair := swFlowTables.Oldest(); pair != nil; pair = pair.Next() {
		initialFT.Extend(pair.Value)
	}
	if initialFT.Entries().Len() == 0 {
		return je.sym.ZERO
	}
	netkatFT := initialFT.ToNetKATStr(
		je.sym.AND,
		je.sym.EQ,
		je.sym.ASSIGN,
		je.sym.OR,
	)
	networkEnc := "(" + netkatFT + ")" + je.sym.AND + "(" + link + ")"
	return networkEnc + je.sym.AND + "(" + networkEnc + ")" + je.sym.STAR
}

func (je JsonEncoder) collectSwFlowTableUpdates(
	dnkSw *DNKSwitch,
	updates om.OrderedMap[int64, string],
	ei en.EncodingInfo,
) {
	for pair := updates.Oldest(); pair != nil; pair = pair.Next() {
		swId, update := pair.Key, pair.Value
		swIndxStr := ei.GetSwIndexStr(swId)
		swUpdate := newDNKDirectUpdate(en.UP_CHANNEL_NAME+swIndxStr, update, false)
		dnkSw.DirectUpdates = append(dnkSw.DirectUpdates, swUpdate)
	}
}

func (je JsonEncoder) encodeControllers(ei en.EncodingInfo) map[string]string {
	fmtConts := make(map[string]string)
	link := je.encodeLinks(ei)
	for i, update := range ei.UsedContUpdates {
		cName := en.CONTROLLER_BASE_NAME + fmt.Sprintf("%d", i)
		wholeNetworkUpdates := je.buildWholeNetworkUpdates(
			link,
			update.FlowTables,
			ei,
		)
		fmtConts[cName] = je.encodeContUpdates(cName, wholeNetworkUpdates, ei)
	}
	return fmtConts
}

func (je JsonEncoder) buildWholeNetworkUpdates(
	link string,
	flowTables om.OrderedMap[int64, *convert.FlowTable],
	ei en.EncodingInfo,
) om.OrderedMap[int64, string] {
	swFTs := ei.CopyUsedSwitchesFTs()
	updates := *om.New[int64, string]()
	for pair := flowTables.Oldest(); pair != nil; pair = pair.Next() {
		swId, flowTable := pair.Key, pair.Value
		swFTs.Set(swId, flowTable)
		updates.Set(swId, je.buildNetworkEncoding(swFTs, link))
	}
	return updates
}

func (je JsonEncoder) encodeContUpdates(
	cName string,
	updates om.OrderedMap[int64, string],
	ei en.EncodingInfo,
) string {
	ftCommStrs := []string{}
	for pair := updates.Oldest(); pair != nil; pair = pair.Next() {
		swIndexStr := ei.GetSwIndexStr(pair.Key)
		commStr := je.encodeUpdateComm(swIndexStr, pair.Value, je.sym.SEND)
		if commStr == "" {
			continue
		}

		ftCommStrs = append(ftCommStrs, commStr)
	}

	ftCommStrs = append(ftCommStrs, cName)
	return je.buildSeqExpr(ftCommStrs)
}

func (je JsonEncoder) encodeUpdateComm(
	swIndexStr string,
	policy string,
	commSym string,
) string {
	policy = en.ASCII_QUOTE + policy + en.ASCII_QUOTE
	return "(" + en.UP_CHANNEL_NAME + swIndexStr + commSym + policy + ")"
}
