/*
Package hint allows to define computations outside of a circuit.

Usually, it is expected that computations in circuits are performed on
variables. However, in some cases defining the computations in circuits may be
complicated or computationally expensive. By using hints, the computations are
performed outside of the circuit on integers (compared to the frontend.Variable
values inside the circuits) and the result of a hint function is assigned to a
newly created variable in a circuit.

As the computations are perfomed outside of the circuit, then the correctness of
the result is not guaranteed. This also means that the result of a hint function
is unconstrained by default, leading to failure while composing circuit proof.
Thus, it is the circuit developer responsibility to verify the correctness hint
result by adding necessary constraints in the circuit.

As an example, lets say the hint function computes a factorization of a
semiprime n:

    p, q <- hint(n) st. p * q = n

into primes p and q. Then, the circuit developer needs to assert in the circuit
that p*q indeed equals to n:

    n == p * q.

However, if the hint function is incorrectly defined (e.g. in the previous
example, it returns 1 and n instead of p and q), then the assertion may still
hold, but the constructed proof is semantically invalid. Thus, the user
constructing the proof must be extremely cautious when using hints.

Using hint functions in circuits

To use a hint function in a circuit, the developer first needs to define a hint
function hintFn according to the Function type. Then, in a circuit, the
developer applies the hint function with frontend.API.NewHint(hintFn, vars...),
where vars are the variables the hint function will be applied to (and
correspond to the argument inputs in the Function type) which returns a new
unconstrained variable. The returned variable must be constrained using
frontend.API.Assert[.*] methods.

As explained, the hints are essentially black boxes from the circuit point of
view and thus the defined hints in circuits are not used when constructing a
proof. To allow the particular hint functions to be used during proof
construction, the user needs to supply a backend.ProverOption indicating the
enabled hints. Such options can be optained by a call to
backend.WithHints(hintFns...), where hintFns are the corresponding hint
functions.

Using hint functions in gadgets

Similar considerations apply for hint functions used in gadgets as in
user-defined circuits. However, listing all hint functions used in a particular
gadget for constructing backend.ProverOption puts high overhead for the user to
enable all necessary hints.

For that, this package also provides a registry of trusted hint functions. When
a gadget registers a hint function, then it is automatically enabled during
proof computation and the prover does not need to provide a corresponding
proving option.

In the init() method of the gadget, call the method Register(hintFn) method on
the hint function hintFn to register a hint function in the package registry.
*/
package hint

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math/big"
	"reflect"
	"runtime"

	"github.com/consensys/gnark-crypto/ecc"
)

// ID is a unique identifier for a hint function used for lookup.
type ID uint32

// Function defines how a hint is computed from the inputs. The hint value is
// stored in res. If the hint is computable, then the function must return a nil
// error and non-nil error otherwise.
type Function func(curveID ecc.ID, inputs []*big.Int, res []*big.Int) error

// AnnotatedFunction defines an annotated hint function.
type AnnotatedFunction interface {
	// UUID returns an unique identifier for the hint function. UUID is used for
	// lookup of the hint function.
	UUID() ID

	// Call is invoked by the framework to obtain the result from inputs. It is
	// ensured that the number of inputs is exactly TotalInputs() if it is
	// non-negative. If TotalInputs() is negative, then the length of inputs is
	// not bounded. The length of res is TotalOutputs() and every element is
	// already initialized (but not necessarily to zero as the elements may be
	// obtained from cache). A returned non-nil error will be propagated.
	Call(curveID ecc.ID, inputs []*big.Int, res []*big.Int) error

	// TotalInputs returns the total number of inputs accepted by the function.
	// If the returned value is negative, then the function takes any number of
	// inputs. If it is zero, then the function does not take any inputs.
	TotalInputs() int

	// TotalOutputs returns the total number of outputs by the function on
	// nInputs number of inputs. The number of outputs must be at least one and
	// the framework errors otherwise.
	TotalOutputs(nInputs int) int

	// String returns a human-readable description of the function used in logs
	// and debug messages.
	String() string
}

// fixedArgumentsFunction defines a function where the number of inputs and
// outputs is fixed.
type fixedArgumentsFunction struct {
	fn   Function
	nIn  int
	nOut int
}

// NewFixedHint returns an AnnotatedFunction where the number of inputs and outputs
// is constant. UUID is computed by combining fn, nIn and nOut and thus it is
// legal to defined multiple AnnotatedFunctions on the same fn with different
// nIn and nOut.
func NewFixedHint(fn Function, nIn, nOut int) AnnotatedFunction {
	return &fixedArgumentsFunction{
		fn:   fn,
		nIn:  nIn,
		nOut: nOut,
	}
}

func (h *fixedArgumentsFunction) Call(curveID ecc.ID, inputs []*big.Int, res []*big.Int) error {
	if len(inputs) != h.nIn {
		return fmt.Errorf("input has %d elements, expected %d", len(inputs), h.nIn)
	}
	if len(res) != h.nOut {
		return fmt.Errorf("result has %d elements, expected %d", len(res), h.nOut)
	}
	return h.fn(curveID, inputs, res)
}

func (h *fixedArgumentsFunction) TotalInputs() int {
	return h.nIn
}

func (h *fixedArgumentsFunction) TotalOutputs(_ int) int {
	return h.nOut
}

func (h *fixedArgumentsFunction) UUID() ID {
	var buf [8]byte
	hf := fnv.New32a()
	name := runtime.FuncForPC(reflect.ValueOf(h.fn).Pointer()).Name()
	// using a name for identifying different hints should be enough as we get a
	// solve-time error when there are duplicate hints with the same signature.
	hf.Write([]byte(name))
	// also feed in the number of input and output variables. This allows to use
	// the same hint function with different signatures.
	binary.BigEndian.PutUint64(buf[:], uint64(h.nIn))
	hf.Write(buf[:])
	binary.BigEndian.PutUint64(buf[:], uint64(h.nOut))
	hf.Write(buf[:])
	return ID(hf.Sum32())
}

func (h *fixedArgumentsFunction) String() string {
	fnptr := reflect.ValueOf(h.fn).Pointer()
	name := runtime.FuncForPC(fnptr).Name()
	return fmt.Sprintf("%s([%d]*big.Int, [%d]*big.Int) at (%x)", name, h.TotalInputs(), h.TotalOutputs(0), fnptr)
}
