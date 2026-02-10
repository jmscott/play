package main

//  Note: no mutex arounf functions add_*()!!

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"hash/crc64"
	"slices"
	"sort"
)


/*
 *  Elements of a set may be bool, uint63, string, set and arrays of
 *  elements.  Naming an element creates another element that is distinct
 *  from the bare element.
 */
type set struct {

	/*
	 *  Bare elements like:
	 *
	 *	{
	 *		true, false,
	 *		123, 456,
	 *		"hello, world",
	 *		"good bye, cruel world",
	 *		{ 0, 1, 2}
	 *	}
	 */
	bare_bool		map[bool]bool
	bare_uint64		map[uint64]bool
	bare_string		map[string]bool
}

func new_set() *set {
	return &set{
			bare_bool:	make(map[bool]bool),
			bare_uint64:	make(map[uint64]bool),
			bare_string:	make(map[string]bool),
	}
}

func (s *set) add_bool(element bool) error {

	_, exists := s.bare_bool[element]
	if exists {
		return errors.New("element exists")
	}
	s.bare_bool[element] = true
	return nil
}

func (s *set) add_uint64(element uint64) error {

	_, exists := s.bare_uint64[element]
	if exists {
		return errors.New("element exists")
	}
	s.bare_uint64[element] = true
	return nil
}

func (s *set) add_string(element string) error {

	if element == "" {
		return errors.New("can not add empty string")
	}
	_, exists := s.bare_string[element]
	if exists {
		return errors.New("element exists")
	}
	s.bare_string[element] = true
	return nil
}

func (s1 *set) equals(s2 *set) bool {

	return s1.crc64() == s2.crc64()
}

func (s *set) sha256() []byte {

	h := sha256.New()

	//  add bools to sha245

	if s.bare_bool[false] == true {
		h.Write([]byte{0x0})
	}
	if s.bare_bool[true] == true {
		h.Write([]byte{0x1})
	}

	//  hash the untagged uint64s
	ui64 := make([]uint64, len(s.bare_uint64))
	i := 0
	for k, _ := range s.bare_uint64 {
		ui64[i] = k
		i++
	}
	slices.Sort(ui64)
	buf := make([]byte, 8)
	for _, v := range ui64 {
		binary.BigEndian.PutUint64(buf[:], uint64(v))
		h.Write(buf[:8])
	}

	//  hash the untagged strings
	strs := make([]string, len(s.bare_string))
	i = 0
	for k, _ := range s.bare_string {
		strs[i] = k
		i++
	}
	sort.Strings(strs)
	for _, v := range strs {
		h.Write([]byte(v))
	}
	return h.Sum(nil)
}

func (s *set) crc64() uint64 {

	tab := crc64.MakeTable(crc64.ECMA)
	h := crc64.New(tab)

	h.Write(s.sha256())

	return h.Sum64()
}
