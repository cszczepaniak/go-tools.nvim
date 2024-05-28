package file

type Position struct {
	Line int `json:"ln"`
	Col  int `json:"col"`
}

type Range struct {
	Start Position `json:"start"`
	Stop  Position `json:"stop"`
}

func (r Range) isOneLine() bool {
	return r.Start.Line == r.Stop.Line
}

func (r Range) containsCol(col int) bool {
	return r.Start.Col <= col && col <= r.Stop.Col
}

func (r Range) containsLine(col int) bool {
	return r.Start.Line <= col && col <= r.Stop.Line
}

func (r Range) ContainsPos(pos Position) bool {
	return (r.isOneLine() && r.containsCol(pos.Col)) || r.containsLine(pos.Line)
}

type Replacement struct {
	Range Range    `json:"rng"`
	Text  string   `json:"text,omitempty"`
	Lines []string `json:"lns,omitempty"`
}
