// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	rtabi "internal/abi"
	"slices"
	"testing"
	"unsafe"
)

// TestRoundUpPowerOfTwo checks the capacity reconstruction used when a
// pointer-free derived slice is returned to the shared []byte pool.
func TestRoundUpPowerOfTwo(t *testing.T) {
	for _, tc := range []struct {
		n    int
		want int
	}{
		{240, 256},
		{256, 256},
		{257, 512},
	} {
		if got := roundUpPowerOfTwo(tc.n); got != tc.want {
			t.Errorf("roundUpPowerOfTwo(%d) = %d, want %d", tc.n, got, tc.want)
		}
	}
}

// TestCacheByteAllocatorClearOnFree checks full-capacity clearing before pooling.
func TestCacheByteAllocatorClearOnFree(t *testing.T) {
	c := new(Cache)

	a := c.AllocIntSlice(1)
	av := a[:cap(a)]
	for i := range av {
		av[i] = -1
	}
	c.FreeIntSlice(a)
	for i, v := range av {
		if v != 0 {
			t.Fatalf("FreeIntSlice did not clear backing[%d] = %d, want 0", i, v)
		}
	}
}

func TestCacheByteSliceClearOnFree(t *testing.T) {
	c := new(Cache)

	a := c.AllocByteSlice(1)
	av := a[:cap(a)]
	for i := range av {
		av[i] = 0xff
	}
	c.FreeByteSlice(a)
	for i, v := range av {
		if v != 0 {
			t.Fatalf("FreeByteSlice did not clear backing[%d] = %d, want 0", i, v)
		}
	}
}

// TestCachePoolPointerInvariant guards the no-scan byte pool.
func TestCachePoolPointerInvariant(t *testing.T) {
	maxAlign := rtabi.TypeFor[Limit]().Align()
	pointerFree := []struct {
		name string
		typ  *rtabi.Type
	}{
		{"Limit", rtabi.TypeFor[Limit]()},
		{"knownBitsEntry", rtabi.TypeFor[knownBitsEntry]()},
		{"int", rtabi.TypeFor[int]()},
		{"int32", rtabi.TypeFor[int32]()},
		{"int8", rtabi.TypeFor[int8]()},
		{"bool", rtabi.TypeFor[bool]()},
		{"ID", rtabi.TypeFor[ID]()},
		{"uint", rtabi.TypeFor[uint]()},
	}
	for _, tc := range pointerFree {
		if tc.typ.Pointers() {
			t.Errorf("%s must remain pointer-free (PtrBytes=%d)",
				tc.name, tc.typ.PtrBytes)
		}
		if tc.typ.Size() == 0 || tc.typ.Size() > 256 {
			t.Errorf("%s is %d bytes; byte-pool elements must be 1..256 bytes",
				tc.name, tc.typ.Size())
		}
		if tc.typ.Align() > maxAlign {
			t.Errorf("%s requires alignment %d; byte pool provides %d",
				tc.name, tc.typ.Align(), maxAlign)
		}
	}

	// Keep the checked list matched to the generated derived types.
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

// TestCacheByteAllocatorAlignment checks alignment of byte-backed slices.
func TestCacheByteAllocatorAlignment(t *testing.T) {
	c := new(Cache)
	for _, n := range []int{1, 300} {
		s := c.AllocLimitSlice(n)
		if d := uintptr(unsafe.Pointer(unsafe.SliceData(s))); d%unsafe.Alignof(s[0]) != 0 {
			t.Errorf("Limit backing %#x is not aligned to %d", d, unsafe.Alignof(s[0]))
		}
		c.FreeLimitSlice(s)
	}
}
