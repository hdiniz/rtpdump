package rtp

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/hdiniz/rtpdump/util"
)

var RtpCapureFilter string = "udp and not (" +
	"udp port 53 or " + // DNS
	"udp port 138 or " + // NETBIOS
	"udp port 67 or " + // BOOTSTRAP
	"udp port 68 or " + // BOOTSTRAP
	"udp port 1900 or " + // SSDP
	//"udp port 4500 or " + // Allow IKE for decrypt
	"udp port 500 or " + // IKE
	"udp port 123 or " + // NTP
	"port 5060" +
	")"

type RtpLayer struct {
	ReceivedAt            time.Time
	Header                []byte
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
}

func (l RtpLayer) String() string {
	return fmt.Sprintf(
		"received:%s,v:%d,pad:%t,ext:%t,cc:%d,mark:%t,type:%d,seq:%d,ts:%d,ssrc:0x%x(%d)",
		util.TimeToStr(l.ReceivedAt),
		l.Version,
		l.Padding,
		l.Extension,
		l.CC,
		l.Marker,
		l.PayloadType,
		l.SequenceNumber,
		l.Timestamp,
		l.Ssrc,
		l.Ssrc,
	)
}

// Improve
func (l RtpLayer) RtpPacket() *RtpPacket {
	return &RtpPacket{
		ReceivedAt:            l.ReceivedAt,
		Version:               l.Version,
		Padding:               l.Padding,
		Extension:             l.Extension,
		CC:                    l.CC,
		Marker:                l.Marker,
		PayloadType:           l.PayloadType,
		SequenceNumber:        l.SequenceNumber,
		Timestamp:             l.Timestamp,
		Ssrc:                  l.Ssrc,
		Csrc:                  l.Csrc,
		ExtensionHeaderId:     l.ExtensionHeaderId,
		ExtensionHeaderLength: l.ExtensionHeaderLength,
		ExtensionHeader:       l.ExtensionHeader,
		Payload:               l.Payload,
	}
}

func (l RtpLayer) LayerType() gopacket.LayerType {
	return RtpLayerType
}

func (l RtpLayer) LayerContents() []byte {
	return l.Header
}

func (l RtpLayer) LayerPayload() []byte {
	return l.Payload
}

var RtpLayerType = gopacket.RegisterLayerType(
	2001,
	gopacket.LayerTypeMetadata{
		"RtpLayerType",
		gopacket.DecodeFunc(decodeRtpLayer),
	},
)

func decodeRtpLayer(data []byte, p gopacket.PacketBuilder) error {
	if len(data) < 12 {
		return errors.New("RTP header should contain at least 12 octets")
	}

	var rtp RtpLayer
	rtp.Version = int(data[0]&0xC0) >> 6

	if rtp.Version != 2 {
		return errors.New("Indicated RTP version != 2")
	}

	rtp.Padding = (data[0] & 0x20) == 0x20

	rtp.Extension = (data[0] & 0x10) == 0x10
	rtp.CC = int((data[0] & 0x0F))
	rtp.Marker = (data[1] & 0x80) == 0x80
	rtp.PayloadType = int((data[1] & 0x7F))
	rtp.SequenceNumber = uint16(data[2])<<8 + uint16(data[3])
	rtp.Timestamp = uint32(data[4])<<24 + uint32(data[5])<<16 + uint32(data[6])<<8 + uint32(data[7])
	rtp.Ssrc = uint32(data[8])<<24 + uint32(data[9])<<16 + uint32(data[10])<<8 + uint32(data[11])
	offset := 12
	if rtp.CC > 0 {
		if len(data[offset:]) < rtp.CC*4 {
			return errors.New("Not enough octets left in RTP header to satisfy CC")
		}
		rtp.Csrc = make([]uint32, rtp.CC)
		for i := 0; i < rtp.CC; i++ {
			rtp.Csrc[i] = uint32(data[offset+i])<<24 + uint32(data[offset+1+i])<<16 + uint32(data[offset+2+i])<<8 + uint32(data[offset+3+i])
		}
		offset += rtp.CC * 4
	}

	if rtp.Extension {
		if len(data[offset:]) < 4 {
			return errors.New("Not enough octets left in RTP header to satisfy ExtensionHeaderId and ExtensionHeaderLength")
		}
		rtp.ExtensionHeaderId = uint16(data[offset])<<8 + uint16(data[offset+1])
		offset += 2
		rtp.ExtensionHeaderLength = uint16(data[offset])<<8 + uint16(data[offset+1])
		offset += 2
	}

	if rtp.ExtensionHeaderLength > 0 {
		if len(data[offset:]) < 4*int(rtp.ExtensionHeaderLength) {
			return errors.New("Not enough octets left in RTP header to satisfy indicated Extensions")
		}
		rtp.ExtensionHeader = make([]byte, 4*int(rtp.ExtensionHeaderLength))
		rtp.ExtensionHeader = data[offset : offset+4*int(rtp.ExtensionHeaderLength)]
		offset += 4 * int(rtp.ExtensionHeaderLength)
	}

	if len(data[offset:]) == 0 {
		return errors.New("No payload contained in RTP")
	}

	if rtp.Padding {
		padLen := int(data[len(data)-1])
		if padLen <= 0 || padLen > len(data[offset:]) {
			return errors.New("Invalid padding lenght")
		}

		rtp.Payload = data[offset : len(data)-padLen]
	} else {
		rtp.Payload = data[offset:]
	}
	p.AddLayer(&rtp)
	return p.NextDecoder(gopacket.LayerTypePayload)
}
