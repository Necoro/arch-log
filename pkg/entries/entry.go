package entries

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("package could not be found remotely")

type Entry struct {
	CommitTime time.Time
	Summary    string
	Message    string
	Author     string
}
