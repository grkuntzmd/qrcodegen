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

type bitBuffer []byte

func (bb *bitBuffer) appendBits(value int, length int8) {
	if length > 31 || value>>length != 0 {
		panic("value out of range")
	}

	for i := length - 1; i >= 0; i-- { // Append data bit by bit.
		*bb = append(*bb, byte(value>>i&1))
	}
}
