// Copyright 2021 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compiled

import (
	"errors"
	"math/big"
	"strings"
)

// errNoValue triggered when trying to access a variable that was not allocated
var errNoValue = errors.New("can't determine API input value")

// Variable represent a linear expression of wires
type Variable []Term

// Clone returns a copy of the underlying slice
func (v Variable) Clone() Variable {
	res := make(Variable, len(v))
	copy(res, v)
	return res
}

// Len return the lenght of the Variable (implements Sort interface)
func (v Variable) Len() int {
	return len(v)
}

// Equals returns true if both SORTED expressions are the same
//
// pre conditions: l and o are sorted
func (v Variable) Equal(o Variable) bool {
	if len(v) != len(o) {
		return false
	}
	if (v == nil) != (o == nil) {
		return false
	}
	for i := 0; i < len(v); i++ {
		if v[i] != o[i] {
			return false
		}
	}
	return true
}

// Swap swaps terms in the Variable (implements Sort interface)
func (v Variable) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// Less returns true if variableID for term at i is less than variableID for term at j (implements Sort interface)
func (v Variable) Less(i, j int) bool {
	_, iID, iVis := v[i].Unpack()
	_, jID, jVis := v[j].Unpack()
	if iVis == jVis {
		return iID < jID
	}
	return iVis > jVis
}

func (v Variable) string(sbb *strings.Builder, coeffs []big.Int) {
	for i := 0; i < len(v); i++ {
		v[i].string(sbb, coeffs)
		if i+1 < len(v) {
			sbb.WriteString(" + ")
		}
	}
}

// assertIsSet panics if the variable is unset
// this may happen if inside a Define we have
// var a variable
// cs.Mul(a, 1)
// since a was not in the circuit struct it is not a secret variable
func (v Variable) AssertIsSet() {

	if len(v) == 0 {
		panic(errNoValue)
	}

}

// isConstant returns true if the variable is ONE_WIRE * coeff
func (v Variable) IsConstant() bool {
	if len(v) != 1 {
		return false
	}
	_, vID, visibility := v[0].Unpack()
	return vID == 0 && visibility == Public
}
