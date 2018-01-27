package manifest

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

const (
	opEqual          = "="
	opNotEqual       = "!="
	opLess           = "<"
	opLessOrEqual    = "<="
	opGreater        = ">"
	opGreaterOrEqual = ">="
	opIn             = "~"
	opNotIn          = "!~"
)

// Constraint can contain interpolations in form ${ns.key}.
// Right field can also begins with compare operation: "<", ">" or "~" (in).
type Constraint map[string]string

// Returns clone of constraint
func (c Constraint) Clone() (res Constraint) {
	res = Constraint{}
	for left, right := range c {
		res[left] = right
	}
	return res
}

// Merge returns constraint merged with given constraints
func (c Constraint) Merge(constraint ...Constraint) (res Constraint) {
	res = Constraint{}
	for _, cons := range append(constraint, c) {
		for left, right := range cons {
			res[left] = right
		}
	}
	return res
}

// FilterOut returns Constraint without pairs which contains references with given prefixes
func (c Constraint) FilterOut(prefix ...string) (res Constraint) {
	res = Constraint{}
	var fields []string
LOOP:
	for left, right := range c {
		fields = ExtractEnv(left + right)
		for _, p := range prefix {
			for _, field := range fields {
				if strings.HasPrefix(field, p) {
					continue LOOP
				}
			}
		}
		res[left] = right
	}
	return res
}

// Check constraint against given environment
func (c Constraint) Check(env map[string]string) (err error) {
	for left, right := range c {
		leftV := Interpolate(left, env)
		rightV := Interpolate(right, env)
		if !check(leftV, rightV) {
			return fmt.Errorf(`constraint failed: "%s":"%s" ("%s":"%s")`, leftV, rightV, left, right)
		}
	}
	return nil
}

func check(left, right string) (res bool) {
	// try to get op
	split := strings.SplitN(right, " ", 2)
	if len(split) != 2 {
		// bad split - just compare
		return left == right
	}
	op := split[0]
	switch op {
	case opEqual:
		return left == split[1]
	case opNotEqual:
		return left != split[1]
	case opLess, opLessOrEqual, opGreater, opGreaterOrEqual:
		right = split[1]
		var cmpRes int
		leftN, leftErr := strconv.ParseFloat(left, 64)
		rightN, rightErr := strconv.ParseFloat(right, 64)
		if leftErr == nil && rightErr == nil {
			// ok, we have numbers
			cmpRes = big.NewFloat(leftN).Cmp(big.NewFloat(rightN))
		} else {
			cmpRes = strings.Compare(left, right)
		}
		switch op {
		case opLess:
			return cmpRes == -1
		case opLessOrEqual:
			return cmpRes <= 0
		case opGreater:
			return cmpRes == 1
		case opGreaterOrEqual:
			return cmpRes >= 0
		}
	case opIn, opNotIn:
		leftSplit := strings.Split(left, ",")
		rightSplit := strings.Split(split[1], ",")
		var found int
	LOOP:
		for _, rightChunk := range rightSplit {
			for _, leftChunk := range leftSplit {
				if strings.TrimSpace(leftChunk) == strings.TrimSpace(rightChunk) {
					found++
					continue LOOP
				}
			}
		}
		switch op {
		case opIn:
			return found == len(leftSplit)
		case opNotIn:
			return found == 0
		}
	default:
		// ordinary string
		return left == right
	}
	return false
}
