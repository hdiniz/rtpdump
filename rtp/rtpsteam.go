package rtp

import (
    "fmt"
    "time"
    "github.com/hdiniz/rtpdump/util"
)

type RtpStream struct {

    // Public
    Ssrc uint32
    PayloadType int
    SrcIP, DstIP string
    SrcPort, DstPort uint
    StartTime, EndTime time.Time

    // Internal - improve
    FirstTimestamp uint32
    FirstSeq uint16
    Cycle uint
    CurSeq uint16

    // Calculated
    TotalExpectedPackets uint
    LostPackets uint
    MeanJitter float32
    MeanBandwidth float32

    RtpPackets []*RtpPacket
}

func (r RtpStream) String() string {
    return fmt.Sprintf(
        "%s\t%s |\t%s \t%d\t->\t%s\t%d\t| t:%d\tc:%d\tssrc:0x%X",
        util.TimeToStr(r.StartTime),
        util.TimeToStr(r.EndTime),
        r.SrcIP,
        r.SrcPort,
        r.DstIP,
        r.DstPort,
        r.PayloadType,
        len(r.RtpPackets),
        r.Ssrc)
}

func (r *RtpStream) AddPacket(rtp *RtpPacket) {

    if rtp.SequenceNumber <= r.CurSeq {
        return
    }

    r.EndTime = rtp.ReceivedAt
    r.CurSeq = rtp.SequenceNumber
    r.TotalExpectedPackets = uint(r.CurSeq - r.FirstSeq)
    r.LostPackets = r.TotalExpectedPackets - uint(len(r.RtpPackets))

    r.RtpPackets = append(r.RtpPackets, rtp)
}
