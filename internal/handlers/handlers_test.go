package handlers

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	params             []postData
	expectedStatusCode int
}{
	{"home", "/", "GET", []postData{}, 200},
	{"about", "/about", "GET", []postData{}, 200},
	{"gq", "/generals-quarters", "GET", []postData{}, 200},
	{"cs", "/colonels-suite", "GET", []postData{}, 200},
	{"sa", "/search-availability", "GET", []postData{}, 200},
	{"contact", "/contact", "GET", []postData{}, 200},
	{"mr", "/make-reservation", "GET", []postData{}, 200},
	{"post-search-avail", "/search-availability", "POST", []postData{
		{key: "start", value: "16-07-2002"},
		{key: "end", value: "2020-01-02"},
	}, 200},
	{"post-search-avail-json", "/search-availability-json", "POST", []postData{
		{key: "start", value: "16-07-2002"},
		{key: "end", value: "2020-01-02"},
	}, 200},
	{"post-make-reservation", "/make-reservation", "POST", []postData{
		{key: "first_name", value: "Bruce"},
		{key: "last_name", value: "Wayne"},
		{key: "email", value: "bwayne@batman.loc"},
		{key: "phone", value: "999-999-9999"},
	}, 200},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, test := range theTests {
		if test.method == "GET" {
			response, err := ts.Client().Get(ts.URL + test.url)
			if err != nil {
				t.Log(err)
				t.Fatal()
			}

			if response.StatusCode != test.expectedStatusCode {
				t.Errorf("For %s, expected status %d but got status %d", test.name, test.expectedStatusCode, response.StatusCode)
			}
		} else {
			values := url.Values{}
			for _, v := range test.params {
				values.Add(v.key, v.value)
			}
			response, err := ts.Client().PostForm(ts.URL+test.url, values)
			if err != nil {
				t.Log(err)
				t.Fatal()
			}
			if response.StatusCode != test.expectedStatusCode {
				t.Errorf("For %s, expected status %d but got status %d", test.name, test.expectedStatusCode, response.StatusCode)
			}
		}
	}
}
