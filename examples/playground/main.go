package main

import (
	"log"
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/examples/cubic"
	"github.com/consensys/gnark/frontend"
)

func main() {
	var circuit cubic.Circuit

	ccs, err := frontend.Compile(ecc.BN254, backend.GROTH16, &circuit)
	if err != nil {
		log.Fatal(err)
	}

	fR1CS, err := os.Create("r1cs.html")
	if err != nil {
		log.Fatal(err)
	}

	if err := ccs.ToHTML(fR1CS); err != nil {
		log.Fatal(err)
	}

	fR1CS.Close()

}
