//go:build go1.22 || go1.22rc1

package ordered

// seq2 is iter.Seq2 without an import.
type seq2[K, V any] func(yield func(K, V) bool)

// All yields key/value pairs from Map for use in a for-range loop with
// GOEXPERIMENT=rangefunc.
func (m *Map[K, V]) All() seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, k := range m.keys {
			if !yield(k, m.m[k]) {
				return
			}
		}
	}
}
