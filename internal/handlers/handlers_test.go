package handlers

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	{"non-existent", "/ooga-booga", "GET", http.StatusNotFound},
	{"login", "/login", "GET", http.StatusOK},
	{"logout", "/logout", "GET", http.StatusOK},
	{"dashboard", "/admin/dashboard", "GET", http.StatusOK},
	{"new-res", "/admin/reservations-new", "GET", http.StatusOK},
	{"all-res", "/admin/reservations-all", "GET", http.StatusOK},
	{"show-res", "/admin/reservations/new/1/show", "GET", http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		response, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal()
		}

		if response.StatusCode != e.expectedStatusCode {
			t.Errorf("For %s, expected status %d but got status %d", e.name, e.expectedStatusCode, response.StatusCode)
		}
	}
}

var reservationTests = []struct {
	name               string
	reservation        models.Reservation
	expectedStatusCode int
	expectedHTML       string
	expectedLocation   string
}{
	{
		name: "reservation-in-session",
		reservation: models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		expectedStatusCode: http.StatusOK,
		expectedHTML:       `action="/make-reservation"`,
	},
	{
		name:               "reservation-not-in-session",
		reservation:        models.Reservation{},
		expectedStatusCode: http.StatusSeeOther,
		expectedHTML:       "",
		expectedLocation:   "/",
	},
	{
		name: "non-existent-room",
		reservation: models.Reservation{
			RoomID: 2,
			Room: models.Room{
				ID:       2,
				RoomName: "Non-existent room",
			},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedHTML:       "",
		expectedLocation:   "/",
	},
}

func TestReservation(t *testing.T) {
	for _, test := range reservationTests {
		req, _ := http.NewRequest("GET", "/make-reservation", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		if test.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", test.reservation)
		}

		handler := http.HandlerFunc(Repo.Reservation)

		handler.ServeHTTP(rr, req)

		if rr.Code != test.expectedStatusCode {
			t.Errorf("failed %s: expected status %d, got status %d", test.name, test.expectedStatusCode, rr.Code)
		}

		if test.expectedHTML != "" {
			html := rr.Body.String()
			if !strings.Contains(html, test.expectedHTML) {
				t.Errorf("failed %s: expected to find %s, but didn't", test.name, test.expectedHTML)
			}
		}

		if test.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != test.expectedLocation {
				t.Errorf("failed %s: expected location %s, got location %s", test.name, test.expectedLocation, actualLoc.String())
			}
		}
	}
}

var postReservationTests = []struct {
	name               string
	postData           url.Values
	expectedStatusCode int
	expectedLocation   string
	expectedHTML       string
}{
	{
		name: "valid-post",
		postData: url.Values{
			"start_date": {"01-01-2050"},
			"end_date":   {"01-02-2050"},
			"first_name": {"Joseph"},
			"last_name":  {"Clyde"},
			"email":      {"jclyde@bookings.loc"},
			"phone":      {"123123123"},
			"room_id":    {"1"},
		},
		expectedStatusCode: http.StatusOK,
		expectedLocation:   "/reservation-summary",
	},
	{
		name:               "missing-post-body",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
	{
		name: "invalid-start-date",
		postData: url.Values{
			"start_date": {"invalid"},
			"end_date":   {"01-02-2050"},
			"first_name": {"Joseph"},
			"last_name":  {"Clyde"},
			"email":      {"jclyde@bookings.loc"},
			"phone":      {"123123123"},
			"room_id":    {"1"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
	{
		name: "invalid-end-date",
		postData: url.Values{
			"start_date": {"01-01-2050"},
			"end_date":   {"invalid"},
			"first_name": {"Joseph"},
			"last_name":  {"Clyde"},
			"email":      {"jclyde@bookings.loc"},
			"phone":      {"123123123"},
			"room_id":    {"1"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
	{
		name: "invalid-room-id",
		postData: url.Values{
			"start_date": {"01-01-2050"},
			"end_date":   {"01-02-2050"},
			"first_name": {"Joseph"},
			"last_name":  {"Clyde"},
			"email":      {"jclyde@bookings.loc"},
			"phone":      {"123123123"},
			"room_id":    {"invalid"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
	{
		name: "non-existent-room",
		postData: url.Values{
			"start_date": {"01-01-2050"},
			"end_date":   {"01-02-2050"},
			"first_name": {"Joseph"},
			"last_name":  {"Clyde"},
			"email":      {"jclyde@bookings.loc"},
			"phone":      {"123123123"},
			"room_id":    {"2"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
	{
		name: "invalid-data",
		postData: url.Values{
			"start_date": {"01-01-2050"},
			"end_date":   {"01-02-2050"},
			"first_name": {"J"},
			"last_name":  {"Clyde"},
			"email":      {"jclyde@bookings.loc"},
			"phone":      {"123123123"},
			"room_id":    {"1"},
		},
		expectedStatusCode: http.StatusOK,
		expectedHTML:       `action="/make-reservation"`,
	},
	{
		name: "failed-to-insert-reservation",
		postData: url.Values{
			"start_date": {"01-01-2050"},
			"end_date":   {"01-02-2050"},
			"first_name": {"Joseph"},
			"last_name":  {"Clyde"},
			"email":      {"jclyde@bookings.loc"},
			"phone":      {"123123123"},
			"room_id":    {"3"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
	{
		name: "failed-to-insert-room-restriction",
		postData: url.Values{
			"start_date": {"01-01-2050"},
			"end_date":   {"01-02-2050"},
			"first_name": {"Joseph"},
			"last_name":  {"Clyde"},
			"email":      {"jclyde@bookings.loc"},
			"phone":      {"123123123"},
			"room_id":    {"1000"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
}

func TestPostReservation(t *testing.T) {
	for _, test := range postReservationTests {
		var req *http.Request

		if test.postData != nil {
			req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(test.postData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/make-reservation", nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostReservation)

		handler.ServeHTTP(rr, req)

		if rr.Code != test.expectedStatusCode {
			t.Errorf("failed %s: expected status %d, got status %d", test.name, test.expectedStatusCode, rr.Code)
		}

		if test.expectedHTML != "" {
			html := rr.Body.String()
			if !strings.Contains(html, test.expectedHTML) {
				t.Errorf("failed %s: expected to find %s, but didn't", test.name, test.expectedHTML)
			}
		}

		if test.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != test.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got %s", test.name, test.expectedLocation, actualLoc.String())
			}
		}
	}
}

var postAvailabilityTests = []struct {
	name               string
	postData           url.Values
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "valid-post",
		postData: url.Values{
			"start": {"01-01-2050"},
			"end":   {"01-02-2050"},
		},
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "missing-post-data",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
	{
		name: "invalid-start-date",
		postData: url.Values{
			"start": {""},
			"end":   {"01-02-2050"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
	{
		name: "invalid-end-date",
		postData: url.Values{
			"start": {"01-01-2050"},
			"end":   {""},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
	{
		name: "error-searching-availability",
		postData: url.Values{
			"start": {"01-01-3000"},
			"end":   {"01-02-2050"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
	{
		name: "error-searching-availability",
		postData: url.Values{
			"start": {"01-01-3000"},
			"end":   {"01-02-2050"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
	{
		name: "rooms-not-available",
		postData: url.Values{
			"start": {"01-01-2050"},
			"end":   {"01-01-2100"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/search-availability",
	},
}

func TestPostAvailability(t *testing.T) {
	for _, e := range postAvailabilityTests {
		var req *http.Request

		if e.postData != nil {
			req, _ = http.NewRequest("POST", "/post-availability", strings.NewReader(e.postData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/post-availability", nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostAvailability)

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected status %d, got status %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if e.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}
	}
}

var reservationSummaryTests = []struct {
	name               string
	reservation        models.Reservation
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "reservation-in-session",
		reservation: models.Reservation{
			ID:     1,
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "reservation-not-in-session",
		reservation:        models.Reservation{},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
}

func TestReservationSummary(t *testing.T) {
	for _, e := range reservationSummaryTests {
		req, _ := http.NewRequest("GET", "/reservation-summary", nil)

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		handler := http.HandlerFunc(Repo.ReservationSummary)

		if e.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", e.reservation)
		}

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected status %d, got status %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if e.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}
	}
}

var chooseRoomTests = []struct {
	name               string
	reservation        models.Reservation
	url                string
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "valid-reservation-in-session",
		reservation: models.Reservation{
			ID:     1,
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		url:                "/choose-room/1",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/make-reservation",
	},
	{
		name: "invalid-room-id-url",
		reservation: models.Reservation{
			ID:     1,
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		url:                "/choose-room/invalid",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
	{
		name:               "reservation-not-in-session",
		reservation:        models.Reservation{},
		url:                "/choose-room/1",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
}

func TestChooseRoom(t *testing.T) {
	for _, e := range chooseRoomTests {
		req, _ := http.NewRequest("GET", e.url, nil)

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.RequestURI = e.url

		if e.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", e.reservation)
		}

		handler := http.HandlerFunc(Repo.ChooseRoom)

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected status %d, got status %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if e.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}
	}
}

var bookRoomTests = []struct {
	name               string
	urlQuery           string
	reservation        models.Reservation
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name:               "success-booking-room",
		urlQuery:           "/book-room?s=01-01-2050&e=01-02-2050&id=1",
		reservation:        models.Reservation{},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/make-reservation",
	},
	{
		name:               "invalid-url-query",
		urlQuery:           "/book-room",
		reservation:        models.Reservation{},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
	{
		name:               "invalid-room-id",
		urlQuery:           "/book-room?s=01-01-2050&e=01-02-2050&id=2",
		reservation:        models.Reservation{},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
}

func TestBookRoom(t *testing.T) {
	for _, e := range bookRoomTests {
		req, _ := http.NewRequest("GET", e.urlQuery, nil)

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		session.Put(ctx, "reservation", e.reservation)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.BookRoom)

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected %d, got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if e.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}
	}
}

var loginTests = []struct {
	name               string
	email              string
	expectedStatusCode int
	expectedHTML       string
	expectedLocation   string
}{
	{
		"valid-credentials",
		"test@test.loc",
		http.StatusSeeOther,
		"",
		"/",
	},
	{
		"invalid-credentials",
		"invalid@invalid.loc",
		http.StatusSeeOther,
		"",
		"/login",
	},
	{
		"invalid-data",
		"invalid",
		http.StatusOK,
		`action="/login"`,
		"",
	},
}

func TestLogin(t *testing.T) {
	for _, test := range loginTests {
		postedData := url.Values{}
		postedData.Add("email", test.email)
		postedData.Add("password", "password")

		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostShowLogin)

		handler.ServeHTTP(rr, req)

		if rr.Code != test.expectedStatusCode {
			t.Errorf("failed %s: expected code %d, but got %d", test.name, test.expectedStatusCode, rr.Code)
		}

		if test.expectedLocation != "" {
			actualLocation, _ := rr.Result().Location()
			if actualLocation.String() != test.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got location %s", test.name, test.expectedLocation, actualLocation.String())
			}
		}

		if test.expectedHTML != "" {
			html := rr.Body.String()
			if !strings.Contains(html, test.expectedHTML) {
				t.Errorf("failed %s: expected to find %s but did not", test.name, test.expectedHTML)
			}
		}
	}
}

var adminResCalendarTests = []struct {
	name               string
	urlQuery           string
	expectedStatusCode int
	expectedHTML       string
}{
	{
		name:               "valid-url-query",
		urlQuery:           "/admin/reservations-calendar?y=2021&m=5",
		expectedStatusCode: http.StatusOK,
		expectedHTML:       `action="/admin/reservations-calendar"`,
	},
}

func TestAdminReservationsCalendar(t *testing.T) {
	for _, e := range adminResCalendarTests {
		req, _ := http.NewRequest("GET", e.urlQuery, nil)

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		handler := http.HandlerFunc(Repo.AdminReservationsCalendar)

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected status %d, got status %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if e.expectedHTML != "" {
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s, but didn't", e.name, e.expectedHTML)
			}
		}
	}
}

var adminPostShowReservationTests = []struct {
	name               string
	uri                string
	postData           url.Values
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "valid-post-from-res-calendar",
		uri:  "/admin/reservations/cal/1/show",
		postData: url.Values{
			"first_name": {"Joseph"},
			"last_name":  {"Clyde"},
			"email":      {"joseph@clyde.com"},
			"phone":      {"1234567890"},
			"month":      {"7"},
			"year":       {"2002"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/admin/reservations-calendar?y=2002&m=7",
	},
	{
		name: "valid-post-from-other-src",
		uri:  "/admin/reservations/all/1/show",
		postData: url.Values{
			"first_name": {"Joseph"},
			"last_name":  {"Clyde"},
			"email":      {"joseph@clyde.com"},
			"phone":      {"1234567890"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/admin/reservations-all",
	},
}

func TestAdminPostShowReservation(t *testing.T) {
	for _, e := range adminPostShowReservationTests {
		var req *http.Request

		if e.postData != nil {
			req, _ = http.NewRequest("POST", "/login", strings.NewReader(e.postData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/login", nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.RequestURI = e.uri

		handler := http.HandlerFunc(Repo.AdminPostShowReservation)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected status %d, got status %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if e.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}
	}
}

var adminPostReservationsCalendarTests = []struct {
	name               string
	uri                string
	postData           url.Values
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "valid-post",
		uri:  "/admin/reservations-calendar?y=2030&m=7",
		postData: url.Values{
			"add_block_2_7-25-2030":    {"1"},
			"remove_block_1_7-20-2030": {"1"},
			"m":                        {"7"},
			"y":                        {"2030"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/admin/reservations-calendar?y=2030&m=7",
	},
}

func TestAdminPostReservationsCalendar(t *testing.T) {
	for _, e := range adminPostReservationsCalendarTests {
		var req *http.Request

		if e.postData != nil {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", strings.NewReader(e.postData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		blockMap := make(map[string]int)
		blockMap["7-30-2030"] = 1
		blockMap["7-20-2030"] = 1

		session.Put(ctx, "block_map_1", blockMap)

		req.RequestURI = e.uri

		handler := http.HandlerFunc(Repo.AdminPostReservationsCalendar)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected status %d, got status %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if e.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}
	}
}

// var adminProcessResTests = []struct {
// 	name               string
// 	query              string
// 	expectedStatusCode int
// 	expectedLocation   string
// }{
// 	{
// 		name:               "process-res-from-cal",
// 		query:              "?y=2050&m=7",
// 		expectedStatusCode: http.StatusSeeOther,
// 		expectedLocation:   "/admin/reservations-calendar?y=2050&m=7",
// 	},
// 	{
// 		name:               "process-res-from-other-src",
// 		query:              "",
// 		expectedStatusCode: http.StatusSeeOther,
// 		expectedLocation:   "/admin/reservations-new",
// 	},
// }

// func TestAdminProcessReservation(t *testing.T) {
// 	for _, e := range adminProcessResTests {
// 		req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/cal/1/do%s", e.query), nil)

// 		ctx := getCtx(req)
// 		req = req.WithContext(ctx)

// 		handler := http.HandlerFunc(Repo.AdminProcessReservation)

// 		rr := httptest.NewRecorder()

// 		handler.ServeHTTP(rr, req)

// 		if rr.Code != e.expectedStatusCode {
// 			t.Errorf("failed %s: expected status %d, got status %d", e.name, e.expectedStatusCode, rr.Code)
// 		}

// 		if e.expectedLocation != "" {
// 			actualLoc, _ := rr.Result().Location()

// 			if actualLoc.String() != e.expectedLocation {
// 				t.Errorf("failed %s: expected location %s, got location %s", e.name, e.expectedLocation, actualLoc.String())
// 			}
// 		}
// 	}
// }

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
