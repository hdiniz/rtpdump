package rtp

import (
  "errors"
  "time"
  "github.com/hdiniz/rtpdump/log"
  "github.com/google/gopacket"
  "github.com/google/gopacket/layers"
  "github.com/google/gopacket/pcap"
)


type RtpReader struct {
  handle *pcap.Handle
  rtpStreamsMap map[uint32]*RtpStream
  rtpStreamsSorted []*RtpStream
}

func NewRtpReader(path string) (reader *RtpReader, err error) {
  reader = &RtpReader{}
  reader.rtpStreamsMap = make(map[uint32]*RtpStream)
  err = reader.openPcapFile(path)
  return
}

func (r *RtpReader) openPcapFile(path string) (err error) {
  r.handle, err = pcap.OpenOffline(path)
  if err != nil {
      log.Error("Failed to open pcap file")
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

func (r *RtpReader) Close() {
  r.handle.Close()
}

func (r *RtpReader) GetStreams() ([]*RtpStream) {
  packetSource := gopacket.NewPacketSource(r.handle, r.handle.LinkType())
  for packet := range packetSource.Packets() {
      r.parsePacket(packet)
  }
  return r.rtpStreamsSorted
}


func (r *RtpReader) parsePacket(packet gopacket.Packet) error {
    receivedAt := packet.Metadata().CaptureInfo.Timestamp
    ipLayer := packet.Layer(layers.LayerTypeIPv4)
    udpLayer := packet.Layer(layers.LayerTypeUDP)

    if ipLayer != nil && udpLayer != nil {

        ip, _ := ipLayer.(*layers.IPv4)
        udp, _ := udpLayer.(*layers.UDP)

        if udp.SrcPort % 2 != 0 || udp.DstPort % 2 != 0 {
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
            r.processRtpPacket(receivedAt, ip, udp, rtp)
        } else {
            log.Debug("Not able to decode RTP layer")
            return errors.New("Not able to decode RTP layer")
        }
    } else {
        log.Debug("Not able to decode Network/Transport layer")
        return errors.New("Not able to decode Network/Transport layers")
    }
    return nil
}

func (r *RtpReader) processRtpPacket(receivedAt time.Time, ip *layers.IPv4, udp *layers.UDP, rtp *RtpLayer) {
    rtp.ReceivedAt = receivedAt

    s, ok := r.rtpStreamsMap[rtp.Ssrc]
    if !ok {
        s = &RtpStream{
            SrcIP:ip.SrcIP.String(),
            SrcPort:uint(udp.SrcPort),
            DstIP:ip.DstIP.String(),
            DstPort:uint(udp.DstPort),
            Ssrc: rtp.Ssrc,
            PayloadType: rtp.PayloadType,
            FirstSeq: rtp.SequenceNumber,
            FirstTimestamp: rtp.Timestamp,
            StartTime: receivedAt,
        }
        r.rtpStreamsMap[rtp.Ssrc] = s
        r.rtpStreamsSorted = append(r.rtpStreamsSorted, s)
    }
    s.AddPacket(rtp.RtpPacket())
}
