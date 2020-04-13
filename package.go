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
 *
 * Modeled after https://github.com/nayuki/QR-Code-generator.
 * See https://www.thonky.com/qr-code-tutorial/introduction and
 * https://en.wikipedia.org/wiki/QR_code for an explanation of how QR codes
 * are formatted.
 */

package qrcodegen

var (
	alignmentPatternPositions [41][]byte

	eccCodeWordsPerBlock [4][41]int = [4][41]int{
		// Version: (note that index 0 is for padding, and is set to an illegal
		// value)
		//       0,  1,  2,  3,  4,  5,  6,  7,  8,  9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40      Error correction level
		{-1, 7, 10, 15, 20, 26, 18, 20, 24, 30, 18, 20, 24, 26, 30, 22, 24, 28, 30, 28, 28, 28, 28, 30, 30, 26, 28, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30},  // Low
		{-1, 10, 16, 26, 18, 24, 16, 18, 22, 22, 26, 30, 22, 22, 24, 24, 28, 28, 26, 26, 26, 26, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28}, // Medium
		{-1, 13, 22, 18, 26, 18, 24, 18, 22, 20, 24, 28, 26, 24, 20, 30, 24, 28, 28, 26, 30, 28, 30, 30, 30, 30, 28, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30}, // Quartile
		{-1, 17, 28, 22, 16, 22, 28, 26, 26, 24, 28, 24, 28, 22, 24, 24, 30, 28, 28, 26, 28, 30, 24, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30}, // High
	}

	numDataCodewords [4][41]int

	numErrorCorrectionBlocks [4][41]int = [4][41]int{
		// Version: (note that index 0 is for padding, and is set to an illegal
		// value)
		//       0, 1, 2, 3, 4, 5, 6, 7, 8, 9,10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40      Error correction level
		{-1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 4, 4, 4, 4, 4, 6, 6, 6, 6, 7, 8, 8, 9, 9, 10, 12, 12, 12, 13, 14, 15, 16, 17, 18, 19, 19, 20, 21, 22, 24, 25},              // Low
		{-1, 1, 1, 1, 2, 2, 4, 4, 4, 5, 5, 5, 8, 9, 9, 10, 10, 11, 13, 14, 16, 17, 17, 18, 20, 21, 23, 25, 26, 28, 29, 31, 33, 35, 37, 38, 40, 43, 45, 47, 49},     // Medium
		{-1, 1, 1, 2, 2, 4, 4, 6, 6, 8, 8, 8, 10, 12, 16, 12, 17, 16, 18, 21, 20, 23, 23, 25, 27, 29, 34, 34, 35, 38, 40, 43, 45, 48, 51, 53, 56, 59, 62, 65, 68},  // Quartile
		{-1, 1, 1, 2, 4, 4, 4, 5, 6, 8, 8, 11, 11, 16, 16, 18, 16, 19, 21, 25, 25, 25, 34, 30, 32, 35, 37, 40, 42, 45, 48, 51, 54, 57, 60, 63, 66, 70, 74, 77, 81}, // High
	}

	numRawDataModules [41]int

	reedSolomonDivisors = make(map[int][]byte)
)

func init() {
	// Initialize the numRawDataModules table for each version number [1, 40].
	// numRawDataModules contains the number of data bits that can be stored in
	// a QR code for each version number, after all function modules are
	// excluded. This includes remainder bits, so it might not be a multiple of
	// 8. The result is in the range [208, 29648].
	for v := 1; v <= 40; v++ {
		result := (16*v+128)*v + 64
		if v >= 2 {
			numAlign := v/7 + 2
			result -= (25*numAlign-10)*numAlign - 55 // Subtract alignment patterns.
			if v >= 7 {
				result -= 36 // Subtract version information.
			}
		}
		if result < 208 || result > 29648 {
			panic("numRawDataModules miscalculated")
		}
		numRawDataModules[v] = result
	}

	// Initialize the numDataCodewords table for each version number and error
	// correction level. numDataCodewords contains the number of 8-bit data (not
	// error correction) codewords contained in any QR code of the given version
	// number and error correction level, with remainder bits discarded.
	for e := Low; e <= High; e++ {
		for v := 1; v <= 40; v++ {
			numDataCodewords[e][v] = numRawDataModules[v]/8 - eccCodeWordsPerBlock[e][v]*numErrorCorrectionBlocks[e][v]
		}
	}

	// Precompute the Reed-Solomon divisor polynomials.
	words := make([]int, 0, 4*40)
	for e := 0; e < 4; e++ {
		for v := 1; v <= 40; v++ {
			words = append(words, eccCodeWordsPerBlock[e][v])
		}
	}

	for _, w := range words {
		if _, ok := reedSolomonDivisors[w]; ok {
			continue
		}

		reedSolomonDivisors[w] = reedSolomonComputeDivisor(w)
	}

	// Initialize the alignmentPatternPositions each based on a version in the
	// range (1, 40).
	for v := 1; v <= 40; v++ {
		alignmentPatternPositions[v] = getAlignmentPatternPositions(Version(v))
	}
}

func abs(a int) int {
	if a >= 0 {
		return a
	}

	return -a
}

func bToI(b bool) int {
	if b {
		return 1
	}

	return 0
}

func bToModule(b bool) module {
	if b {
		return 1
	}

	return 0
}

func getBit(x, i int) int {
	return x >> i & 1
}

func getBitAsBool(x, i int) bool {
	return x>>i&1 == 1
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
