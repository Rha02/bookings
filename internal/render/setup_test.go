package render

import (
	"encoding/gob"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Rha02/bookings-app/internal/config"
	"github.com/Rha02/bookings-app/internal/models"
	"github.com/alexedwards/scs/v2"
)

var session *scs.SessionManager

var testApp config.AppConfig

type myResponseWriter struct{}

func (tw *myResponseWriter) Header() http.Header {
	var h http.Header
	return h
}

func (tw *myResponseWriter) WriteHeader(i int) {

}

func (tw *myResponseWriter) Write(b []byte) (int, error) {
	length := len(b)
	return length, nil
}

func TestMain(m *testing.M) {
	// What am I to put in the session:
	gob.Register(models.Reservation{})

	//Change this to true when in production, keep it false when in development
	testApp.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false

	testApp.Session = session

	app = &testApp

	os.Exit(m.Run())
}
