// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssacompile

import (
	rtabi "internal/abi"
	"slices"
	"testing"
	"unsafe"

	"cmd/compile/internal/ssa"
)

// TestLimitFactPoolPointerInvariant guards the no-scan byte pool.
func TestLimitFactPoolPointerInvariant(t *testing.T) {
	maxAlign := rtabi.TypeFor[ssa.Limit]().Align()
	pointerFree := []struct {
		name string
		typ  *rtabi.Type
	}{
		{"limitFact", rtabi.TypeFor[limitFact]()},
	}
	for _, tc := range pointerFree {
		if tc.typ.Pointers() {
			t.Errorf("%s must remain pointer-free (PtrBytes=%d)", tc.name, tc.typ.PtrBytes)
		}
		if tc.typ.Size() == 0 || tc.typ.Size() > 256 {
			t.Errorf("%s is %d bytes; byte-pool elements must be 1..256 bytes", tc.name, tc.typ.Size())
		}
		if tc.typ.Align() > maxAlign {
			t.Errorf("%s requires alignment %d; byte pool provides %d", tc.name, tc.typ.Align(), maxAlign)
		}
	}

	got := make([]string, len(pointerFree))
	for i, tc := range pointerFree {
		got[i] = tc.name
	}
	slices.Sort(got)
	want := slices.Clone(byteSlicePoolElemTypes)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("checked types %v do not match byte-pool types %v", got, want)
	}
}

// TestLimitFactAppendAllocator checks variadic, in-place, and pooled growth.
func TestLimitFactAppendAllocator(t *testing.T) {
	ft := &factsTable{cache: new(ssa.Cache)}

	facts := ft.appendLimitFactSlice(nil, limitFact{vid: 1}, limitFact{vid: 2})
	data := unsafe.SliceData(facts)
	facts = ft.appendLimitFactSlice(facts, limitFact{vid: 3})
	if unsafe.SliceData(facts) != data {
		t.Error("appendLimitFactSlice reallocated despite spare capacity")
	}
	for len(facts) < cap(facts) {
		facts = ft.appendLimitFactSlice(facts, limitFact{vid: ssa.ID(len(facts) + 1)})
	}
	facts = ft.appendLimitFactSlice(facts, limitFact{vid: ssa.ID(len(facts) + 1)})
	for i, fact := range facts {
		want := ssa.ID(i + 1)
		if fact.vid != want {
			t.Errorf("appendLimitFactSlice[%d].vid = %d, want %d", i, fact.vid, want)
		}
	}
	ft.freeLimitFactSlice(facts)
}
