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

package cubic

import (
	"testing"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test"
)

func TestCubicEquation(t *testing.T) {
	assert := test.NewAssert(t)

	var cubicCircuit Circuit

	assert.ProverFailed(&cubicCircuit, &Circuit{
		X: frontend.Value(42),
		Y: frontend.Value(42),
	})

	assert.ProverSucceeded(&cubicCircuit, &Circuit{
		X: frontend.Value(3),
		Y: frontend.Value(35),
	})

}
