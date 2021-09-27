/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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
	"errors"
	"math"

	"github.com/onflow/atree"
	"github.com/onflow/cadence/runtime/common"
)

func ByteArrayValueToByteSlice(value Value) ([]byte, error) {
	array, ok := value.(*ArrayValue)
	if !ok {
		return nil, errors.New("value is not an array")
	}

	result := make([]byte, array.Count())

	iterator, err := array.array.Iterator()
	if err != nil {
		return nil, ExternalError{err}
	}

	index := 0
	for {
		var atreeValue atree.Value
		atreeValue, err = iterator.Next()
		if err != nil {
			return nil, ExternalError{err}
		}
		if atreeValue == nil {
			return result, nil
		}

		// atree.Array iterator returns low-level atree.Value,
		// convert to high-level interpreter.Value
		value := MustConvertStoredValue(atreeValue)

		b, err := ByteValueToByte(value)
		if err != nil {
			return nil, err
		}

		result[index] = b
		index++
	}
}

func ByteValueToByte(element Value) (byte, error) {
	var b byte

	switch element := element.(type) {
	case BigNumberValue:
		bigInt := element.ToBigInt()
		if !bigInt.IsUint64() {
			return 0, errors.New("value is not in byte range (0-255)")
		}

		integer := bigInt.Uint64()

		if integer > math.MaxUint8 {
			return 0, errors.New("value is not in byte range (0-255)")
		}

		b = byte(integer)

	case NumberValue:
		integer := element.ToInt()

		if integer < 0 || integer > math.MaxUint8 {
			return 0, errors.New("value is not in byte range (0-255)")
		}

		b = byte(integer)

	default:
		return 0, errors.New("value is not an integer")
	}

	return b, nil
}

var byteArrayTypeInfo = encodeArrayTypeInfo(ByteArrayStaticType)

func ByteSliceToByteArrayValue(interpreter *Interpreter, buf []byte) *ArrayValue {
	values := make([]Value, len(buf))
	for i, b := range buf {
		values[i] = UInt8Value(b)
	}

	return NewArrayValueWithTypeInfo(
		interpreter,
		ByteArrayStaticType,
		common.Address{},
		byteArrayTypeInfo,
		values...,
	)
}
