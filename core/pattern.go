package core

import "sort"

type Pattern interface {
	Name() string
	Category() string
	Matches(line string) bool
}

var registry []Pattern

func Register(p Pattern) {
	registry = append(registry, p)
}

func All() []Pattern {
	sorted := make([]Pattern, len(registry))
	copy(sorted, registry)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name() < sorted[j].Name()
	})
	return sorted
}
