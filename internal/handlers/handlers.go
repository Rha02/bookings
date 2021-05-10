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
	"github.com/Rha02/bookings/internal/models"
	"github.com/Rha02/bookings/internal/render"
	"github.com/Rha02/bookings/internal/repository"
	"github.com/Rha02/bookings/internal/repository/dbrepo"
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
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't find room")
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
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
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	layout := "01-02-2006"

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	startDate, err := time.Parse(layout, sd)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse start date")
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, ed)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse end date")
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid data")
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
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
	}

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		http.Error(rw, "invalid form data", http.StatusSeeOther)
		render.Template(rw, r, "make-reservation.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})

		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't insert reservation into database")
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
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
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
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

	http.Redirect(rw, r, "/reservation-summary", http.StatusSeeOther)
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
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "01-02-2006"

	startDate, err := time.Parse(layout, start)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse start date")
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, end)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse end date")
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse start date")
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
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
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
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
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
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
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
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
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
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
		log.Println(err)
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
