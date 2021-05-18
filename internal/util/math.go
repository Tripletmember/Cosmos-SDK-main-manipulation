package util

func Uint64Min(a, b uint64) uint64 {
	if b < a { return b }
	return a
}
