package file

type Contents struct {
	AbsPath  string
	Contents []byte
}

func (c Contents) BytesInRange(start, stop int) []byte {
	if start > stop {
		start, stop = stop, start
	}

	start = min(start, len(c.Contents))
	stop = min(stop, len(c.Contents))

	return c.Contents[start:stop]
}

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
	return (r.isOneLine() && r.Start.Line == pos.Line && r.containsCol(pos.Col)) || r.containsLine(pos.Line)
}

type Replacement struct {
	Range Range    `json:"rng"`
	Lines []string `json:"lns,omitempty"`
}
