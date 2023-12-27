package jitterbuffer

import (
  "math"
  "errors"
	"github.com/pion/rtp"
)

type JitterBufferState uint16

type JitterBufferEvent string
const (
	Buffering JitterBufferState = iota
	Emitting  
)
const (
  StartBuffering JitterBufferEvent = "startBuffering"
  BeginPlayback = "playing"
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
	packets             [math.MaxUint16 + 1]*rtp.Packet
	last_sequence       uint16
  buffer_length       uint16
  playout_head        uint16
	state               JitterBufferState
	sample_rate         uint16
	payload_sample_rate int
	max_depth           int
	stats               JitterBufferStats
  listeners           []JitterBufferEventListener
}

type JitterBufferStats struct {
	out_of_order_count uint32
	empty_count        uint32
	overflow_count     uint32
	jitter             float32
	max_jitter         float32
}

func New(opts ...Option) *JitterBuffer {
	jb := &JitterBuffer{state: Buffering, stats: JitterBufferStats{0, 0, 0, .0, 0}}
	for _, o := range opts {
		o(jb)
	}
	return jb
}

func (jb *JitterBuffer) Listen(event JitterBufferEvent, cb JitterBufferEventListener) {
  jb.listeners = append(jb.listeners, cb)
}

func (jb *JitterBuffer) Push(packet *rtp.Packet) {
	jb.packets[packet.SequenceNumber] = packet
	if packet.SequenceNumber != jb.last_sequence+1 &&
		// we are wrapping around
		(jb.last_sequence != math.MaxUint16 && packet.SequenceNumber == 0) {
		jb.stats.overflow_count++
	}
  jb.last_sequence = packet.SequenceNumber
  jb.buffer_length ++
  jb.updateState()
}

func (jb *JitterBuffer) emit(event JitterBufferEvent) {
  for _, l := range jb.listeners {
    l(event, jb)
  }
}

func (jb *JitterBuffer) updateState() {
  // For now, we only look at the number of packets captured in the play buffer
  if (jb.buffer_length >= 50) {
    jb.state = Emitting
    jb.emit(BeginPlayback)
  }
}

func (jb *JitterBuffer) Peek(playoutHead bool) (*rtp.Packet, error) {
  if jb.buffer_length < 1 {
    return nil, errors.New("Invalid Peek: Empty jitter buffer")
  }
  if playoutHead {
    return jb.packets[jb.playout_head], nil
  }
  return jb.packets[jb.last_sequence], nil
}
