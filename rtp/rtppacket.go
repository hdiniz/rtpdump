package rtp

import (
  "time"
)

type RtpPacket struct {
  ReceivedAt time.Time
  Version int
  Padding bool
  Extension bool
  CC int
  Marker bool
  PayloadType int
  SequenceNumber uint16
  Timestamp uint32
  Ssrc uint32
  Csrc []uint32
  ExtensionHeaderId uint16
  ExtensionHeaderLength uint16
  ExtensionHeader []byte
  Payload []byte
}
