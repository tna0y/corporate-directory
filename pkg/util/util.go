package util

type ArgMinResult struct {
	Pos   int
	Value int
}

func ArgMin(a, b ArgMinResult) ArgMinResult {
	if a.Value < b.Value {
		return a
	}
	return b
}
