package encoder

// Original code borrowed from https://github.com/PuerkitoBio/martini-api-example
// TextEncoder and XmlEncoder has been removed. If someone really needs it, let me know.

// Supported tags:
// 	 - "out" if it sets to "false", value won't be set to field
import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"reflect"
	"time"
)

// An Encoder implements an encoding format of values to be sent as response to
// requests on the API endpoints.
type Encoder interface {
	Encode(v ...interface{}) ([]byte, error)
}

// Because `panic`s are caught by martini's Recovery handler, it can be used
// to return server-side errors (500). Some helpful text message should probably
// be sent, although not the technical error (which is printed in the log).
func Must(data []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return data
}

type JsonEncoder struct {
	PrettyPrint bool
}

// @todo test for slice of structure support
func (self JsonEncoder) Encode(v ...interface{}) ([]byte, error) {
	var data interface{} = v
	var result interface{}

	if v == nil {
		// So that empty results produces `[]` and not `null`
		data = []interface{}{}
	} else if len(v) == 1 {
		data = v[0]
	}

	t := reflect.TypeOf(data)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Slice:
		result = iterateSlice(reflect.ValueOf(data)).Interface()

	case reflect.Struct:
		result = copyStruct(reflect.ValueOf(data)).Interface()

	default:
		result = data
	}

	if self.PrettyPrint {
		return json.MarshalIndent(result, "", "	")
	} else {
		return json.Marshal(result)
	}
}

type XmlEncoder struct{}

// Since we don't use xml as a binding source, we don't need to use
// copyStruct here. Just Marshal.
func (_ XmlEncoder) Encode(v ...interface{}) ([]byte, error) {
	var data interface{} = v
	var buffer bytes.Buffer

	if v == nil {
		data = []interface{}{}
	} else if len(v) == 1 {
		data = v[0]
	}

	if _, err := buffer.Write([]byte(xml.Header)); err != nil {
		return []byte{}, err
	}

	b, err := xml.Marshal(data)
	if err != nil {
		return []byte{}, err
	}

	buffer.Write(b)

	return buffer.Bytes(), nil
}

func copyStruct(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	result := reflect.New(v.Type()).Elem()

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		vfield := v.Field(i)

		if tag := t.Field(i).Tag.Get("out"); tag == "false" {
			continue
		}

		if vfield.Type() == reflect.TypeOf(time.Time{}) {
			result.Field(i).Set(vfield)
			continue
		}

		if vfield.Kind() == reflect.Interface && vfield.Interface() != nil {
			vfield = vfield.Elem()

			for vfield.Kind() == reflect.Ptr {
				vfield = vfield.Elem()
			}

			result.Field(i).Set(copyStruct(vfield))
			continue
		}

		if vfield.Kind() == reflect.Struct || vfield.Kind() == reflect.Ptr {
			r := copyStruct(vfield)

			if result.Field(i).Kind() == reflect.Ptr {
				result.Field(i).Set(r.Addr())
			} else {
				result.Field(i).Set(r)
			}

			continue
		}

		if vfield.Kind() == reflect.Array || vfield.Kind() == reflect.Slice {
			result.Field(i).Set(iterateSlice(vfield))
			continue
		}

		if result.Field(i).CanSet() {
			result.Field(i).Set(vfield)
		}
	}

	return result
}

func iterateSlice(v reflect.Value) reflect.Value {
	result := reflect.MakeSlice(v.Type(), 0, v.Len())

	for i := 0; i < v.Len(); i++ {
		value := v.Index(i)

		if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
			result = reflect.Append(result, iterateSlice(value))
			continue
		}

		vi := value
		if value.Kind() == reflect.Struct {
			vi = copyStruct(value)
		}

		if value.Kind() == reflect.Ptr {
			result = reflect.Append(result, vi.Addr())
		} else {
			result = reflect.Append(result, vi)
		}
	}

	return result
}
