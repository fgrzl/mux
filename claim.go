package mux

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type Claim interface {
	// Returns the name of the claim (e.g., "sub", "iss", "aud").
	Name() string

	// Returns the value of the claim as an string.
	Value() string

	// Returns the values of the claim as an []string.
	Values(sep string) []string

	// Returns the value of the claim as an int, if applicable.
	IntValue() (int, bool)

	// Returns the value of the claim as an int32, if applicable.
	Int32Value() (int32, bool)

	// Returns the value of the claim as an int64, if applicable.
	Int64Value() (int64, bool)

	// Returns the value of the claim as a bool, if applicable.
	BoolValue() (bool, bool)

	// Returns the value of the claim as a bool, if applicable.
	UUIDValue() (uuid.UUID, bool)
}

func NewClaim(name, value string) Claim {
	return &claim{
		name:  name,
		value: value,
	}
}

type claim struct {
	name  string
	value string
}

func (c *claim) Name() string {
	return c.name
}

func (c *claim) Value() string {
	return c.value
}

// Values implements Claim.
func (c *claim) Values(sep string) []string {
	return strings.Split(c.value, sep)
}

func (c *claim) IntValue() (int, bool) {
	r, ok := c.Int64Value()
	return int(r), ok
}

func (c *claim) Int32Value() (int32, bool) {
	r, err := strconv.ParseInt(c.value, 10, 32)
	if err != nil {

		return 0, false
	}
	return int32(r), true
}

func (c *claim) Int64Value() (int64, bool) {
	r, err := strconv.ParseInt(c.value, 10, 64)
	if err != nil {
		return 0, false
	}
	return r, true
}

func (c *claim) BoolValue() (bool, bool) {
	r, err := strconv.ParseBool(c.value)
	if err != nil {
		return false, false
	}
	return r, true
}

func (c *claim) UUIDValue() (uuid.UUID, bool) {
	r, err := uuid.Parse(c.value)
	if err != nil {
		return r, false
	}
	return r, true
}
