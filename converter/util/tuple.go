package util

type Tuple[A any, B any] struct {
	Fst A
	Snd B
}

type I64Tup = Tuple[int64, int64]

func NewI64Tup(fst, snd int64) I64Tup {
	return Tuple[int64, int64]{
		Fst: fst,
		Snd: snd,
	}
}
