// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package jitterbuffer

import (
	"errors"
)

type PQ[T any] struct {
  next *node[T]
  length uint16
}

type node[T any] struct {
	val *T
	next  *node[T]
	prev  *node[T]
	prio  uint16
}

func NewQueue[T any]() *PQ[T] {
	return &PQ[T]{
		next:   nil,
    length: 0,
	}
}

func newNode[T any](val *T, prio uint16) *node[T] {
	return &node[T]{
    val: val,
		prev:   nil,
		next:   nil,
		prio:   prio,
	}
}
func (q *PQ[T]) Find(sqNum uint16) (*T, error) {
	if q.next.prio == sqNum {
		return q.next.val, nil
	}

	if sqNum < q.next.prio {
	  return nil, errors.New("No previous sequence")
	}
  next:= q.next
  for next!= nil {
    if next.prio == sqNum {
      return next.val, nil
    }
    next = next.next
  }
  return nil, errors.New("Priority item not found")
}

func (q *PQ[T]) Push(val *T, prio uint16) {
	newPq := newNode(val, prio)
  if q.next == nil {
    q.next = newPq
    q.length ++
    return
  }
  if prio < q.next.prio {
    newPq.next = q.next
    q.next.prev = newPq
    q.next = newPq
    q.length ++
    return
  }
	head := q.next
	prev := q.next
	for head != nil {
		if prio <= head.prio {
			break
		}
		prev = head
		head = head.next
	}
	if head == nil {
		if prev != nil {
      prev.next = newPq
    }
		newPq.prev = prev
	} else {
		newPq.next = head
		newPq.prev = prev
		if prev != nil {
			prev.next = newPq
		}
		head.prev = newPq
	}
  q.length ++
}

func (q *PQ[T]) Length() uint16 {
	return q.length
}

func (q *PQ[T]) Pop() (*T, error) {
  if q.next == nil {
    return nil,  errors.New("Attempt to pop without a current value")
  }
  val := q.next.val
  q.length --
  q.next = q.next.next
  return val, nil
}
