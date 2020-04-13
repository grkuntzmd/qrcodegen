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
