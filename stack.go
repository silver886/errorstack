package errorstack

import (
	"errors"
	"reflect"
)

// Copy creates a copy of error stack.
func (s *Stack) Copy() *Stack {
	return &Stack{
		s.errs,
	}
}

// Attach pushes error stack into the stack.
func (s *Stack) Attach(errs ...error) *Stack {
	for _, err := range errs {
		s.errs = append(s.errs, Convert(err).errs...)
	}
	return s
}

// Push pushes errors into the stack.
func (s *Stack) Push(errs ...error) *Stack {
	s.errs = append(s.errs, errs...)
	return s
}

// pop pops an error off the stack.
func (s *Stack) pop() error {
	var err error
	if len(s.errs) != 0 {
		err = s.errs[len(s.errs)-1]
		s.errs = s.errs[:len(s.errs)-1]
	}
	return err
}

// Pop pops errors off the stack, errs start from last.
func (s *Stack) Pop(errs []error) *Stack {
	var tmpErrs []error
	for range errs {
		tmpErrs = append(tmpErrs, s.pop())
	}
	copy(errs, tmpErrs)
	return s
}

// Level gets the total level of the error stack has.
// If the stack is empty, will return 0.
func (s *Stack) Level() int {
	return len(s.errs)
}

// Empty reports whether the error stack is empty.
func (s *Stack) Empty() bool {
	if len(s.errs) == 0 {
		return true
	}
	return false
}

// Get gets the error of given level.
func (s *Stack) Get(level int) error {
	if level > 0 && len(s.errs) != 0 && level < len(s.errs)+1 {
		return s.errs[level-1]
	}
	return nil
}

// Root gets the root error.
func (s *Stack) Root() error {
	if len(s.errs) != 0 {
		return s.errs[0]
	}
	return nil
}

// Last gets the last error.
func (s *Stack) Last() error {
	if len(s.errs) != 0 {
		return s.errs[len(s.errs)-1]
	}
	return nil
}

// Find gets the level of given error appears within the first max times.
// If max < 0, there is no limit on the number of times.
// If not found, will return empty array.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
func (s *Stack) Find(err error, max int) []int {
	if err != nil && max != 0 {
		var levels []int
		s.walk(true, func(l int, e error) error {
			if len(levels) == max {
				return ErrSkip
			}
			for {
				if isComparable := reflect.TypeOf(e).Comparable(); isComparable {
					if e == err {
						levels = append(levels, l)
						return nil
					}
				} else if eIs, ok := e.(interface{ Is(error) bool }); ok {
					if eIs.Is(err) {
						levels = append(levels, l)
						return nil
					}
				}

				if eWrap, ok := e.(interface{ Unwrap() error }); ok {
					if eWrap := eWrap.Unwrap(); eWrap != nil {
						e = eWrap
					} else {
						return nil
					}
				} else {
					return nil
				}
			}
		})
		return levels
	}
	return nil
}

// First gets the first level of given error appears.
// If not found, will return 0.
func (s *Stack) First(err error) int {
	if levels := s.Find(err, 1); len(levels) > 0 {
		return levels[0]
	}
	return 0
}

// Has reports whether the error stack contains given error.
func (s *Stack) Has(err error) bool {
	if l := s.First(err); l > 0 {
		return true
	}
	return false
}

// Is reports whether the error stack is given error.
func (s *Stack) Is(err error) bool {
	if s.Root() == err {
		return true
	}
	return false
}

var errNotMatch = errors.New("skip this error")

// FirstStack gets the first level of given error stack appears.
// If not found, will return 0.
func (s *Stack) FirstStack(err *Stack) int {
	level, slevel, errlevel := 0, len(s.errs), len(err.errs)
	s.walk(true, func(ls int, es error) error {
		if slevel < errlevel || err.walk(true, func(l int, e error) error {
			if s.errs[ls+l-2] == e {
				if l == 1 {
					level = ls
				}
				return nil
			}
			return errNotMatch
		}) == nil {
			return ErrSkip
		}
		level = 0
		slevel--
		return nil
	})
	return level
}

// HasStack reports whether the error stack contains given error stack.
func (s *Stack) HasStack(err *Stack) bool {
	if l := s.FirstStack(err); l > 0 {
		return true
	}
	return false
}

// IsStack reports whether the error stack is given error stack.
func (s *Stack) IsStack(err *Stack) bool {
	if err.walk(true, func(l int, e error) error {
		if s.errs[l-1] == e {
			return nil
		}
		return errNotMatch
	}) == nil {
		return true
	}
	return false
}
