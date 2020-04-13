/*
 * Copyright © 2020, G.Ralph Kuntz, MD.
 *
 * Licensed under the Apache License, Version 2.0(the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIC
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package qrcodegen

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// QRSegment represents a single segment in a QR code. A QR code may contain
// more than one segment (numeric, alphanumeric, byte, kanji, or ECI).
type QRSegment struct {
	Mode            // The mode of this segment (numeric, alphanumeric, byte, kanji, or ECI).
	NumChars int    // The length of this segments unencoded data.
	Data     []byte // The encoded data for this segment.
}

const alphanumericCharset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:"

var (
	alphanumericRegexp = regexp.MustCompile(`^[A-Z0-9 $%*+./:-]*$`)
	numericRegexp      = regexp.MustCompile(`^[0-9]*$`)
)

func getTotalBits(segs []*QRSegment, version Version) int {
	result := int64(0)
	for _, seg := range segs {
		ccBits := seg.Mode.numCharCountBits(version)
		if seg.NumChars >= 1<<ccBits {
			return -1 // The segment's length does not fit the field's bit width.
		}

		result += int64(4 + int(ccBits) + len(seg.Data))
		if result > math.MaxInt32 {
			return -1 // The sum will overflow an integer type.
		}
	}

	return int(result)
}

// MakeAlphanumeric creates an alphanumeric segment from the given text
// (uppercase letters, digits, some symbols).
func MakeAlphanumeric(text string) *QRSegment {
	if !alphanumericRegexp.MatchString(text) {
		panic("string contains non-alphanumeric characters")
	}

	bb := make(bitBuffer, 0, len(text)*5+(len(text)+1)/2)
	var i int
	for i = 0; i <= len(text)-2; i += 2 { // Process groups of 2 characters.
		temp := strings.Index(alphanumericCharset, text[i:i+1]) * 45
		temp += strings.Index(alphanumericCharset, text[i+1:i+2])
		bb.appendBits(temp, 11)
	}

	if i < len(text) { // 1 character remaining.
		bb.appendBits(strings.Index(alphanumericCharset, text[i:i+1]), 6)
	}

	return &QRSegment{
		Mode:     Alphanumeric,
		NumChars: len(text),
		Data:     bb,
	}
}

// MakeBytes encodes a byte slice into a QR segment of type Byte.
func MakeBytes(data []byte) *QRSegment {
	bb := make(bitBuffer, 0, len(data)*8)
	for _, b := range data {
		bb.appendBits(int(b), 8)
	}

	return &QRSegment{
		Mode:     Byte,
		NumChars: len(data),
		Data:     bb,
	}
}

// MakeECI creates a segment representing an extended channel interpretation
// (ECI) designator with the specified value.
func MakeECI(assignValue int) (*QRSegment, error) {
	bb := make(bitBuffer, 0, 24)
	if assignValue < 1<<7 {
		bb.appendBits(assignValue, 8)
	} else if assignValue < 1<<14 {
		bb.appendBits(2, 2)
		bb.appendBits(assignValue, 14)
	} else if assignValue < 1_000_000 {
		bb.appendBits(6, 3)
		bb.appendBits(assignValue, 21)
	} else {
		return nil, fmt.Errorf("ECI assignment out of range")
	}

	return &QRSegment{
		Mode:     ECI,
		NumChars: 0,
		Data:     bb,
	}, nil
}

// MakeNumeric creates a numeric segment from the given digit string.
func MakeNumeric(digits string) *QRSegment {
	if !numericRegexp.MatchString(digits) {
		panic("string contains non-numeric characters")
	}

	bb := make(bitBuffer, 0, len(digits)*3+(len(digits)+2)/3)
	for i := 0; i < len(digits); {
		n := min(len(digits)-i, 3)
		d, _ := strconv.Atoi(digits[i : i+n]) // We can safely ignore the possible conversion error because we have confirmed that the string contains only digits in the regexp above.
		bb.appendBits(d, int8(n*3+1))
		i += n
	}

	return &QRSegment{
		Mode:     Numeric,
		NumChars: len(digits),
		Data:     bb,
	}
}

// MakeSegments encodes text into a QR segment, selecting the most efficient
// mode that can be used (numeric, alphanumeric, byte, or kanji).
func MakeSegments(text string) []*QRSegment {
	if len(text) == 0 {
		return []*QRSegment{}
	}

	if numericRegexp.MatchString(text) {
		return []*QRSegment{MakeNumeric(text)}
	}

	if alphanumericRegexp.MatchString(text) {
		return []*QRSegment{MakeAlphanumeric(text)}
	}

	return []*QRSegment{MakeBytes([]byte(text))}
}
