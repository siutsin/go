// Code generated from _gen/allocators.go using 'go generate'; DO NOT EDIT.

package ssa

import (
	"math/bits"
	"sync"
	"unsafe"
)

// roundUpPowerOfTwo returns the smallest power of two
// greater than or equal to n. n must be positive.
func roundUpPowerOfTwo(n int) int {
	return 1 << bits.Len(uint(n-1))
}

var poolFreeValueSlice [27]sync.Pool

func (c *Cache) AllocValueSlice(n int) []*Value {
	var s []*Value
	n2 := n
	if n2 < 32 {
		n2 = 32
	}
	b := bits.Len(uint(n2 - 1))
	v := poolFreeValueSlice[b-5].Get()
	if v == nil {
		s = make([]*Value, 1<<b)
	} else {
		sp := v.(*[]*Value)
		s = *sp
		*sp = nil
		c.hdrValueSlice = append(c.hdrValueSlice, sp)
	}
	s = s[:n]
	return s
}
func (c *Cache) FreeValueSlice(s []*Value) {
	clear(s)
	b := bits.Len(uint(cap(s)) - 1)
	var sp *[]*Value
	if len(c.hdrValueSlice) == 0 {
		sp = new([]*Value)
	} else {
		sp = c.hdrValueSlice[len(c.hdrValueSlice)-1]
		c.hdrValueSlice[len(c.hdrValueSlice)-1] = nil
		c.hdrValueSlice = c.hdrValueSlice[:len(c.hdrValueSlice)-1]
	}
	*sp = s
	poolFreeValueSlice[b-5].Put(sp)
}

var poolFreeByteSlice [28]sync.Pool

func (c *Cache) AllocByteSlice(n int) []byte {
	var s []byte
	n2 := n
	if n2 < 256 {
		n2 = 256
	}
	b := bits.Len(uint(n2 - 1))
	v := poolFreeByteSlice[b-8].Get()
	if v == nil {
		s = make([]byte, 1<<b)
	} else {
		sp := v.(*[]byte)
		s = *sp
		*sp = nil
		c.hdrByteSlice = append(c.hdrByteSlice, sp)
	}
	s = s[:n]
	return s
}
func (c *Cache) FreeByteSlice(s []byte) {
	clear(s[:cap(s)])
	b := bits.Len(uint(cap(s)) - 1)
	var sp *[]byte
	if len(c.hdrByteSlice) == 0 {
		sp = new([]byte)
	} else {
		sp = c.hdrByteSlice[len(c.hdrByteSlice)-1]
		c.hdrByteSlice[len(c.hdrByteSlice)-1] = nil
		c.hdrByteSlice = c.hdrByteSlice[:len(c.hdrByteSlice)-1]
	}
	*sp = s
	poolFreeByteSlice[b-8].Put(sp)
}

var poolFreeSparseSet [27]sync.Pool

func (c *Cache) AllocSparseSet(n int) *SparseSet {
	var s *SparseSet
	n2 := n
	if n2 < 32 {
		n2 = 32
	}
	b := bits.Len(uint(n2 - 1))
	v := poolFreeSparseSet[b-5].Get()
	if v == nil {
		s = NewSparseSet(1 << b)
	} else {
		s = v.(*SparseSet)
	}
	return s
}
func (c *Cache) FreeSparseSet(s *SparseSet) {
	s.Clear()
	b := bits.Len(uint(s.cap()) - 1)
	poolFreeSparseSet[b-5].Put(s)
}

var poolFreeSparseMap [27]sync.Pool

func (c *Cache) AllocSparseMap(n int) *sparseMap {
	var s *sparseMap
	n2 := n
	if n2 < 32 {
		n2 = 32
	}
	b := bits.Len(uint(n2 - 1))
	v := poolFreeSparseMap[b-5].Get()
	if v == nil {
		s = NewSparseMap(1 << b)
	} else {
		s = v.(*sparseMap)
	}
	return s
}
func (c *Cache) FreeSparseMap(s *sparseMap) {
	s.Clear()
	b := bits.Len(uint(s.cap()) - 1)
	poolFreeSparseMap[b-5].Put(s)
}

var poolFreeSparseMapPos [27]sync.Pool

func (c *Cache) AllocSparseMapPos(n int) *SparseMapPos {
	var s *SparseMapPos
	n2 := n
	if n2 < 32 {
		n2 = 32
	}
	b := bits.Len(uint(n2 - 1))
	v := poolFreeSparseMapPos[b-5].Get()
	if v == nil {
		s = newSparseMapPos(1 << b)
	} else {
		s = v.(*SparseMapPos)
	}
	return s
}
func (c *Cache) FreeSparseMapPos(s *SparseMapPos) {
	s.Clear()
	b := bits.Len(uint(s.cap()) - 1)
	poolFreeSparseMapPos[b-5].Put(s)
}
func (c *Cache) AllocBlockSlice(n int) []*Block {
	var base *Value
	var derived *Block
	if unsafe.Sizeof(base)%unsafe.Sizeof(derived) != 0 {
		panic("bad")
	}
	scale := unsafe.Sizeof(base) / unsafe.Sizeof(derived)
	b := c.AllocValueSlice(int((uintptr(n) + scale - 1) / scale))
	derivedCap := cap(b) * int(scale)
	data := (**Block)(unsafe.Pointer(unsafe.SliceData(b)))
	return unsafe.Slice(data, derivedCap)[:n]
}
func (c *Cache) FreeBlockSlice(s []*Block) {
	if cap(s) == 0 {
		return
	}
	var base *Value
	var derived *Block
	scale := unsafe.Sizeof(base) / unsafe.Sizeof(derived)
	baseCap := int((uintptr(cap(s)) + scale - 1) / scale)
	data := (**Value)(unsafe.Pointer(unsafe.SliceData(s)))
	b := unsafe.Slice(data, baseCap)
	c.FreeValueSlice(b)
}
func (c *Cache) AllocLimitSlice(n int) []Limit {
	var base byte
	var derived Limit
	if unsafe.Sizeof(derived)%unsafe.Sizeof(base) != 0 {
		panic("bad")
	}
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	b := c.AllocByteSlice(n * int(scale))
	derivedCap := cap(b) / int(scale)
	data := (*Limit)(unsafe.Pointer(unsafe.SliceData(b)))
	return unsafe.Slice(data, derivedCap)[:n]
}
func (c *Cache) FreeLimitSlice(s []Limit) {
	if cap(s) == 0 {
		return
	}
	var base byte
	var derived Limit
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	byteCap := cap(s) * int(scale)
	byteCap = roundUpPowerOfTwo(byteCap)
	data := (*byte)(unsafe.Pointer(unsafe.SliceData(s)))
	b := unsafe.Slice(data, byteCap)
	c.FreeByteSlice(b)
}
func (c *Cache) AllocIntSlice(n int) []int {
	var base byte
	var derived int
	if unsafe.Sizeof(derived)%unsafe.Sizeof(base) != 0 {
		panic("bad")
	}
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	b := c.AllocByteSlice(n * int(scale))
	derivedCap := cap(b) / int(scale)
	data := (*int)(unsafe.Pointer(unsafe.SliceData(b)))
	return unsafe.Slice(data, derivedCap)[:n]
}
func (c *Cache) FreeIntSlice(s []int) {
	if cap(s) == 0 {
		return
	}
	var base byte
	var derived int
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	byteCap := cap(s) * int(scale)
	byteCap = roundUpPowerOfTwo(byteCap)
	data := (*byte)(unsafe.Pointer(unsafe.SliceData(s)))
	b := unsafe.Slice(data, byteCap)
	c.FreeByteSlice(b)
}
func (c *Cache) AllocInt32Slice(n int) []int32 {
	var base byte
	var derived int32
	if unsafe.Sizeof(derived)%unsafe.Sizeof(base) != 0 {
		panic("bad")
	}
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	b := c.AllocByteSlice(n * int(scale))
	derivedCap := cap(b) / int(scale)
	data := (*int32)(unsafe.Pointer(unsafe.SliceData(b)))
	return unsafe.Slice(data, derivedCap)[:n]
}
func (c *Cache) FreeInt32Slice(s []int32) {
	if cap(s) == 0 {
		return
	}
	var base byte
	var derived int32
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	byteCap := cap(s) * int(scale)
	byteCap = roundUpPowerOfTwo(byteCap)
	data := (*byte)(unsafe.Pointer(unsafe.SliceData(s)))
	b := unsafe.Slice(data, byteCap)
	c.FreeByteSlice(b)
}
func (c *Cache) AllocInt8Slice(n int) []int8 {
	var base byte
	var derived int8
	if unsafe.Sizeof(derived)%unsafe.Sizeof(base) != 0 {
		panic("bad")
	}
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	b := c.AllocByteSlice(n * int(scale))
	derivedCap := cap(b) / int(scale)
	data := (*int8)(unsafe.Pointer(unsafe.SliceData(b)))
	return unsafe.Slice(data, derivedCap)[:n]
}
func (c *Cache) FreeInt8Slice(s []int8) {
	if cap(s) == 0 {
		return
	}
	var base byte
	var derived int8
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	byteCap := cap(s) * int(scale)
	byteCap = roundUpPowerOfTwo(byteCap)
	data := (*byte)(unsafe.Pointer(unsafe.SliceData(s)))
	b := unsafe.Slice(data, byteCap)
	c.FreeByteSlice(b)
}
func (c *Cache) AllocBoolSlice(n int) []bool {
	var base byte
	var derived bool
	if unsafe.Sizeof(derived)%unsafe.Sizeof(base) != 0 {
		panic("bad")
	}
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	b := c.AllocByteSlice(n * int(scale))
	derivedCap := cap(b) / int(scale)
	data := (*bool)(unsafe.Pointer(unsafe.SliceData(b)))
	return unsafe.Slice(data, derivedCap)[:n]
}
func (c *Cache) FreeBoolSlice(s []bool) {
	if cap(s) == 0 {
		return
	}
	var base byte
	var derived bool
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	byteCap := cap(s) * int(scale)
	byteCap = roundUpPowerOfTwo(byteCap)
	data := (*byte)(unsafe.Pointer(unsafe.SliceData(s)))
	b := unsafe.Slice(data, byteCap)
	c.FreeByteSlice(b)
}
func (c *Cache) AllocIDSlice(n int) []ID {
	var base byte
	var derived ID
	if unsafe.Sizeof(derived)%unsafe.Sizeof(base) != 0 {
		panic("bad")
	}
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	b := c.AllocByteSlice(n * int(scale))
	derivedCap := cap(b) / int(scale)
	data := (*ID)(unsafe.Pointer(unsafe.SliceData(b)))
	return unsafe.Slice(data, derivedCap)[:n]
}
func (c *Cache) FreeIDSlice(s []ID) {
	if cap(s) == 0 {
		return
	}
	var base byte
	var derived ID
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	byteCap := cap(s) * int(scale)
	byteCap = roundUpPowerOfTwo(byteCap)
	data := (*byte)(unsafe.Pointer(unsafe.SliceData(s)))
	b := unsafe.Slice(data, byteCap)
	c.FreeByteSlice(b)
}
func (c *Cache) AllocUintSlice(n int) []uint {
	var base byte
	var derived uint
	if unsafe.Sizeof(derived)%unsafe.Sizeof(base) != 0 {
		panic("bad")
	}
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	b := c.AllocByteSlice(n * int(scale))
	derivedCap := cap(b) / int(scale)
	data := (*uint)(unsafe.Pointer(unsafe.SliceData(b)))
	return unsafe.Slice(data, derivedCap)[:n]
}
func (c *Cache) FreeUintSlice(s []uint) {
	if cap(s) == 0 {
		return
	}
	var base byte
	var derived uint
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	byteCap := cap(s) * int(scale)
	byteCap = roundUpPowerOfTwo(byteCap)
	data := (*byte)(unsafe.Pointer(unsafe.SliceData(s)))
	b := unsafe.Slice(data, byteCap)
	c.FreeByteSlice(b)
}
func (c *Cache) AllocKnownBitsEntriesSlice(n int) []knownBitsEntry {
	var base byte
	var derived knownBitsEntry
	if unsafe.Sizeof(derived)%unsafe.Sizeof(base) != 0 {
		panic("bad")
	}
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	b := c.AllocByteSlice(n * int(scale))
	derivedCap := cap(b) / int(scale)
	data := (*knownBitsEntry)(unsafe.Pointer(unsafe.SliceData(b)))
	return unsafe.Slice(data, derivedCap)[:n]
}
func (c *Cache) FreeKnownBitsEntriesSlice(s []knownBitsEntry) {
	if cap(s) == 0 {
		return
	}
	var base byte
	var derived knownBitsEntry
	scale := unsafe.Sizeof(derived) / unsafe.Sizeof(base)
	byteCap := cap(s) * int(scale)
	byteCap = roundUpPowerOfTwo(byteCap)
	data := (*byte)(unsafe.Pointer(unsafe.SliceData(s)))
	b := unsafe.Slice(data, byteCap)
	c.FreeByteSlice(b)
}

var byteSlicePoolElemTypes = []string{"Limit", "int", "int32", "int8", "bool", "ID", "uint", "knownBitsEntry"}
