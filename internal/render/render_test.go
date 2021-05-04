package render

import (
	"net/http"
	"testing"

	"github.com/Rha02/bookings-app/internal/models"
)

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	session.Put(r.Context(), "flash", "123")

	result := AddDefaultData(&td, r)

	if result.Flash != "123" {
		t.Error("Flash value of 123 not found in session")
	}
}

func TestTemplate(t *testing.T) {
	pathToTemplates = "./../../templates"

	tc, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}

	app.TemplateCache = tc

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	var rw myResponseWriter

	err = Template(&rw, r, "home.page.html", &models.TemplateData{})
	if err != nil {
		t.Error("Error writing template to browser")
	}

	err = Template(&rw, r, "non-existent.page.html", &models.TemplateData{})
	if err == nil {
		t.Error("Rendered a template that does not exist!")
	}

	app.UseCache = true
	err = Template(&rw, r, "home.page.html", &models.TemplateData{})
	if err != nil {
		t.Error("Rendered a template that does not exist!")
	}
}

func getSession() (*http.Request, error) {
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))

	r = r.WithContext(ctx)

	return r, nil
}

func TestNewTemplates(t *testing.T) {
	NewRenderer(app)
}

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "./../../templates"
	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}
}
