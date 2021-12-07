package hint

import (
	"fmt"
	"sync"
)

var registry = make(map[ID]AnnotatedFunction)
var registryM sync.RWMutex

// Register registers an annotated hint function in the global registry. All
// registered hint functions can be retrieved with a call to GetAll(). It is an
// error to register a single function twice and results in a panic.
func Register(hintFn AnnotatedFunction) {
	registryM.Lock()
	defer registryM.Unlock()
	key := hintFn.UUID()
	if _, ok := registry[key]; ok {
		panic(fmt.Sprintf("function %s registered twice", hintFn))
	}
	registry[key] = hintFn
}

// GetAll returns all registered hint functions.
func GetAll() []AnnotatedFunction {
	registryM.RLock()
	defer registryM.RUnlock()
	ret := make([]AnnotatedFunction, 0, len(registry))
	for _, v := range registry {
		ret = append(ret, v)
	}
	return ret
}
