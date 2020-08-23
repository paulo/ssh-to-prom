package main

import (
	"net"
	"time"
)

// FailedConnEvent represents a reportable failed connection ocurrence
type FailedConnEvent struct {
	Username  string
	Timestamp time.Time
	IPAddress net.IP
	Port      int
	Country   string
}
