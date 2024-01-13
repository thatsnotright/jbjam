// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package jitterbuffer

import (
	"errors"
	"github.com/pion/rtp"
	"math"
	"sync"
)

type JitterBufferState uint16

type JitterBufferEvent string

const (
	Buffering JitterBufferState = iota
	Emitting
)
const (
	StartBuffering  JitterBufferEvent = "startBuffering"
	BeginPlayback                     = "playing"
	BufferUnderflow                   = "underflow"
	BufferOverflow                    = "overflow"
)

func (jbs JitterBufferState) String() string {
	switch jbs {
	case Buffering:
		return "Buffering"
	case Emitting:
		return "Emitting"
	}
	return "unknown"
}

type JitterBufferOption struct {
	initialLatency uint32
}
type Option func(jb *JitterBuffer)
type JitterBufferEventListener func(event JitterBufferEvent, jb *JitterBuffer)

type JitterBuffer struct {
	packets             *PriorityQueue
	last_sequence       uint16
	playout_head        uint16
	playout_ready       bool
	state               JitterBufferState
	sample_rate         uint16
	payload_sample_rate int
	max_depth           int
	stats               JitterBufferStats
	listeners           []JitterBufferEventListener
	mutex               sync.Mutex
}

type JitterBufferStats struct {
	out_of_order_count uint32
	empty_count        uint32
	underflow_count    uint32
	overflow_count     uint32
	jitter             float32
	max_jitter         float32
}

// New will initialize a jitter buffer and its associated statistics
func New(opts ...Option) *JitterBuffer {
	jb := &JitterBuffer{state: Buffering, stats: JitterBufferStats{0, 0, 0, 0, .0, .0}, packets: NewQueue()}
	for _, o := range opts {
		o(jb)
	}
	return jb
}

// The jitter buffer may emit events correspnding, interested listerns should
// look at JitterBufferEvent for available events
func (jb *JitterBuffer) Listen(event JitterBufferEvent, cb JitterBufferEventListener) {
	jb.listeners = append(jb.listeners, cb)
}

func (jb *JitterBuffer) updateStats(last_packet_seq_no uint16) {
	// If we have at least one packet, and the next packet being pushed in is not
	// at the expected sequence number increment the out of order count
	if jb.packets.Length() > 0 && last_packet_seq_no != ((jb.last_sequence+1)%math.MaxUint16) {
		jb.stats.out_of_order_count++
	}
	jb.last_sequence = last_packet_seq_no

}

// Push an RTP packet into the jitter buffer, this does not clone
// the data so if the memory is expected to be reused, the caller should
// take this in to account and pass a copy of the packet they wish to buffer
func (jb *JitterBuffer) Push(packet *rtp.Packet) {
	jb.mutex.Lock()
	defer jb.mutex.Unlock()
	if jb.packets.Length() > 100 {
		jb.stats.overflow_count++
		jb.emit(BufferOverflow)
	}
	if !jb.playout_ready && jb.packets.Length() == 0 {
		jb.playout_head = packet.SequenceNumber
	}
	jb.updateStats(packet.SequenceNumber)
	jb.packets.Push(packet, packet.SequenceNumber)
	jb.updateState()
}

func (jb *JitterBuffer) emit(event JitterBufferEvent) {
	for _, l := range jb.listeners {
		l(event, jb)
	}
}

func (jb *JitterBuffer) updateState() {
	// For now, we only look at the number of packets captured in the play buffer
	if jb.packets.Length() >= 50 {
		jb.state = Emitting
		jb.emit(BeginPlayback)
	}
}

// Peek at the packet which is either:
//
//	At the playout head when we are emitting, and the playoutHead flag is true
//
// or else
//
//	At the last sequence received
func (jb *JitterBuffer) Peek(playoutHead bool) (*rtp.Packet, error) {
	jb.mutex.Lock()
	defer jb.mutex.Unlock()
	if jb.packets.Length() < 1 {
		return nil, errors.New("Invalid Peek: Empty jitter buffer")
	}
	if playoutHead && jb.state == Emitting {
		return jb.packets.Find(jb.playout_head)
	}
	return jb.packets.Find(jb.last_sequence)
}

// Pop an RTP packet from the jitter buffer at the current playout head
func (jb *JitterBuffer) Pop() (*rtp.Packet, error) {
	jb.mutex.Lock()
	defer jb.mutex.Unlock()
	if jb.state != Emitting {
		return nil, errors.New("Attempt to pop while buffering")
	}
	packet, err := jb.packets.PopAt(jb.playout_head)
	if err != nil {
		jb.stats.underflow_count++
		jb.emit(BufferUnderflow)
		return (*rtp.Packet)(nil), err
	}
	jb.playout_head = (jb.playout_head + 1) % math.MaxUint16
	jb.updateState()
	return packet, nil
}
