// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2019 Dapper Labs, Inc.
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
 *
 * Based on https://github.com/wk8/go-ordered-map, Copyright Jean Rougé
 *
 */

package sema

import "container/list"

// StringValueDeclarationOrderedMap
//
type StringValueDeclarationOrderedMap struct {
	pairs map[string]*StringValueDeclarationPair
	list  *list.List
}

// NewStringValueDeclarationOrderedMap creates a new StringValueDeclarationOrderedMap.
func NewStringValueDeclarationOrderedMap() *StringValueDeclarationOrderedMap {
	return &StringValueDeclarationOrderedMap{
		pairs: make(map[string]*StringValueDeclarationPair),
		list:  list.New(),
	}
}

// Clear removes all entries from this ordered map.
func (om *StringValueDeclarationOrderedMap) Clear() {
	om.list.Init()
	// NOTE: Range over map is safe, as it is only used to delete entries
	for key := range om.pairs { //nolint:maprangecheck
		delete(om.pairs, key)
	}
}

// Get returns the value associated with the given key.
// Returns nil if not found.
// The second return value indicates if the key is present in the map.
func (om *StringValueDeclarationOrderedMap) Get(key string) (result ValueDeclaration, present bool) {
	var pair *StringValueDeclarationPair
	if pair, present = om.pairs[key]; present {
		return pair.Value, present
	}
	return
}

// GetPair returns the key-value pair associated with the given key.
// Returns nil if not found.
func (om *StringValueDeclarationOrderedMap) GetPair(key string) *StringValueDeclarationPair {
	return om.pairs[key]
}

// Set sets the key-value pair, and returns what `Get` would have returned
// on that key prior to the call to `Set`.
func (om *StringValueDeclarationOrderedMap) Set(key string, value ValueDeclaration) (oldValue ValueDeclaration, present bool) {
	var pair *StringValueDeclarationPair
	if pair, present = om.pairs[key]; present {
		oldValue = pair.Value
		pair.Value = value
		return
	}

	pair = &StringValueDeclarationPair{
		Key:   key,
		Value: value,
	}
	pair.element = om.list.PushBack(pair)
	om.pairs[key] = pair

	return
}

// Delete removes the key-value pair, and returns what `Get` would have returned
// on that key prior to the call to `Delete`.
func (om *StringValueDeclarationOrderedMap) Delete(key string) (oldValue ValueDeclaration, present bool) {
	var pair *StringValueDeclarationPair
	pair, present = om.pairs[key]
	if !present {
		return
	}

	om.list.Remove(pair.element)
	delete(om.pairs, key)
	oldValue = pair.Value

	return
}

// Len returns the length of the ordered map.
func (om *StringValueDeclarationOrderedMap) Len() int {
	return len(om.pairs)
}

// Oldest returns a pointer to the oldest pair.
func (om *StringValueDeclarationOrderedMap) Oldest() *StringValueDeclarationPair {
	return listElementToStringValueDeclarationPair(om.list.Front())
}

// Newest returns a pointer to the newest pair.
func (om *StringValueDeclarationOrderedMap) Newest() *StringValueDeclarationPair {
	return listElementToStringValueDeclarationPair(om.list.Back())
}

// Foreach iterates over the entries of the map in the insertion order, and invokes
// the provided function for each key-value pair.
func (om *StringValueDeclarationOrderedMap) Foreach(f func(key string, value ValueDeclaration)) {
	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		f(pair.Key, pair.Value)
	}
}

// ForeachWithError iterates over the entries of the map in the insertion order,
// and invokes the provided function for each key-value pair.
// If the passed function returns an error, iteration breaks and the error is returned.
func (om *StringValueDeclarationOrderedMap) ForeachWithError(f func(key string, value ValueDeclaration) error) error {
	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		err := f(pair.Key, pair.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

// StringValueDeclarationPair
//
type StringValueDeclarationPair struct {
	Key   string
	Value ValueDeclaration

	element *list.Element
}

// Next returns a pointer to the next pair.
func (p *StringValueDeclarationPair) Next() *StringValueDeclarationPair {
	return listElementToStringValueDeclarationPair(p.element.Next())
}

// Prev returns a pointer to the previous pair.
func (p *StringValueDeclarationPair) Prev() *StringValueDeclarationPair {
	return listElementToStringValueDeclarationPair(p.element.Prev())
}

func listElementToStringValueDeclarationPair(element *list.Element) *StringValueDeclarationPair {
	if element == nil {
		return nil
	}
	return element.Value.(*StringValueDeclarationPair)
}
