package notify

import (
	"fmt"
	"maps"
	"slices"
)

type Registry struct{ byKind map[string]Channel }

func NewRegistry() *Registry { return &Registry{byKind: map[string]Channel{}} }

// Register panics on a duplicate kind: that is a programming error at startup.
func (r *Registry) Register(ch Channel) {
	if _, dup := r.byKind[ch.Kind()]; dup {
		panic(fmt.Sprintf("notify: channel kind %q registered twice", ch.Kind()))
	}

	r.byKind[ch.Kind()] = ch
}

func (r *Registry) Get(kind string) (Channel, bool) {
	ch, ok := r.byKind[kind]

	return ch, ok
}

// Kinds is never nil so it serializes as [] rather than null.
func (r *Registry) Kinds() []string {
	return slices.AppendSeq([]string{}, slices.Values(slices.Sorted(maps.Keys(r.byKind))))
}
