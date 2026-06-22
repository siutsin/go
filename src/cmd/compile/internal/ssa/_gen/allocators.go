// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// Pointer-free derived slices share ByteSlice's no-scan backing arrays.
// Pointer-bearing slices must use scanned pools; []*Block can share []*Value
// because their elements have compatible pointer layouts.

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"strings"
)

type allocator struct {
	name     string // name for alloc/free functions
	typ      string // the type they return/accept
	mak      string // code to make a new object (takes power-of-2 size as fmt arg)
	capacity string // code to calculate the capacity of an object. Should always report a power of 2.
	resize   string // code to shrink to sub-power-of-two size (takes size as fmt arg)
	clear    string // code for clearing object before putting it on the free list
	minLog   int    // log_2 of minimum allocation size
	maxLog   int    // log_2 of maximum allocation size
}

type derived struct {
	name   string // name for alloc/free functions
	typ    string // the type they return/accept
	base   string // underlying allocator
	append bool   // generate an append helper that grows through this allocator
}

func genAllocators() {
	allocators := []allocator{
		{
			name:     "ValueSlice",
			typ:      "[]*Value",
			capacity: "cap(%s)",
			mak:      "make([]*Value, %s)",
			resize:   "%s[:%s]",
			clear:    "clear(%s)",
			minLog:   5,
			maxLog:   32,
		},
		{
			// ByteSlice backs pointer-free derived slices. Its no-scan backing
			// must never hold pointers. Tests guard this invariant.
			//
			// The 256-byte minimum bypasses the tiny allocator. Regular heap
			// allocations satisfy the alignment required by every derived type.
			// Tests check each type's size and alignment and the actual backing
			// address. Limiting element sizes to 256 bytes also makes power-of-two
			// bucket reconstruction exact.
			name:     "ByteSlice",
			typ:      "[]byte",
			capacity: "cap(%s)",
			mak:      "make([]byte, %s)",
			resize:   "%s[:%s]",
			clear:    "clear(%[1]s[:cap(%[1]s)])",
			minLog:   8,
			maxLog:   36,
		},
		{
			name:     "SparseSet",
			typ:      "*sparseSet",
			capacity: "%s.cap()",
			mak:      "newSparseSet(%s)",
			resize:   "", // larger-sized sparse sets are ok
			clear:    "%s.clear()",
			minLog:   5,
			maxLog:   32,
		},
		{
			name:     "SparseMap",
			typ:      "*sparseMap",
			capacity: "%s.cap()",
			mak:      "newSparseMap(%s)",
			resize:   "", // larger-sized sparse maps are ok
			clear:    "%s.clear()",
			minLog:   5,
			maxLog:   32,
		},
		{
			name:     "SparseMapPos",
			typ:      "*sparseMapPos",
			capacity: "%s.cap()",
			mak:      "newSparseMapPos(%s)",
			resize:   "", // larger-sized sparse maps are ok
			clear:    "%s.clear()",
			minLog:   5,
			maxLog:   32,
		},
	}
	limitType := "[]limit"
	if splitPhase >= phase0Export {
		limitType = "[]Limit"
		allocators[2].typ = "*SparseSet"
		allocators[2].mak = "NewSparseSet(%s)"
		allocators[2].clear = "%s.Clear()"
		allocators[3].mak = "NewSparseMap(%s)"
		allocators[3].clear = "%s.Clear()"
		allocators[4].typ = "*SparseMapPos"
		allocators[4].clear = "%s.Clear()"
	}
	compilerAllocators := []allocator{
		{
			name:     "RegStateSlice",
			typ:      "[]regState",
			capacity: "cap(%s)",
			mak:      "make([]regState, %s)",
			resize:   "%s[:%s]",
			clear:    "clear(%[1]s[:cap(%[1]s)])",
			minLog:   3,
			maxLog:   32,
		},
		{
			// These pools hold outer slice headers, so their backing arrays must
			// be scanned. Separate typed pools avoid unsafe reinterpretation.
			// Clearing the full capacity drops all inner backing arrays.
			name:     "EndRegSliceSlice",
			typ:      "[][]endReg",
			capacity: "cap(%s)",
			mak:      "make([][]endReg, %s)",
			resize:   "%s[:%s]",
			clear:    "clear(%[1]s[:cap(%[1]s)])",
			minLog:   3,
			maxLog:   32,
		},
		{
			name:     "StartRegSliceSlice",
			typ:      "[][]startReg",
			capacity: "cap(%s)",
			mak:      "make([][]startReg, %s)",
			resize:   "%s[:%s]",
			clear:    "clear(%[1]s[:cap(%[1]s)])",
			minLog:   3,
			maxLog:   32,
		},
		{
			name:     "IDSliceSlice",
			typ:      "[][]ssa.ID",
			capacity: "cap(%s)",
			mak:      "make([][]ssa.ID, %s)",
			resize:   "%s[:%s]",
			clear:    "clear(%[1]s[:cap(%[1]s)])",
			minLog:   3,
			maxLog:   32,
		},
	}
	deriveds := []derived{
		{
			name: "BlockSlice",
			typ:  "[]*Block",
			base: "ValueSlice",
		},
		{
			name: "LimitSlice",
			typ:  limitType,
			base: "ByteSlice",
		},
		{
			name: "IntSlice",
			typ:  "[]int",
			base: "ByteSlice",
		},
		{
			name: "Int32Slice",
			typ:  "[]int32",
			base: "ByteSlice",
		},
		{
			name: "Int8Slice",
			typ:  "[]int8",
			base: "ByteSlice",
		},
		{
			name: "BoolSlice",
			typ:  "[]bool",
			base: "ByteSlice",
		},
		{
			name:   "IDSlice",
			typ:    "[]ID",
			base:   "ByteSlice",
			append: true,
		},
		{
			name: "UintSlice",
			typ:  "[]uint",
			base: "ByteSlice",
		},
		{
			name: "KnownBitsEntriesSlice",
			typ:  "[]knownBitsEntry",
			base: "ByteSlice",
		},
	}
	compilerDeriveds := []derived{
		{
			name:   "LimitFactSlice",
			typ:    "[]limitFact",
			base:   "ByteSlice",
			append: true,
		},
	}

	w := new(bytes.Buffer)
	fmt.Fprintf(w, "// Code generated from _gen/allocators.go using 'go generate'; DO NOT EDIT.\n")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "package %s\n", splitCorePkg)

	fmt.Fprintln(w, "import (")
	fmt.Fprintln(w, "\"math/bits\"")
	fmt.Fprintln(w, "\"sync\"")
	fmt.Fprintln(w, "\"unsafe\"")
	fmt.Fprintln(w, ")")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "// roundUpPowerOfTwo returns the smallest power of two")
	fmt.Fprintln(w, "// greater than or equal to n. n must be positive.")
	fmt.Fprintln(w, "func roundUpPowerOfTwo(n int) int {")
	fmt.Fprintln(w, "return 1 << bits.Len(uint(n-1))")
	fmt.Fprintln(w, "}")
	for _, a := range allocators {
		genAllocator(w, a, "Cache", splitTitle)
	}
	for _, d := range deriveds {
		for _, base := range allocators {
			if base.name == d.base {
				genDerived(w, d, base, "Cache", splitTitle, "c", splitTitle)
				if d.append {
					genAppend(w, d.name, d.typ, "Cache", splitTitle)
				}
				break
			}
		}
	}

	// Emit ByteSlice-derived element names so the pointer-layout test stays
	// synchronized with the generator.
	var byteSliceElems []string
	for _, d := range deriveds {
		if d.base == "ByteSlice" {
			byteSliceElems = append(byteSliceElems, fmt.Sprintf("%q", d.typ[2:]))
		}
	}
	fmt.Fprintf(w, "\nvar byteSlicePoolElemTypes = []string{%s}\n", strings.Join(byteSliceElems, ", "))
	// gofmt result
	b := w.Bytes()
	var err error
	b, err = format.Source(b)
	if err != nil {
		fmt.Printf("%s\n", w.Bytes())
		panic(err)
	}

	mkdirOutFile(allocatorsFile)
	if err := os.WriteFile(outFile(allocatorsFile), b, 0666); err != nil {
		log.Fatalf("can't write output: %v\n", err)
	}

	w.Reset()
	fmt.Fprintf(w, "// Code generated from _gen/allocators.go using 'go generate'; DO NOT EDIT.\n")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "package ssacompile")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "import (")
	fmt.Fprintln(w, "\"cmd/compile/internal/ssa\"")
	fmt.Fprintln(w, "\"math/bits\"")
	fmt.Fprintln(w, "\"sync\"")
	fmt.Fprintln(w, "\"unsafe\"")
	fmt.Fprintln(w, ")")
	for _, a := range compilerAllocators {
		genAllocator(w, a, "compilerCache", identity)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "// roundUpPowerOfTwo returns the smallest power of two")
	fmt.Fprintln(w, "// greater than or equal to n. n must be positive.")
	fmt.Fprintln(w, "func roundUpPowerOfTwo(n int) int {")
	fmt.Fprintln(w, "return 1 << bits.Len(uint(n-1))")
	fmt.Fprintln(w, "}")
	for _, d := range compilerDeriveds {
		for _, base := range allocators {
			if base.name == d.base {
				genDerived(w, d, base, "factsTable", identity, "c.cache", splitTitle)
				if d.append {
					genAppend(w, d.name, d.typ, "factsTable", identity)
				}
				break
			}
		}
	}
	// Keep the package's pointer-layout test synchronized with the generator.
	var compilerByteSliceElems []string
	for _, d := range compilerDeriveds {
		if d.base == "ByteSlice" {
			compilerByteSliceElems = append(compilerByteSliceElems, fmt.Sprintf("%q", d.typ[2:]))
		}
	}
	fmt.Fprintf(w, "\nvar byteSlicePoolElemTypes = []string{%s}\n", strings.Join(compilerByteSliceElems, ", "))
	b, err = format.Source(w.Bytes())
	if err != nil {
		fmt.Printf("%s\n", w.Bytes())
		panic(err)
	}

	const compilerAllocatorsFile = "../ssacompile/allocators.go"
	mkdirOutFile(compilerAllocatorsFile)
	if err := os.WriteFile(outFile(compilerAllocatorsFile), b, 0666); err != nil {
		log.Fatalf("can't write output: %v\n", err)
	}
}
func genAllocator(w io.Writer, a allocator, cacheType string, title func(string) string) {
	fmt.Fprintf(w, "var poolFree%s [%d]sync.Pool\n", a.name, a.maxLog-a.minLog)
	fmt.Fprintf(w, "func (c *%s) %s%s(n int) %s {\n", cacheType, title("alloc"), a.name, a.typ)
	fmt.Fprintf(w, "var s %s\n", a.typ)
	fmt.Fprintf(w, "n2 := n\n")
	fmt.Fprintf(w, "if n2 < %d { n2 = %d }\n", 1<<a.minLog, 1<<a.minLog)
	fmt.Fprintf(w, "b := bits.Len(uint(n2-1))\n")
	fmt.Fprintf(w, "v := poolFree%s[b-%d].Get()\n", a.name, a.minLog)
	fmt.Fprintf(w, "if v == nil {\n")
	fmt.Fprintf(w, "  s = %s\n", fmt.Sprintf(a.mak, "1<<b"))
	fmt.Fprintf(w, "} else {\n")
	if a.typ[0] == '*' {
		fmt.Fprintf(w, "s = v.(%s)\n", a.typ)
	} else {
		fmt.Fprintf(w, "sp := v.(*%s)\n", a.typ)
		fmt.Fprintf(w, "s = *sp\n")
		fmt.Fprintf(w, "*sp = nil\n")
		fmt.Fprintf(w, "c.hdr%s = append(c.hdr%s, sp)\n", a.name, a.name)
	}
	fmt.Fprintf(w, "}\n")
	if a.resize != "" {
		fmt.Fprintf(w, "s = %s\n", fmt.Sprintf(a.resize, "s", "n"))
	}
	fmt.Fprintf(w, "return s\n")
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "func (c *%s) %s%s(s %s) {\n", cacheType, title("free"), a.name, a.typ)
	fmt.Fprintf(w, "%s\n", fmt.Sprintf(a.clear, "s"))
	fmt.Fprintf(w, "b := bits.Len(uint(%s) - 1)\n", fmt.Sprintf(a.capacity, "s"))
	if a.typ[0] == '*' {
		fmt.Fprintf(w, "poolFree%s[b-%d].Put(s)\n", a.name, a.minLog)
	} else {
		fmt.Fprintf(w, "var sp *%s\n", a.typ)
		fmt.Fprintf(w, "if len(c.hdr%s) == 0 {\n", a.name)
		fmt.Fprintf(w, "  sp = new(%s)\n", a.typ)
		fmt.Fprintf(w, "} else {\n")
		fmt.Fprintf(w, "  sp = c.hdr%s[len(c.hdr%s)-1]\n", a.name, a.name)
		fmt.Fprintf(w, "  c.hdr%s[len(c.hdr%s)-1] = nil\n", a.name, a.name)
		fmt.Fprintf(w, "  c.hdr%s = c.hdr%s[:len(c.hdr%s)-1]\n", a.name, a.name, a.name)
		fmt.Fprintf(w, "}\n")
		fmt.Fprintf(w, "*sp = s\n")
		fmt.Fprintf(w, "poolFree%s[b-%d].Put(sp)\n", a.name, a.minLog)
	}
	fmt.Fprintf(w, "}\n")
}

func genAppend(w io.Writer, name, typ, cacheType string, title func(string) string) {
	if typ[:2] != "[]" {
		panic(fmt.Sprintf("bad append type: %s", typ))
	}
	elem := typ[2:]
	appendName := title("append") + name
	allocName := title("alloc") + name
	freeName := title("free") + name
	fmt.Fprintf(w, "// %s appends elems to s, growing through %s when needed.\n", appendName, cacheType)
	fmt.Fprintf(w, "// s must be nil or retain the start pointer and capacity returned by\n")
	fmt.Fprintf(w, "// %s/%s; only its length may change.\n", allocName, appendName)
	fmt.Fprintf(w, "// Use only the returned slice; growth returns s's old backing to the pool.\n")
	fmt.Fprintf(w, "func (c *%s) %s(s %s, elems ...%s) %s {\n", cacheType, appendName, typ, elem, typ)
	fmt.Fprintf(w, "oldLen := len(s)\n")
	fmt.Fprintf(w, "n := oldLen + len(elems)\n")
	fmt.Fprintf(w, "if n <= cap(s) {\n")
	fmt.Fprintf(w, "  s = s[:n]\n")
	fmt.Fprintf(w, "  copy(s[oldLen:], elems)\n")
	fmt.Fprintf(w, "  return s\n")
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "ns := c.%s(n)\n", allocName)
	fmt.Fprintf(w, "copy(ns, s)\n")
	fmt.Fprintf(w, "copy(ns[oldLen:], elems)\n")
	fmt.Fprintf(w, "c.%s(s)\n", freeName)
	fmt.Fprintf(w, "return ns\n")
	fmt.Fprintf(w, "}\n")
}

func genDerived(w io.Writer, d derived, base allocator, cacheType string, title func(string) string, baseReceiver string, baseTitle func(string) string) {
	if d.typ[:2] != "[]" || base.typ[:2] != "[]" {
		panic(fmt.Sprintf("bad derived types: %s %s", d.typ, base.typ))
	}
	// ByteSlice scales in bytes per derived element; typed bases scale in
	// derived elements per base element.
	byteBase := base.typ == "[]byte"
	fmt.Fprintf(w, "func (c *%s) %s%s(n int) %s {\n", cacheType, title("alloc"), d.name, d.typ)
	fmt.Fprintf(w, "var base %s\n", base.typ[2:])
	fmt.Fprintf(w, "var derived %s\n", d.typ[2:])
	if byteBase {
		fmt.Fprintf(w, "if unsafe.Sizeof(derived)%%unsafe.Sizeof(base) != 0 { panic(\"bad\") }\n")
		fmt.Fprintf(w, "scale := unsafe.Sizeof(derived)/unsafe.Sizeof(base)\n")
		fmt.Fprintf(w, "b := %s.%s%s(n*int(scale))\n", baseReceiver, baseTitle("alloc"), base.name)
	} else {
		fmt.Fprintf(w, "if unsafe.Sizeof(base)%%unsafe.Sizeof(derived) != 0 { panic(\"bad\") }\n")
		fmt.Fprintf(w, "scale := unsafe.Sizeof(base)/unsafe.Sizeof(derived)\n")
		fmt.Fprintf(w, "b := %s.%s%s(int((uintptr(n)+scale-1)/scale))\n", baseReceiver, baseTitle("alloc"), base.name)
	}
	if byteBase {
		fmt.Fprintf(w, "derivedCap := cap(b)/int(scale)\n")
	} else {
		fmt.Fprintf(w, "derivedCap := cap(b)*int(scale)\n")
	}
	fmt.Fprintf(w, "data := (*%s)(unsafe.Pointer(unsafe.SliceData(b)))\n", d.typ[2:])
	fmt.Fprintf(w, "return unsafe.Slice(data, derivedCap)[:n]\n")
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "func (c *%s) %s%s(s %s) {\n", cacheType, title("free"), d.name, d.typ)
	fmt.Fprintf(w, "if cap(s) == 0 { return }\n")
	// Free converts the derived slice back to the base slice by capacity
	// rather than length. A caller may return a zero-length slice that still
	// owns pooled storage.
	fmt.Fprintf(w, "var base %s\n", base.typ[2:])
	fmt.Fprintf(w, "var derived %s\n", d.typ[2:])
	if byteBase {
		fmt.Fprintf(w, "scale := unsafe.Sizeof(derived)/unsafe.Sizeof(base)\n")
		// cap(s) omits trailing bytes that do not form a complete element.
		// Round up to recover the original power-of-two byte bucket. The
		// pointer-invariant test limits element size so this is exact.
		fmt.Fprintf(w, "byteCap := cap(s)*int(scale)\n")
		fmt.Fprintf(w, "byteCap = roundUpPowerOfTwo(byteCap)\n")
		fmt.Fprintf(w, "data := (*%s)(unsafe.Pointer(unsafe.SliceData(s)))\n", base.typ[2:])
		fmt.Fprintf(w, "b := unsafe.Slice(data, byteCap)\n")
	} else {
		fmt.Fprintf(w, "scale := unsafe.Sizeof(base)/unsafe.Sizeof(derived)\n")
		fmt.Fprintf(w, "baseCap := int((uintptr(cap(s))+scale-1)/scale)\n")
		fmt.Fprintf(w, "data := (*%s)(unsafe.Pointer(unsafe.SliceData(s)))\n", base.typ[2:])
		fmt.Fprintf(w, "b := unsafe.Slice(data, baseCap)\n")
	}
	fmt.Fprintf(w, "%s.%s%s(b)\n", baseReceiver, baseTitle("free"), base.name)
	fmt.Fprintf(w, "}\n")
}
