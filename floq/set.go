package main

//  Note: no mutex arounf functions add_*()!!

import (
	"errors"
)

type element struct {
	bool
	is_bool		bool

	uint64
	is_uint64	bool

	string
	is_string	bool

	set
	is_set		bool

	array		[]element
	is_array	bool
}

/*
 *  Elements of a set may be bool, uint63, string, set and arrays of
 *  elements.  Naming an element creates another element that is distinct
 *  from the bare element.
 */
type set struct {
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
		return errors.New("add_bool: exists")
	}
	s.bare_bool[element] = true
	return nil
}

func (s *set) add_uint64(element uint64) error {

	_, exists := s.bare_uint64[element]
	if exists {
		return errors.New("add_uint64: exists")
	}
	s.bare_uint64[element] = true
	return nil
}

func (s *set) add_string(element string) error {

	if element == "" {
		return errors.New("add_string: can not add empty string")
	}
	_, exists := s.bare_string[element]
	if exists {
		return errors.New("add_string: exists")
	}
	s.bare_string[element] = true
	return nil
}
