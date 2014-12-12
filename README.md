# encoder [![wercker status](https://app.wercker.com/status/170727695eeb0c8fef3220cd7585c855 "wercker status")](https://app.wercker.com/project/bykey/170727695eeb0c8fef3220cd7585c855)

This is a simple wrapper to the json and xml Marshallers with some filter capabilities. Unlike 'render' package it doesn't write anything, just returns marshalled byte array.

E.g.:

```go
type Some struct {
	Login    string        `json:"login"`
	Password string        `json:"password,omitempty"`
	Avatar   string        `json:"avatar"`
}

// Adding Filter method
func (this Some) Filter() interface{} {
	this.Password = "" // will be omitted in Marshaller
	this.Avatar = "http://some-origin/" + this.Login
	return this
}
```

Filter method will be invoked automatically. If a slice or an array passed, for each `Some` structure in slice.

#### Example

```go
package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	"log"
	"net/http"
	"strconv"
)

type Some struct {
	Login    string `json:"login"`
	Password string `json:"password,omitempty" xml:",omitempty"`
	Url      string `json:"url"`
}

func (this Some) Filter() interface{} {
	this.Password = ""
	this.Url = "http://some-origin/" + this.Login
	return this
}

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	m := martini.New()
	route := martini.NewRouter()

	m.Use(func(c martini.Context, w http.ResponseWriter, r *http.Request) {
		// Use indentations. &pretty=1
		pretty, _ := strconv.ParseBool(r.FormValue("pretty"))
		// Use null instead of empty object for json &null=1
		null, _ := strconv.ParseBool(r.FormValue("null"))
		// Some content negotiation
		switch r.Header.Get("Content-Type") {
		case "application/xml":
			c.MapTo(encoder.XmlEncoder{PrettyPrint: pretty}, (*encoder.Encoder)(nil))
			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		default:
			c.MapTo(encoder.JsonEncoder{PrettyPrint: pretty, PrintNull: null}, (*encoder.Encoder)(nil))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
	})

	route.Get("/user", func(enc encoder.Encoder) (int, []byte) {
		result := Some{"user1", "passwordhash", "/user1"}
		return http.StatusOK, encoder.Must(enc.Encode(result))
	})

	route.Get("/users", func(enc encoder.Encoder) (int, []byte) {
		result := []Some{
			Some{"user1", "somehash", "/user1"},
			Some{"user2", "somehash", "/user2"},
		}

		return http.StatusOK, encoder.Must(enc.Encode(result))
	})

	m.Action(route.Handle)

	log.Println("Waiting for connections...")

	if err := http.ListenAndServe(":8000", m); err != nil {
		log.Fatal(err)
	}
}
```

So, the result will be as follows:

```sh
~ curl -XGET http://localhost:8000/users\?pretty\=1\&null\=1
[
    {
        "login": "user1",
        "url": "http://some-origin/user1"
    },
    {
        "login": "user2",
        "url": "http://some-origin/user2"
    }
]
```

