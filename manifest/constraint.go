package manifest

import (
	"fmt"
	"math/big"
	"sort"
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

// Extract constraint fields by namespaces
func (c Constraint) ExtractFields() (res map[string][]string) {
	res = map[string][]string{}
	collected := map[string]struct{}{}
	for k, v := range c {
		for _, f := range ExtractEnv(k + v) {
			collected[f] = struct{}{}
		}
	}
	for k := range collected {
		split := strings.SplitN(k, ".", 2)
		if len(split) == 2 {
			res[split[0]] = append(res[split[0]], split[1])
		}
	}
	for _, v := range res {
		sort.Strings(v)
	}
	return
}

// Merge returns constraint merged with given fields
func (c Constraint) Merge(constraint ...Constraint) (res Constraint) {
	res = Constraint{}
	for _, cons := range append(constraint, c) {
		for k, v := range cons {
			res[k] = v
		}
	}
	return
}

// Ignore returns constraint without pairs which contains variables with given names
func (c Constraint) Ignore(name ...string) (res Constraint) {
	res = Constraint{}
	var findRes []string
LOOP:
	for k, v := range c {
		findRes = envRe.FindAllString(k+" "+v, -1)
		for _, chunk := range findRes {
			chunk = chunk[2 : len(chunk)-1]
			for _, candidate := range name {
				if candidate == chunk {
					continue LOOP
				}
			}
			res[k] = v
		}
	}
	return
}

func (c Constraint) Check(env map[string]string) (err error) {
	for left, right := range c {
		leftV := Interpolate(left, env)
		rightV := Interpolate(right, env)
		if !check(leftV, rightV) {
			err = fmt.Errorf("constraint failed %s != %s (%s:%s)", leftV, rightV, left, right)
			return
		}
	}
	return
}

func check(left, right string) (res bool) {
	// try to get op
	split := strings.SplitN(right, " ", 2)
	if len(split) != 2 {
		// just compare and return
		res = left == right
		return
	}
	op := split[0]
	switch op {
	case opEqual:
		res = left == split[1]
	case opNotEqual:
		res = left != split[1]
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
			res = cmpRes == -1
		case opLessOrEqual:
			res = cmpRes <= 0
		case opGreater:
			res = cmpRes == 1
		case opGreaterOrEqual:
			res = cmpRes >= 0
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
			res = found == len(leftSplit)
		case opNotIn:
			res = found == 0
		}
	default:
		// ordinary string
		res = left == right
	}
	return
}
