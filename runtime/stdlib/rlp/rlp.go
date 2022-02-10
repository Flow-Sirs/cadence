/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2022 Dapper Labs, Inc.
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

package rlp

import (
	"encoding/binary"
	"errors"
	"math"
)

const (
	ByteRangeStart        = 0x00 // not in use, here only for inclusivity
	ByteRangeEnd          = 0x7f
	ShortStringRangeStart = 0x80
	ShortStringRangeEnd   = 0xb7
	LongStringRangeStart  = 0xb8
	LongStringRangeEnd    = 0xbf
	ShortListRangeStart   = 0xc0
	ShortListRangeEnd     = 0xf7
	LongListRangeStart    = 0xf8
	LongListRangeEnd      = 0xff // not in use, here only for inclusivity
	MaxShortLengthAllowed = 55
	MaxLongLengthAllowed  = math.MaxInt64
)

var (
	ErrEmptyInput              = errors.New("input data is empty")
	ErrInvalidStartIndex       = errors.New("invalid start index")
	ErrIncompleteInput         = errors.New("incomplete input! not enough bytes to read")
	ErrNonCanonicalInput       = errors.New("non-canonical encoded input")
	ErrDataSizeTooLarge        = errors.New("data size is larger than what is supported")
	ErrListSizeMismatch        = errors.New("list size doesn't match the size of items")
	ErrInputContainsExtraBytes = errors.New("input contains extra bytes")
	ErrTypeMismatch            = errors.New("type extracted from input doesn't match the function")
)

func ReadSize(inp []byte, startIndex int) (isString bool, dataStartIndex, dataSize int, err error) {
	if len(inp) == 0 {
		return false, 0, 0, ErrEmptyInput
	}

	// check startIndex is in the range
	if startIndex >= len(inp) {
		return false, 0, 0, ErrInvalidStartIndex
	}

	firstByte := inp[startIndex]
	startIndex++

	// single character space - first byte holds the data itslef
	if firstByte <= ByteRangeEnd {
		return true, startIndex - 1, 1, nil
	}

	// short string space (0-55 bytes long string)
	// firstByte minus the start range for the short string returns the data size
	// valid range of firstByte is [0x80, 0xB7].
	if firstByte <= ShortStringRangeEnd {
		strLen := uint(firstByte - ShortStringRangeStart)
		return true, startIndex, int(strLen), nil
	}

	// short list space
	// firstByte minus the start range for the short list would return the data size
	if firstByte >= ShortListRangeStart && firstByte <= ShortListRangeEnd {
		strLen := uint(firstByte - ShortListRangeStart)
		return false, startIndex, int(strLen), nil
	}

	// string and list long space

	var bytesToReadForLen uint
	// long string mode (55+ long strings)
	// firstByte minus the end range of short string, returns the number of bytes
	if firstByte >= LongStringRangeStart && firstByte <= LongStringRangeEnd {
		bytesToReadForLen = uint(firstByte - ShortStringRangeEnd)
		isString = true
	}

	// long list mode
	if firstByte >= LongListRangeStart {
		bytesToReadForLen = uint(firstByte - ShortListRangeEnd)
		isString = false
	}

	// check atleast there is one more byte to read
	if startIndex >= len(inp) {
		return false, 0, 0, ErrIncompleteInput
	}

	// bytesToReadForLen with value of zero never happens
	// optimization for a single extra byte for size
	if bytesToReadForLen == 1 {
		strLen := uint(inp[startIndex])
		startIndex++
		if strLen <= MaxShortLengthAllowed {
			// encoding is not canonical, unnecessary bytes used for encoding
			// should have encoded as a short string
			return false, 0, 0, ErrNonCanonicalInput
		}
		return isString, startIndex, int(strLen), nil
	}

	// several bytes case

	// allocate 8 bytes
	lenData := make([]byte, 8)
	// but copy to lower part only
	// note that its not possible for bytesToReadForLen to go beyond 8
	start := int(8 - bytesToReadForLen)

	// if any trailing zero bytes, unnecessary bytes were used for encoding
	// checking only the first byte is sufficient
	if inp[startIndex] == 0 {
		return false, 0, 0, ErrNonCanonicalInput
	}

	endIndex := startIndex + int(bytesToReadForLen)
	// check endIndex is in the range
	if endIndex > len(inp) {
		return false, 0, 0, ErrIncompleteInput
	}

	copy(lenData[start:], inp[startIndex:endIndex])
	startIndex += int(bytesToReadForLen)
	strLen := uint(binary.BigEndian.Uint64(lenData))

	if strLen <= MaxShortLengthAllowed {
		// encoding is not canonical, unnecessary bytes used for encoding
		// should have encoded as a short string
		return false, 0, 0, ErrNonCanonicalInput
	}
	if strLen > MaxLongLengthAllowed {
		return false, 0, 0, ErrDataSizeTooLarge
	}
	return isString, startIndex, int(strLen), nil
}

func DecodeString(inp []byte, startIndex int) (str []byte, err error) {
	// read data size info
	isString, dataStartIndex, dataSize, err := ReadSize(inp, startIndex)
	if err != nil {
		return nil, err
	}
	// check type
	if !isString {
		return nil, ErrTypeMismatch
	}
	// single character special case
	if dataSize == 1 && startIndex == dataStartIndex {
		if len(inp) > 1 {
			return nil, ErrInputContainsExtraBytes
		}
		return []byte{inp[dataStartIndex]}, nil
	}

	// collect and return string
	dataEndIndex := dataStartIndex + dataSize
	if dataEndIndex > len(inp) {
		return nil, ErrIncompleteInput
	}

	if len(inp) > dataEndIndex {
		return nil, ErrInputContainsExtraBytes
	}

	return inp[dataStartIndex:dataEndIndex], nil
}

func DecodeList(inp []byte, startIndex int) (encodedItems [][]byte, err error) {
	// read data size info
	isString, dataStartIndex, listDataSize, err := ReadSize(inp, startIndex)
	if err != nil {
		return nil, err
	}

	// check type
	if isString {
		return nil, ErrTypeMismatch
	}

	itemStartIndex := dataStartIndex
	bytesRead := 0
	retList := make([][]byte, 0)

	for bytesRead < listDataSize {

		_, itemDataStartIndex, itemSize, err := ReadSize(inp, itemStartIndex)
		if err != nil {
			return nil, err
		}
		// collect encoded item
		itemEndIndex := itemDataStartIndex + itemSize
		if itemEndIndex > len(inp) {
			return nil, ErrIncompleteInput
		}
		retList = append(retList, inp[itemDataStartIndex:itemEndIndex])
		bytesRead += itemEndIndex - itemStartIndex
		itemStartIndex = itemEndIndex
	}
	if bytesRead != listDataSize {
		return nil, ErrListSizeMismatch
	}

	if len(inp) > itemStartIndex {
		return nil, ErrInputContainsExtraBytes
	}

	return retList, nil
}