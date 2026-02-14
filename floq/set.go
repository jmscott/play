package main

//  Note: no mutex around functions add_*()!!

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc64"
	"slices"
	"sort"
	"strconv"
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
	bare_set		map[uint64]bool

	name_bool		map[string](map[bool]bool)
}

func new_set() *set {
	return &set{
			bare_bool:	make(map[bool]bool),
			bare_uint64:	make(map[uint64]bool),
			bare_string:	make(map[string]bool),
			bare_set:	make(map[uint64]bool),
	}
}

func (s *set) add_bare_bool(ele bool) error {

	_, exists := s.bare_bool[ele]
	if exists {
		return fmt.Errorf("bool element (%t) already exists", ele)
	}
	s.bare_bool[ele] = true
	return nil
}

func (s *set) add_bare_uint64(ele uint64) error {

	_, exists := s.bare_uint64[ele]
	if exists {
		return fmt.Errorf("uint64 (%d) element already exists", ele)
	}
	s.bare_uint64[ele] = true
	return nil
}

func (s *set) add_bare_string(ele string) error {

	if ele == "" {
		return errors.New("can not add empty string")
	}
	_, exists := s.bare_string[ele]
	if exists {
		return fmt.Errorf(
				"string element \"%s\" already exists",
				string_brief(ele, 5, true),
		)
	}
	s.bare_string[ele] = true
	return nil
}

func (s *set) count() uint64 {

	return uint64(
		len(s.bare_bool) +
		len(s.bare_uint64) +
		len(s.bare_string) +
		len(s.bare_set))
}

func (s *set) add_bare_set(ele *set) error {

	_, exists := s.bare_set[ele.crc64()]
	if exists {
		return errors.New("element already in set")
	}
	s.bare_set[ele.crc64()] = true
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

	//  add untagged uint64 to sha256
	//  build array, sort, then hashsum 8 bytes per uint64
	//  in array order

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

	//  hash the untagged strings.
	//  build array of strings, sort, then hashsum bytes of each string in
	//  array order.

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

	//  hash crc64s of bare_set elements

	crc := make([]uint64, len(s.bare_set))
	i = 0
	for k, _ := range s.bare_set {
		crc[i] = k
		i++
	}
	slices.Sort(crc)
	for _, v := range crc {
		binary.BigEndian.PutUint64(buf[:], uint64(v))
		h.Write(buf[:8])
	}

	return h.Sum(nil)
}

//  write crc64 as string with trailing ellipse if truncated

func (s *set) crc64_brief(clen int, ellipse bool) string {

	return string_brief(strconv.FormatUint(s.crc64(), 10), clen, ellipse)
}

func (s *set) crc64() uint64 {

	tab := crc64.MakeTable(crc64.ECMA)
	h := crc64.New(tab)

	h.Write(s.sha256())

	return h.Sum64()
}

func (s *set) String() string {
	
	var str string

	str = strconv.FormatUint(s.crc64(), 10)

	l := len(s.bare_bool)
	if l > 0 {
		str += " bb=" + strconv.Itoa(l)
	}

	l = len(s.bare_uint64)
	if l > 0 {
		str += " bui=" + strconv.Itoa(l)
	}

	l = len(s.bare_string)
	if l > 0 {
		str += " bs=" + strconv.Itoa(l)
	}

	l = len(s.bare_set)
	if l > 0 {
		str += " bset=" + strconv.Itoa(l)
	}

	return str
}
