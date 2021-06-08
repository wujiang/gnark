package main

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/examples/cubic"
	"github.com/consensys/gnark/frontend"
)

//go:generate go run generate.go
func main() {
	var circuit, witness cubic.Circuit

	// witness part is temporary for PLONK, while the Setup is not split into 2.
	// witness.X.Assign(3)
	witness.Y.Assign(35)

	for _, b := range backend.Implemented() {
		if b == backend.PLONK {
			continue // TODO @gbotrel not ready yet.
		}
		for _, curve := range ecc.Implemented() {
			circuitID := filepath.Join(b.String(), curve.String(), "cubic")
			os.MkdirAll(circuitID, 0700)

			log.Println("compiling", circuitID)
			ccs, err := frontend.Compile(curve, b, &circuit)
			if err != nil {
				log.Fatal(err)
			}
			writeGnarkObject(ccs, filepath.Join(circuitID, "cubic"+".ccs"))

			if b == backend.GROTH16 {
				log.Println("groth16 setup", circuitID)
				pk, vk, err := groth16.Setup(ccs)
				if err != nil {
					log.Fatal(err)
				}
				writeGnarkObject(pk, filepath.Join(circuitID, "cubic"+".pk"))
				writeGnarkObject(vk, filepath.Join(circuitID, "cubic"+".vk"))
			} else if b == backend.PLONK {
				log.Println("plonk setup", circuitID)
				// TODO @gbotrel @thomas --> problem here, Setup should be split into witness dependent / independent part.
				// TODO looks ugly
				// sparseR1CS := ccs.(*cs.SparseR1CS)
				// nbConstraints := len(sparseR1CS.Constraints)
				// nbVariables := sparseR1CS.NbInternalVariables + sparseR1CS.NbPublicVariables + sparseR1CS.NbSecretVariables
				// var s, size int
				// if nbConstraints > nbVariables {
				// 	s = nbConstraints
				// } else {
				// 	s = nbVariables
				// }
				// size = 1
				// for ; size < s; size *= 2 {
				// }
				// var alpha fr.Element
				// alpha.SetRandom()
				// kate := kzg.NewScheme(size, alpha)

				publicData, err := plonk.Setup(ccs, nil, &witness)
				if err != nil {
					log.Fatal(err)
				}
				writeGnarkObject(publicData, filepath.Join(circuitID, "cubic"+".data"))
			}

		}

	}

}

func writeGnarkObject(o io.WriterTo, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	_, err = o.WriteTo(file)
	file.Close()
	if err != nil {
		log.Fatal(err)
	}
}