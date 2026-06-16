// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssagen

import (
	"testing"

	"cmd/compile/internal/ir"
	"cmd/compile/internal/ssa"
)

func newDefvarsMap(n int) map[ir.Node]*ssa.Value {
	m := make(map[ir.Node]*ssa.Value)
	for range n {
		m[ssaMarker("test")] = nil
	}
	return m
}

func TestCacheDefvarsMaps(t *testing.T) {
	atLimit := newDefvarsMap(maxCachedDefvarsEntries)
	overLimit := newDefvarsMap(maxCachedDefvarsEntries + 1)
	got := cacheDefvarsMaps(nil, []map[ir.Node]*ssa.Value{nil, atLimit, overLimit})
	if len(got) != 1 || got[0] == nil {
		t.Fatalf("cacheDefvarsMaps returned %d maps; want 1", len(got))
	}
	if len(got[0]) != 0 {
		t.Fatalf("retained map has len %d; want 0", len(got[0]))
	}
	if len(overLimit) != maxCachedDefvarsEntries+1 {
		t.Fatalf("oversized map was cleared; len %d, want %d", len(overLimit), maxCachedDefvarsEntries+1)
	}

	defvars := make([]map[ir.Node]*ssa.Value, maxCachedDefvarsMaps+2)
	for id := 1; id < len(defvars); id++ {
		defvars[id] = newDefvarsMap(1)
	}
	got = cacheDefvarsMaps(nil, defvars)
	if len(got) != maxCachedDefvarsMaps {
		t.Fatalf("cacheDefvarsMaps returned %d maps; want %d", len(got), maxCachedDefvarsMaps)
	}
	if len(defvars[maxCachedDefvarsMaps+1]) != 1 {
		t.Fatal("map past count limit was cleared")
	}

	candidate := newDefvarsMap(1)
	got = cacheDefvarsMaps(got, []map[ir.Node]*ssa.Value{nil, candidate})
	if len(got) != maxCachedDefvarsMaps {
		t.Fatalf("full cache grew to %d maps; want %d", len(got), maxCachedDefvarsMaps)
	}
	if len(candidate) != 1 {
		t.Fatal("map added to full cache was cleared")
	}
}
