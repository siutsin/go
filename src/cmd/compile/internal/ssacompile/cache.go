// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssacompile

import "cmd/compile/internal/ssa"

type compilerCache struct {
	visited      map[ssa.Edge]bool
	latticeCells map[*ssa.Value]lattice
	defUse       map[*ssa.Value][]*ssa.Value
	defBlock     map[*ssa.Value][]*ssa.Block

	// Reusable regalloc scratch, retained within fixed per-worker bounds.
	// The retained inner backing arrays contain pointer-free liveInfo and
	// desiredStateEntry elements.
	regallocLive    [][]liveInfo
	regallocDesired []desiredState

	// Free headers used to put slices in sync.Pools without allocation.
	hdrRegStateSlice      []*[]regState
	hdrEndRegSliceSlice   []*[][]endReg
	hdrStartRegSliceSlice []*[][]startReg
	hdrIDSliceSlice       []*[][]ssa.ID

	// Reusable prove scratch scoped to one function. Cleanup clears the
	// pointer-bearing map after each use; Compile drops it after the final use.
	proveOrderings map[ssa.ID]*ordering
}

func getCompilerCache(c *ssa.Cache) *compilerCache {
	cache, _ := c.CompilerCache.(*compilerCache)
	if cache == nil {
		cache = new(compilerCache)
		c.CompilerCache = cache
	}
	return cache
}
