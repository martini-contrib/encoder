package encoder

import (
	"encoding/json"
	// "encoding/xml"
	"errors"
	"reflect"
	"testing"
)

type TestCase struct {
	Name         string
	Encoder      interface{}
	Unmarshaller func([]byte, interface{}) error
	Ref          string
}

var Cases = []TestCase{
	{
		Name:         "JsonEncoderTest",
		Encoder:      &JsonEncoder{},
		Unmarshaller: json.Unmarshal,
	},
	// TODO.
	// Problem with slice, it needs a root element for xml.
	//
	// {
	// 	Name:         "XmlEncoderTest",
	// 	Encoder:      &XmlEncoder{},
	// 	Unmarshaller: xml.Unmarshal,
	// },
}

type Profile struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password,omitempty" xml:",omitempty"`
	Avatar   string `json:"avatar"`
}

func (this Profile) Filter() interface{} {
	this.Password = ""
	this.Avatar = "//origin/" + this.Avatar
	return this
}

type testData struct {
	name     string
	src      interface{}
	expected interface{}
}

var data = []testData{
	{
		name:     "byVal",
		src:      Profile{Id: "1", Name: "Buster", Password: "hideme", Avatar: "xxx"},
		expected: Profile{Id: "1", Name: "Buster", Password: "", Avatar: "//origin/xxx"},
	},
	{
		name:     "byPtr",
		src:      &Profile{Id: "1", Name: "Buster", Password: "hideme", Avatar: "xxx"},
		expected: Profile{Id: "1", Name: "Buster", Password: "", Avatar: "//origin/xxx"},
	},
	{
		name: "slice",
		src: []Profile{
			{Id: "1", Name: "Buster", Password: "hideme", Avatar: "xxx"},
			{Id: "2", Name: "SomeOne", Password: "hideme", Avatar: "yyy"},
		},
		expected: []Profile{
			{Id: "1", Name: "Buster", Password: "", Avatar: "//origin/xxx"},
			{Id: "2", Name: "SomeOne", Password: "", Avatar: "//origin/yyy"},
		},
	},
}

func callEncode(i reflect.Value, arg interface{}) (data []byte, err error) {
	method := i.MethodByName("Encode")

	if !method.IsValid() {
		return nil, errors.New("method 'Encode' not found")
	}

	result := method.Call([]reflect.Value{reflect.ValueOf(arg)})
	if len(result) != 2 {
		return nil, errors.New("unexpected emelents count in result")
	}

	if result[0].Interface() != nil {
		data = result[0].Interface().([]byte)
	}

	if result[1].Interface() != nil {
		err = result[1].Interface().(error)
	}

	return data, err
}

func TestAll(t *testing.T) {
	for _, test := range Cases {
		for _, d := range data {
			result, err := callEncode(reflect.ValueOf(test.Encoder), d.src)
			if err != nil {
				t.Fatal(err)
			}

			if reflect.TypeOf(d.src).Kind() == reflect.Slice {
				dst := make([]Profile, 0)
				if err := test.Unmarshaller(result, &dst); err != nil {
					t.Fatal("Test:", test.Name, d.name, "raised Unmarshaling error:", err)
				}

				if !reflect.DeepEqual(dst, d.expected) {
					t.Fatalf("%s:%s fails!\nExpected:%v\nGot:    %v\n", test.Name, d.name, d.expected, dst)
				}
			} else {
				dst := Profile{}
				if err := test.Unmarshaller(result, &dst); err != nil {
					t.Fatal("Test:", test.Name, d.name, "raised Unmarshaling error:", err)
				}

				if !reflect.DeepEqual(dst, d.expected) {
					t.Fatalf("%s:%s fails!\nExpected:%v\nGot:    %v\n", test.Name, d.name, d.expected, dst)
				}
			}
		}
	}
}
