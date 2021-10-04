/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package interpreter

import (
	"github.com/onflow/atree"
)

// HashableValue is an immutable value that can be hashed
//
type HashableValue interface {
	Value
	HashInput(interpreter *Interpreter, getLocationRange func() LocationRange, scratch []byte) []byte
}

func newHashInputProvider(interpreter *Interpreter, getLocationRange func() LocationRange) atree.HashInputProvider {
	return func(value atree.Value, scratch []byte) ([]byte, error) {
		hashInput := MustConvertStoredValue(value).(HashableValue).
			HashInput(interpreter, getLocationRange, scratch)
		return hashInput, nil
	}
}

// !!! *WARNING* !!!
//
// Only add new types by:
// - replacing existing placeholders (`_`) with new types
// - appending new types
//
// Only remove types by:
// - replace existing types with a placeholder `_`
//
// DO *NOT* REPLACE EXISTING TYPES!
// DO *NOT* ADD NEW TYPES IN BETWEEN!

const (
	HashInputTypeBool = iota
	HashInputTypeString
	HashInputTypeEnum
	HashInputTypeAddress
	HashInputTypePath
	_
	_
	_
	_
	_
	// Int*
	HashInputTypeInt
	HashInputTypeInt8
	HashInputTypeInt16
	HashInputTypeInt32
	HashInputTypeInt64
	HashInputTypeInt128
	HashInputTypeInt256
	_

	// UInt*
	HashInputTypeUInt
	HashInputTypeUInt8
	HashInputTypeUInt16
	HashInputTypeUInt32
	HashInputTypeUInt64
	HashInputTypeUInt128
	HashInputTypeUInt256
	_

	// Word*
	_
	HashInputTypeWord8
	HashInputTypeWord16
	HashInputTypeWord32
	HashInputTypeWord64
	_ // future: Word128
	_ // future: Word256
	_

	// Fix*
	_
	_ // future: Fix8
	_ // future: Fix16
	_ // future: Fix32
	HashInputTypeFix64
	_ // future: Fix128
	_ // future: Fix256
	_

	// UFix*
	_
	_ // future: UFix8
	_ // future: UFix16
	_ // future: UFix32
	HashInputTypeUFix64
	_ // future: UFix128
	_ // future: UFix256
	_
)