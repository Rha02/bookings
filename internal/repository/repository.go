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

	AllReservations() ([]models.Reservation, error)
	AllNewReservations() ([]models.Reservation, error)
	GetReservationByID(id int) (models.Reservation, error)
	UpdateReservation(r models.Reservation) error
	DeleteReservation(id int) error
	UpdateProcessedForReservation(id, processed int) error

	AllRooms() ([]models.Room, error)

	GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error)

	InsertBlockForRoom(id int, startDate time.Time) error
	DeleteBlockByID(id int) error

	InsertRoomRestriction(r models.RoomRestriction) error
	CheckAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error)
	SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error)
	GetRoomByID(id int) (models.Room, error)
}
