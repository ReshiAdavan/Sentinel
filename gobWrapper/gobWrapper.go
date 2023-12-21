package gobWrapper

// Wrapper around Go's encoding/gob that checks and warns about capitalization.

import (
	"encoding/gob"
	"fmt"
	"io"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"
)

var mu sync.Mutex // Mutex for synchronizing access to global variables
var errorCount int // Tracks the number of capitalization errors encountered
var checked map[reflect.Type]bool // Keeps track of already checked types

type Encoder struct {
	gob *gob.Encoder // Embeds gob.Encoder to handle the actual encoding
}

// NewEncoder creates a new Encoder that writes to the provided io.Writer.
func NewEncoder(w io.Writer) *Encoder {
	enc := &Encoder{}
	enc.gob = gob.NewEncoder(w)
	return enc
}

// Encode wraps gob.Encoder's Encode method, adding capitalization checks.
func (enc *Encoder) Encode(e interface{}) error {
	checkValue(e)
	return enc.gob.Encode(e)
}

// EncodeValue wraps gob.Encoder's EncodeValue method, adding capitalization checks.
func (enc *Encoder) EncodeValue(value reflect.Value) error {
	checkValue(value.Interface())
	return enc.gob.EncodeValue(value)
}

type Decoder struct {
	gob *gob.Decoder // Embeds gob.Decoder to handle the actual decoding
}

// NewDecoder creates a new Decoder that reads from the provided io.Reader.
func NewDecoder(r io.Reader) *Decoder {
	dec := &Decoder{}
	dec.gob = gob.NewDecoder(r)
	return dec
}

// Decode wraps gob.Decoder's Decode method, adding checks for capitalization and default values.
func (dec *Decoder) Decode(e interface{}) error {
	checkValue(e)
	checkDefault(e)
	return dec.gob.Decode(e)
}

// Register wraps gob.Register, adding a capitalization check for the value.
func Register(value interface{}) {
	checkValue(value)
	gob.Register(value)
}

// RegisterName wraps gob.RegisterName, adding a capitalization check for the value.
func RegisterName(name string, value interface{}) {
	checkValue(value)
	gob.RegisterName(name, value)
}

// checkValue performs capitalization checks on the provided value.
func checkValue(value interface{}) {
	checkType(reflect.TypeOf(value))
}

// checkType checks the type for capitalization issues and stores checked types to avoid repetition.
func checkType(t reflect.Type) {
	k := t.Kind()

	mu.Lock()
	if checked == nil {
		checked = map[reflect.Type]bool{}
	}
	if checked[t] {
		mu.Unlock()
		return
	}
	checked[t] = true
	mu.Unlock()

	switch k {
	case reflect.Struct:
		// Check each field of the struct for capitalization.
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			rune, _ := utf8.DecodeRuneInString(f.Name)
			if !unicode.IsUpper(rune) {
				fmt.Printf("gobWrapper warning: lower-case field %v of %v won't work over RPC or in persist/snapshot\n",
					f.Name, t.Name())
				mu.Lock()
				errorCount += 1
				mu.Unlock()
			}
			checkType(f.Type)
		}
	case reflect.Slice, reflect.Array, reflect.Ptr:
		// Check the element type of slices, arrays, and pointers.
		checkType(t.Elem())
	case reflect.Map:
		// Check both the key and value types of maps.
		checkType(t.Elem())
		checkType(t.Key())
	}
}

// checkDefault warns if the value contains non-default values, which can be problematic in RPC.
func checkDefault(value interface{}) {
	if value == nil {
		return
	}
	checkDefault1(reflect.ValueOf(value), 1, "")
}

// checkDefault1 is a recursive helper for checkDefault to check fields at various depths.
func checkDefault1(value reflect.Value, depth int, name string) {
	if depth > 3 {
		return
	}

	t := value.Type()
	k := t.Kind()

	switch k {
	case reflect.Struct:
		// Check each field of the struct.
		for i := 0; i < t.NumField(); i++ {
			vv := value.Field(i)
			name1 := t.Field(i).Name
			if name != "" {
				name1 = name + "." + name1
			}
			checkDefault1(vv, depth+1, name1)
		}
	case reflect.Ptr:
		// Check the dereferenced pointer if not nil.
		if !value.IsNil() {
			checkDefault1(value.Elem(), depth+1, name)
		}
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.Float32, reflect.Float64,
		reflect.String:
		// Warn if a primitive type field is not in its default state.
		if !reflect.DeepEqual(reflect.Zero(t).Interface(), value.Interface()) {
			mu.Lock()
			if errorCount < 1 {
				what := name
				if what == "" {
					what = t.Name()
				}
				fmt.Printf("gobWrapper warning: Decoding into a non-default variable/field %v may not work\n",
					what)
			}
			errorCount += 1
			mu.Unlock()
		}
	}
}
