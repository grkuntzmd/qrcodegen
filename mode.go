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
