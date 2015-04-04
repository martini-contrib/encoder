package encoder

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"reflect"
)

type Encoder interface {
	Encode(v interface{}) ([]byte, error)
}

type Filter interface {
	Filter() interface{}
}

func Must(data []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return data
}

type JsonEncoder struct {
	PrettyPrint bool
	PrintNull   bool // if true, empty object will be 'null', '{}' instead (by default)
}

func (e JsonEncoder) Encode(obj interface{}) ([]byte, error) {
	obj, _ = filter(obj)
	if obj == nil && !e.PrintNull {
		obj = struct{}{}
	}

	if e.PrettyPrint {
		return json.MarshalIndent(obj, "", "    ")
	} else {
		return json.Marshal(obj)
	}
}

type XmlEncoder struct {
	PrettyPrint bool
}

func (e XmlEncoder) Encode(obj interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	var err error
	var b []byte

	obj, _ = filter(obj)

	if _, err := buffer.Write([]byte(xml.Header)); err != nil {
		return []byte{}, err
	}

	if e.PrettyPrint {
		b, err = xml.MarshalIndent(obj, "", "    ")
	} else {
		b, err = xml.Marshal(obj)
	}

	if err != nil {
		return []byte{}, err
	}

	buffer.Write(b)
	return buffer.Bytes(), nil
}

func filter(obj interface{}) (interface{}, error) {
	v := reflect.ValueOf(obj)
	k := v.Kind()

	if k == reflect.Ptr && v.IsNil() {
		return nil, nil
	}

	if k == reflect.Array || k == reflect.Slice {
		result := make([]interface{}, 0)

		for i := 0; i < v.Len(); i++ {
			obj := v.Index(i).Interface()
			if f, ok := obj.(Filter); ok {
				result = append(result, f.Filter())
			} else {
				result = append(result, obj)
			}
		}
		return result, nil
	} else {
		if f, ok := obj.(Filter); ok {
			return f.Filter(), nil
		}

		return obj, nil
	}

	return nil, nil
}
