package jitterbuffer

import (
	"github.com/pion/interceptor"
	"github.com/pion/rtp"
)

type JitterBufferState uint16

const (
	Buffering JitterBufferState = iota
	Emitting  JitterBufferState
)

func (jbs JitterBufferState) String() string {
	switch jbs {
	case Buffering:
		return "Buffering"
	case Emitting:
		return "Emitting"
	}
}

type JitterBufferOption struct {
	initialLatency unit16
}

type Option func(jb *JitterBuffer)

type JitterBuffer struct {
	packets             [math.MaxUint16 + 1]*rtp.Packet
	last_sequence       uint16
	state               JitterBufferState
	sample_rate         uint16
	payload_sample_rate int
	max_depth           int
	stats               JitterBufferStats
}

type JitterBufferStats struct {
	out_of_order_count uint32
	empty_count        uint32
	overflow_count     uint32
	jitter             float32
	max_jitter         float32
}

func New(opts ...Option) JitterBuffer {
	jb := &JitterBuffer{state: JitterBufferState.Buffering, stats: {0, 0, 0, .0, 0}}
	for _, o := range opts {
		o(jb)
	}
	return jb
}

func (jb *JitterBuffer) Push(packet *rtp.Packet) {
	jb.packets[packet.SequenceNumber] = packet
	if packet.SequenceNumber != jb.last_sequence+1 &&
		// we are wrapping around
		(jb.last_sequnce != math.MaxUint16 && packet.SequenceNumber == 0) {
		jb.stats.overflow_count++
	}
}