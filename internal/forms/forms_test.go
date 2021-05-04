package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest("POST", "/anything", nil)

	form := New(r.PostForm)

	isValid := form.Valid()
	if !isValid {
		t.Error("Expected a form to be valid, but got invalid")
	}
}

func TestForm_Required(t *testing.T) {
	postData := url.Values{}

	form := New(postData)

	form.Required("field1", "field2")
	if form.Valid() {
		t.Error("Expected an invalid form, but got valid")
	}

	postData.Add("field1", "kachow")
	postData.Add("field2", "kachow2")

	form = New(postData)

	form.Required("field1", "field2")
	if !form.Valid() {
		t.Error("Expected a valid form, but got invalid")
	}
}

func TestForm_Has(t *testing.T) {
	postData := url.Values{}
	form := New(postData)

	if form.Has("field") {
		t.Error("Expected a form to have a missing field, but the form has that field")
	}

	postData.Add("field", "example field")

	form = New(postData)

	if !form.Has("field") {
		t.Error("Expected the form to have a field, but it's missing")
	}
}

func TestForm_MinLength(t *testing.T) {
	postData := url.Values{}
	postData.Add("field1", "abc")

	form := New(postData)

	form.MinLength("field1", 4)

	if form.Valid() {
		t.Error("Expected the MinLength check to fail, but it passed")
	}

	isError := form.Errors.Get("field1")

	if isError == "" {
		t.Error("Should have an error, but didn't get one")
	}

	postData.Set("field1", "abcde")

	form = New(postData)

	if !form.MinLength("field1", 4) {
		t.Error("Expected the MinLength check to pass, but it failed")
	}

	isError = form.Errors.Get("field1")

	if isError != "" {
		t.Error("Should not have an error, but got one")
	}
}

func TestForm_IsEmail(t *testing.T) {
	postData := url.Values{}
	postData.Add("field1", "invalid email")

	form := New(postData)

	form.IsEmail("field1")

	if form.Valid() {
		t.Error("Expected the form email to be invalid, but got valid")
	}

	postData.Set("field1", "valid@email.loc")

	form = New(postData)

	if !form.Valid() {
		t.Error("Expected the form to be valid, but got invalid")
	}
}
