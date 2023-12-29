package ordered_test

import (
	stdcmp "cmp"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/ordered"
)

func ExampleMap() {
	// Create a map of string keys and integer elements, ordered by lexical
	// comparison of the string keys.
	m := ordered.NewMap[string, int](stdcmp.Compare)
	m.Set("foo", 1)
	m.Set("bar", 2)
	m.Set("baz", 3)

	// Modify some elements of the map.
	m.Set("foo", 10)
	m.Delete("bar")

	// Look up an element that does not exist.
	if v, ok := m.TryGet("notfound"); ok {
		fmt.Printf("found notfound: %d!\n", v)
	}

	// Iterate over the elements of the map, in order. Read-only accesses are
	// permitted while an iterator is open, but writes will result in a panic
	// until mi.Close is called. For more basic iteration, see Map.Range.
	mi := m.Iter()
	defer mi.Close()

	fmt.Println("length:", m.Len())
	for kv := mi.Next(); kv != nil; kv = mi.Next() {
		fmt.Printf("- %s: %d\n", kv.Key, kv.Value)
	}

	// Output:
	// length: 2
	// - baz: 3
	// - foo: 10
}

func TestMapBasics(t *testing.T) {
	// Initial map of 3 elements.
	m := testMap()

	if diff := cmp.Diff(3, m.Len()); diff != "" {
		t.Fatalf("unexpected initial length (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(1, m.Get("foo")); diff != "" {
		t.Fatalf("unexpected foo value (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(0, m.Get("notfound")); diff != "" {
		t.Fatalf("unexpected notfound value (-want +got):\n%s", diff)
	}

	// foo exists, notfound does not.
	_, aOK := m.TryGet("foo")
	if diff := cmp.Diff(true, aOK); diff != "" {
		t.Fatalf("unexpected foo OK value (-want +got):\n%s", diff)
	}

	_, bOK := m.TryGet("notfound")
	if diff := cmp.Diff(false, bOK); diff != "" {
		t.Fatalf("unexpected notfound OK value (-want +got):\n%s", diff)
	}

	// foo is updated.
	m.Set("foo", 10)
	if diff := cmp.Diff(10, m.Get("foo")); diff != "" {
		t.Fatalf("unexpected updated foo value (-want +got):\n%s", diff)
	}

	// Delete all but one key.
	m.Delete("bar")
	m.Delete("baz")

	if diff := cmp.Diff(1, m.Len()); diff != "" {
		t.Fatalf("unexpected post-delete length (-want +got):\n%s", diff)
	}

	// Clear the remaining keys.
	m.Reset()
	if diff := cmp.Diff(0, m.Get("foo")); diff != "" {
		t.Fatalf("unexpected final foo value (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(0, m.Len()); diff != "" {
		t.Fatalf("unexpected final length (-want +got):\n%s", diff)
	}
}

func TestMapRange(t *testing.T) {
	m := testMap()

	var (
		want = []string{"bar", "baz", "foo"}
		got  []string
	)

	for _, kv := range m.Range() {
		got = append(got, kv.Key)

		// Reads okay during iteration.
		if diff := cmp.Diff(kv.Value, m.Get(kv.Key)); diff != "" {
			t.Fatalf("unexpected value for key %q (-want +got):\n%s", kv.Key, diff)
		}
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected keys (-want +got):\n%s", diff)
	}
}

func TestMapIterate(t *testing.T) {
	m := testMap()
	m.Set("zzz", 999)

	var (
		want = []string{"bar", "baz", "foo"}
		got  []string
	)

	mi := m.Iter()
	defer mi.Close()
	for i, kv := 0, mi.Next(); i < 3 && kv != nil; i, kv = i+1, mi.Next() {
		got = append(got, kv.Key)

		// Reads okay during iteration.
		if diff := cmp.Diff(kv.Value, m.Get(kv.Key)); diff != "" {
			t.Fatalf("unexpected value for key %q (-want +got):\n%s", kv.Key, diff)
		}
	}

	for i := 0; i < 10; i++ {
		if mi.Next() != nil {
			t.Fatalf("next returned non-nil for completed iterator")
		}
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected keys (-want +got):\n%s", diff)
	}
}

func TestMapIterateMultiple(t *testing.T) {
	m := testMap()

	var (
		want = []string{"bar", "bar", "baz", "baz", "foo", "foo"}
		got  []string
	)

	mi0, mi1 := m.Iter(), m.Iter()
	defer mi0.Close()
	defer mi1.Close()

	for kv0, kv1 := mi0.Next(), mi1.Next(); kv0 != nil && kv1 != nil; kv0, kv1 = mi0.Next(), mi1.Next() {
		got = append(got, kv0.Key)
		got = append(got, kv1.Key)
	}

	for i := 0; i < 10; i++ {
		if mi0.Next() != nil {
			t.Fatalf("next returned non-nil for completed iterator")
		}
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected keys (-want +got):\n%s", diff)
	}
}

func TestMapIterateEmpty(t *testing.T) {
	m := ordered.NewMap[string, int](stdcmp.Compare)

	// Never called, no keys.
	for range m.Range() {
		t.Fatal("range loop entered for empty map")
	}

	mi := m.Iter()
	defer mi.Close()
	for i := 0; i < 10; i++ {
		if mi.Next() != nil {
			t.Fatalf("next returned non-nil for empty map")
		}
	}
}

func TestMapZeroPanics(t *testing.T) {
	var m0 *ordered.Map[string, int]
	if !panics(t, func() { m0.Len() }) {
		t.Fatal("expected nil map panic, but got none")
	}

	var m1 ordered.Map[string, int]
	if !panics(t, func() { m1.Len() }) {
		t.Fatal("expected zero map panic, but got none")
	}

	if !panics(t, func() { ordered.NewMap[string, int](nil) }) {
		t.Fatal("expected nil less panic, but got none")
	}
}

func TestMapMethodPanics(t *testing.T) {
	tests := []struct {
		name string
		fn   func(m *ordered.Map[string, int])
	}{
		{
			name: "iter close twice",
			fn: func(m *ordered.Map[string, int]) {
				mi := m.Iter()
				mi.Close()
				mi.Close()
			},
		},
		{
			name: "iter set",
			fn: func(m *ordered.Map[string, int]) {
				_ = m.Iter()
				m.Set("panic", 0)
			},
		},
		{
			name: "iter delete",
			fn: func(m *ordered.Map[string, int]) {
				_ = m.Iter()
				m.Delete("panic")
			},
		},
		{
			name: "iter reset",
			fn: func(m *ordered.Map[string, int]) {
				_ = m.Iter()
				m.Reset()
			},
		},
		{
			name: "iter nil",
			fn: func(_ *ordered.Map[string, int]) {
				var mi *ordered.MapIterator[string, int]
				mi.Next()
			},
		},
		{
			name: "iter zero",
			fn: func(_ *ordered.Map[string, int]) {
				var mi ordered.MapIterator[string, int]
				mi.Next()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := testMap()
			if !panics(t, func() { tt.fn(m) }) {
				t.Fatal("expected a panic, but got none")
			}
		})
	}
}

func testMap() *ordered.Map[string, int] {
	m := ordered.NewMap[string, int](stdcmp.Compare)
	m.Set("foo", 1)
	m.Set("bar", 2)
	m.Set("baz", 3)

	return m
}

func panics(t *testing.T, fn func()) (panics bool) {
	defer func() {
		r := recover()
		if r == nil {
			panics = false
			return
		}

		t.Logf("panic: %s", r)
		panics = true
	}()

	fn()
	return false
}
