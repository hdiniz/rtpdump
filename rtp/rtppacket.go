package rtp

import (
	"fmt"
	"time"

	"github.com/hdiniz/rtpdump/util"
)

type RtpPacket struct {
	ReceivedAt            time.Time
	Version               int
	Padding               bool
	Extension             bool
	CC                    int
	Marker                bool
	PayloadType           int
	SequenceNumber        uint16
	Timestamp             uint32
	Ssrc                  uint32
	Csrc                  []uint32
	ExtensionHeaderId     uint16
	ExtensionHeaderLength uint16
	ExtensionHeader       []byte
	Payload               []byte
	Data                  []byte
}

func (r RtpPacket) String() string {
	return fmt.Sprintf("%s - %d - %d",
		util.TimeMsToStr(r.ReceivedAt),
		r.SequenceNumber,
		r.Timestamp,
	)
}
