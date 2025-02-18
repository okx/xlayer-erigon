/*
   Copyright 2021 Erigon contributors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package common

import "slices"

type OrderedList[T any] struct {
	list        []T
	isOrdered   bool
	compareFunc func(a, b T) int
}

func (l *OrderedList[T]) Add(item T) {
	l.isOrdered = false
	l.list = append(l.list, item)
}

func (l *OrderedList[T]) Sort() {
	slices.SortFunc(l.list, l.compareFunc)
	l.isOrdered = true
}

func (l *OrderedList[T]) containsBinarySearch(item T) bool {
	upper := len(l.list)
	lower := 0
	for lower < upper {
		mid := (upper + lower) / 2
		cmp := l.compareFunc(item, l.list[mid])
		if cmp == 0 {
			return true
		}
		if cmp > 0 {
			lower = mid + 1
		} else {
			upper = mid
		}
	}
	return false
}

func (l *OrderedList[T]) containsLinear(item T) bool {
	for _, i := range l.list {
		if l.compareFunc(i, item) == 0 {
			return true
		}
	}
	return false
}

func (l *OrderedList[T]) Contains(item T) bool {
	if !l.isOrdered {
		l.containsLinear(item)
	}
	return l.containsBinarySearch(item)
}

func (l *OrderedList[T]) Size() int {
	return len(l.list)
}

func (l *OrderedList[T]) Items() []T {
	return l.list
}
