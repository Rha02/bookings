package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/Rha02/bookings-app/internal/config"
)

var app *config.AppConfig

//NewHelpers sets up app config for helpers
func NewHelpers(a *config.AppConfig) {
	app = a
}

func ClientError(rw http.ResponseWriter, status int) {
	app.InfoLog.Println("Client error with status of", status)
	http.Error(rw, http.StatusText(status), status)
}

func ServerError(rw http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(rw, http.StatusText(500), 500)
}
