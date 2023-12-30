package jitterbuffer

import (
	"github.com/pion/rtp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJitterBuffer(t *testing.T) {
	assert := assert.New(t)
	t.Run("Appends packets in order", func(t *testing.T) {
		jb := New()
		assert.Equal(jb.last_sequence, uint16(0))
		jb.Push(&rtp.Packet{Header: rtp.Header{SequenceNumber: 5000, Timestamp: 500}, Payload: []byte{0x02}})
		jb.Push(&rtp.Packet{Header: rtp.Header{SequenceNumber: 5001, Timestamp: 501}, Payload: []byte{0x02}})
		jb.Push(&rtp.Packet{Header: rtp.Header{SequenceNumber: 5002, Timestamp: 502}, Payload: []byte{0x02}})

		assert.Equal(jb.last_sequence, uint16(5002))

		jb.Push(&rtp.Packet{Header: rtp.Header{SequenceNumber: 5012, Timestamp: 512}, Payload: []byte{0x02}})

		assert.Equal(jb.last_sequence, uint16(5012))
	})
}
