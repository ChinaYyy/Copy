package expvar2

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
)

// RemoveAll removes all exported variables.
// This is for tests only.
func RemoveAll() {
	varKeysMu.Lock()
	defer varKeysMu.Unlock()

	for _, k := range varKeys {
		vars.Delete(k)
	}
	varKeys = nil
}

func TestNil(t *testing.T) {
	RemoveAll()

	val := Get("missing")
	if val != nil {
		t.Errorf("got %v, want nil", val)
	}
}

func TestInt(t *testing.T) {
	RemoveAll()

	reqs := NewInt("requests")
	if i := reqs.Value(); i != 0 {
		t.Errorf("reqs.Value() = %v, want 0", i)
	}

	if reqs != Get("requests").(*Int) {
		t.Errorf("Get() failed.")
	}

	reqs.Add(1)
	reqs.Add(3)
	if s := reqs.Value(); s != 4 {
		t.Errorf("reqs.String() = %q, want \"4\"", s)
	}

	reqs.Set(-2)
	if s := reqs.Value(); s != -2 {
		t.Errorf("reqs.String() = %q, want \"-2\"", s)
	}
}

func BenchmarkIntAdd(b *testing.B) {
	var v Int

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v.Add(1)
		}
	})
}

func BenchmarkIntSet(b *testing.B) {
	var v Int
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v.Set(1)
		}
	})
}

func TestFloat(t *testing.T) {
	RemoveAll()

	reqs := NewFloat("requests-float")
	if reqs.f != 0.0 {
		t.Errorf("reqs.f = %v, want 0", reqs.f)
	}

	if reqs != Get("requests-float").(*Float) {
		t.Errorf("Get() failed.")
	}

	reqs.Add(1.5)
	reqs.Add(1.25)

	if v := reqs.Value(); v != 2.75 {
		t.Errorf("reqs.Value() = %v, want 2.75", v)
	}

	reqs.Add(-2)
	if v := reqs.String(); v != "0.75" {
		t.Errorf("reqs.Value() = %v, want 0.75", v)
	}
}

func TestString(t *testing.T) {
	RemoveAll()
	name := NewString("my-name")
	if s := name.Value(); s != "" {
		t.Errorf(`NewString("my-name").Value() = %q, want""`, s)
	}

	name.Set("Mike")
	if s, want := name.String(), `"Mike"`; s != want {
		t.Errorf(`after name.Set("Mike"), name.String() = %q, want %q`, s, want)
	}
	if s, want := name.Value(), "Mike"; s != want {
		t.Errorf(`after name.Set("Mike"), name.Value() = %q, want %q`, s, want)
	}

	// Make sure we produce same JSON output
	name.Set("<")
	if s, want := name.String(), "\"\\u003c\""; s != want {
		t.Errorf(`after name.Set("<"), name.String() = %q, want %q`, s, want)
	}
}

func TestMapInit(t *testing.T) {
	RemoveAll()
	colors := NewMap("bike-shed-colors")
	colors.Add("red", 1)
	colors.Add("blue", 1)
	colors.Add("chartreuse", 1)

	n := 0
	colors.Do(func(KeyValue) {
		n++
	})
	if n != 3 {
		t.Errorf("after three Add calls with distinct keys, Do should invoke f 3 times; get %v", n)
	}

	colors.Init()
	n = 0
	colors.Do(func(KeyValue) { n++ })
	if n != 0 {
		t.Errorf("after Init, Do should invoke f 0 times; got %v", n)
	}
}

func TestFunc(t *testing.T) {
	RemoveAll()
	var x interface{} = []string{"a", "b"}
	f := Func(func() interface{} { return x })
	if s, exp := f.String(), `["a","b"]`; s != exp {
		t.Errorf("f.String() = %q, want %q", s, exp)
	}
	if v := f.Value(); !reflect.DeepEqual(v, x) {
		t.Errorf(`f.Value()=%q, want %q`, v, x)
	}

	x = 17
	if s, exp := f.String(), `17`; s != exp {
		t.Errorf(`f.String()=%q, want %q`, s, exp)
	}
}

func TestHandler(t *testing.T) {
	RemoveAll()
	m := NewMap("map1")
	m.Add("a", 1)
	m.Add("z", 2)
	m2 := NewMap("map2")

	for i := 0; i < 9; i++ {
		m2.Add(strconv.Itoa(i), int64(i))
		fmt.Println("--", i, m2.keys)
	}
	fmt.Println("all:", m2.keys)
	rr := httptest.NewRecorder()
	rr.Body = new(bytes.Buffer)
	expvarHandler(rr, nil)

	want := `{
"map1": {"a": 1, "z": 2},
"map2": {"0": 0, "1": 1, "2": 2, "3": 3, "4": 4, "5": 5, "6": 6, "7": 7, "8": 8}
}
`
	if got := rr.Body.String(); got != want {
		t.Errorf("HTTP handler wrote: \n%s\n%s", got, want)
	}
}
