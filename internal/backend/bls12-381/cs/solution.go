// Copyright 2020 ConsenSys Software Inc.
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

// Code generated by gnark DO NOT EDIT

package cs

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync"

	"github.com/consensys/gnark/backend/hint"
	"github.com/consensys/gnark/internal/backend/compiled"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"

	curve "github.com/consensys/gnark-crypto/ecc/bls12-381"
)

// ErrUnsatisfiedConstraint can be generated when solving a R1CS
var ErrUnsatisfiedConstraint = errors.New("constraint is not satisfied")

// solution represents elements needed to compute
// a solution to a R1CS or SparseR1CS
type solution struct {
	values, coefficients []fr.Element
	solved               []bool
	nbSolved             int
	mHintsFunctions      map[hint.ID]hint.AnnotatedFunction
}

func newSolution(nbWires int, hintFunctions []hint.AnnotatedFunction, coefficients []fr.Element) (solution, error) {
	s := solution{
		values:          make([]fr.Element, nbWires),
		coefficients:    coefficients,
		solved:          make([]bool, nbWires),
		mHintsFunctions: make(map[hint.ID]hint.AnnotatedFunction, len(hintFunctions)),
	}

	for _, h := range hintFunctions {
		if _, ok := s.mHintsFunctions[h.UUID()]; ok {
			return solution{}, fmt.Errorf("duplicate hint function %s", h)
		}
		s.mHintsFunctions[h.UUID()] = h
	}

	return s, nil
}

func (s *solution) set(id int, value fr.Element) {
	if s.solved[id] {
		panic("solving the same wire twice should never happen.")
	}
	s.values[id] = value
	s.solved[id] = true
	s.nbSolved++
}

func (s *solution) isValid() bool {
	return s.nbSolved == len(s.values)
}

// computeTerm computes coef*variable
func (s *solution) computeTerm(t compiled.Term) fr.Element {
	cID, vID, _ := t.Unpack()
	if cID != 0 && !s.solved[vID] {
		panic("computing a term with an unsolved wire")
	}
	switch cID {
	case compiled.CoeffIdZero:
		return fr.Element{}
	case compiled.CoeffIdOne:
		return s.values[vID]
	case compiled.CoeffIdTwo:
		var res fr.Element
		res.Double(&s.values[vID])
		return res
	case compiled.CoeffIdMinusOne:
		var res fr.Element
		res.Neg(&s.values[vID])
		return res
	default:
		var res fr.Element
		res.Mul(&s.coefficients[cID], &s.values[vID])
		return res
	}
}

// solveHint compute solution.values[vID] using provided solver hint
func (s *solution) solveWithHint(vID int, h *compiled.Hint) error {
	// skip if the wire is already solved by a call to the same hint
	// function on the same inputs
	if s.solved[vID] {
		return nil
	}
	// ensure hint function was provided
	f, ok := s.mHintsFunctions[h.ID]
	if !ok {
		return errors.New("missing hint function")
	}

	// compute values for all inputs.
	inputs := make([]*big.Int, len(h.Inputs))
	for i := 0; i < len(inputs); i++ {
		inputs[i] = bigIntPool.Get().(*big.Int)
		inputs[i].SetUint64(0)
	}
	lambda := bigIntPool.Get().(*big.Int)

	for i := 0; i < len(h.Inputs); i++ {
		// input is a linear expression, we must compute the value
		for j := 0; j < len(h.Inputs[i]); j++ {
			ciID, viID, visibility := h.Inputs[i][j].Unpack()
			if visibility == compiled.Virtual {
				// we have a constant, just take the coefficient value
				s.coefficients[ciID].ToBigIntRegular(lambda)
				inputs[i].Add(inputs[i], lambda)
				continue
			}
			if !s.solved[viID] {
				// release objects into pool
				bigIntPool.Put(lambda)
				for i := 0; i < len(inputs); i++ {
					bigIntPool.Put(inputs[i])
				}
				return errors.New("expected wire to be instantiated while evaluating hint")
			}
			v := s.computeTerm(h.Inputs[i][j])
			v.ToBigIntRegular(lambda)
			inputs[i].Add(inputs[i], lambda)
		}
	}

	outputs := make([]*big.Int, f.TotalOutputs(len(inputs)))
	// use lambda as the result.
	outputs[0] = lambda
	for i := 1; i < len(outputs); i++ {
		outputs[i] = bigIntPool.Get().(*big.Int)
	}

	// ensure our inputs are mod q
	q := fr.Modulus()
	for i := 0; i < len(inputs); i++ {
		// note since we're only doing additions up there, we may want to avoid the use of Mod
		// here in favor of Cmp & Sub
		inputs[i].Mod(inputs[i], q)
	}

	err := f.Call(curve.ID, inputs, outputs)

	var v fr.Element
	for i := range outputs {
		v.SetBigInt(outputs[i])
		s.set(h.Wires[i], v)
	}

	// release objects into pool
	for i := 0; i < len(inputs); i++ {
		bigIntPool.Put(inputs[i])
	}

	for i := 0; i < len(outputs); i++ {
		bigIntPool.Put(outputs[i])
	}

	if err != nil {
		return err
	}

	return nil
}

func (s *solution) printLogs(w io.Writer, logs []compiled.LogEntry) {
	if w == nil {
		return
	}

	for i := 0; i < len(logs); i++ {
		logLine := s.logValue(logs[i])
		_, _ = io.WriteString(w, logLine)
	}
}

const unsolvedVariable = "<unsolved>"

func (s *solution) logValue(log compiled.LogEntry) string {
	var toResolve []interface{}
	var (
		isEval       bool
		eval         fr.Element
		missingValue bool
	)
	for j := 0; j < len(log.ToResolve); j++ {
		if log.ToResolve[j] == compiled.TermDelimitor {
			// this is a special case where we want to evaluate the following terms until the next delimitor.
			if !isEval {
				isEval = true
				missingValue = false
				eval.SetZero()
				continue
			}
			isEval = false
			if missingValue {
				toResolve = append(toResolve, unsolvedVariable)
			} else {
				// we have to append our accumulator
				toResolve = append(toResolve, eval.String())
			}
			continue
		}
		cID, vID, visibility := log.ToResolve[j].Unpack()

		if isEval {
			// we are evaluating
			if visibility == compiled.Virtual {
				// just add the constant
				eval.Add(&eval, &s.coefficients[cID])
				continue
			}
			if !s.solved[vID] {
				missingValue = true
				continue
			}
			tv := s.computeTerm(log.ToResolve[j])
			eval.Add(&eval, &tv)
			continue
		}

		if visibility == compiled.Virtual {
			// it's just a constant
			if cID == compiled.CoeffIdMinusOne {
				toResolve = append(toResolve, "-1")
			} else {
				toResolve = append(toResolve, s.coefficients[cID].String())
			}
			continue
		}
		if !(cID == compiled.CoeffIdMinusOne || cID == compiled.CoeffIdOne) {
			toResolve = append(toResolve, s.coefficients[cID].String())
		}
		if !s.solved[vID] {
			toResolve = append(toResolve, unsolvedVariable)
		} else {
			toResolve = append(toResolve, s.values[vID].String())
		}
	}
	return fmt.Sprintf(log.Format, toResolve...)
}

var bigIntPool = sync.Pool{
	New: func() interface{} {
		return new(big.Int)
	},
}
