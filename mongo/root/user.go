package root

import uuid "github.com/satori/go.uuid"

type User struct {
	Id       string    `json:"id"`
	UUID     uuid.UUID `json:"uuid"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	//Name     string `json:"name"`
	//EMail    string `json:"email"`
	//Zipcode  string `json:"zipcode"`
}

type UserService interface {
	CreateUser(u *User) error
	GetByUsername(username string) (*User, error)
}
