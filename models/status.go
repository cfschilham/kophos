package models

import "fmt"

type Status int
const (
	Valid Status = iota + 1
	Invalid
)

func (s Status) String() string {
	seasons := [...]string{"valid", "invalid"}
	if s < Valid || s > Invalid {
		return fmt.Sprintf("Status(%d)", int(s))
	}
	return seasons[s-1]
}
