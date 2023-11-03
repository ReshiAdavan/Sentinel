package gobWrapper

/*
 * Wrapper around Go's encoding/gob, that
 * Checks and warns about capitalization.
 */

import (
	"encoding/gob"
	"fmt"
	"io"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"
)

var mu sync.Mutex
var errorCount int // for TestCapital
var checked map[reflect.Type]bool

type Encoder struct {
	gob *gob.Encoder
}

func NewEncoder(w io.Writer) *Encoder {
	enc := &Encoder{}
	enc.gob = gob.NewEncoder(w)
	return enc
}

func (enc *Encoder) Encode(e interface{}) error {
	checkValue(e)
	return enc.gob.Encode(e)
}

func (enc *Encoder) EncodeValue(value reflect.Value) error {
	checkValue(value.Interface())
	return enc.gob.EncodeValue(value)
}

type Decoder struct {
	gob *gob.Decoder
}

func NewDecoder(r io.Reader) *Decoder {
	dec := &Decoder{}
	dec.gob = gob.NewDecoder(r)
	return dec
}

func (dec *Decoder) Decode(e interface{}) error {
	checkValue(e)
	checkDefault(e)
	return dec.gob.Decode(e)
}

func Register(value interface{}) {
	checkValue(value)
	gob.Register(value)
}

func RegisterName(name string, value interface{}) {
	checkValue(value)
	gob.RegisterName(name, value)
}

func checkValue(value interface{}) {
	checkType(reflect.TypeOf(value))
}

func checkType(t reflect.Type) {
	k := t.Kind()

	mu.Lock()
	// only complain once, and avoid recursion.
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
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			rune, _ := utf8.DecodeRuneInString(f.Name)
			if !unicode.IsUpper(rune) {
				// ta da
				fmt.Printf("gobWrapper warning: lower-case field %v of %v won't work over RPC or in persist/snapshot\n",
					f.Name, t.Name())
				mu.Lock()
				errorCount += 1
				mu.Unlock()
			}
			checkType(f.Type)
		}
		return
	case reflect.Slice, reflect.Array, reflect.Ptr:
		checkType(t.Elem())
		return
	case reflect.Map:
		checkType(t.Elem())
		checkType(t.Key())
		return
	default:
		return
	}
}

/*
 * Warn if the value contains non-default values, as it would if one sent an RPC but the reply
 struct was already modified. 
 * If the RPC reply contains default values, GOB won't overwrite the non-default value.
 */

func checkDefault(value interface{}) {
	if value == nil {
		return
	}
	checkDefault1(reflect.ValueOf(value), 1, "")
}

func checkDefault1(value reflect.Value, depth int, name string) {
	if depth > 3 {
		return
	}

	t := value.Type()
	k := t.Kind()

	switch k {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			vv := value.Field(i)
			name1 := t.Field(i).Name
			if name != "" {
				name1 = name + "." + name1
			}
			checkDefault1(vv, depth+1, name1)
		}
		return
	case reflect.Ptr:
		if value.IsNil() {
			return
		}
		checkDefault1(value.Elem(), depth+1, name)
		return
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.Float32, reflect.Float64,
		reflect.String:
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
		return
	}
}
