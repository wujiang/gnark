package multihint

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/hint"
	"github.com/consensys/gnark/frontend"
)

var multiHint = hint.NewFixedHint(func(curveID ecc.ID, inputs, res []*big.Int) error {
	res[0].Mul(inputs[0], big.NewInt(2))
	res[1].Mul(inputs[1], big.NewInt(2))
	return nil
}, 2, 2)

type Circuit struct {
	A0, A1 frontend.Variable `gnark:",secret"`
	B0, B1 frontend.Variable `gnark:",public"`
}

func (c *Circuit) Define(api frontend.API) error {
	b, err := api.NewHint(multiHint, c.A0, c.A1)
	if err != nil {
		return fmt.Errorf("multiHint: %w", err)
	}
	a0 := api.Mul(c.A0, 2)
	a1 := api.Mul(c.A1, 2)
	api.AssertIsEqual(b[0], a0)
	api.AssertIsEqual(b[1], a1)
	api.AssertIsEqual(b[0], c.B0)
	api.AssertIsEqual(b[1], c.B1)
	return nil
}

type ExpCircuit struct {
	X, E frontend.Variable
	Y    frontend.Variable `gnark:",public"`
}

func (circuit *ExpCircuit) Define(cs frontend.API) error {
	o := frontend.Variable(1)
	b := cs.ToBinary(circuit.E, 4)

	var i int
	for i < len(b) {
		o = cs.Mul(o, o)
		mu := cs.Mul(o, circuit.X)
		o = cs.Select(b[len(b)-1-i], mu, o)
		i++
	}
	cs.AssertIsEqual(circuit.Y, o)
	return nil
}

type rangeCheckCircuit struct {
	X        frontend.Variable
	Y, Bound frontend.Variable `gnark:",public"`
}

func (circuit *rangeCheckCircuit) Define(cs frontend.API) error {
	c1 := cs.Mul(circuit.X, circuit.Y)
	c2 := cs.Mul(c1, circuit.Y)
	c3 := cs.Add(circuit.X, circuit.Y)
	cs.AssertIsLessOrEqual(c2, circuit.Bound)
	cs.AssertIsLessOrEqual(c3, circuit.Bound) // c3 is from a linear expression only

	return nil
}

type zeroCircuit struct {
	X, Y frontend.Variable
}

func (circuit *zeroCircuit) Define(cs frontend.API) error {
	a := cs.IsZero(circuit.X)
	cs.AssertIsEqual(a, 1)
	b := cs.IsZero(circuit.Y)
	cs.AssertIsEqual(b, 1)

	return nil
}
