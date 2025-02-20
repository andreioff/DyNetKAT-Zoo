package encode

const (
	LINK_TERM_NAME       = "L"
	FLOW_MOD_CHANNEL     = "FM"
	PACKET_IN_CHANNEL    = "PI"
	PACKET_OUT_CHANNEL   = "PO"
	SW_BASE_NAME         = "SW"
	CONTROLLER_BASE_NAME = "C"
	UP_CHANNEL_NAME      = "UP"
	HELP_CHANNEL_NAME    = "HELP"
)

type SymbolEncoding struct {
	// NetKAT symbols
	ONE    string // identity symbol
	ZERO   string // drop packet symbol
	EQ     string // equal
	OR     string
	AND    string
	NEG    string // negation
	STAR   string // recursive symbol
	ASSIGN string // packet field assignment

	// DyNetKAT symbols
	BOT    string // Bot symbol (aka do nothing)
	SEQ    string // Sequential composition
	RECV   string // Receive on channel
	SEND   string // Send over channel
	PAR    string // Parallel composition
	DEF    string // Defines
	NONDET string // non-deterministic choice symbol
}

type NetworkEncoder interface {
	SymbolEncoding() SymbolEncoding
	Encode(EncodingInfo) (string, error)
	ProactiveSwitch() bool
}
