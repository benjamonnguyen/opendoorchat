package model

import "time"

type User struct {
	Id        string    `json:"id,omitempty"        bson:"_id,omitempty"`
	FirstName string    `json:"firstName,omitempty"`
	LastName  string    `json:"lastName,omitempty"`
	Email     string    `json:"email,omitempty"`
	Username  string    `json:"username,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}
