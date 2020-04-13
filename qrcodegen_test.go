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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppendBitsToBuffer(t *testing.T) {
	bb := make(bitBuffer, 0)

	bb.appendBits(0, 0)
	assert.Equal(t, 0, len(bb))

	bb.appendBits(1, 1)
	assert.Equal(t, 1, len(bb))
	assert.Equal(t, []byte{1}, []byte(bb))

	bb.appendBits(0, 1)
	assert.Equal(t, 2, len(bb))
	assert.Equal(t, []byte{1, 0}, []byte(bb))

	bb.appendBits(5, 3)
	assert.Equal(t, 5, len(bb))
	assert.Equal(t, []byte{1, 0, 1, 0, 1}, []byte(bb))

	bb.appendBits(6, 3)
	assert.Equal(t, 8, len(bb))
	assert.Equal(t, []byte{1, 0, 1, 0, 1, 1, 1, 0}, []byte(bb))
}

func TestNumDataCodewords(t *testing.T) {
	cases := [][3]int{
		{3, 1, 44},
		{3, 2, 34},
		{3, 3, 26},
		{6, 0, 136},
		{7, 0, 156},
		{9, 0, 232},
		{9, 1, 182},
		{12, 3, 158},
		{15, 0, 523},
		{16, 2, 325},
		{19, 3, 341},
		{21, 0, 932},
		{22, 0, 1006},
		{22, 1, 782},
		{22, 3, 442},
		{24, 0, 1174},
		{24, 3, 514},
		{28, 0, 1531},
		{30, 3, 745},
		{32, 3, 845},
		{33, 0, 2071},
		{33, 3, 901},
		{35, 0, 2306},
		{35, 1, 1812},
		{35, 2, 1286},
		{36, 3, 1054},
		{37, 3, 1096},
		{39, 1, 2216},
		{40, 1, 2334},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("TestNumDataCodewords %v", tc), func(t *testing.T) {
			assert.Equal(t, numDataCodewords[tc[1]][tc[0]], tc[2])
		})
	}
}

func TestNumRawDataModules(t *testing.T) {
	cases := [][2]int{
		{1, 208},
		{2, 359},
		{3, 567},
		{6, 1383},
		{7, 1568},
		{12, 3728},
		{15, 5243},
		{18, 7211},
		{22, 10068},
		{26, 13652},
		{32, 19723},
		{37, 25568},
		{40, 29648},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("TestNumRawDataModules %v", tc), func(t *testing.T) {
			assert.Equal(t, numRawDataModules[tc[0]], tc[1])
		})
	}
}

func TestReedSolomonComputeDivisor(t *testing.T) {
	var generator []byte

	generator = reedSolomonComputeDivisor(1)
	assert.True(t, generator[0] == 0x01)

	generator = reedSolomonComputeDivisor(2)
	assert.True(t, generator[0] == 0x03)
	assert.True(t, generator[1] == 0x02)

	generator = reedSolomonComputeDivisor(5)
	assert.True(t, generator[0] == 0x1F)
	assert.True(t, generator[1] == 0xC6)
	assert.True(t, generator[2] == 0x3F)
	assert.True(t, generator[3] == 0x93)
	assert.True(t, generator[4] == 0x74)

	generator = reedSolomonComputeDivisor(30)
	assert.True(t, generator[0] == 0xD4)
	assert.True(t, generator[1] == 0xF6)
	assert.True(t, generator[5] == 0xC0)
	assert.True(t, generator[12] == 0x16)
	assert.True(t, generator[13] == 0xD9)
	assert.True(t, generator[20] == 0x12)
	assert.True(t, generator[27] == 0x6A)
	assert.True(t, generator[29] == 0x96)
}

func TestReedSolomonComputeRemainder(t *testing.T) {
	{
		data := []byte{0}
		generator := reedSolomonComputeDivisor(3)
		remainder := reedSolomonComputeRemainder(data, generator)
		assert.Equal(t, 3, len(remainder))
		for i := 0; i < 3; i++ {
			assert.Equal(t, byte(0), remainder[i])
		}
	}
	{
		data := []byte{0, 1}
		generator := reedSolomonComputeDivisor(3)
		remainder := reedSolomonComputeRemainder(data, generator)
		assert.Equal(t, 3, len(remainder))
		for i := 0; i < 3; i++ {
			assert.Equal(t, generator[i], remainder[i])
		}
	}
	{
		data := []byte{0x03, 0x3A, 0x60, 0x12, 0xC7}
		generator := reedSolomonComputeDivisor(5)
		remainder := reedSolomonComputeRemainder(data, generator)
		assert.Equal(t, 5, len(remainder))
		expected := []byte{0xCB, 0x36, 0x16, 0xFA, 0x9D}
		for i := 0; i < 3; i++ {
			assert.Equal(t, expected[i], remainder[i])
		}
	}
	{
		data := []byte{
			0x38, 0x71, 0xDB, 0xF9, 0xD7, 0x28, 0xF6, 0x8E, 0xFE, 0x5E,
			0xE6, 0x7D, 0x7D, 0xB2, 0xA5, 0x58, 0xBC, 0x28, 0x23, 0x53,
			0x14, 0xD5, 0x61, 0xC0, 0x20, 0x6C, 0xDE, 0xDE, 0xFC, 0x79,
			0xB0, 0x8B, 0x78, 0x6B, 0x49, 0xD0, 0x1A, 0xAD, 0xF3, 0xEF,
			0x52, 0x7D, 0x9A,
		}
		generator := reedSolomonComputeDivisor(30)
		remainder := reedSolomonComputeRemainder(data, generator)
		assert.Equal(t, 30, len(remainder))
		assert.Equal(t, byte(0xCE), remainder[0])
		assert.Equal(t, byte(0xF0), remainder[1])
		assert.Equal(t, byte(0x31), remainder[2])
		assert.Equal(t, byte(0xDE), remainder[3])
		assert.Equal(t, byte(0xE1), remainder[8])
		assert.Equal(t, byte(0xCA), remainder[12])
		assert.Equal(t, byte(0xE3), remainder[17])
		assert.Equal(t, byte(0x85), remainder[19])
		assert.Equal(t, byte(0x50), remainder[20])
		assert.Equal(t, byte(0xBE), remainder[24])
		assert.Equal(t, byte(0xB3), remainder[29])
	}
}

func TestReedSolomonMultiply(t *testing.T) {
	cases := [][3]byte{
		{0x00, 0x00, 0x00},
		{0x01, 0x01, 0x01},
		{0x02, 0x02, 0x04},
		{0x00, 0x6E, 0x00},
		{0xB2, 0xDD, 0xE6},
		{0x41, 0x11, 0x25},
		{0xB0, 0x1F, 0x11},
		{0x05, 0x75, 0xBC},
		{0x52, 0xB5, 0xAE},
		{0xA8, 0x20, 0xA4},
		{0x0E, 0x44, 0x9F},
		{0xD4, 0x13, 0xA0},
		{0x31, 0x10, 0x37},
		{0x6C, 0x58, 0xCB},
		{0xB6, 0x75, 0x3E},
		{0xFF, 0xFF, 0xE2},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("TestReedSolomonMultiply %v", tc), func(t *testing.T) {
			assert.Equal(t, tc[2], reedSolomonMultiply(tc[0], tc[1]))
		})
	}
}

func TestDrawFunctionPatterns(t *testing.T) {
	for version := Version(1); version <= 40; version++ {
		size := int(version)*4 + 17
		qrCode := QRCode{
			Version:    version,
			Size:       size,
			Modules:    make([][]Module, size),
			isFunction: make([][]bool, size),
		}

		for i := 0; i < size; i++ {
			qrCode.Modules[i] = make([]Module, size)
			qrCode.isFunction[i] = make([]bool, size)
		}

		qrCode.drawFunctionPatterns()

		hasBlack := false
		hasWhite := false
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				color := qrCode.Modules[y][x]
				if color == 1 {
					hasBlack = true
				} else {
					hasWhite = true
				}
			}
		}
		assert.True(t, hasBlack)
		assert.True(t, hasWhite)
	}
}

func TestGetAlignmentPatternPositions(t *testing.T) {
	cases := [][9]int{
		{1, 0, -1, -1, -1, -1, -1, -1, -1},
		{2, 2, 6, 18, -1, -1, -1, -1, -1},
		{3, 2, 6, 22, -1, -1, -1, -1, -1},
		{6, 2, 6, 34, -1, -1, -1, -1, -1},
		{7, 3, 6, 22, 38, -1, -1, -1, -1},
		{8, 3, 6, 24, 42, -1, -1, -1, -1},
		{16, 4, 6, 26, 50, 74, -1, -1, -1},
		{25, 5, 6, 32, 58, 84, 110, -1, -1},
		{32, 6, 6, 34, 60, 86, 112, 138, -1},
		{33, 6, 6, 30, 58, 86, 114, 142, -1},
		{39, 7, 6, 26, 54, 82, 110, 138, 166},
		{40, 7, 6, 30, 58, 86, 114, 142, 170},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("TestGetAlignmentPatternPositions %v", tc), func(t *testing.T) {
			pos := alignmentPatternPositions[tc[0]]
			assert.Equal(t, tc[1], len(pos))
			for i := 0; i < len(pos); i++ {
				assert.Equal(t, tc[i+2], int(pos[i]))
			}
		})
	}
}

func TestIsAlphanumeric(t *testing.T) {
	cases := []struct {
		answer bool
		text   string
	}{
		{true, ""},
		{true, "0"},
		{true, "A"},
		{false, "a"},
		{true, " "},
		{true, "."},
		{true, "*"},
		{false, ","},
		{false, "|"},
		{false, "@"},
		{true, "XYZ"},
		{false, "XYZ!"},
		{true, "79068"},
		{true, "+123 ABC$"},
		{false, "\x01"},
		{false, "\x7F"},
		{false, "\x80"},
		{false, "\xC0"},
		{false, "\xFF"},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("TestIsAlphanumeric %v", tc), func(t *testing.T) {
			assert.Equal(t, tc.answer, alphanumericRegexp.MatchString(tc.text))
		})
	}
}

func TestIsNumeric(t *testing.T) {
	cases := []struct {
		answer bool
		text   string
	}{
		{true, ""},
		{true, "0"},
		{false, "A"},
		{false, "a"},
		{false, " "},
		{false, "."},
		{false, "*"},
		{false, ","},
		{false, "|"},
		{false, "@"},
		{false, "XYZ"},
		{false, "XYZ!"},
		{true, "79068"},
		{false, "+123 ABC$"},
		{false, "\x01"},
		{false, "\x7F"},
		{false, "\x80"},
		{false, "\xC0"},
		{false, "\xFF"},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("TestIsNumeric %v", tc), func(t *testing.T) {
			assert.Equal(t, tc.answer, numericRegexp.MatchString(tc.text))
		})
	}
}

func TestMakeBytes(t *testing.T) {
	{
		seg := MakeBytes([]byte{})
		assert.Equal(t, Byte, seg.Mode)
		assert.Equal(t, 0, seg.NumChars)
		assert.Equal(t, 0, len(seg.Data))
		assert.Equal(t, []byte{}, seg.Data)
	}
	{
		seg := MakeBytes([]byte{0x00})
		assert.Equal(t, Byte, seg.Mode)
		assert.Equal(t, 1, seg.NumChars)
		assert.Equal(t, 8, len(seg.Data))
		assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, seg.Data)
	}
	{
		seg := MakeBytes([]byte{0xEF, 0xBB, 0xBF})
		assert.Equal(t, Byte, seg.Mode)
		assert.Equal(t, 3, seg.NumChars)
		assert.Equal(t, 24, len(seg.Data))
		assert.Equal(t, []byte{0x1, 0x1, 0x1, 0x0, 0x1, 0x1, 0x1, 0x1, 0x1, 0x0, 0x1, 0x1, 0x1, 0x0, 0x1, 0x1, 0x1, 0x0, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}, seg.Data)
	}
}

func TestMakeNumeric(t *testing.T) {
	cases := []struct {
		text      string
		length    int
		bitLength int
		bytes     []byte
	}{
		{"", 0, 0, []byte{}},
		{"9", 1, 4, []byte{0x1, 0x0, 0x0, 0x1}},
		{"81", 2, 7, []byte{0x1, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1}},
		{"673", 3, 10, []byte{0x1, 0x0, 0x1, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x1}},
		{"3141592653", 10, 34, []byte{0x0, 0x1, 0x0, 0x0, 0x1, 0x1, 0x1, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x1, 0x1, 0x1,
			0x1, 0x1, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x1, 0x0, 0x0, 0x1, 0x1}},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("TestMakeNumeric %v", tc), func(t *testing.T) {
			seg := MakeNumeric(tc.text)
			assert.Equal(t, Numeric, seg.Mode)
			assert.Equal(t, tc.length, seg.NumChars)
			assert.Equal(t, tc.bitLength, len(seg.Data))
			assert.Equal(t, tc.bytes, seg.Data)
		})
	}
}

func TestMakeAlphanumeric(t *testing.T) {
	cases := []struct {
		text      string
		length    int
		bitLength int
		bytes     []byte
	}{
		{"", 0, 0, []byte{}},
		{"A", 1, 6, []byte{0x0, 0x0, 0x1, 0x0, 0x1, 0x0}},
		{"%:", 2, 11, []byte{0x1, 0x1, 0x0, 0x1, 0x1, 0x0, 0x1, 0x1, 0x0, 0x1, 0x0}},
		{"Q R", 3, 17, []byte{0x1, 0x0, 0x0, 0x1, 0x0, 0x1, 0x1, 0x0, 0x1, 0x1, 0x0, 0x0, 0x1, 0x1, 0x0, 0x1, 0x1}},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("TestMakeAlphanumeric %v", tc), func(t *testing.T) {
			seg := MakeAlphanumeric(tc.text)
			assert.Equal(t, Alphanumeric, seg.Mode)
			assert.Equal(t, tc.length, seg.NumChars)
			assert.Equal(t, tc.bitLength, len(seg.Data))
			assert.Equal(t, tc.bytes, seg.Data)
		})
	}
}

func TestMakeEci(t *testing.T) {
	cases := []struct {
		input     int
		length    int
		bitLength int
		bytes     []byte
	}{
		{127, 0, 8, []byte{0x0, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}},
		{10345, 0, 16, []byte{0x1, 0x0, 0x1, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x1, 0x1, 0x0, 0x1, 0x0, 0x0, 0x1}},
		{999999, 0, 24, []byte{0x1, 0x1, 0x0, 0x0, 0x1, 0x1, 0x1, 0x1, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("TestMakeEci %v", tc), func(t *testing.T) {
			seg, err := MakeECI(tc.input)
			assert.Nil(t, err)
			assert.Equal(t, ECI, seg.Mode)
			assert.Equal(t, tc.length, seg.NumChars)
			assert.Equal(t, tc.bitLength, len(seg.Data))
			assert.Equal(t, tc.bytes, seg.Data)
		})
	}
}

func TestGetTotalBits(t *testing.T) {
	{
		assert.Equal(t, 0, getTotalBits([]*QRSegment{}, 1))
		assert.Equal(t, 0, getTotalBits([]*QRSegment{}, 40))
	}
	{
		segs := []*QRSegment{{Mode: Byte, NumChars: 3, Data: make([]byte, 24)}}
		assert.Equal(t, 36, getTotalBits(segs, 2))
		assert.Equal(t, 44, getTotalBits(segs, 10))
		assert.Equal(t, 44, getTotalBits(segs, 30))
	}
	{
		segs := []*QRSegment{
			{Mode: ECI, NumChars: 0, Data: make([]byte, 8)},
			{Mode: Numeric, NumChars: 7, Data: make([]byte, 24)},
			{Mode: Alphanumeric, NumChars: 1, Data: make([]byte, 6)},
			{Mode: kanji, NumChars: 4, Data: make([]byte, 52)},
		}
		assert.Equal(t, 133, getTotalBits(segs, 9))
		assert.Equal(t, 139, getTotalBits(segs, 21))
		assert.Equal(t, 145, getTotalBits(segs, 27))
	}
	{
		segs := []*QRSegment{{Mode: Byte, NumChars: 4093, Data: make([]byte, 32744)}}
		assert.Equal(t, -1, getTotalBits(segs, 1))
		assert.Equal(t, 32764, getTotalBits(segs, 10))
		assert.Equal(t, 32764, getTotalBits(segs, 27))
	}
	{
		segs := []*QRSegment{
			{Mode: Numeric, NumChars: 2047, Data: make([]byte, 6824)},
			{Mode: Numeric, NumChars: 2047, Data: make([]byte, 6824)},
			{Mode: Numeric, NumChars: 2047, Data: make([]byte, 6824)},
			{Mode: Numeric, NumChars: 2047, Data: make([]byte, 6824)},
			{Mode: Numeric, NumChars: 1617, Data: make([]byte, 5390)},
		}
		assert.Equal(t, -1, getTotalBits(segs, 1))
		assert.Equal(t, 32766, getTotalBits(segs, 10))
		assert.Equal(t, 32776, getTotalBits(segs, 27))
	}
	{
		segs := []*QRSegment{
			{Mode: kanji, NumChars: 255, Data: make([]byte, 3315)},
			{Mode: kanji, NumChars: 255, Data: make([]byte, 3315)},
			{Mode: kanji, NumChars: 255, Data: make([]byte, 3315)},
			{Mode: kanji, NumChars: 255, Data: make([]byte, 3315)},
			{Mode: kanji, NumChars: 255, Data: make([]byte, 3315)},
			{Mode: kanji, NumChars: 255, Data: make([]byte, 3315)},
			{Mode: kanji, NumChars: 255, Data: make([]byte, 3315)},
			{Mode: kanji, NumChars: 255, Data: make([]byte, 3315)},
			{Mode: kanji, NumChars: 255, Data: make([]byte, 3315)},
			{Mode: Alphanumeric, NumChars: 511, Data: make([]byte, 2811)},
		}
		assert.Equal(t, 32767, getTotalBits(segs, 9))
		assert.Equal(t, 32787, getTotalBits(segs, 26))
		assert.Equal(t, 32807, getTotalBits(segs, 40))
	}
}
