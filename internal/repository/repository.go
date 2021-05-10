package repository

import (
	"time"

	"github.com/Rha02/bookings/internal/models"
)

type DatabaseRepo interface {
	AllUsers() bool
	GetUserByID(id int) (models.User, error)
	UpdateUser(u models.User) error

	Authenticate(email, testPassword string) (int, string, error)

	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(r models.RoomRestriction) error
	CheckAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error)
	SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error)
	GetRoomByID(id int) (models.Room, error)
}
