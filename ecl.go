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

// ECL represents the error correction level of the QR code.
type ECL int8

// ECL values.
const (
	Low      ECL = iota // Low error correction level (recovers 7% of data).
	Medium              // Medium error correction level (recovers 15% of data).
	Quartile            // Quartile error correction level (recovers 25% of data).
	High                // High error correction level (recovers 30% of data).
)

func (e ECL) formatBits() int {
	switch e {
	case Low:
		return 1
	case Medium:
		return 0
	case Quartile:
		return 3
	case High:
		return 2
	default:
		panic("unknown ECC level")
	}
}
