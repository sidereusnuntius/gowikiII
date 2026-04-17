package model

type HostStatus uint8

const (
	HostUnknown HostStatus = iota
	Fetched
	Peer
	PeerRejected
	Blocked
)

type Host struct {
	ID      int64
	Host    string
	Status  HostStatus
	IsWiki  bool
	ActorID int64
}
