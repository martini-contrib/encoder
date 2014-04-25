package encoder

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"reflect"
	"testing"
	"time"
)

type (
	sample struct {
		Id         int         `json:"id"`
		Name       string      `json:"name"`
		Password   string      `json:"password,omitempty" out:"false" xml:"-"`
		Registered time.Time   `json:"registered"`
		Ptr        *profile    `json:"profile_ptr"`
		IfStruct   interface{} `json:"profile_as_interface_struct"`
		IfPtr      interface{} `json:"profile_as_interface_ptr"`
		Profile    profile     `json:"profile"`
		Msgs       []message   `json:"messages"`
		// MsgsPtr  []*message  `json:"messages_ptrs"` // not implemented yet
	}

	profile struct {
		FieldVisible     string `json:"field_visible"`
		FieldHiddenValue string `json:"field_hidden_value"     out:"false"  xml:",omitempty"`
		FieldHidden      string `json:"field_hidden,omitempty" out:"false"  xml:"-"`
	}

	message struct {
		FieldVisible     string `json:"field_visible"`
		FieldHiddenValue string `json:"field_hidden_value"     out:"false"  xml:",omitempty"`
		FieldHidden      string `json:"field_hidden,omitempty" out:"false"  xml:"-"`
	}
)

var refJson = `
{
	"id": 1,
    
    "messages": [
		{
			"field_hidden_value": "",
            "field_visible": "123"
        },
        {
            "field_hidden_value": "",
            "field_visible": "345"
        }
    ],

    "name": "foo",

    "registered" : "2006-01-02T15:04:05Z",

    "profile": {
        "field_hidden_value": "",
        "field_visible": "ccc"
    },

    "profile_as_interface_ptr": {
        "field_hidden_value": "",
        "field_visible": "yyy"
    },

    "profile_as_interface_struct": {
        "field_hidden_value": "",
        "field_visible": "vvv"
    },

    "profile_ptr": {
        "field_hidden_value": "",
        "field_visible": ""
    }
}
`
var refXml = `
<?xml version="1.0" encoding="UTF-8"?>
<sample><Id>1</Id><Name>foo</Name><Registered>2006-01-02T15:04:05Z</Registered><Ptr><FieldVisible>xxx</FieldVisible><FieldHiddenValue>zxc</FieldHiddenValue></Ptr><IfStruct><FieldVisible>vvv</FieldVisible></IfStruct><IfPtr><FieldVisible>yyy</FieldVisible></IfPtr><Profile><FieldVisible>ccc</FieldVisible></Profile><Msgs><FieldVisible>123</FieldVisible></Msgs><Msgs><FieldVisible>345</FieldVisible></Msgs></sample>`

var (
	source = sample{
		Id:         1,
		Name:       "foo",
		Password:   "should be hidden and omitted",
		Registered: mustParse("2006-01-02T15:04:05Z"),

		Ptr: &profile{
			FieldVisible:     "xxx",
			FieldHiddenValue: "zxc",
			FieldHidden:      "should be hidden and omitted",
		},

		IfPtr: &profile{
			FieldVisible:     "yyy",
			FieldHiddenValue: "",
			FieldHidden:      "should be hidden and omitted",
		},

		IfStruct: profile{
			FieldVisible:     "vvv",
			FieldHiddenValue: "",
			FieldHidden:      "should be hidden and omitted",
		},
		Profile: profile{
			FieldVisible:     "ccc",
			FieldHiddenValue: "",
			FieldHidden:      "should be hidden and omitted",
		},
		Msgs: []message{
			message{FieldVisible: "123", FieldHiddenValue: "", FieldHidden: "should be hidden and omitted"},
			message{FieldVisible: "345", FieldHiddenValue: "", FieldHidden: "should be hidden and omitted"},
		},
		// MsgsPtr: []*message{
		// 	&message{FieldVisible: "123", FieldHiddenValue: "", FieldHidden: "should be hidden and omitted"},
		// 	&message{FieldVisible: "345", FieldHiddenValue: "", FieldHidden: "should be hidden and omitted"},
		// },
	}
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
		Ref:          refJson,
	},
	{
		Name:         "XmlEncoderTest",
		Encoder:      &XmlEncoder{},
		Unmarshaller: xml.Unmarshal,
		Ref:          refXml,
	},
}

func mustParse(d string) time.Time {
	if result, err := time.Parse(time.RFC3339, d); err != nil {
		panic(err)
	} else {
		return result
	}
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
		result, err := callEncode(reflect.ValueOf(test.Encoder), source)
		if err != nil {
			t.Fatal(err)
		}

		// decode reference json into sample structure
		refSample := &sample{}
		if err = test.Unmarshaller([]byte(test.Ref), refSample); err != nil {
			t.Fatal(test.Name, err)
		}

		// decode result back to struct
		dst := &sample{}
		if err := test.Unmarshaller(result, dst); err != nil {
			t.Fatal("Unmarshal error:", err)
		}

		if !reflect.DeepEqual(refSample, dst) {
			t.Fatalf("%s fails, 'refSample' is not equal to 'dst'\n%v\n%v\n", test.Name, refSample, dst)
		}
	}
}
