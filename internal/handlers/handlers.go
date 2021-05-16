package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Rha02/bookings/internal/config"
	"github.com/Rha02/bookings/internal/driver"
	"github.com/Rha02/bookings/internal/forms"
	"github.com/Rha02/bookings/internal/helpers"
	"github.com/Rha02/bookings/internal/models"
	"github.com/Rha02/bookings/internal/render"
	"github.com/Rha02/bookings/internal/repository"
	"github.com/Rha02/bookings/internal/repository/dbrepo"
	"github.com/go-chi/chi/v5"
)

//Repo the repository used by the handlers
var Repo *Repository

//Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

//NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

//NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the home page handler
func (m *Repository) Home(rw http.ResponseWriter, r *http.Request) {
	m.DB.AllUsers()
	render.Template(rw, r, "home.page.html", &models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(rw http.ResponseWriter, r *http.Request) {
	render.Template(rw, r, "about.page.html", &models.TemplateData{})
}

// Generals renders the room page
func (m *Repository) Generals(rw http.ResponseWriter, r *http.Request) {
	render.Template(rw, r, "generals.page.html", &models.TemplateData{})
}

// Colonels renders the room page
func (m *Repository) Colonels(rw http.ResponseWriter, r *http.Request) {
	render.Template(rw, r, "colonels.page.html", &models.TemplateData{})
}

// Reservation renders the make a reservation page
func (m *Repository) Reservation(rw http.ResponseWriter, r *http.Request) {
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't find room")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	sd := res.StartDate.Format("01-02-2006")
	ed := res.EndDate.Format("01-02-2006")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(rw, r, "make-reservation.page.html", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse form!")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	layout := "01-02-2006"

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	startDate, err := time.Parse(layout, sd)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse start date")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	endDate, err := time.Parse(layout, ed)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse end date")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid data")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid data")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Phone:     r.Form.Get("phone"),
		Email:     r.Form.Get("email"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    roomID,
		Room:      room,
	}

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		stringMap := make(map[string]string)
		stringMap["start_date"] = r.Form.Get("start_date")
		stringMap["end_date"] = r.Form.Get("end_date")

		render.Template(rw, r, "make-reservation.page.html", &models.TemplateData{
			Form:      form,
			Data:      data,
			StringMap: stringMap,
		})

		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't insert reservation into database")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't insert room restriction")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	htmlMessage := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong><br>
		Dear %s, <br>
		This is to confirm your reservation from %s to %s
	`, reservation.FirstName, reservation.StartDate.Format(layout), reservation.EndDate.Format(layout))

	msg := models.MailData{
		To:       reservation.Email,
		From:     "server@bookings.loc",
		Subject:  "Reservation Confirmation",
		Content:  htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

	htmlMessage = fmt.Sprintf(`
		<strong>Reservation Notification</strong><br>
		A reservation was made for room %s from %s to %s
	`, reservation.Room.RoomName, reservation.StartDate.Format(layout), reservation.EndDate.Format(layout))

	msg = models.MailData{
		To:      "server@bookings.loc",
		From:    "server@bookings.loc",
		Subject: "Reservation Confirmation",
		Content: htmlMessage,
	}

	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(rw, r, "/reservation-summary", http.StatusOK)
}

// Availability renders the search availability page
func (m *Repository) Availability(rw http.ResponseWriter, r *http.Request) {
	render.Template(rw, r, "search-availability.page.html", &models.TemplateData{})
}

// Contact renders the contact page
func (m *Repository) Contact(rw http.ResponseWriter, r *http.Request) {
	render.Template(rw, r, "contact.page.html", &models.TemplateData{})
}

// PostAvailability
func (m *Repository) PostAvailability(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse form!")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "01-02-2006"

	startDate, err := time.Parse(layout, start)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse start date")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	endDate, err := time.Parse(layout, end)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse end date")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't search for availability")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(rw, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(rw, r, "choose-room.page.html", &models.TemplateData{
		Data: data,
	})
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handles request for availability and send JSON response
func (m *Repository) AvailabilityJSON(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Internal Server Error",
		}

		out, _ := json.MarshalIndent(resp, "", "  ")
		rw.Header().Set("Content-Type", "application/json")
		rw.Write(out)
		return
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	layout := "01-02-2006"

	startDate, _ := time.Parse(layout, sd)

	endDate, _ := time.Parse(layout, ed)

	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, err := m.DB.CheckAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Error connecting to database",
		}

		out, _ := json.MarshalIndent(resp, "", "  ")
		rw.Header().Set("Content-Type", "application/json")
		rw.Write(out)
		return
	}

	resp := jsonResponse{
		OK:        available,
		Message:   "Available!",
		RoomID:    strconv.Itoa(roomID),
		StartDate: sd,
		EndDate:   ed,
	}

	out, _ := json.MarshalIndent(resp, "", "     ")

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(out)
}

//ReservationSummary displays the reservation summary page
func (m *Repository) ReservationSummary(rw http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})

	data["reservation"] = reservation

	sd := reservation.StartDate.Format("01-02-2006")
	ed := reservation.EndDate.Format("01-02-2006")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(rw, r, "reservation-summary.page.html", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

//ChooseRoom displays list of available room
func (m *Repository) ChooseRoom(rw http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	roomID, err := strconv.Atoi(exploded[2])
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't get room id from url")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(rw, r, "/make-reservation", http.StatusSeeOther)
}

//BookRoom takes URL parameters, builds sessional data, and takes user to make res screen
func (m *Repository) BookRoom(rw http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't get id from the url query")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	layout := "01-02-2006"

	startDate, _ := time.Parse(layout, sd)

	endDate, _ := time.Parse(layout, ed)

	var res models.Reservation

	room, err := m.DB.GetRoomByID(id)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't get room from the database")
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	res.Room.RoomName = room.RoomName

	res.RoomID = id
	res.StartDate = startDate
	res.EndDate = endDate

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(rw, r, "/make-reservation", http.StatusSeeOther)
}

func (m *Repository) ShowLogin(rw http.ResponseWriter, r *http.Request) {
	render.Template(rw, r, "login.page.html", &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (m *Repository) PostShowLogin(rw http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		render.Template(rw, r, "login.page.html", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(rw, r, "/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully!")

	http.Redirect(rw, r, "/", http.StatusSeeOther)
}

//logout logs a user out
func (m *Repository) Logout(rw http.ResponseWriter, r *http.Request) {
	m.App.Session.Destroy(r.Context())

	m.App.Session.RenewToken(r.Context())

	http.Redirect(rw, r, "/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(rw http.ResponseWriter, r *http.Request) {
	render.Template(rw, r, "admin-dashboard.page.html", &models.TemplateData{})
}

//AdminNewReservations shows all new reservations in admin tool
func (m *Repository) AdminNewReservations(rw http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllNewReservations()
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(rw, r, "admin-new-reservations.page.html", &models.TemplateData{
		Data: data,
	})
}

//AdminAllReservations shows all reservations that were processed
func (m *Repository) AdminAllReservations(rw http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(rw, r, "admin-all-reservations.page.html", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) AdminReservationsCalendar(rw http.ResponseWriter, r *http.Request) {
	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))

		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")
	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	//get the first and last days of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	data["rooms"] = rooms

	for _, x := range rooms {
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for d := firstOfMonth; !d.After(lastOfMonth); d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("01-02-2006")] = 0
			blockMap[d.Format("01-02-2006")] = 0
		}

		restrictions, err := m.DB.GetRestrictionsForRoomByDate(x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(rw, err)
			return
		}

		for _, y := range restrictions {
			if y.ReservationID > 0 {
				for d := y.StartDate; !d.After(y.EndDate); d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("01-02-2006")] = y.ReservationID
				}
			} else {
				blockMap[y.StartDate.Format("01-02-2006")] = y.ID
			}
		}

		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)
	}

	render.Template(rw, r, "admin-reservations-calendar.page.html", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

//AdminShowReservation shows the reservation in the admin tool
func (m *Repository) AdminShowReservation(rw http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	src := exploded[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	stringMap["month"] = month
	stringMap["year"] = year

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(rw, r, "admin-reservation-show.page.html", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
		Form:      forms.New(nil),
	})
}

func (m *Repository) AdminPostShowReservation(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	exploded := strings.Split(r.RequestURI, "/")

	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	src := exploded[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	month := r.Form.Get("month")
	year := r.Form.Get("year")

	m.App.Session.Put(r.Context(), "flash", "Changes saved")

	if year == "" {
		http.Redirect(rw, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(rw, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

//AdminProcessReservation marks a reservation as processed
func (m *Repository) AdminProcessReservation(rw http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	log.Println(id)
	src := chi.URLParam(r, "src")

	log.Println(src)

	err := m.DB.UpdateProcessedForReservation(id, 1)
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "flash", "Reservation marked as processed")

	if year == "" {
		http.Redirect(rw, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(rw, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

func (m *Repository) AdminDeleteReservation(rw http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	err := m.DB.DeleteReservation(id)
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "flash", "Reservation deleted")

	if year == "" {
		http.Redirect(rw, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(rw, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

//AdminPostReservationsCalendar handles post of reservations calendar
func (m *Repository) AdminPostReservationsCalendar(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	rooms, _ := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	form := forms.New(r.PostForm)

	for _, room := range rooms {
		curMap := m.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", room.ID)).(map[string]int)

		for name, value := range curMap {
			if val, ok := curMap[name]; ok {
				if val > 0 {
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", room.ID, name)) {
						err := m.DB.DeleteBlockByID(value)
						if err != nil {
							helpers.ServerError(rw, err)
							return
						}
					}
				}
			}
		}
	}

	for name := range r.PostForm {
		if strings.HasPrefix(name, "add_block") {
			exploded := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(exploded[2])
			t, _ := time.Parse("01-02-2006", exploded[3])

			err := m.DB.InsertBlockForRoom(roomID, t)
			if err != nil {
				helpers.ServerError(rw, err)
				return
			}
		}
	}

	m.App.Session.Put(r.Context(), "flash", "Changes saved")

	http.Redirect(rw, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}
