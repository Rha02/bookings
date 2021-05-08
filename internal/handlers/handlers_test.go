package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Rha02/bookings/internal/models"
)

// type postData struct {
// 	key   string
// 	value string
// }

var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET", 200},
	{"about", "/about", "GET", 200},
	{"gq", "/generals-quarters", "GET", 200},
	{"cs", "/colonels-suite", "GET", 200},
	{"sa", "/search-availability", "GET", 200},
	{"contact", "/contact", "GET", 200},
	// {"mr", "/make-reservation", "GET", []postData{}, 200},
	// {"post-search-avail", "/search-availability", "POST", []postData{
	// 	{key: "start", value: "16-07-2002"},
	// 	{key: "end", value: "2020-01-02"},
	// }, 200},
	// {"post-search-avail-json", "/search-availability-json", "POST", []postData{
	// 	{key: "start", value: "16-07-2002"},
	// 	{key: "end", value: "2020-01-02"},
	// }, 200},
	// {"post-make-reservation", "/make-reservation", "POST", []postData{
	// 	{key: "first_name", value: "Bruce"},
	// 	{key: "last_name", value: "Wayne"},
	// 	{key: "email", value: "bwayne@batman.loc"},
	// 	{key: "phone", value: "999-999-9999"},
	// }, 200},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, test := range theTests {
		response, err := ts.Client().Get(ts.URL + test.url)
		if err != nil {
			t.Log(err)
			t.Fatal()
		}

		if response.StatusCode != test.expectedStatusCode {
			t.Errorf("For %s, expected status %d but got status %d", test.name, test.expectedStatusCode, response.StatusCode)
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.Reservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Reservation handlers returned wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	//test case where reservation is not in session
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handlers returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test case where room does not exist
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	reservation.RoomID = 100

	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handlers returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_PostReservation(t *testing.T) {
	reqBody := "start_date=01-01-2050"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=01-02-2050")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=Joseph")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Clyde")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=jclyde@book.loc")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=1231234")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handlers returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	//test for missing post body
	req, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handlers returned wrong response code for missing post body: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test for invalid start date
	reqBody = "start_date=invalid"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=01-02-2050")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=Joseph")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Clyde")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=jclyde@book.loc")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=1231234")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handlers returned wrong response code for invalid start date: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test for invalid end date
	reqBody = "start_date=01-01-2050"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=invalid")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=Joseph")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Clyde")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=jclyde@book.loc")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=1231234")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handlers returned wrong response code for invalid end date: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test for invalid room id
	reqBody = "start_date=01-01-2050"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=01-02-2050")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=Joseph")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Clyde")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=jclyde@book.loc")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=1231234")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=invalid")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handlers returned wrong response code for invalid room id: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test for invalid data
	reqBody = "start_date=01-01-2050"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=01-02-2050")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=J")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Clyde")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=jclyde@book.loc")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=1231234")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handlers returned wrong response code for invalid data: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	//test for failure to insert reservation into database
	reqBody = "start_date=01-01-2050"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=01-02-2050")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=Joseph")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Clyde")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=jclyde@book.loc")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=1231234")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=2")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler failed to fail inserting reservation: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test for failure to insert room restriction into database
	reqBody = "start_date=01-01-2050"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=01-02-2050")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=Joseph")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Clyde")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=jclyde@book.loc")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=1231234")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1000")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler failed to fail inserting reservation: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_PostAvailability(t *testing.T) {
	reqBody := "start=01-01-2050"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=01-02-2050")

	req, _ := http.NewRequest("POST", "/post-availability", strings.NewReader(reqBody))

	ctx := getCtx(req)

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostAvailability)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Got wrong status code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	//test case where body is null
	req, _ = http.NewRequest("POST", "/post-availability", nil)

	ctx = getCtx(req)

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostAvailability)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Got wrong status code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test case where start date is invalid
	reqBody = "start="
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=01-02-2050")

	req, _ = http.NewRequest("POST", "/post-availability", strings.NewReader(reqBody))

	ctx = getCtx(req)

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostAvailability)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Got wrong status code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test case where end date is invalid
	reqBody = "start=01-01-2050"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=")

	req, _ = http.NewRequest("POST", "/post-availability", strings.NewReader(reqBody))

	ctx = getCtx(req)

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostAvailability)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Got wrong status code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test case where error occurs checking room availability
	reqBody = "start=01-01-3000"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=01-02-2050")

	req, _ = http.NewRequest("POST", "/post-availability", strings.NewReader(reqBody))

	ctx = getCtx(req)

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostAvailability)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Got wrong status code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test case where no rooms are available
	reqBody = "start=01-01-2050"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=01-01-2100")

	req, _ = http.NewRequest("POST", "/post-availability", strings.NewReader(reqBody))

	ctx = getCtx(req)

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostAvailability)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Got wrong status code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepository_ReservationSummary(t *testing.T) {
	req, _ := http.NewRequest("GET", "/reservation-summary", nil)

	ctx := getCtx(req)
	req = req.WithContext(ctx)

	handler := http.HandlerFunc(Repo.ReservationSummary)

	reservation := models.Reservation{
		ID:     1,
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	session.Put(ctx, "reservation", reservation)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Wrong status code: expected %d, got %d", http.StatusOK, rr.Code)
	}

	//test case where can't get reservation out of session
	req, _ = http.NewRequest("GET", "/reservation-summary", nil)

	ctx = getCtx(req)
	req = req.WithContext(ctx)

	handler = http.HandlerFunc(Repo.ReservationSummary)

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Wrong status code: expected %d, got %d", http.StatusTemporaryRedirect, rr.Code)
	}
}

func TestRepository_ChooseRoom(t *testing.T) {
	req, _ := http.NewRequest("GET", "/choose-room/1", nil)

	ctx := getCtx(req)
	req = req.WithContext(ctx)

	req.RequestURI = "/choose-room/1"

	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.ChooseRoom)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Wrong status code: expected %d, got %d", http.StatusSeeOther, rr.Code)
	}

	//test case where room id is invalid
	req, _ = http.NewRequest("GET", "/choose-room/invalid", nil)

	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.RequestURI = "/choose-room/invalid"

	reservation = models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.ChooseRoom)

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Wrong status code: expected %d, got %d", http.StatusTemporaryRedirect, rr.Code)
	}

	//test case where reservation is not in session
	req, _ = http.NewRequest("GET", "/choose-room/1", nil)

	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.RequestURI = "/choose-room/1"

	handler = http.HandlerFunc(Repo.ChooseRoom)

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Wrong status code: expected %d, got %d", http.StatusTemporaryRedirect, rr.Code)
	}
}

func TestRepository_BookRoom(t *testing.T) {
	req, _ := http.NewRequest("GET", "/book-room?s=01-01-2050&e=01-02-2050&id=1", nil)

	ctx := getCtx(req)
	req = req.WithContext(ctx)

	reservation := models.Reservation{
		ID: 1,
	}

	session.Put(ctx, "reservation", reservation)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.BookRoom)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Wrong status: expected %d, got %d", http.StatusSeeOther, rr.Code)
	}

	//test case where url query is invalid
	req, _ = http.NewRequest("GET", "/book-room", nil)

	ctx = getCtx(req)
	req = req.WithContext(ctx)

	reservation = models.Reservation{
		ID: 1,
	}

	session.Put(ctx, "reservation", reservation)

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.BookRoom)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Wrong status: expected %d, got %d", http.StatusTemporaryRedirect, rr.Code)
	}

	//test case where can't get room from database
	req, _ = http.NewRequest("GET", "/book-room?s=01-01-2050&e=01-02-2050&id=100", nil)

	ctx = getCtx(req)
	req = req.WithContext(ctx)

	reservation = models.Reservation{
		ID: 1,
	}

	session.Put(ctx, "reservation", reservation)

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.BookRoom)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Wrong status: expected %d, got %d", http.StatusTemporaryRedirect, rr.Code)
	}
}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}

// func TestRepository_AvailabilityJSON(t *testing.T) {
// 	reqBody := "start=01-01-2050"
// 	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=01-02-2050")
// 	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")

// 	req, err := http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))
// 	if err != nil {
// 		t.Error("Error trying to create a request")
// 	}

// 	ctx := getCtx(req)
// 	req = req.WithContext(ctx)

// 	req.Header.Set("Content-Type", "x-www-form-urlencoded")

// 	handler := http.HandlerFunc(Repo.AvailabilityJSON)

// 	rr := httptest.NewRecorder()

// 	handler.ServeHTTP(rr, req)

// 	var j jsonResponse

// 	err = json.Unmarshal([]byte(rr.Body.Bytes()), &j)
// 	if err != nil {
// 		t.Error("Failed to parse json")
// 	}

// 	if j.OK != true {
// 		t.Error("Expected the test to pass, but it didn't")
// 	}

// 	//test for missing body
// 	req, _ = http.NewRequest("POST", "/search-availability-json", nil)

// 	ctx = getCtx(req)
// 	req = req.WithContext(ctx)

// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

// 	handler = http.HandlerFunc(Repo.AvailabilityJSON)

// 	rr = httptest.NewRecorder()

// 	handler.ServeHTTP(rr, req)

// 	err = json.Unmarshal([]byte(rr.Body.Bytes()), &j)
// 	if err != nil {
// 		t.Error("Failed to parse form")
// 	}

// 	if j.OK != false {
// 		t.Error("Expected to fail parsing form, instead it passed")
// 	}

// 	//test for error checking availability
// 	reqBody = "start=01-01-2050"
// 	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=01-02-2050")
// 	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=100")

// 	log.Println(reqBody)

// 	req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))

// 	ctx = getCtx(req)
// 	req = req.WithContext(ctx)

// 	req.Header.Set("Content-Type", "x-www-form-urlencoded")

// 	handler = http.HandlerFunc(Repo.AvailabilityJSON)

// 	rr = httptest.NewRecorder()

// 	handler.ServeHTTP(rr, req)

// 	err = json.Unmarshal([]byte(rr.Body.Bytes()), &j)
// 	if err != nil {
// 		t.Error("Failed to parse form")
// 	}

// 	if j.OK != false {
// 		t.Error("Expected to fail searching for availability, but it didn't")
// 	}
// }
