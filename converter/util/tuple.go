package util

type Tuple[A any, B any] struct {
	Fst A
	Snd B
}

type (
	I64Tup = Tuple[int64, int64]
	StrTup = Tuple[string, string]
)

func NewI64Tup(fst, snd int64) I64Tup {
	return I64Tup{
		Fst: fst,
		Snd: snd,
	}
}

func NewStrTup(fst, snd string) StrTup {
	return StrTup{
		Fst: fst,
		Snd: snd,
	}
}
