// Package iata provides utilities for working with IATA standards in the
// context of air cargo logistics.
//
// Currently it includes parsing and validation of Master Air Waybill (MAWB)
// numbers according to IATA Resolution 600a.
package iata

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	errMawbNumberContainsNonASCII = errors.New("mawb number contains non-ASCII characters")
	errMawbNumberWrongLength      = errors.New("mawb number must be 11 characters long")
)

// MawbLength the length of a Master Air Waybill Number as per IATA RESOLUTION
// 600a.
const MawbLength = 11

// MawbNumber represents the number of a Master Air Waybill Number (MAWB) as
// defined by IATA RESOLUTION 600a.
type MawbNumber struct {
	airlineCode int
	serial      int
}

// ParseMawb parses a string into a MawbNumber.
func ParseMawb(number string) (MawbNumber, error) {
	sanitizedMawb := strings.ReplaceAll(number, "-", "")

	sanitizedMawb = strings.ReplaceAll(sanitizedMawb, " ", "")

	if utf8.RuneCountInString(sanitizedMawb) != len(sanitizedMawb) {
		return MawbNumber{}, errMawbNumberContainsNonASCII
	}

	if len(sanitizedMawb) != MawbLength {
		return MawbNumber{}, errMawbNumberWrongLength
	}

	airlineCode, err := strconv.Atoi(sanitizedMawb[:3])
	if err != nil {
		return MawbNumber{}, fmt.Errorf("failed to parse airline code: %w", err)
	}

	serial, err := strconv.Atoi(sanitizedMawb[3:])
	if err != nil {
		return MawbNumber{}, fmt.Errorf("failed to parse serial: %w", err)
	}

	return MawbNumber{
		airlineCode: airlineCode,
		serial:      serial,
	}, nil
}

// AirlineCode returns the airline code of the MAWB number.
func (s MawbNumber) AirlineCode() int {
	return s.airlineCode
}

// Serial returns the serial number of the MAWB number.
func (s MawbNumber) Serial() int {
	return s.serial
}

// String returns the string representation of the MAWB number.
func (s MawbNumber) String() string {
	return fmt.Sprintf("%03d%08d", s.airlineCode, s.serial)
}
