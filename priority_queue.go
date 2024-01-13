// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package jitterbuffer

import (
	"errors"
	"github.com/pion/rtp"
)

type PriorityQueue struct {
	next   *node
	length uint16
}

type node struct {
	val  *rtp.Packet
	next *node
	prev *node
	prio uint16
}

func NewQueue() *PriorityQueue {
	return &PriorityQueue{
		next:   nil,
		length: 0,
	}
}

func newNode(val *rtp.Packet, prio uint16) *node {
	return &node{
		val:  val,
		prev: nil,
		next: nil,
		prio: prio,
	}
}
func (q *PriorityQueue) Find(sqNum uint16) (*rtp.Packet, error) {
	if q.next.prio == sqNum {
		return q.next.val, nil
	}

	if sqNum < q.next.prio {
		return nil, errors.New("No previous sequence")
	}
	next := q.next
	for next != nil {
		if next.prio == sqNum {
			return next.val, nil
		}
		next = next.next
	}
	return nil, errors.New("Priority item not found")
}

func (q *PriorityQueue) Push(val *rtp.Packet, prio uint16) {
	newPq := newNode(val, prio)
	if q.next == nil {
		q.next = newPq
		q.length++
		return
	}
	if prio < q.next.prio {
		newPq.next = q.next
		q.next.prev = newPq
		q.next = newPq
		q.length++
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
	q.length++
}

func (q *PriorityQueue) Length() uint16 {
	return q.length
}

func (q *PriorityQueue) Pop() (*rtp.Packet, error) {
	if q.next == nil {
		return nil, errors.New("Attempt to pop without a current value")
	}
	val := q.next.val
	q.length--
	q.next = q.next.next
	return val, nil
}

func (q *PriorityQueue) PopAt(sqNum uint16) (*rtp.Packet, error) {
	if q.next == nil {
		return nil, errors.New("Attempt to pop without a current value")
	}
	if q.next.prio == sqNum {
		val := q.next.val
		q.next = q.next.next
		return val, nil
	}
	pos := q.next
	prev := q.next.prev
	for pos != nil {
		if pos.prio == sqNum {
			val := pos.val
			prev.next = pos.next
			if prev.next != nil {
				prev.next.prev = prev
			}
			return val, nil
		}
		prev = pos
		pos = pos.next
	}
	return nil, errors.New("sequence not found")
}

func (q *PriorityQueue) PopAtTimestamp(timestamp uint32) (*rtp.Packet, error) {
	if q.next == nil {
		return nil, errors.New("Attempt to pop without a current value")
	}
	if q.next.val.Timestamp == timestamp {
		val := q.next.val
		q.next = q.next.next
		return val, nil
	}
	pos := q.next
	prev := q.next.prev
	for pos != nil {
		if pos.val.Timestamp == timestamp{
			val := pos.val
			prev.next = pos.next
			if prev.next != nil {
				prev.next.prev = prev
			}
			return val, nil
		}
		prev = pos
		pos = pos.next
	}
	return nil, errors.New("timestamp not found")
}
