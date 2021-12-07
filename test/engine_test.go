package test

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/hint"
	"github.com/consensys/gnark/frontend"
)

type hintCircuit struct {
	A, B frontend.Variable
}

func (circuit *hintCircuit) Define(api frontend.API) error {
	res, err := api.NewHint(hint.IthBit, circuit.A, 3)
	if err != nil {
		return fmt.Errorf("IthBit circuitA 3: %w", err)
	}
	a3b := res[0]
	res, err = api.NewHint(hint.IthBit, circuit.A, 25)
	if err != nil {
		return fmt.Errorf("IthBit circuitA 25: %w", err)
	}
	a25b := res[0]
	res, err = api.NewHint(hint.IsZero, circuit.A)
	if err != nil {
		return fmt.Errorf("IsZero CircuitA: %w", err)
	}
	aisZero := res[0]
	res, err = api.NewHint(hint.IsZero, circuit.B)
	if err != nil {
		return fmt.Errorf("IsZero, CircuitB")
	}
	bisZero := res[0]

	api.AssertIsEqual(aisZero, 0)
	api.AssertIsEqual(bisZero, 1)
	api.AssertIsEqual(a3b, 1)
	api.AssertIsEqual(a25b, 0)

	return nil
}

func TestBuiltinHints(t *testing.T) {
	for _, curve := range ecc.Implemented() {
		if err := IsSolved(&hintCircuit{}, &hintCircuit{
			A: (0b1000),
			B: (0),
		}, curve); err != nil {
			t.Fatal(err)
		}

		if err := IsSolved(&hintCircuit{}, &hintCircuit{
			A: (0b10),
			B: (1),
		}, curve); err == nil {
			t.Fatal("witness shouldn't solve circuit")
		}
	}

}
