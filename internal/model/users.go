package model

import (
	"time"
)

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

type User struct {
	Id        int64
	Username  string
	Email     string
	Password  []byte
	Verified  bool
	IsAdmin   bool
	CreatedAt time.Time
}
