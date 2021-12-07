// Copyright 2020 ConsenSys AG
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

// Package backend implements Zero Knowledge Proof systems: it consumes circuit compiled with gnark/frontend.
package backend

import (
	"io"
	"os"

	"github.com/consensys/gnark/backend/hint"
)

// ID represent a unique ID for a proving scheme
type ID uint16

const (
	UNKNOWN ID = iota
	GROTH16
	PLONK
)

// Implemented return the list of proof systems implemented in gnark
func Implemented() []ID {
	return []ID{GROTH16, PLONK}
}

// String returns the string representation of a proof system
func (id ID) String() string {
	switch id {
	case GROTH16:
		return "groth16"
	case PLONK:
		return "plonk"
	default:
		return "unknown"
	}
}

// NewProverOption returns a default ProverOption with given options applied
func NewProverOption(opts ...func(opt *ProverOption) error) (ProverOption, error) {
	opt := ProverOption{LoggerOut: os.Stdout, HintFunctions: hint.GetAll()}
	for _, option := range opts {
		if err := option(&opt); err != nil {
			return ProverOption{}, err
		}
	}
	return opt, nil
}

// ProverOption is shared accross backends to parametrize calls to xxx.Prove(...)
type ProverOption struct {
	Force         bool                     // default to false
	HintFunctions []hint.AnnotatedFunction // default to nil (use only solver std hints)
	LoggerOut     io.Writer                // default to os.Stdout
}

// IgnoreSolverError is a ProverOption that indicates that the Prove algorithm
// should complete, even if constraint system is not solved.
// In that case, Prove will output an invalid Proof, but will execute all algorithms
// which is useful for test and benchmarking purposes
func IgnoreSolverError(opt *ProverOption) error {
	opt.Force = true
	return nil
}

// WithHints is a Prover option that specifies additional hint functions to be used
// by the constraint solver
func WithHints(hintFunctions ...hint.AnnotatedFunction) func(opt *ProverOption) error {
	return func(opt *ProverOption) error {
		opt.HintFunctions = append(opt.HintFunctions, hintFunctions...)
		return nil
	}
}

// WithOutput is a Prover option that specifies an io.Writer as destination for logs printed by
// api.Println(). If set to nil, no logs are printed.
func WithOutput(w io.Writer) func(opt *ProverOption) error {
	return func(opt *ProverOption) error {
		opt.LoggerOut = w
		return nil
	}
}
