package model

import (
	"time"
)

type Actor struct {
	ID            int64
	URI           string
	Type          string
	Username      string
	Host          string
	DisplayName   string
	Summary       string
	Inbox         string
	SharedInbox   string
	SharedInboxID int64
	Outbox        string
	Followers     string
	Following     string
	PublicKey     PublicKey
	PrivateKey    []byte
	URL           string
	UserID        int64
	Published     time.Time
	Updated       time.Time
}

type KeyType uint8

const (
	RSAKey KeyType = iota
)

type PublicKey struct {
	ID       int64
	URI      string
	OwnerID  int64
	OwnerIRI string
	Type     KeyType
	Pem      []byte
}

type PrivateKey struct {
	ID           int
	OwnerID      int64
	Type         KeyType
	Pem          []byte
	PublicKeyIRI string
}
