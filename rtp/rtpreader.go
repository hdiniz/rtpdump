package rtp

import (
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/hdiniz/rtpdump/esp"
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
		receivedAt := packet.Metadata().CaptureInfo.Timestamp
		r.decodePacket(receivedAt, packet)
	}
	/* if no packets were found, try raw link layer */
	if len(r.rtpStreamsSorted) <= 0 {
		r.reOpenPcapFile()
		packetSource = gopacket.NewPacketSource(r.handle, layers.LinkTypeRaw)
		for packet := range packetSource.Packets() {
			receivedAt := packet.Metadata().CaptureInfo.Timestamp
			r.decodePacket(receivedAt, packet)
		}
	}
	return r.rtpStreamsSorted
}

func (r *RtpReader) decodeIPv4Packet(receivedAt time.Time, packet gopacket.Packet, ipLayerType gopacket.Layer) error {
	if ipLayerType == nil {
		log.Sdebug("LayerPayload v4: %s", hex.Dump(ipLayerType.LayerPayload()))
		return errors.New("Not able to decode ipv4 packet")
	}
	ipLayer := ipLayerType.(*layers.IPv4)
	udpLayer, _ := packet.Layer(layers.LayerTypeUDP).(*layers.UDP)
	if udpLayer == nil {
		return errors.New("Not UDP Packet")
	}
	return r.decodeUDPLayer(receivedAt, packet, ipLayer.SrcIP.String(), ipLayer.DstIP.String(), udpLayer)
}

func (r *RtpReader) decodeIPv6Packet(receivedAt time.Time, packet gopacket.Packet, ipLayerType gopacket.Layer) error {
	if ipLayerType == nil {
		log.Sdebug("LayerPayload v6: %s", hex.Dump(ipLayerType.LayerPayload()))
		return errors.New("Not able to decode ipv6 packet")
	}
	ipLayer := ipLayerType.(*layers.IPv6)
	udpLayer, _ := packet.Layer(layers.LayerTypeUDP).(*layers.UDP)
	if udpLayer == nil {
		return errors.New("Not UDP Packet")
	}
	return r.decodeUDPLayer(receivedAt, packet, ipLayer.SrcIP.String(), ipLayer.DstIP.String(), udpLayer)
}

func (r *RtpReader) decodeUDPLayer(receivedAt time.Time, packet gopacket.Packet, src string, dst string, udp *layers.UDP) error {
	if udp.SrcPort%2 != 0 || udp.DstPort%2 != 0 {
		return errors.New("Likely RTCP packet")
	}

	if udp.SrcPort == 4500 || udp.DstPort == 4500 {
		espPacket := gopacket.NewPacket(udp.Payload, layers.LayerTypeIPSecESP, gopacket.Default)
		espLayer := espPacket.Layer(layers.LayerTypeIPSecESP).(*layers.IPSecESP)
		return r.decodeESPLayer(receivedAt, packet, espLayer)
	}

	rtpPacket := gopacket.NewPacket(
		udp.Payload,
		RtpLayerType,
		gopacket.Default,
	)
	rtpLayer := rtpPacket.Layer(RtpLayerType)
	rtp, _ := rtpLayer.(*RtpLayer)
	if rtpLayer == nil || rtp == nil {
		return errors.New("Not able to decode RTP layer")
	}
	return r.processRtpPacket(receivedAt, src, dst, udp, rtp)
}

func (r *RtpReader) decodeESPLayer(receivedAt time.Time, packet gopacket.Packet, espLayer *layers.IPSecESP) error {
	espPacket := esp.DecodeESPLayer(packet, espLayer)
	if espPacket != nil {
		return r.decodePacket(receivedAt, espPacket)
	}
	return errors.New("Not able to decode ESP")
}

func (r *RtpReader) decodePacket(receivedAt time.Time, packet gopacket.Packet) error {
	//log.Sdebug("decodePacket: %s", packet.Dump())
	networkLayer := packet.Layer(layers.LayerTypeIPv4)
	if networkLayer != nil {
		return r.decodeIPv4Packet(receivedAt, packet, networkLayer)
	}
	networkLayer = packet.Layer(layers.LayerTypeIPv6)
	if networkLayer != nil {
		return r.decodeIPv6Packet(receivedAt, packet, networkLayer)
	}
	return errors.New("Failed to decode packet")
}

func (r *RtpReader) processRtpPacket(receivedAt time.Time, src string, dst string, udp *layers.UDP, rtp *RtpLayer) error {
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
	return nil
}
