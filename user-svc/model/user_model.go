package model

import (
	"fmt"
	"time"
)

type User struct {
	Id        string    `json:"id,omitempty"        bson:"_id,omitempty"`
	FirstName string    `json:"firstName,omitempty" bson:"firstName,omitempty"`
	LastName  string    `json:"lastName,omitempty"  bson:"lastName,omitempty"`
	Email     string    `json:"email,omitempty"     bson:"email,omitempty"`
	Username  string    `json:"username,omitempty"  bson:"username,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

func (u User) Name() string {
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}
