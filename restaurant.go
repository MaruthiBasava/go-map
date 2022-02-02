package main

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type RestaurantThumbnailInput struct {
	ThumbnailURL string
	Position     int
}

type RestaurantThumbnail struct {
	restaurantID uuid.UUID
	thumbnailID  uuid.UUID
	thumbnailURL string
	position     int
	dateCreated  int64
	dateUpdated  int64
}

func (r RestaurantThumbnail) RestaurantID() uuid.UUID {
	return r.restaurantID
}

func (r RestaurantThumbnail) ThumbnailID() uuid.UUID {
	return r.thumbnailID
}

func (r RestaurantThumbnail) ThumbnailURL() string {
	return r.thumbnailURL
}

func (r RestaurantThumbnail) Position() int {
	return r.position
}

func (r RestaurantThumbnail) DateCreated() int64 {
	return r.dateCreated
}

func (r RestaurantThumbnail) DateUpdated() int64 {
	return r.dateUpdated
}

type RestaurantInput struct {
	RestaurantOwnerID uuid.UUID
	PhoneNumber       string
	Street            string
	City              string
	State             string
	Zipcode           int
	Restaurant        string
	Thumbnails        []RestaurantThumbnailInput
}

type RestaurantUpdate struct {
	PhoneNumber         string
	Restaurant          string
	NewThumbnails       []RestaurantThumbnailInput
	RemovedThumbnailIDs []uuid.UUID
}

type Restaurant struct {
	restaurantID        uuid.UUID
	restaurantOwnerID   uuid.UUID
	phoneNumber         string
	street              string
	city                string
	state               string
	zipcode             int
	restaurant          string
	thumbnails          []*RestaurantThumbnail
	removedThumbnailIDs []uuid.UUID
	dateCreated         int64
	dateUpdated         int64
}

func NewRestaurant(r RestaurantInput) *Restaurant {

	restID := uuid.NewV4()

	rest := &Restaurant{
		restaurantID:        restID,
		restaurantOwnerID:   r.RestaurantOwnerID,
		phoneNumber:         r.PhoneNumber,
		street:              r.Street,
		city:                r.City,
		state:               r.State,
		zipcode:             r.Zipcode,
		restaurant:          r.Restaurant,
		thumbnails:          make([]*RestaurantThumbnail, 0),
		removedThumbnailIDs: make([]uuid.UUID, 0),
	}

	rest.AddThumbnails(r.Thumbnails)

	return rest
}

func (r Restaurant) RestaurantID() uuid.UUID {
	return r.restaurantID
}

func (r Restaurant) RestaurantOwnerID() uuid.UUID {
	return r.restaurantOwnerID
}

func (r Restaurant) PhoneNumber() string {
	return r.phoneNumber
}

func (r Restaurant) Street() string {
	return r.street
}

func (r Restaurant) City() string {
	return r.city
}

func (r Restaurant) State() string {
	return r.state
}

func (r Restaurant) Zipcode() int {
	return r.zipcode
}

func (r Restaurant) Restaurant() string {
	return r.restaurant
}

func (r Restaurant) Thumbnails() []*RestaurantThumbnail {
	return r.thumbnails
}

func (r Restaurant) DateCreated() int64 {
	return r.dateCreated
}

func (r Restaurant) DateUpdated() int64 {
	return r.dateUpdated
}

func (r Restaurant) RemovedThumbnailIDs() []uuid.UUID {
	return r.removedThumbnailIDs
}

func (r *Restaurant) ChangePhoneNumberTo(pn string) error {
	r.phoneNumber = pn
	r.dateUpdated = time.Now().Unix()
	return nil
}

func (r *Restaurant) ChangeRestaurantNameTo(rest string) error {
	r.restaurant = rest
	r.dateUpdated = time.Now().Unix()

	return nil
}

func (r *Restaurant) pairThumbnail(tbi RestaurantThumbnailInput) *RestaurantThumbnail {

	rtuuid := uuid.NewV4()

	dateCreated := time.Now().Unix()

	rt := &RestaurantThumbnail{
		restaurantID: r.restaurantID,
		thumbnailID:  rtuuid,
		thumbnailURL: tbi.ThumbnailURL,
		position:     tbi.Position,
		dateCreated:  dateCreated,
		dateUpdated:  dateCreated,
	}

	return rt
}

func (r *Restaurant) AddThumbnails(tbis []RestaurantThumbnailInput) {

	tmbs := make([]*RestaurantThumbnail, len(tbis))

	for i, tbi := range tbis {
		tmb := r.pairThumbnail(tbi)
		tmbs[i] = tmb
	}

	r.thumbnails = append(r.thumbnails, tmbs...)
	r.dateUpdated = time.Now().Unix()

}

func (r *Restaurant) RemoveThumbnails(tmbids []uuid.UUID) {

	tmbs := r.thumbnails
	var newtmbs []*RestaurantThumbnail

	for _, tmb := range tmbs {
		for _, tmbid := range tmbids {
			if tmb.thumbnailID != tmbid {
				newtmbs = append(newtmbs, tmb)
			}
		}
	}

	r.thumbnails = newtmbs
	r.removedThumbnailIDs = tmbids
	r.dateUpdated = time.Now().Unix()
}

func (r *Restaurant) ReorderThumbnails(tbis []*RestaurantThumbnail) {

	for _, tmb := range r.thumbnails {
		for _, tbi := range tbis {
			if tmb.thumbnailID == tbi.thumbnailID {
				tmb.position = tbi.position
				tmb.dateUpdated = time.Now().Unix()
			}
		}
	}

	r.dateUpdated = time.Now().Unix()

}

// func (r *Restaurant) PairNewMenuCatgory(categoryName string) *MenuCategory {
// 	return newMenuCategory(r.restaurantID, categoryName)
// }

// func (r *Restaurant) PairNewMenuLayout() *MenuLayout {
// 	return MenuLayoutForRestaurant(r.restaurantID)
// }
