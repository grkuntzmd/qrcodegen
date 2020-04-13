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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeparate(t *testing.T) {
	segs := MakeSegments("SUDOKU://")
	segs = append(segs, MakeSegments("007020004930000600600300000000000050200010008006900400003700900020050001000008000")...)
	qrCode, err := EncodeSegments(segs, Low)
	assert.Nil(t, err)
	_ = qrCode
	// str, err := qrCode.ToSVGString(2, false)
	// assert.Nil(t, err)
	// browser.OpenReader(strings.NewReader(str))
}

func TestSingle(t *testing.T) {
	qrCode, err := EncodeText("SUDOKU://007020004930000600600300000000000050200010008006900400003700900020050001000008000", Low)
	assert.Nil(t, err)
	_ = qrCode
	// str, err := qrCode.ToSVGString(2, false)
	// assert.Nil(t, err)
	// browser.OpenReader(strings.NewReader(str))
}
