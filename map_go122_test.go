//go:build go1.22 || go1.22rc1

package ordered_test

import (
	"fmt"

	"github.com/mdlayher/ordered"
)

func ExampleMapAll() {
	m := ordered.NewMap[string, int](ordered.Less[string])
	m.Set("foo", 1)
	m.Set("bar", 2)
	m.Set("baz", 3)

	for k, v := range m.All() {
		fmt.Println(k, v)
	}

	// Output:
	// bar 2
	// baz 3
	// foo 1
}
