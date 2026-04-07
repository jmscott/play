package main

//  Note: no mutex around functions add_*()!!

import (
	"hash"
	"sort"
	"slices"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash/crc64"
	"strconv"
)

/*
 *  Elements of a set may be named bool, uint63, string, set and arrays of
 *  elements.  No two elements have the same name.  Two elements with different
 *  name but same value are distinct.
 *
 *	{
 *		a:"a",
 *		b:"a"
 *	}
 *
 *  the set contains two elements.
 */
type set struct {

	/*
	 *  Map values into their existence
	 */
	bare_bool		map[bool]bool
	bare_uint64		map[uint64]bool
	bare_string		map[string]bool
	bare_set		map[crc_64]bool
	bare_array		map[crc_64]bool

	/*
	 *  Map named elements onto their values
	 */
	name_bool		map[string]bool
	name_uint64		map[string]uint64
	name_string		map[string]string
	name_set		map[string]*set
	name_array		map[string][]string
}

//  crc of sets and arrays for equality op

type crc_64 uint64

func new_set() *set {
	return &set{
			bare_bool:	make(map[bool]bool),
			bare_uint64:	make(map[uint64]bool),
			bare_string:	make(map[string]bool),
			bare_set:	make(map[crc_64]bool),
			bare_array:	make(map[crc_64]bool),

			name_bool:	make(map[string]bool),
			name_uint64:	make(map[string]uint64),
			name_string:	make(map[string]string),
			name_set:	make(map[string]*set),
			name_array:	make(map[string][]string),
	}
}

//  add a named uint64 element to a set, erroring if element already exists

func (s *set) add_name_uint64(name string, ele uint64) error {

	if s.has_name(name) {
		return s.error("named element (%s) exists: %d", name, ele)
	}

	s.name_uint64[name] = ele
	return nil
}

//  add a uint64 element to a set, erroring if element already exists

func (s *set) add_bare_uint64(ele uint64) error {

	_, exists := s.bare_uint64[ele]
	if exists {
		return s.error("uint64: element exists: %d", ele)
	}
	s.bare_uint64[ele] = true
	return nil
}

//  add a uint64 element to a set, erroring if element already exists

func (s *set) add_uint64(name string, ele uint64) error {

	if name == "" {
		return s.add_bare_uint64(ele)
	}
	return s.add_name_uint64(name, ele)
}

//  add a named string element to a set, erroring if element already exists

func (s *set) add_name_string(name string, ele string) error {

	if s.has_name(name) {
		return fmt.Errorf("string: named element exists: %s", name)
	}

	s.name_string[name] = ele
	return nil
}

func (s *set) add_string(name, ele string) error {
	
	if name == "" {
		return s.add_bare_string(ele)
	}
	return s.add_name_string(name, ele)
}

//  make an error associated with a particlar set

func (s *set) error(format string, args ...interface{}) error {

	return fmt.Errorf("set: " + format, args...)  
}

//  add a string element to a set, error if element already exists

func (s *set) add_bare_string(ele string) error {

	_, exists := s.bare_string[ele]
	if exists {
		return s.error("bare string element exists: %s", ele)
	}
	s.bare_string[ele] = true
	return nil
}

//  add a named set element to a set, error if element already exists

func (s *set) add_name_set(name string, ele *set) error {

	if s.has_name(name) {
		return fmt.Errorf("set: named element exists: %s", name)
	}

	s.name_set[name] = ele
	return nil
}

//  add a named bool element to a set, error if bool already exists

func (s *set) add_name_bool(name string, ele bool) error {

	if s.has_name(name) {
		return fmt.Errorf("bool: named element exists: %s", name)
	}

	s.name_bool[name] = ele
	return nil
}

//  add a bool element to a set, error if bool already exists

func (s *set) add_bare_bool(ele bool) error {

	_, exists := s.bare_bool[ele]
	if exists {
		return s.error("named element exists: %t", ele)
	}
	s.bare_bool[ele] = true
	return nil
}

//  add a named bool element to a set, error if bool already exists

func (s *set) add_bool(name string, ele bool) error {

	if name == "" {
		return s.add_bare_bool(ele)
	}
	return s.add_name_bool(name, ele)
}

//  added a named array []string element to a set, error if array already exists

func (s *set) add_name_array(name string, ele []string) error {
	if s.has_name(name) {
		return fmt.Errorf("array: named element exists: %s", name)
	}
	s.name_array[name] = ele
	return nil
}

//  add an array []string element to a set, error if array already exists

func (s *set) add_bare_array(ele []string) error {
	crc := s.crc64_array(ele)
	_, exists := s.bare_array[crc]
	if exists {
		return s.error("array element exists: %s", ele)
	}
	s.bare_array[crc] = true
	return nil
}

//  add a named or bare array[]string element to a set, error if already exists.

func (s *set) add_array(name string, ele []string) error {

	if name == "" {
		return s.add_bare_array(ele)
	}
	return s.add_name_array(name, ele)
}

//  crc64 checksum of an array []string used inequality

func (s *set) crc64_array(strings []string) crc_64 {
	sha := sha256.New()

        buf := make([]byte, 8)
	for i, str := range strings {
		binary.BigEndian.PutUint64(buf[:], uint64(i))
		sha.Write(buf)
		sha.Write([]byte(str))
	}
	return crc_64(
		crc64.Checksum(
			sha.Sum(nil),
			crc64.MakeTable(crc64.ECMA),
		),
	)
}

//  count of elements in set.  no recursive descent.

func (s *set) count() uint64 {

	return uint64(
		len(s.bare_bool) +
		len(s.bare_uint64) +
		len(s.bare_string) +
		len(s.bare_set) +
		len(s.bare_array) +

		len(s.name_bool) +
		len(s.name_uint64) +
		len(s.name_string) +
		len(s.name_set) +
		len(s.name_array))
}

//  is set s1 equal to set s2

func (s1 *set) equals(s2 *set) bool {

	return s1.crc64() == s2.crc64()
}

func (s *set) sum_bool(h hash.Hash) {

	//  write the bare bools, false, then true

	if s.bare_bool[false] == true {
		h.Write([]byte{0x0})
	}
	if s.bare_bool[true] == true {
		h.Write([]byte{0x1})
	}

	//  write the named bools in lexical order of name

	type nv struct {
		Name   string
		Value bool
	}

	var ss []nv
	for n, v := range s.name_bool {
		ss = append(ss, nv{n, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Name < ss[j].Name
	})

	for i := 0;  i < len(ss);  i++ {
		var buf [8]byte

		binary.BigEndian.PutUint64(buf[:], uint64(i))
		h.Write(buf[:8])

		h.Write([]byte(ss[i].Name))
		if ss[i].Value == true {
			h.Write([]byte{0x1})
		} else {
			h.Write([]byte{0x0})
		}
	}
}

func (s *set) sum_uint64(h hash.Hash) {

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
	for i, v := range ui64 {
		binary.BigEndian.PutUint64(buf[:], uint64(i))
		h.Write(buf[:8])

		binary.BigEndian.PutUint64(buf[:], uint64(v))
		h.Write(buf[:8])
	}

	//  write the named uint64 in lexical order of name

	type nv struct {
		Name	string
		Value	uint64
	}

	var ss []nv
	for n, v := range s.name_uint64 {
		ss = append(ss, nv{n, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Name < ss[j].Name
	})

	for i := 0;  i < len(ss);  i++ {
		var buf [8]byte

		binary.BigEndian.PutUint64(buf[:], uint64(i))
		h.Write(buf[:8])
		
		h.Write([]byte(ss[i].Name))

		binary.BigEndian.PutUint64(buf[:], uint64(ss[i].Value))
		h.Write(buf[:8])
	}
}

func (s *set) sum_string(h hash.Hash) {

	//  hash the bare strings.
	//  build array of strings, sort, then hashsum bytes of each string in
	//  array order.

	strs := make([]string, len(s.bare_string))
	i := 0
	for k, _ := range s.bare_string {
		strs[i] = k
		i++
	}
	sort.Strings(strs)
	for i, v := range strs {
		var buf [8]byte

		binary.BigEndian.PutUint64(buf[:], uint64(i))
		h.Write(buf[:8])

		h.Write([]byte(v))
	}

	//  write the named strings in lexical order of name

	type nv struct {
		Name	string
		Value	string
	}

	var ss []nv
	for n, v := range s.name_string {
		ss = append(ss, nv{n, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Name < ss[j].Name
	})

	for i := 0;  i < len(ss);  i++ {
		var buf [8]byte

		binary.BigEndian.PutUint64(buf[:], uint64(i))
		h.Write(buf[:8])
		
		h.Write([]byte(ss[i].Name))
		h.Write([]byte(ss[i].Value))
	}
}

func (s *set) sum_set(h hash.Hash) {

	//  hash crc64s of bare_set elements

	crc := make([]crc_64, len(s.bare_set))
	i := 0
	for c64, _  := range s.bare_set {
		crc[i] = c64
		i++
	}

	//  sum crcs of elements in sorted order

	slices.Sort(crc)
	for i, v := range crc {
		var buf [8]byte

		binary.BigEndian.PutUint64(buf[:], uint64(i))
		h.Write(buf[:])

		binary.BigEndian.PutUint64(buf[:], uint64(v))
		h.Write(buf[:])
	}

	//  sum crc64 of named sets

	type nv struct {
		Name	string
		Value	crc_64
	}

	var ss []nv
	for n, v := range s.name_set {
		ss = append(ss, nv{n, v.crc64()})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Name < ss[j].Name
	})

	for i := 0;  i < len(ss);  i++ {
		var buf [8]byte

		binary.BigEndian.PutUint64(buf[:], uint64(i))
		h.Write(buf[:])
		
		h.Write([]byte(ss[i].Name))
		binary.BigEndian.PutUint64(buf[:], uint64(ss[i].Value))
		h.Write(buf[:])
	}
}

//  calculate sha256 of set

func (s *set) sha256() []byte {

	h := sha256.New()

	//  sum bare and named elements of sets
	s.sum_bool(h)
	s.sum_uint64(h)
	s.sum_string(h)
	s.sum_set(h)

	return h.Sum(nil)
}

//  write crc64 as string with trailing ellipse if truncated

func (s *set) crc64_brief(clen int, ellipse bool) string {

	return string_brief(
		strconv.FormatUint(uint64(s.crc64()), 10), clen, ellipse)
}

func (s *set) crc64() crc_64 {

	tab := crc64.MakeTable(crc64.ECMA)
	h := crc64.New(tab)

	h.Write(s.sha256())

	return crc_64(h.Sum64())
}

func (s *set) String() string {
	
	var str string

	str = strconv.FormatUint(uint64(s.crc64()), 10)

	l := len(s.name_bool)
	if l > 0 {
		str += " b=" + strconv.Itoa(l)
	}

	l = len(s.name_uint64)
	if l > 0 {
		str += " ui=" + strconv.Itoa(l)
	}

	l = len(s.name_string)
	if l > 0 {
		str += " str=" + strconv.Itoa(l)
	}

	l = len(s.name_set)
	if l > 0 {
		str += " set=" + strconv.Itoa(l)
	}

	l = len(s.name_array)
	if l > 0 {
		str += " arr=" + strconv.Itoa(l)
	}

	return str
}

func (s *set) has_name(name string) bool {
	
	if _, ok := s.name_bool[name];  ok == true {
		return true
	}
	if _, ok := s.name_uint64[name];  ok == true {
		return true
	}
	if _, ok := s.name_string[name];  ok == true {
		return true
	}
	if _, ok := s.name_set[name];  ok == true {
		return true
	}
	if _, ok := s.name_array[name];  ok == true {
		return true
	}
	return false
}
