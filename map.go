package ordered

import (
	"sync/atomic"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// An op is a read-only or read-write operation, used to annotate invariant
// checks.
//enumcheck:exhaustive
type op int

// Possible op types.
const (
	_ op = iota
	ro
	rw
)

// A Map is like a map[K]V, but offers deterministic iteration order by applying
// a comparison function against all keys. A Map must be constructed using
// NewMap or its methods will panic.
//
// Maps are not safe for concurrent use.
type Map[K comparable, V any] struct {
	// Atomic: whether or not a MapIterator is live for this Map.
	iter int32

	// A sorted list of keys stored in the map and the function to compare those
	// keys.
	keys []K
	less func(a, b K) bool

	// The actual underlying map storage.
	m map[K]V
}

// Less is a comparison function for key types which are ordered. It is a
// convenience function for comparing primitive types with NewMap.
func Less[K constraints.Ordered](a, b K) bool { return a < b }

// NewMap creates a *Map[K, V] which uses the comparison function less to order
// the keys in the map. less must not be nil or NewMap will panic. For primitive
// types, Less can be used as a comparison function.
func NewMap[K comparable, V any](less func(a, b K) bool) *Map[K, V] {
	if less == nil {
		panic("ordered: NewMap must use a non-nil less function")
	}

	return &Map[K, V]{
		m:    make(map[K]V),
		less: less,
	}
}

// Get gets the value V for a given key K, returning the zero value of V if K is
// not found.
func (m *Map[K, V]) Get(k K) V {
	m.check(ro)
	return m.m[k]
}

// TryGet tries to get the value V for a given key K, returning false if K is
// not found.
func (m *Map[K, V]) TryGet(k K) (V, bool) {
	m.check(ro)
	v, ok := m.m[k]
	return v, ok
}

// Len returns the number of elements in the Map.
func (m *Map[K, V]) Len() int {
	m.check(ro)
	return len(m.keys)
}

// Set inserts or updates the value V for a given key K.
func (m *Map[K, V]) Set(k K, v V) {
	m.check(rw)

	if _, ok := m.m[k]; !ok {
		// Always sort when a new key is inserted.
		m.keys = append(m.keys, k)
		slices.SortFunc(m.keys, m.less)
	}

	m.m[k] = v
}

// Delete deletes the value for a given key K.
func (m *Map[K, V]) Delete(k K) {
	m.check(rw)

	i := slices.Index(m.keys, k)
	if i != -1 {
		// Found this key, remove it from the order index.
		m.keys = slices.Delete(m.keys, i, i+1)
	}

	delete(m.m, k)
}

// Reset clears the underlying storage for a Map by removing all elements,
// enabling the allocated capacity to be reused.
func (m *Map[K, V]) Reset() {
	m.check(rw)

	m.keys = m.keys[:0]
	maps.Clear(m.m)
}

// check checks the Map's invariants for a given operation type.
func (m *Map[K, V]) check(op op) {
	if m == nil || m.less == nil {
		panic("ordered: a Map must be constructed using NewMap")
	}

	if op == rw && atomic.LoadInt32(&m.iter) != 0 {
		panic("ordered: write to Map while MapIterator is not closed")
	}
}

// A KeyValue is a key/value pair produced by a MapIterator or Map.Range call.
type KeyValue[K comparable, V any] struct {
	Key   K
	Value V
}

// Range produces a slice of all KeyValue pairs from Map for use in a for range
// loop. See Map.Iter for more fine-grained iteration control.
func (m *Map[K, V]) Range() []KeyValue[K, V] {
	m.check(ro)

	kvs := make([]KeyValue[K, V], 0, len(m.keys))
	for _, k := range m.keys {
		kvs = append(kvs, KeyValue[K, V]{
			Key:   k,
			Value: m.m[k],
		})
	}

	return kvs
}

// A MapIterator is an iteration cursor over a Map. A MapIterator must be
// constructed using Map.Iter or its methods will panic.
//
// When a MapIterator is created, any methods which write to a Map (Delete,
// Reset, Set) will panic. Reads during iteration are permitted. To complete
// iteration and permit further writes, call MapIterator.Close. Multiple
// MapIterators can be used at once over the same Map, but write methods will
// panic until all MapIterators are closed. After a call to Close, the
// MapIterator can no longer be used.
//
// For more basic iteration use cases, see Map.Range.
type MapIterator[K comparable, V any] struct {
	m *Map[K, V]
	i int
}

// Iter produces a MapIterator which allows fine-grained iteration over a Map.
func (m *Map[K, V]) Iter() *MapIterator[K, V] {
	m.check(ro)

	// Add another iterator to the stack.
	atomic.AddInt32(&m.iter, 1)
	return &MapIterator[K, V]{m: m}
}

// Close releases a MapIterator's resources, enabling further writes to a Map.
func (mi *MapIterator[K, V]) Close() {
	mi.check()

	// Remove an iterator from the stack. If this number goes below zero, panic
	// due to misuse.
	if atomic.AddInt32(&mi.m.iter, -1) < 0 {
		panic("ordered: call to MapIterator.Close while MapIterator is not open")
	}

	mi = nil
}

// Next returns the next KeyValue pair from a Map. If Next returns nil, no more
// KeyValue pairs are present. Next is intended to be used in a for loop, in the
// format:
//
//  mi := m.Iter()
//  defer mi.Close()
//  for kv := mi.Next(); kv != nil; kv = mi.Next() {
//      // use kv
//  }
func (mi *MapIterator[K, V]) Next() *KeyValue[K, V] {
	mi.check()

	if mi.i >= len(mi.m.keys) {
		// No more keys.
		return nil
	}

	k := mi.m.keys[mi.i]
	mi.i++

	return &KeyValue[K, V]{
		Key:   k,
		Value: mi.m.m[k],
	}
}

// check checks the MapIterator's invariants.
func (mi *MapIterator[K, V]) check() {
	if mi == nil || mi.m == nil {
		panic("ordered: a MapIterator must be constructed using Map.Iter")
	}
}
