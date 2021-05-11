package dbrepo

import (
	"errors"
	"log"
	"time"

	"github.com/Rha02/bookings/internal/models"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	if res.RoomID == 2 {
		return 0, errors.New("this is a false error")
	}
	return 1, nil
}

func (m *testDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	if r.RoomID == 1000 {
		return errors.New("some error")
	}
	return nil
}

//SearchAvailability returns true if availability exists for roomID, and false if no availability
func (m *testDBRepo) CheckAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	log.Println(roomID)
	if roomID == 100 {
		log.Println("kachow again")
		return false, errors.New("some error")
	}
	return false, nil
}

//SearchAvailabilityForAllRooms returns a slice of available rooms, if any, for given dates
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room

	invalidDate, _ := time.Parse("01-02-2006", "01-01-3000") //pseudo invalid date

	if start == invalidDate || end == invalidDate {
		return rooms, errors.New("some error")
	}

	unavailDate, _ := time.Parse("01-02-2006", "01-01-2100") //pseudo date where rooms are unavailable

	if start == unavailDate || end == unavailDate {
		return rooms, nil
	}

	room := models.Room{
		ID:       1,
		RoomName: "General's Quarters",
	}

	rooms = append(rooms, room)

	return rooms, nil
}

func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var room models.Room

	if id > 2 {
		return room, errors.New("some error")
	}

	return room, nil
}

func (m *testDBRepo) GetUserByID(id int) (models.User, error) {
	var u models.User

	return u, nil
}

func (m *testDBRepo) UpdateUser(u models.User) error {
	return nil
}

func (m *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	return 1, "", nil
}

func (m *testDBRepo) AllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}

//AllNewReservations returns a slice of all reservations
func (m *testDBRepo) AllNewReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}

func (m *testDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	var res models.Reservation

	return res, nil
}

func (m *testDBRepo) UpdateReservation(r models.Reservation) error {
	return nil
}

func (m *testDBRepo) DeleteReservation(id int) error {
	return nil
}

func (m *testDBRepo) UpdateProcessedForReservation(id, processed int) error {
	return nil
}
