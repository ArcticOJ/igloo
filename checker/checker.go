package checker

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

func Check(output string, expected io.ReadCloser) (bool, string) {
	outFile, e := os.Open(output)
	if e != nil {
		return false, "failed to open output file"
	}
	defer outFile.Close()
	defer expected.Close()
	cntOfMatches, cntOfOut := 0, 0
	r1 := bufio.NewReader(outFile)
	r2 := bufio.NewReader(expected)
	for {
		outEof, expEof := false, false
		_l1, p1, e := r1.ReadLine()
		if e != nil {
			if e == io.EOF {
				outEof = true
			} else {
				return false, "error whilst reading output file"
			}
		}
		for p1 && !outEof {
			_l, _p, _e := r1.ReadLine()
			if _e != nil {
				return false, "error whilst reading output file"
			}
			p1 = _p
			_l1 = append(_l1, _l...)
		}
		_l2, p2, e := r2.ReadLine()
		if e != nil {
			if e == io.EOF {
				expEof = true
			} else {
				return false, "error whilst reading expected output file"
			}
		}
		for p2 && !expEof {
			_l, _p, _e := r2.ReadLine()
			if _e != nil {
				return false, "error whilst reading expected output file"
			}
			p2 = _p
			_l2 = append(_l2, _l...)
		}
		if (_l1 == nil && _l2 == nil) || (outEof == true && expEof == true) {
			break
		}
		if (_l1 == nil) != (_l2 == nil) {
			return false, "invalid output"
		}
		cntOfOut++
		l1 := bytes.TrimRight(_l1, " \r\n")
		l2 := bytes.TrimRight(_l2, " \r\n")
		if bytes.Equal(l1, l2) {
			cntOfMatches++
		}
	}
	return cntOfMatches == cntOfOut, fmt.Sprintf("matched %d out of %d lines", cntOfMatches, cntOfOut)
}
