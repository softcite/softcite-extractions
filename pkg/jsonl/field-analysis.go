package jsonl

import (
	"fmt"
	"math"
	"strings"
)

// MaxEnum is the largest number of unique values to track before not trying to
// interpret the field as an enum.
const MaxEnum = 20

// Field represents JSON atomic types: number, boolean, string, and null.
// Does not support fields which may hold different types
// (e.g. number and string), but does allow fields which sometimes have null
// values.
type Field interface {
	Add(obj any) (Field, error)
	String() string
}

// EmptyField represents a field which is never filled in.
// Adding any object to a EmptyField returns a non-EmptyField.
type EmptyField struct{}

// Add turns the EmptyField into an appropriate field based on the passed type.
func (nf *EmptyField) Add(obj any) (Field, error) {
	switch o := obj.(type) {
	case nil:
		return nf, nil
	case bool:
		var f Field = &BoolField{}
		f, err := f.Add(o)
		if err != nil {
			return nil, err
		}
		return f, nil
	case float64:
		var f Field = &NumberField{
			Seen: make(map[float64]int),
		}
		f, err := f.Add(o)
		if err != nil {
			return nil, err
		}
		return f, nil
	case string:
		var f Field = &StringField{
			Seen: make(map[string]int),
		}
		f, err := f.Add(o)
		if err != nil {
			return nil, err
		}
		return f, nil
	default:
		return nil, fmt.Errorf("unknown type %T added to %T", o, nf)
	}
}

func (nf *EmptyField) String() string {
	return "empty"
}

// BoolField indicates the field only ever holds "true" or "false" JSON boolean
// values.
type BoolField struct {
	True  int
	False int
}

func (f *BoolField) Add(obj any) (Field, error) {
	switch o := obj.(type) {
	case nil:
		return f, nil
	case bool:
		if o {
			f.True++
		} else {
			f.False++
		}
		return f, nil
	default:
		return nil, fmt.Errorf("unknown type %T added to %T", o, f)
	}
}

func (f *BoolField) String() string {
	return fmt.Sprintf("true:%d;false:%d", f.True, f.False)
}

// A NumberField only holds JSON numbers. Keeps track of the properties of the
// numbers passed in to determine the types of numbers used.
type NumberField struct {
	// Integral tracks if all instances of this field are integers.
	Integral bool
	// Float32 tracks if all instances of this field can fit in a 32-bit floating
	// point type. Note that integers greater than about 2^23 cannot fit in
	// 32-bit floats.
	Float32 bool

	// Min and Max allow determining whether the number is unsigned, or, for
	// integers, the smallest type which can hold all seen values.
	Min, Max float64

	// Seen tracks the unique numbers passed to this field.
	// Used for detecting if this is an enumerated field where only a few
	// unique values are passed.
	// Stops collecting values after it contains more than MaxEnum entries.
	Seen map[float64]int
}

func (f *NumberField) Add(obj any) (Field, error) {
	switch o := obj.(type) {
	case nil:
		return f, nil
	case float64:
		if len(f.Seen) > 0 {
			f.Integral = f.Integral && isIntegral(o)
			f.Float32 = f.Float32 && isFloat32(o)

			if o < f.Min {
				f.Min = o
			} else if o > f.Max {
				f.Max = o
			}
		} else {
			f.Integral = isIntegral(o)
			f.Float32 = isFloat32(o)

			f.Min = o
			f.Max = o
		}

		if len(f.Seen) <= MaxEnum {
			f.Seen[o]++
		}
		return f, nil
	default:
		return nil, fmt.Errorf("unknown type %T added to %T", o, f)
	}
}

func isIntegral(f float64) bool {
	return math.Round(f) == f
}

const (
	Float64FractionLength = 52
	Float32FractionLength = 23
	Float64Mask           = (1 << (Float64FractionLength - Float32FractionLength)) - 1

	// DropRight is the number of decimal places at the end to discard when testing for
	// if the value is a float32.
	DropRight = 3
)

func isFloat32(f float64) bool {
	n := math.Float64bits(f)
	n &= Float64Mask

	// The number can be represented as a float32 without loss of precision as
	// it uses none of the float64-specific fraction bits.
	// Does not handle exponents out of the range of float32.
	return n == 0
}

func (f *NumberField) String() string {
	result := strings.Builder{}
	if f.Integral {
		if f.Min < 0 {
			if f.Max <= math.MaxInt8 {
				result.WriteString("int8")
			} else if f.Max <= math.MaxInt16 {
				result.WriteString("int16")
			} else if f.Max <= math.MaxInt32 {
				result.WriteString("int32")
			} else {
				result.WriteString("int64")
			}
		} else {
			if f.Max <= math.MaxUint8 {
				result.WriteString("uint8")
			} else if f.Max <= math.MaxUint16 {
				result.WriteString("uint16")
			} else if f.Max <= math.MaxUint32 {
				result.WriteString("uint32")
			} else {
				result.WriteString("uint64")
			}
		}
	} else {
		if f.Float32 {
			result.WriteString("float32")
		} else {
			result.WriteString("float64")
		}
	}
	result.WriteString(";")
	if f.Integral {
		result.WriteString(fmt.Sprintf("%d;%d", int(f.Min), int(f.Max)))
	} else {
		result.WriteString(fmt.Sprintf("%f;%f", f.Min, f.Max))
	}

	if len(f.Seen) <= MaxEnum {
		for k, v := range f.Seen {
			if f.Integral {
				result.WriteString(fmt.Sprintf("%d:%d;", int(k), v))
			} else {
				result.WriteString(fmt.Sprintf("%f:%d;", k, v))
			}
		}
	}

	return result.String()
}

// A StringField only holds JSON string values.
type StringField struct {
	// Seen attempts to determine if the field is actually an enum with a small
	// number of unique values.
	Seen map[string]int
}

func (f *StringField) Add(obj any) (Field, error) {
	switch o := obj.(type) {
	case nil:
		return f, nil
	case string:
		if len(f.Seen) <= MaxEnum {
			f.Seen[o]++
		}
		return f, nil
	default:
		return nil, fmt.Errorf("unknown type %T added to %T", o, f)
	}
}

func (f *StringField) String() string {
	result := strings.Builder{}
	if len(f.Seen) <= MaxEnum {
		result.WriteString(fmt.Sprintf("enum;%d;", len(f.Seen)))
		for k, v := range f.Seen {
			result.WriteString(fmt.Sprintf("%s:%d;", k, v))
		}
	} else {
		result.WriteString("string;")
	}

	return result.String()
}
