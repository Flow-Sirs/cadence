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

package rlp_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/onflow/cadence/runtime/stdlib/rlp"
)

func TestRLPReadSize(t *testing.T) {
	tests := []struct {
		input          []byte
		startIndex     int
		isString       bool
		dataStartIndex int
		dataSize       int
		expectedErr    error
	}{
		// string test

		// empty data
		{[]byte{}, 0, false, 0, 0, rlp.ErrEmptyInput},
		// out of range index
		{[]byte{0x00}, 1, false, 0, 0, rlp.ErrInvalidStartIndex},
		// first char
		{[]byte{0x00}, 0, true, 0, 1, nil},
		// next char
		{[]byte{0x01}, 0, true, 0, 1, nil},
		// last char
		{[]byte{0x7f}, 0, true, 0, 1, nil},
		// empty string
		{[]byte{0x80}, 0, true, 1, 0, nil},
		// start of short string
		{[]byte{0x81}, 0, true, 1, 1, nil},
		// end of short string
		{[]byte{0xb7}, 0, true, 1, 55, nil},
		// start of long string (reading next byte to find out the size and decoding of that byte is smaller than 55)
		{[]byte{0xb8, 0x01}, 0, false, 0, 0, rlp.ErrNonCanonicalInput},
		{[]byte{0xb8, 0x37}, 0, false, 0, 0, rlp.ErrNonCanonicalInput},
		// not having the next byte to read
		{[]byte{0xb8}, 0, false, 0, 0, rlp.ErrIncompleteInput},
		// first valid long string entry (string len 56, first two bytes used for string size)
		{[]byte{0xb8, 0x38}, 0, true, 2, 56, nil},
		{[]byte{0xb8, 0x39}, 0, true, 2, 57, nil},
		// end of long string with only 1 extra byte (string len 255)
		{[]byte{0xb8, 0xff}, 0, true, 2, 255, nil},
		// long string (string len 258)
		{[]byte{0xb9, 0x01, 0x02}, 0, true, 3, 258, nil},
		// not enough bytes to read
		{[]byte{0xb9, 0x01}, 0, false, 0, 0, rlp.ErrIncompleteInput},
		// trailing zero bytes are not allowed (ie. the size has to be bigger than 255)
		{[]byte{0xb8, 0x00, 0xff}, 0, false, 0, 0, rlp.ErrNonCanonicalInput},
		// several bytes
		{[]byte{0xba, 0x01, 0x00, 0x00}, 0, true, 4, 65536, nil},
		// end of large string (max number of bytes)
		{[]byte{0xbf, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 0, true, 9, 9223372036854775807, nil},
		// we don't support data size larger than 9223372036854775807
		{[]byte{0xbf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 0, false, 0, 0, rlp.ErrDataSizeTooLarge},

		// list test

		// empty list
		{[]byte{0xc0}, 0, false, 1, 0, nil},
		// short list with 1 byte of data
		{[]byte{0xc1}, 0, false, 1, 1, nil},
		// short list with 55 bytes of data
		{[]byte{0xf7}, 0, false, 1, 55, nil},
		// start of long list (reading next byte to find out the size and decoding of that byte is smaller than 55)
		{[]byte{0xf8, 0x01}, 0, false, 0, 0, rlp.ErrNonCanonicalInput},
		{[]byte{0xf8, 0x37}, 0, false, 0, 0, rlp.ErrNonCanonicalInput},
		// not having the next byte to read
		{[]byte{0xf8}, 0, false, 0, 0, rlp.ErrIncompleteInput},
		// first valid long string entry (string len 56, first two bytes used for string size)
		{[]byte{0xf8, 0x38}, 0, false, 2, 56, nil},
		{[]byte{0xf8, 0x39}, 0, false, 2, 57, nil},
		// end of long string with only 1 extra byte (string len 255)
		{[]byte{0xf8, 0xff}, 0, false, 2, 255, nil},
		// long list (len 256) 2 extra bytes
		{[]byte{0xf9, 0x01, 0x00}, 0, false, 3, 256, nil},
		// not enough bytes to read
		{[]byte{0xf9, 0x01}, 0, false, 0, 0, rlp.ErrIncompleteInput},
		// trailing zero bytes are not allowed (ie. the size has to be bigger than 255)
		{[]byte{0xf8, 0x00, 0xff}, 0, false, 0, 0, rlp.ErrNonCanonicalInput},
		// several bytes
		{[]byte{0xfa, 0x01, 0x00, 0x00}, 0, false, 4, 65536, nil},
		// end of large list (max number of bytes)
		{[]byte{0xff, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 0, false, 9, rlp.MaxLongLengthAllowed, nil},
		// we don't support data size larger than 9223372036854775807
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 0, false, 0, 0, rlp.ErrDataSizeTooLarge},
	}

	for _, test := range tests {
		isString, dataStartIndex, dataSize, err := rlp.ReadSize(test.input, test.startIndex)
		if test.expectedErr != nil {
			require.Error(t, err)
			require.Equal(t, test.expectedErr, err)
		} else {
			require.NoError(t, err)
		}
		require.Equal(t, test.isString, isString)
		require.Equal(t, test.dataStartIndex, dataStartIndex)
		require.Equal(t, test.dataSize, dataSize)
	}

}

func TestDecodeString(t *testing.T) {
	tests := []struct {
		expectedOutput []byte
		encoded        []byte
		expectedErr    error
	}{
		{
			[]byte(""), // empty string
			[]byte{0x80},
			nil,
		},
		{
			nil,
			[]byte{0xc0}, // empty list
			rlp.ErrTypeMismatch,
		},
		{
			[]byte("A"),
			[]byte{0x41},
			nil,
		},
		{ // extra data for char
			[]byte("A"),
			[]byte{0x41, 0x01},
			rlp.ErrInputContainsExtraBytes,
		},
		{
			[]byte("dog"),
			[]byte{0x83, 0x64, 0x6f, 0x67},
			nil,
		},
		{
			nil,
			[]byte{0x83}, // requires data bytes
			rlp.ErrIncompleteInput,
		},
		{
			nil,
			[]byte{0x83, 0x64, 0x6f}, // requires 4 bytes
			rlp.ErrIncompleteInput,
		},
		{
			nil,
			[]byte{0x83, 0x64, 0x6f, 0x67, 0x01}, // an extra byte
			rlp.ErrInputContainsExtraBytes,
		},
		{
			[]byte("this is a test lo0o0o0o0o0ong string with 55 characters"),
			[]byte{0xb7, // one byte size
				0x74, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20,
				0x61, 0x20, 0x74, 0x65, 0x73, 0x74, 0x20, 0x6c,
				0x6f, 0x30, 0x6f, 0x30, 0x6f, 0x30, 0x6f, 0x30,
				0x6f, 0x30, 0x6f, 0x6e, 0x67, 0x20, 0x73, 0x74,
				0x72, 0x69, 0x6e, 0x67, 0x20, 0x77, 0x69, 0x74,
				0x68, 0x20, 0x35, 0x35, 0x20, 0x63, 0x68, 0x61,
				0x72, 0x61, 0x63, 0x74, 0x65, 0x72, 0x73},
			nil,
		},
		{
			[]byte("this is a test lo0o0o0o0o0o0ng string with 56 characters"),
			[]byte{0xb8, 0x38, // an extra byte for size
				0x74, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20,
				0x61, 0x20, 0x74, 0x65, 0x73, 0x74, 0x20, 0x6c,
				0x6f, 0x30, 0x6f, 0x30, 0x6f, 0x30, 0x6f, 0x30,
				0x6f, 0x30, 0x6f, 0x30, 0x6e, 0x67, 0x20, 0x73,
				0x74, 0x72, 0x69, 0x6e, 0x67, 0x20, 0x77, 0x69,
				0x74, 0x68, 0x20, 0x35, 0x36, 0x20, 0x63, 0x68,
				0x61, 0x72, 0x61, 0x63, 0x74, 0x65, 0x72, 0x73},
			nil,
		},
		{
			[]byte("Lorem ipsum dolor sit amet, consectetuer adipiscing elit. " +
				"Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus" +
				" et magnis dis parturient montes, nascetur ridiculous mus."),
			[]byte{0xb8, 0xc0, // two byte sizes
				0x4c, 0x6f, 0x72, 0x65, 0x6d, 0x20, 0x69, 0x70, 0x73, 0x75,
				0x6d, 0x20, 0x64, 0x6f, 0x6c, 0x6f, 0x72, 0x20, 0x73, 0x69,
				0x74, 0x20, 0x61, 0x6d, 0x65, 0x74, 0x2c, 0x20, 0x63, 0x6f,
				0x6e, 0x73, 0x65, 0x63, 0x74, 0x65, 0x74, 0x75, 0x65, 0x72,
				0x20, 0x61, 0x64, 0x69, 0x70, 0x69, 0x73, 0x63, 0x69, 0x6e,
				0x67, 0x20, 0x65, 0x6c, 0x69, 0x74, 0x2e, 0x20, 0x41, 0x65,
				0x6e, 0x65, 0x61, 0x6e, 0x20, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
				0x64, 0x6f, 0x20, 0x6c, 0x69, 0x67, 0x75, 0x6c, 0x61, 0x20,
				0x65, 0x67, 0x65, 0x74, 0x20, 0x64, 0x6f, 0x6c, 0x6f, 0x72,
				0x2e, 0x20, 0x41, 0x65, 0x6e, 0x65, 0x61, 0x6e, 0x20, 0x6d,
				0x61, 0x73, 0x73, 0x61, 0x2e, 0x20, 0x43, 0x75, 0x6d, 0x20,
				0x73, 0x6f, 0x63, 0x69, 0x69, 0x73, 0x20, 0x6e, 0x61, 0x74,
				0x6f, 0x71, 0x75, 0x65, 0x20, 0x70, 0x65, 0x6e, 0x61, 0x74,
				0x69, 0x62, 0x75, 0x73, 0x20, 0x65, 0x74, 0x20, 0x6d, 0x61,
				0x67, 0x6e, 0x69, 0x73, 0x20, 0x64, 0x69, 0x73, 0x20, 0x70,
				0x61, 0x72, 0x74, 0x75, 0x72, 0x69, 0x65, 0x6e, 0x74, 0x20,
				0x6d, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x2c, 0x20, 0x6e, 0x61,
				0x73, 0x63, 0x65, 0x74, 0x75, 0x72, 0x20, 0x72, 0x69, 0x64,
				0x69, 0x63, 0x75, 0x6c, 0x6f, 0x75, 0x73, 0x20, 0x6d, 0x75,
				0x73, 0x2e},
			nil,
		},
		{
			[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
				"Sed imperdiet odio a nibh rutrum blandit. Phasellus porta " +
				"eleifend tellus non consequat. Donec sodales velit in tortor " +
				"iaculis, sollicitudin dignissim orci maximus. Nunc at est sem. Sed congue proin."),
			[]byte{0xb9, 0x01, 0x00, // three bytes for size (256 chars) - checks big endian encoding
				0x4c, 0x6f, 0x72, 0x65, 0x6d, 0x20, 0x69, 0x70, 0x73, 0x75, 0x6d,
				0x20, 0x64, 0x6f, 0x6c, 0x6f, 0x72, 0x20, 0x73, 0x69, 0x74, 0x20,
				0x61, 0x6d, 0x65, 0x74, 0x2c, 0x20, 0x63, 0x6f, 0x6e, 0x73, 0x65,
				0x63, 0x74, 0x65, 0x74, 0x75, 0x72, 0x20, 0x61, 0x64, 0x69, 0x70,
				0x69, 0x73, 0x63, 0x69, 0x6e, 0x67, 0x20, 0x65, 0x6c, 0x69, 0x74,
				0x2e, 0x20, 0x53, 0x65, 0x64, 0x20, 0x69, 0x6d, 0x70, 0x65, 0x72,
				0x64, 0x69, 0x65, 0x74, 0x20, 0x6f, 0x64, 0x69, 0x6f, 0x20, 0x61,
				0x20, 0x6e, 0x69, 0x62, 0x68, 0x20, 0x72, 0x75, 0x74, 0x72, 0x75,
				0x6d, 0x20, 0x62, 0x6c, 0x61, 0x6e, 0x64, 0x69, 0x74, 0x2e, 0x20,
				0x50, 0x68, 0x61, 0x73, 0x65, 0x6c, 0x6c, 0x75, 0x73, 0x20, 0x70,
				0x6f, 0x72, 0x74, 0x61, 0x20, 0x65, 0x6c, 0x65, 0x69, 0x66, 0x65,
				0x6e, 0x64, 0x20, 0x74, 0x65, 0x6c, 0x6c, 0x75, 0x73, 0x20, 0x6e,
				0x6f, 0x6e, 0x20, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x71, 0x75, 0x61,
				0x74, 0x2e, 0x20, 0x44, 0x6f, 0x6e, 0x65, 0x63, 0x20, 0x73, 0x6f,
				0x64, 0x61, 0x6c, 0x65, 0x73, 0x20, 0x76, 0x65, 0x6c, 0x69, 0x74,
				0x20, 0x69, 0x6e, 0x20, 0x74, 0x6f, 0x72, 0x74, 0x6f, 0x72, 0x20,
				0x69, 0x61, 0x63, 0x75, 0x6c, 0x69, 0x73, 0x2c, 0x20, 0x73, 0x6f,
				0x6c, 0x6c, 0x69, 0x63, 0x69, 0x74, 0x75, 0x64, 0x69, 0x6e, 0x20,
				0x64, 0x69, 0x67, 0x6e, 0x69, 0x73, 0x73, 0x69, 0x6d, 0x20, 0x6f,
				0x72, 0x63, 0x69, 0x20, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x73,
				0x2e, 0x20, 0x4e, 0x75, 0x6e, 0x63, 0x20, 0x61, 0x74, 0x20, 0x65,
				0x73, 0x74, 0x20, 0x73, 0x65, 0x6d, 0x2e, 0x20, 0x53, 0x65, 0x64,
				0x20, 0x63, 0x6f, 0x6e, 0x67, 0x75, 0x65, 0x20, 0x70, 0x72, 0x6f,
				0x69, 0x6e, 0x2e},
			nil,
		},
	}

	for _, test := range tests {
		item, err := rlp.DecodeString(test.encoded, 0)
		if test.expectedErr != nil {
			require.Equal(t, test.expectedErr, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, item, test.expectedOutput)
		}
	}
}

func TestDecodeList(t *testing.T) {
	tests := []struct {
		expectedItems [][]byte
		encoded       []byte
		expectedErr   error
	}{
		{
			[][]byte{},
			[]byte{0xc0}, // empty list
			nil,
		},
		{
			nil,
			[]byte{0x80}, // empty string
			rlp.ErrTypeMismatch,
		},
		{
			[][]byte{{}}, // list with an empty list
			[]byte{0xc1, 0xc0},
			nil,
		},
		{
			[][]byte{{}, {}, {}}, // list with several empty list
			[]byte{0xc3, 0xc0, 0xc0, 0xc0},
			nil,
		},
		{
			[][]byte{[]byte("A")}, // single element
			[]byte{0xc1, 0x41},
			nil,
		},
		{
			[][]byte{
				[]byte("ABCDEFG"),
				[]byte("HIJKLMN"),
			}, // two short string elements
			[]byte{
				0xd0,                                     // 1 byte size
				0x87,                                     // size of string
				0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, // content
				0x87,                                     // size of string
				0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, // content
			},
			nil,
		},
		{
			nil,
			[]byte{
				0xcf,                                     // 1 byte size
				0x87,                                     // size of string
				0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, // content
				0x87,                                     // size of string
				0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, // content
			},
			rlp.ErrListSizeMismatch,
		},
		{
			nil,
			[]byte{
				0xd0,                                     // 1 byte size
				0x87,                                     // size of string
				0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, // content
				0x87,                                     // size of string
				0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, // content
				0x01,
			},
			rlp.ErrInputContainsExtraBytes,
		},
		{
			[][]byte{
				[]byte("AB"),
				[]byte("CD"),
				[]byte("EF"),
				[]byte(""),
				[]byte("GH"),
				[]byte(""),
			},
			[]byte{
				0xce, // one byte size
				0x82, 0x41, 0x42,
				0x82, 0x43, 0x44,
				0x82, 0x45, 0x46,
				0x80,
				0x82, 0x47, 0x48,
				0x80,
			},
			nil,
		},

		{
			[][]byte{
				[]byte("A"),
				[]byte("AB"),
				[]byte("ABC"),
				[]byte("ABCD"),
				[]byte("ABCDE"),
				[]byte("ABCDEF"),
				[]byte("ABCDEFG"),
				[]byte("ABCDEFGH"),
				[]byte("ABCDEFGHI"),
				[]byte("ABCDEFGHIJ"),
				[]byte("ABCDEFGHIJK"),
				[]byte("ABCDEFGHIJKL"),
				[]byte("ABCDEFGHIJKLM"),
				[]byte("ABCDEFGHIJKLMN"),
				[]byte("ABCDEFGHIJKLMNO"),
				[]byte("ABCDEFGHIJKLMNOP"),
				[]byte("ABCDEFGHIJKLMNOPQ"),
				[]byte("ABCDEFGHIJKLMNOPQR"),
				[]byte("ABCDEFGHIJKLMNOPQRS"),
				[]byte("ABCDEFGHIJKLMNOPQRST"),
				[]byte("ABCDEFGHIJKLMNOPQRSTU"),
				[]byte("ABCDEFGHIJKLMNOPQRSTUV"),
				[]byte("ABCDEFGHIJKLMNOPQRSTUVW"),
				[]byte("ABCDEFGHIJKLMNOPQRSTUVWX"),
				[]byte("ABCDEFGHIJKLMNOPQRSTUVWXY"),
				[]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ"),
			},
			[]byte{
				0xf9, 0x01, // two byte size
				0x78, 0x41,
				0x82, 0x41, 0x42,
				0x83, 0x41, 0x42, 0x43,
				0x84, 0x41, 0x42, 0x43, 0x44,
				0x85, 0x41, 0x42, 0x43, 0x44, 0x45,
				0x86, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46,
				0x87, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47,
				0x88, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48,
				0x89, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49,
				0x8a, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a,
				0x8b, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b,
				0x8c, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c,
				0x8d, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d,
				0x8e, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e,
				0x8f, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f,
				0x90, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50,
				0x91, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51,
				0x92, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51, 0x52,
				0x93, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51, 0x52, 0x53,
				0x94, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51, 0x52, 0x53, 0x54,
				0x95, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51, 0x52, 0x53, 0x54, 0x55,
				0x96, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56,
				0x97, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57,
				0x98, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58,
				0x99, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59,
				0x9a, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a,
			},
			nil,
		},
	}

	for _, test := range tests {
		item, err := rlp.DecodeList(test.encoded, 0)
		if test.expectedErr != nil {
			require.Equal(t, test.expectedErr, err)
		} else {
			require.NoError(t, err)
			for i, expectedItem := range test.expectedItems {
				require.Equal(t, item[i], expectedItem)
			}
		}
	}
}
