package jitterbuffer 

import (
	"testing"
  "github.com/stretchr/testify/assert"
	//"github.com/pion/rtp"
)

func TestJitterBuffer(t *testing.T) {
  assert := assert.New(t)
  t.Run("Appends packets in order", func (t *testing.T) {
    jb := New()
    assert.Equal(jb.last_sequence, 0)
  })
}
