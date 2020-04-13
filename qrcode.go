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
	"strings"
)

// QRCode represents a QR code symbol, which is a type of two-dimensional
// barcode.
type QRCode struct {
	Version                         // The QR code version, a number in the range [1, 40].
	Size                 int        // The width and height of the square QR code symbol as measured in "modules" (smallest square, either black or white, in a QR code).
	ErrorCorrectionLevel ECC        // The error correction level used in this QR code.
	Mask                            // The type of mask [0, 7] used in this QR code.
	Modules              [][]module // The modules ("pixels") that make up this QR code (black = 1, white = 0)
	IsFunction           [][]bool   // Indicates that a module is a "function" (contains metadata and does not represent part of the message of the QR code).
}

// The maximum and minimum versions (QR code sizes) for a QR code symbol.
// Version = 21 modules, squared, and version 40 = 177 modules, squared.
const (
	MaxVersion = Version(40)
	MinVersion = Version(1)

	// Penalties scores used to determine how likely a mask is to make scanning
	// more error-prone.
	penaltyN1 = 3
	penaltyN2 = 3
	penaltyN3 = 40
	penaltyN4 = 10
)

// EncodeBinary encodes a byte slice into a QR code symbol with the given error correction level.
func EncodeBinary(data []byte, ecl ECC) (*QRCode, error) {
	seg := MakeBytes(data)
	return EncodeSegments([]*QRSegment{seg}, ecl)
}

// EncodeSegments creates the QR code structure from one or more QR segments.
func EncodeSegments(segs []*QRSegment, ecl ECC, options ...func(*segmentEncoder)) (*QRCode, error) {
	s := segmentEncoder{
		boostECL:   true,
		mask:       -1, // Set to automatic mask selection.
		maxVersion: 40,
		minVersion: 1,
	}
	for _, o := range options {
		o(&s)
	}

	if s.minVersion < MinVersion || MaxVersion < s.maxVersion || s.maxVersion < s.minVersion {
		return nil, fmt.Errorf("invalid segment versions")
	}

	if s.mask < -1 || s.mask > 7 {
		return nil, fmt.Errorf("mask value out of range")
	}

	// Find the minimal version number to use.
	version := MinVersion
	var dataUsedBits int
	for {
		dataCapacityBits := numDataCodewords[ecl][version] * 8 // Number of data bits available.
		dataUsedBits = getTotalBits(segs, version)
		if dataUsedBits != -1 && dataUsedBits <= dataCapacityBits {
			break // This version number is suitable.
		}
		if version >= MaxVersion { // All versions in the range could not fit the given data.
			if dataUsedBits != -1 {
				return nil, fmt.Errorf("data length = %d bits, max capacity = %d bits", dataUsedBits, dataCapacityBits)
			}
			return nil, fmt.Errorf("data too long")
		}
		version++
	}

	if dataUsedBits == -1 {
		panic("incorrect data size calculation")
	}

	// Increase the error correction level while the data still fits in the current version number.
	for newEcl := Medium; newEcl <= High; newEcl++ {
		if s.boostECL && dataUsedBits <= numDataCodewords[newEcl][version]*8 {
			ecl = newEcl
		}
	}

	// Concatenate all segments to create the data bit string.
	bb := make(bitBuffer, 0)
	for _, seg := range segs {
		bb.appendBits(int(seg.modeBits), 4)
		bb.appendBits(seg.NumChars, seg.Mode.numCharCountBits(version))
		bb = append(bb, seg.Data...)
	}
	if len(bb) != dataUsedBits {
		panic("incorrect data size calculation")
	}

	// Add the terminator and pad up to a byte if applicable.
	dataCapacityBits := numDataCodewords[ecl][version] * 8
	if len(bb) > dataCapacityBits {
		panic("incorrect data size calculation")
	}
	bb.appendBits(0, int8(min(4, dataCapacityBits-len(bb))))
	bb.appendBits(0, int8((8-len(bb)%8)%8))
	if len(bb)%8 != 0 {
		panic("incorrect data size calculation")
	}

	// Pad with alternating bytes until data capacity is reached.
	for padByte := int16(0xec); len(bb) < dataCapacityBits; padByte ^= 0xec ^ 0x11 {
		bb.appendBits(int(padByte), 8)
	}

	// Pack bits into bytes in big endian order.
	dataCodeWords := make([]byte, len(bb)/8)
	for i := 0; i < len(bb); i++ {
		// if i >= 70 && i <= 73 {
		// 	fmt.Printf("i: %d, dataCodeWords[i>>3]: %d, i >> 3: %d, bb[i]: %d, bb[i] << (7 - i&7): %d: dataCodeWords[i>>3] | bb[i] << (7 - i&7): %d\n",
		// 		i, dataCodeWords[i>>3], i>>3, bb[i], bb[i]<<(7-i&7), dataCodeWords[i>>3]|bb[i]<<(7-i&7))
		// }
		dataCodeWords[i>>3] |= bb[i] << (7 - i&7)
	}

	size := int(version)*4 + 17
	qrCode := QRCode{
		Version:              version,
		Size:                 size,
		ErrorCorrectionLevel: ecl,
		Modules:              make([][]module, size),
		IsFunction:           make([][]bool, size),
	}

	for i := 0; i < size; i++ {
		qrCode.Modules[i] = make([]module, size)
		qrCode.IsFunction[i] = make([]bool, size)
	}

	qrCode.drawFunctionPatterns()
	allCodeWords := qrCode.addECCAndInterleave(dataCodeWords)
	qrCode.drawCodewords(allCodeWords)
	qrCode.Mask = qrCode.handleConstructorMasking(s.mask)

	qrCode.IsFunction = nil

	return &qrCode, nil
}

// EncodeText encodes text as a QR code symbol with the given error correction
// level.
func EncodeText(text string, ecl ECC) (*QRCode, error) {
	segs := MakeSegments(text)
	return EncodeSegments(segs, ecl)
}

func (q *QRCode) addECCAndInterleave(data []byte) []byte {
	if len(data) != numDataCodewords[q.ErrorCorrectionLevel][q.Version] {
		panic("data is not correct length")
	}

	// Calculate the parameter numbers.
	numBlocks := numErrorCorrectionBlocks[q.ErrorCorrectionLevel][q.Version]
	blockECCLen := eccCodeWordsPerBlock[q.ErrorCorrectionLevel][q.Version]
	rawCodeWords := numRawDataModules[q.Version] / 8
	numShortBlocks := numBlocks - rawCodeWords%numBlocks
	shortBlockLen := rawCodeWords / numBlocks

	// Split data into blocks and append ECC to each block.
	blocks := make([][]byte, numBlocks)
	rsDiv := reedSolomonDivisors[blockECCLen]
	for i, k := 0, 0; i < numBlocks; i++ {
		dat := data[k : k+shortBlockLen-blockECCLen+bToI(i >= numShortBlocks)]
		k += len(dat)
		block := make([]byte, shortBlockLen+1)
		copy(block, dat)
		ecc := reedSolomonComputeRemainder(dat, rsDiv)
		copy(block[(len(block)-len(ecc)):], ecc)
		blocks[i] = block
	}

	// Interleave (not concatenate) the bytes from every block into a single
	// sequence.
	result := make([]byte, rawCodeWords)
	for i, k := 0, 0; i < len(blocks[0]); i++ {
		for j := 0; j < len(blocks); j++ {
			// Skip the padding byte in short blocks.
			if i != shortBlockLen-blockECCLen || j >= numShortBlocks {
				result[k] = blocks[j][i]
				k++
			}
		}
	}

	return result
}

func (q *QRCode) String() string {
	var sb strings.Builder
	sb.WriteString("QRCode\n")
	fmt.Fprintf(&sb, "\tVersion: %d\n", q.Version)
	fmt.Fprintf(&sb, "\tSize: %d\n", q.Size)
	fmt.Fprintf(&sb, "\tErrorCorrectionLevel: %d\n", q.ErrorCorrectionLevel)
	fmt.Fprintf(&sb, "\tMask: %d\n", q.Mask)
	sb.WriteString("\tModules\n")
	for y := 0; y < q.Size; y++ {
		sb.WriteString("\t\t")
		for x := 0; x < q.Size; x++ {
			if q.Modules[y][x] == 1 {
				sb.WriteString("░")
			} else {
				sb.WriteString("▓")
			}
			// fmt.Fprintf(&sb, "%d", q.Modules[y][x])
		}
		sb.WriteString("\n")
	}
	// sb.WriteString("\tIsFunction")
	// if q.IsFunction == nil {
	// 	sb.WriteString(" nil\n")
	// } else {
	// 	sb.WriteString("\n")
	// 	for y := 0; y < q.Size; y++ {
	// 		sb.WriteString("\t\t")
	// 		for x := 0; x < q.Size; x++ {
	// 			if q.IsFunction[y][x] {
	// 				sb.WriteString("1")
	// 			} else {
	// 				sb.WriteString("0")
	// 			}
	// 		}
	// 		sb.WriteString("\n")
	// 	}
	// }

	return sb.String()
}

// ToSVGString returns a scalable vector graphics (SVG) representation of the QR
// code.
func (q *QRCode) ToSVGString(border int, includeDocType bool) (string, error) {
	if border < 0 {
		return "", fmt.Errorf("border must be non-negative")
	}

	var sb strings.Builder
	if includeDocType {
		sb.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
		sb.WriteString("<!DOCTYPE svg PUBLIC \"-//W3C//DTD SVG 1.1//EN\" \"http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd\">\n")
	}
	fmt.Fprintf(&sb, "<svg xmlns=\"http://www.w3.org/2000/svg\" version=\"1.1\" viewBox=\"0 0 %[1]d %[1]d\" stroke=\"none\">\n", q.Size+border*2)
	sb.WriteString("\t<rect width=\"100%\" height=\"100%\" fill=\"#FFFFFF\"/>\n")
	sb.WriteString("\t<path d=\"")
	for y := 0; y < q.Size; y++ {
		for x := 0; x < q.Size; x++ {
			if q.Modules[y][x] == 1 {
				if x != 0 && y != 0 {
					sb.WriteString(" ")
				}
				fmt.Fprintf(&sb, "M%d,%dh1v1h-1z", x+border, y+border)
			}
		}
	}
	sb.WriteString("\" fill=\"#000000\"/>\n")
	sb.WriteString("</svg>\n")

	return sb.String(), nil
}

// applyMask XOR's the codeword modules (not functions) in this QR code with the
// given mask. Applying this method twice with the same mask will remove the
// mask.
func (q *QRCode) applyMask(mask Mask) {
	for y := 0; y < q.Size; y++ {
		for x := 0; x < q.Size; x++ {
			var invert bool
			switch mask {
			case 0:
				invert = (x+y)%2 == 0
			case 1:
				invert = y%2 == 0
			case 2:
				invert = x%3 == 0
			case 3:
				invert = (x+y)%3 == 0
			case 4:
				invert = (x/3+y/2)%2 == 0
			case 5:
				invert = x*y%2+x*y%3 == 0
			case 6:
				invert = (x*y%2+x*y%3)%2 == 0
			case 7:
				invert = ((x+y)%2+x*y%3)%2 == 0
			default:
				panic("illegal mask value")
			}
			q.Modules[y][x] ^= module(bToI(invert && !q.IsFunction[y][x]))
		}
	}
}

// drawAlignmentPattern draws a 5*5 alignment pattern, with the center module at
// (x, y).
func (q *QRCode) drawAlignmentPattern(x, y int) {
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			q.setFunctionModule(x+dx, y+dy, max(abs(dx), abs(dy)) != 1)
		}
	}
}

// drawCodewords draws the given sequence of 8-bit codewords (data and error
// correction) onto the entire data area of this QR code. Function modules need
// to be marked off before this is called.
func (q *QRCode) drawCodewords(data []byte) {
	if len(data) != numRawDataModules[q.Version]/8 {
		panic("incorrect data length")
	}

	i := 0 // Bit index into the data.

	// Do the funny zig-zag scan.
	for right := q.Size - 1; right >= 1; right -= 2 {
		if right == 6 {
			right = 5
		}
		for vert := 0; vert < q.Size; vert++ {
			for j := 0; j < 2; j++ {
				x := right - j // Actual x coordinate.
				upward := (right+1)&2 == 0

				var y int
				if upward {
					y = q.Size - 1 - vert
				} else {
					y = vert
				} // Actual y coordinate.

				if !q.IsFunction[y][x] && i < len(data)*8 {
					q.Modules[y][x] = module(getBit(int(data[i>>3]), 7-(i&7)))
					i++
				}
				// If this QR code has any remainder bits (0 to 7), they were
				// assigned as 0/false/white during construction and are left
				// unchanged.
			}
		}
	}

	if i != len(data)*8 {
		panic("incorrect length")
	}
}

// drawFinderPattern draws a 9*9 finder pattern including the border separator,
// with the center module at (x, y).
func (q *QRCode) drawFinderPattern(x, y int) {
	for dy := -4; dy <= 4; dy++ {
		for dx := -4; dx <= 4; dx++ {
			dist := max(abs(dx), abs(dy))
			xx := x + dx
			yy := y + dy
			if 0 <= xx && xx < q.Size && 0 <= yy && yy < q.Size {
				q.setFunctionModule(xx, yy, dist != 2 && dist != 4)
			}
		}
	}
}

// drawFormatBits draws two cooies of the format bits (with its own error
// correction code), based on the given mask and this object's error correction
// level.
func (q *QRCode) drawFormatBits(mask Mask) {
	// Calculate error correction code and pack bits.
	data := q.ErrorCorrectionLevel.formatBits()<<3 | int(mask)
	rem := data
	for i := 0; i < 10; i++ {
		rem = rem<<1 ^ rem>>9*0x537
	}
	bits := data<<10 | rem ^ 0x5412
	if bits>>15 != 0 {
		panic("incorrect format bits calculation")
	}

	// Draw first copy.
	for i := 0; i <= 5; i++ {
		q.setFunctionModule(8, i, getBitAsBool(bits, i))
	}
	q.setFunctionModule(8, 7, getBitAsBool(bits, 6))
	q.setFunctionModule(8, 8, getBitAsBool(bits, 7))
	q.setFunctionModule(7, 8, getBitAsBool(bits, 8))
	for i := 9; i < 15; i++ {
		q.setFunctionModule(14-i, 8, getBitAsBool(bits, i))
	}

	// Draw second copy.
	for i := 0; i < 8; i++ {
		q.setFunctionModule(q.Size-1-i, 8, getBitAsBool(bits, i))
	}
	for i := 8; i < 15; i++ {
		q.setFunctionModule(8, q.Size-15+i, getBitAsBool(bits, i))
	}
	q.setFunctionModule(8, q.Size-8, true) // Always black.
}

// drawFunctionPatterns draws (set to black) all modules that correspond to
// "metadata" for the QR code symbol (non-data modules), such as finder
// patterns, version number, etc.
func (q *QRCode) drawFunctionPatterns() {
	// Draw horizontal and vertical timing patterns.
	for i := 0; i < q.Size; i++ {
		q.setFunctionModule(6, i, i%2 == 0)
		q.setFunctionModule(i, 6, i%2 == 0)
	}

	// Draw 3 finder patterns (all corners except the bottom right; overwrites
	// some timing modules).
	q.drawFinderPattern(3, 3)
	q.drawFinderPattern(q.Size-4, 3)
	q.drawFinderPattern(3, q.Size-4)

	// Draw alignment patterns.
	alignPatPos := alignmentPatternPositions[q.Version]
	numAlign := len(alignPatPos)
	for i := 0; i < numAlign; i++ {
		for j := 0; j < numAlign; j++ {
			// Do not draw on the three finder corners.
			if !(i == 0 && j == 0 || i == 0 && j == numAlign-1 || i == numAlign-1 && j == 0) {
				q.drawAlignmentPattern(int(alignPatPos[i]), int(alignPatPos[j]))
			}
		}
	}

	// Draw configuration data.
	q.drawFormatBits(0)
	q.drawVersion()
}

// drawVersion draws two copies of the version bits (with its own error
// correction code), based on this object's version, iff 7 <= version <= 40.
func (q *QRCode) drawVersion() {
	if q.Version < 7 {
		return
	}

	// Calculate error correction code and pack bits.
	rem := int(q.Version)
	for i := 0; i < 12; i++ {
		rem = rem<<1 ^ rem>>11*0x1F25
	}
	bits := int(q.Version)<<12 | rem
	if bits>>18 != 0 {
		panic("incorrect version calculation")
	}

	// Draw two copies.
	for i := 0; i < 18; i++ {
		bit := getBitAsBool(bits, i)
		a := q.Size - 11 + i%3
		b := i / 3
		q.setFunctionModule(a, b, bit)
		q.setFunctionModule(b, a, bit)
	}
}

// finderPenaltyAddHistory pushes the given value to the front and drops the
// last value.
func (q *QRCode) finderPenaltyAddHistory(currentRunLength int, runHistory *[7]int) {
	if runHistory[0] == 0 {
		currentRunLength += q.Size // Add white border to initial run.
	}

	copy(runHistory[1:], runHistory[0:])
	runHistory[0] = currentRunLength
}

// finderPenaltyCountPatterns finds patterns similar to the finder squares.
func (q *QRCode) finderPenaltyCountPatterns(runHistory *[7]int) int {
	n := runHistory[1]
	if n > q.Size*3 {
		panic("bad run history")
	}
	core := n > 0 && runHistory[2] == n && runHistory[3] == n*3 && runHistory[4] == n && runHistory[5] == n
	return bToI(core && runHistory[0] >= n*4 && runHistory[6] >= n) + bToI(core && runHistory[6] >= n*4 && runHistory[0] >= n)
}

// finderPenaltyTerminateAndCount adds the penalty at the end of a finder-like pattern.
func (q *QRCode) finderPenaltyTerminateAndCount(runColor module, runLength int, runHistory *[7]int) int {
	if runColor == 1 { // Terminate a black run.
		q.finderPenaltyAddHistory(runLength, runHistory)
		runLength = 0
	}
	runLength += q.Size // Add the white border to final run.
	q.finderPenaltyAddHistory(runLength, runHistory)
	return q.finderPenaltyCountPatterns(runHistory)
}

// getAlignmentPatternPositions returns an ascending list of positions of
// alignment patterns for this version number. Each position is in the range [0,
// 177), and are used on both the x and y axes.
func getAlignmentPatternPositions(version Version) []byte {
	if version == 1 {
		return []byte{}
	}

	numAlign := version/7 + 2
	var step int
	if version == 32 { // Special snowflake.
		step = 26
	} else { // step = ceil[(size - 13) / (numALign * 2 - 2)] * 2.
		step = (int(version)*4 + int(numAlign)*2 + 1) / (int(numAlign)*2 - 2) * 2
	}
	result := make([]byte, numAlign)
	result[0] = 6
	for i, pos := len(result)-1, int(version)*4+17-7; i >= 1; i-- {
		result[i] = byte(pos)
		pos -= step
	}

	return result
}

// getPenaltyScore calculates the penalty score based on the state of this QR
// code's current modules. Masking that results in lower penalties are designed
// to improve the chances of a scanner successfuly scanning the QR code.
func (q *QRCode) getPenaltyScore() int {
	result := 0

	// Adjacent modules in a row having the same color, and finder-like
	// patterns.
	for y := 0; y < q.Size; y++ {
		runColor := module(0)
		runX := 0
		var runHistory [7]int
		for x := 0; x < q.Size; x++ {
			if q.Modules[y][x] == runColor {
				runX++
				if runX == 5 {
					result += penaltyN1
				} else if runX > 5 {
					result++
				}
			} else {
				q.finderPenaltyAddHistory(runX, &runHistory)
				if runColor == 0 {
					result += q.finderPenaltyCountPatterns(&runHistory) * penaltyN3
				}
				runColor = q.Modules[y][x]
				runX = 1
			}
		}
		result += q.finderPenaltyTerminateAndCount(runColor, runX, &runHistory) * penaltyN3
	}

	// Adjacent modules in a column having the same color, and finder-like
	// patterns.
	for x := 0; x < q.Size; x++ {
		runColor := module(0)
		runY := 0
		var runHistory [7]int
		for y := 0; y < q.Size; y++ {
			if q.Modules[y][x] == runColor {
				runY++
				if runY == 5 {
					result += penaltyN1
				} else if runY > 5 {
					result++
				}
			} else {
				q.finderPenaltyAddHistory(runY, &runHistory)
				if runColor == 0 {
					result += q.finderPenaltyCountPatterns(&runHistory) * penaltyN3
				}
				runColor = q.Modules[y][x]
				runY = 1
			}
		}
		result += q.finderPenaltyTerminateAndCount(runColor, runY, &runHistory) * penaltyN3
	}

	// 2*2 blocks of modules having the same color.
	for y := 0; y < q.Size-1; y++ {
		for x := 0; x < q.Size-1; x++ {
			color := q.Modules[y][x]
			if color == q.Modules[y][x+1] &&
				color == q.Modules[y+1][x] &&
				color == q.Modules[y+1][x+1] {
				result += penaltyN2
			}
		}
	}

	// Balance of black and white modules.
	black := 0
	for _, rows := range q.Modules {
		for _, color := range rows {
			if color == 1 {
				black++
			}
		}
	}
	total := q.Size * q.Size // Note that the size is always odd, so black / total will never = 1/2.
	// Compute the smallest integer k >= 0 such that (45 - 5 * k)% <= black /
	// total <= (55 + 5 * k)%
	k := (abs(black*20-total*10)+total-1)/total - 1
	result += k * penaltyN4

	return result
}

// handleConstructorMasking is used during construction of the QR code
// structure. This method takes a given mask (or -1 for "auto") and applies the
// mask to the QR code. If auto is chosen, the method selects the mask that
// results in the lowest penalty.
func (q *QRCode) handleConstructorMasking(mask Mask) Mask {
	if mask == -1 { // Automatically choose the best mask.
		minPenalty := math.MaxInt32
		for i := Mask(0); i < 8; i++ {
			q.applyMask(i)
			q.drawFormatBits(i)
			penalty := q.getPenaltyScore()
			if penalty < minPenalty {
				mask = i
				minPenalty = penalty
			}
			q.applyMask(i) // Undoes the mask because of XOR.
		}
	}

	if mask < 0 || 7 < mask {
		panic("illegal mask value")
	}

	q.applyMask(mask)      // Apply the final choice of mask.
	q.drawFormatBits(mask) // Overwrite the old format bits.
	return mask
}

// reedSolomonComputeDivisor creates a Reed-Solomon error correction generator
// polynomial if the given degree.
func reedSolomonComputeDivisor(degree int) []byte {
	if degree < 1 || degree > 255 {
		panic("degree out of range")
	}

	// Polynomial coefficients are stored from highest to lowest power,
	// excluding the leading term, which is always 1. For example, the
	// polynomial x^3 + 255*x^2 + 8x + 93 is stored as the byte array [255, 8,
	// 93].
	result := make([]byte, degree)
	result[degree-1] = 1 // Start off with the monomial x^0.

	// Compute the product polunomial (x - r^0) * (x - r^1) * (x - r^2) * ... *
	// (x - r^(degree - 1)), and drop the heighest monomial term which is always
	// 1*x^degree. Note that r = 0x02, which is a generator element of this
	// field GF(^8/0x11D).
	root := byte(1)
	for i := 0; i < degree; i++ {
		// Multiply the current product by (x - r^i).
		for j := 0; j < len(result); j++ {
			result[j] = byte(reedSolomonMultiply(result[j], root))
			if j+1 < len(result) {
				result[j] ^= result[j+1]
			}
		}
		root = reedSolomonMultiply(root, 0x02)
	}

	return result
}

// reedSolomonMultiply returns the product of the two given field elements
// modulo GF(2^8/0x11D).
func reedSolomonMultiply(x, y byte) byte {
	// Russian peasant multiplication.
	z := 0
	for i := 7; i >= 0; i-- {
		z = z<<1 ^ z>>7*0x11D
		z ^= int(y >> i & 1 * x)
	}

	return byte(z)
}

// reedSolomonComputeRemainder returns the Reed-Solomon error correction
// codeword for the given data and divisor polynomials.
func reedSolomonComputeRemainder(data, divisor []byte) []byte {
	result := make([]byte, len(divisor))
	for _, b := range data { // Polynomial division.
		factor := b ^ result[0]
		copy(result[0:], result[1:])
		result[len(result)-1] = 0
		for i := 0; i < len(result); i++ {
			result[i] ^= byte(reedSolomonMultiply(divisor[i], factor))
		}
	}

	return result
}

func (q *QRCode) setFunctionModule(x, y int, isBlack bool) {
	q.Modules[y][x] = bToModule(isBlack)
	q.IsFunction[y][x] = true
}
