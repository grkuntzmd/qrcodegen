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

// segmentEncoder contains options for EncodeSegments.
type segmentEncoder struct {
	boostECL   bool // Boost error correction level if there is still room in the QR code version that has been chosen.
	mask       Mask
	maxVersion Version
	minVersion Version
}

// WithAutoMask sets the mask value to automatic selection on a segment
// encoding.
func WithAutoMask() func(*segmentEncoder) {
	return func(s *segmentEncoder) {
		s.mask = -1
	}
}

// WithBoostECL causes the segment encoding to automatic increase the error
// correction level if there is room in the chosen version.
func WithBoostECL(boost bool) func(*segmentEncoder) {
	return func(s *segmentEncoder) {
		s.boostECL = boost
	}
}

// WithMask sets the mask value on a segment encoding.
func WithMask(mask Mask) func(*segmentEncoder) {
	return func(s *segmentEncoder) {
		s.mask = mask
	}
}

// WithMaxVersion sets the maximum allows version on a segment encoding.
func WithMaxVersion(version Version) func(*segmentEncoder) {
	return func(s *segmentEncoder) {
		s.minVersion = version
	}
}

// WithMinVersion sets the minimum allows version on a segment encoding.
func WithMinVersion(version Version) func(*segmentEncoder) {
	return func(s *segmentEncoder) {
		s.minVersion = version
	}
}
