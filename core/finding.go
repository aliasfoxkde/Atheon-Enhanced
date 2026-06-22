package core

type Finding struct {
	Pattern string
	File    string
	Line    int
	Content string
}

type Stats struct {
	Files     int
	Bytes     int64
	ElapsedMs int64
}
