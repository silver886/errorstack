package errorstack

import (
	"errors"
	"io"
)

// WalkFunc is the type of the function called for each error visited by Walk.
type WalkFunc func(level int, err error) error

// ErrSkip is used as a return value from WalkFuncs to indicate that
// the error in the call is to be skipped. It is not returned as an
// error by any function.
var ErrSkip = errors.New("skip this error")

// errWalkFunc determines the error from WalkFuncs.
func errWalkFunc(err error) error {
	if err == ErrSkip {
		return nil
	}
	return err
}

// walk walks through the error stack.
// If direction == true, it will start from root.
// If direction == false, it will start from last.
func (s *Stack) walk(direction bool, walkFn WalkFunc) error {
	if direction {
		for i, e := range s.errs {
			if err := walkFn(i, e); err != nil {
				return errWalkFunc(err)
			}
		}
	} else {
		for i := len(s.errs) - 1; i >= 0; i-- {
			if err := walkFn(i, s.errs[i]); err != nil {
				return errWalkFunc(err)
			}
		}
	}
	return nil
}

// Walk walks the error stack from root.
func (s *Stack) Walk(walkFn WalkFunc) error {
	return s.walk(true, walkFn)
}

// printFunc is the type of the function called for each error visited by walkPrint.
type printFunc func(writer io.Writer, verb rune, err error)

// walkPrint walks the error stack from last and write them to writer with verb and separator.
func (s *Stack) walkPrint(writer io.Writer, verb rune, separator string, print printFunc) {
	maxLevel := len(s.errs) - 1
	s.walk(false, func(level int, err error) error {
		if level != maxLevel {
			io.WriteString(writer, separator)
		}
		print(writer, verb, err)
		return nil
	})
}
