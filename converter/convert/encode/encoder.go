package encode

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
	Encode(EncodingInfo) string
	ProactiveSwitch() bool
}
