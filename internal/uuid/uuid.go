package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

type UUID [16]byte

var Nil UUID

func New() UUID {
	var value UUID
	if _, err := rand.Read(value[:]); err != nil {
		panic(err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return value
}

func NewString() string { return New().String() }

func Parse(value string) (UUID, error) {
	var result UUID
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return result, errors.New("invalid UUID format")
	}
	raw := make([]byte, 16)
	if _, err := hex.Decode(raw, []byte(value[:8]+value[9:13]+value[14:18]+value[19:23]+value[24:])); err != nil {
		return result, err
	}
	copy(result[:], raw)
	return result, nil
}

func (u UUID) String() string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(u[:4]), hex.EncodeToString(u[4:6]),
		hex.EncodeToString(u[6:8]), hex.EncodeToString(u[8:10]), hex.EncodeToString(u[10:]))
}

func (u UUID) MarshalText() ([]byte, error) { return []byte(u.String()), nil }
func (u *UUID) UnmarshalText(value []byte) error {
	parsed, err := Parse(string(value))
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}
func (u UUID) MarshalJSON() ([]byte, error) { return json.Marshal(u.String()) }
func (u *UUID) UnmarshalJSON(value []byte) error {
	var text string
	if err := json.Unmarshal(value, &text); err != nil {
		return err
	}
	return u.UnmarshalText([]byte(text))
}
