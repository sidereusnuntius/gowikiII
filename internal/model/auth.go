package model

import "time"

type Session struct {
	Token      string
	Expiration time.Time
	Created    time.Time
	User       User
	ActorID    string
}

type LoginInput struct {
	Email    string
	Password string
}
