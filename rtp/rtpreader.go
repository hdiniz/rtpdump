package rtp

import (
	"errors"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/hdiniz/rtpdump/log"
)

// RtpReader reads
type RtpReader struct {
	handle           *pcap.Handle
	rtpStreamsMap    map[uint32]*RtpStream
	rtpStreamsSorted []*RtpStream
	filePath         string
}

//NewRtpReader creates new reader
func NewRtpReader(path string) (reader *RtpReader, err error) {
	reader = &RtpReader{}
	reader.rtpStreamsMap = make(map[uint32]*RtpStream)
	err = reader.openPcapFile(path)
	return
}

func (r *RtpReader) openPcapFile(path string) (err error) {
	r.filePath = path
	r.handle, err = pcap.OpenOffline(path)
	if err != nil {
		log.Error("Failed to open pcap file, note that pcapng format is not supported\nplease convert to legacy pcap format before using this tool")
		return err
	}
	err = r.handle.SetBPFFilter(RtpCapureFilter)
	if err != nil {
		r.handle.Close()
		log.Error("Failed to set bpf file")
		return err
	}
	return nil
}

func (r *RtpReader) reOpenPcapFile() {
	r.Close()
	r.openPcapFile(r.filePath)
}

//Close rtp reader
func (r *RtpReader) Close() {
	r.handle.Close()
}

//GetStreams returns rtp streams identified
func (r *RtpReader) GetStreams() []*RtpStream {
	packetSource := gopacket.NewPacketSource(r.handle, r.handle.LinkType())
	for packet := range packetSource.Packets() {
		r.parsePacket(packet)
	}
	/* if no packets were found, try raw link layer */
	if len(r.rtpStreamsSorted) <= 0 {
		r.reOpenPcapFile()
		packetSource = gopacket.NewPacketSource(r.handle, layers.LinkTypeRaw)
		for packet := range packetSource.Packets() {
			r.parsePacket(packet)
		}
	}
	return r.rtpStreamsSorted
}

func (r *RtpReader) parsePacket(packet gopacket.Packet) error {
	receivedAt := packet.Metadata().CaptureInfo.Timestamp
	networkLayer := packet.Layer(layers.LayerTypeIPv4)
	isIPv4 := true
	if networkLayer == nil {
		isIPv4 = false
	}

	var ipLayer gopacket.Layer

	if isIPv4 {
		ipLayer = packet.Layer(layers.LayerTypeIPv4)
	} else {
		ipLayer = packet.Layer(layers.LayerTypeIPv6)
	}

	udpLayer := packet.Layer(layers.LayerTypeUDP)

	if ipLayer != nil && udpLayer != nil {
		var ipv4 *layers.IPv4
		var ipv6 *layers.IPv6
		if isIPv4 {
			ipv4, _ = ipLayer.(*layers.IPv4)
		} else {
			ipv6, _ = ipLayer.(*layers.IPv6)
		}

		udp, _ := udpLayer.(*layers.UDP)

		if udp.SrcPort%2 != 0 || udp.DstPort%2 != 0 {
			return errors.New("Likely RTCP packet")
		}

		rtpPacket := gopacket.NewPacket(
			packet.ApplicationLayer().Payload(),
			RtpLayerType,
			gopacket.Default,
		)

		rtpLayer := rtpPacket.Layer(RtpLayerType)
		if rtpLayer != nil {
			rtp, _ := rtpLayer.(*RtpLayer)
			if isIPv4 {
				r.processRtpPacket(receivedAt, ipv4.SrcIP.String(), ipv4.DstIP.String(), udp, rtp)
			} else {
				r.processRtpPacket(receivedAt, ipv6.SrcIP.String(), ipv6.DstIP.String(), udp, rtp)
			}
		} else {
			return errors.New("Not able to decode RTP layer")
		}
	} else {
		return errors.New("Not able to decode Network/Transport layers")
	}
	return nil
}

func (r *RtpReader) processRtpPacket(receivedAt time.Time, src string, dst string, udp *layers.UDP, rtp *RtpLayer) {
	rtp.ReceivedAt = receivedAt

	s, ok := r.rtpStreamsMap[rtp.Ssrc]
	if !ok {
		s = &RtpStream{
			SrcIP:          src,
			SrcPort:        uint(udp.SrcPort),
			DstIP:          dst,
			DstPort:        uint(udp.DstPort),
			Ssrc:           rtp.Ssrc,
			PayloadType:    rtp.PayloadType,
			FirstSeq:       rtp.SequenceNumber,
			FirstTimestamp: rtp.Timestamp,
			StartTime:      receivedAt,
		}
		r.rtpStreamsMap[rtp.Ssrc] = s
		r.rtpStreamsSorted = append(r.rtpStreamsSorted, s)
	}
	s.AddPacket(rtp.RtpPacket())
}
