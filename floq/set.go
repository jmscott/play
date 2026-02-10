package main

//  Note: no mutex arounf functions add_*()!!

import (
	"errors"
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
	 *		"good bye, cruel world"
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

	//  find a bool in set1 not in set2
	for k1, _ := range s1.bare_bool {
		_, exists := s2.bare_bool[k1]
		if exists == false {
			return false
		}
	}

	//  find a bool in set2 not in set1
	for k2, _ := range s2.bare_bool {
		_, exists := s1.bare_bool[k2]
		if exists == false {
			return false
		}
	}

	//  find a uint64 in set1 not in set2
	for k1, _ := range s1.bare_uint64 {
		_, exists := s2.bare_uint64[k1]
		if exists == false {
			return false
		}
	}

	//  find a uint64 in set2 not in set1
	for k2, _ := range s2.bare_uint64 {
		_, exists := s1.bare_uint64[k2]
		if exists == false {
			return false
		}
	}

	//  find a string in set1 not in set2
	for k1, _ := range s1.bare_string {
		_, exists := s2.bare_string[k1]
		if exists == false {
			return false
		}
	}

	//  find a string in set2 not in set1
	for k2, _ := range s2.bare_string {
		_, exists := s1.bare_string[k2]
		if exists == false {
			return false
		}
	}

	return true
}
