/*
 * MIT LICENSE
 *
 * Copyright © 2020, G.Ralph Kuntz, MD.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package qrcodegen

// Mode represents the mode (numeric, alphanumeric, byte, kanji, or ECI) of a
// segment.
type Mode struct {
	modeBits int8
	numBits  [3]int8
}

// Mode values for a segment.
var (
	Numeric      = Mode{0x1, [3]int8{10, 12, 14}}
	Alphanumeric = Mode{0x2, [3]int8{9, 11, 13}}
	Byte         = Mode{0x4, [3]int8{8, 16, 16}}
	kanji        = Mode{0x8, [3]int8{8, 10, 12}}
	ECI          = Mode{0x7, [3]int8{0, 0, 0}}
)

func (m *Mode) numCharCountBits(version Version) int8 {
	return m.numBits[(version+7)/17]
}
