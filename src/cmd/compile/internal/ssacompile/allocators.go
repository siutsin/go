// Code generated from _gen/allocators.go using 'go generate'; DO NOT EDIT.

package ssacompile

import (
	"cmd/compile/internal/ssa"
	"math/bits"
	"sync"
	"unsafe"
)

var poolFreeRegStateSlice [29]sync.Pool

func (c *compilerCache) allocRegStateSlice(n int) []regState {
	var s []regState
	n2 := n
	if n2 < 8 {
		n2 = 8
	}
	b := bits.Len(uint(n2 - 1))
	v := poolFreeRegStateSlice[b-3].Get()
	if v == nil {
		s = make([]regState, 1<<b)
	} else {
		sp := v.(*[]regState)
		s = *sp
		*sp = nil
		c.hdrRegStateSlice = append(c.hdrRegStateSlice, sp)
	}
	s = s[:n]
	return s
}
func (c *compilerCache) freeRegStateSlice(s []regState) {
	clear(s[:cap(s)])
	b := bits.Len(uint(cap(s)) - 1)
	var sp *[]regState
	if len(c.hdrRegStateSlice) == 0 {
		sp = new([]regState)
	} else {
		sp = c.hdrRegStateSlice[len(c.hdrRegStateSlice)-1]
		c.hdrRegStateSlice[len(c.hdrRegStateSlice)-1] = nil
		c.hdrRegStateSlice = c.hdrRegStateSlice[:len(c.hdrRegStateSlice)-1]
	}
	*sp = s
	poolFreeRegStateSlice[b-3].Put(sp)
}

var poolFreeEndRegSliceSlice [29]sync.Pool

func (c *compilerCache) allocEndRegSliceSlice(n int) [][]endReg {
	var s [][]endReg
	n2 := n
	if n2 < 8 {
		n2 = 8
	}
	b := bits.Len(uint(n2 - 1))
	v := poolFreeEndRegSliceSlice[b-3].Get()
	if v == nil {
		s = make([][]endReg, 1<<b)
	} else {
		sp := v.(*[][]endReg)
		s = *sp
		*sp = nil
		c.hdrEndRegSliceSlice = append(c.hdrEndRegSliceSlice, sp)
	}
	s = s[:n]
	return s
}
func (c *compilerCache) freeEndRegSliceSlice(s [][]endReg) {
	clear(s[:cap(s)])
	b := bits.Len(uint(cap(s)) - 1)
	var sp *[][]endReg
	if len(c.hdrEndRegSliceSlice) == 0 {
		sp = new([][]endReg)
	} else {
		sp = c.hdrEndRegSliceSlice[len(c.hdrEndRegSliceSlice)-1]
		c.hdrEndRegSliceSlice[len(c.hdrEndRegSliceSlice)-1] = nil
		c.hdrEndRegSliceSlice = c.hdrEndRegSliceSlice[:len(c.hdrEndRegSliceSlice)-1]
	}
	*sp = s
	poolFreeEndRegSliceSlice[b-3].Put(sp)
}

var poolFreeStartRegSliceSlice [29]sync.Pool

func (c *compilerCache) allocStartRegSliceSlice(n int) [][]startReg {
	var s [][]startReg
	n2 := n
	if n2 < 8 {
		n2 = 8
	}
	b := bits.Len(uint(n2 - 1))
	v := poolFreeStartRegSliceSlice[b-3].Get()
	if v == nil {
		s = make([][]startReg, 1<<b)
	} else {
		sp := v.(*[][]startReg)
		s = *sp
		*sp = nil
		c.hdrStartRegSliceSlice = append(c.hdrStartRegSliceSlice, sp)
	}
	s = s[:n]
	return s
}
func (c *compilerCache) freeStartRegSliceSlice(s [][]startReg) {
	clear(s[:cap(s)])
	b := bits.Len(uint(cap(s)) - 1)
	var sp *[][]startReg
	if len(c.hdrStartRegSliceSlice) == 0 {
		sp = new([][]startReg)
	} else {
		sp = c.hdrStartRegSliceSlice[len(c.hdrStartRegSliceSlice)-1]
		c.hdrStartRegSliceSlice[len(c.hdrStartRegSliceSlice)-1] = nil
		c.hdrStartRegSliceSlice = c.hdrStartRegSliceSlice[:len(c.hdrStartRegSliceSlice)-1]
	}
	*sp = s
	poolFreeStartRegSliceSlice[b-3].Put(sp)
}

var poolFreeIDSliceSlice [29]sync.Pool

func (c *compilerCache) allocIDSliceSlice(n int) [][]ssa.ID {
	var s [][]ssa.ID
	n2 := n
	if n2 < 8 {
		n2 = 8
	}
	b := bits.Len(uint(n2 - 1))
	v := poolFreeIDSliceSlice[b-3].Get()
	if v == nil {
		s = make([][]ssa.ID, 1<<b)
	} else {
		sp := v.(*[][]ssa.ID)
		s = *sp
		*sp = nil
		c.hdrIDSliceSlice = append(c.hdrIDSliceSlice, sp)
	}
	s = s[:n]
	return s
}
func (c *compilerCache) freeIDSliceSlice(s [][]ssa.ID) {
	clear(s[:cap(s)])
	b := bits.Len(uint(cap(s)) - 1)
	var sp *[][]ssa.ID
	if len(c.hdrIDSliceSlice) == 0 {
		sp = new([][]ssa.ID)
	} else {
		sp = c.hdrIDSliceSlice[len(c.hdrIDSliceSlice)-1]
		c.hdrIDSliceSlice[len(c.hdrIDSliceSlice)-1] = nil
		c.hdrIDSliceSlice = c.hdrIDSliceSlice[:len(c.hdrIDSliceSlice)-1]
	}
	*sp = s
	poolFreeIDSliceSlice[b-3].Put(sp)
}

// roundUpPowerOfTwo returns the smallest power of two
// greater than or equal to n. n must be positive.
func roundUpPowerOfTwo(n int) int {
	return 1 << bits.Len(uint(n-1))
}
func (c *factsTable) allocLimitFactSlice(n int) []limitFact {
	var base byte
	var derived limitFact
	if unsafe.Sizeof(derived)%unsafe.Sizeof(base) != 0 {
		panic("bad")
	}
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	b := c.cache.AllocByteSlice(n * int(scale))
	derivedCap := cap(b) / int(scale)
	data := (*limitFact)(unsafe.Pointer(unsafe.SliceData(b)))
	return unsafe.Slice(data, derivedCap)[:n]
}
func (c *factsTable) freeLimitFactSlice(s []limitFact) {
	if cap(s) == 0 {
		return
	}
	var base byte
	var derived limitFact
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	byteCap := cap(s) * int(scale)
	byteCap = roundUpPowerOfTwo(byteCap)
	data := (*byte)(unsafe.Pointer(unsafe.SliceData(s)))
	b := unsafe.Slice(data, byteCap)
	c.cache.FreeByteSlice(b)
}

// appendLimitFactSlice appends elems to s, growing through factsTable when needed.
// s must be nil or retain the start pointer and capacity returned by
// allocLimitFactSlice/appendLimitFactSlice; only its length may change.
// Use only the returned slice; growth returns s's old backing to the pool.
func (c *factsTable) appendLimitFactSlice(s []limitFact, elems ...limitFact) []limitFact {
	oldLen := len(s)
	n := oldLen + len(elems)
	if n <= cap(s) {
		s = s[:n]
		copy(s[oldLen:], elems)
		return s
	}
	ns := c.allocLimitFactSlice(n)
	copy(ns, s)
	copy(ns[oldLen:], elems)
	c.freeLimitFactSlice(s)
	return ns
}

var byteSlicePoolElemTypes = []string{"limitFact"}
