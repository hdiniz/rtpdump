package codecs
import (
  "errors"
  "github.com/hdiniz/rtpdump/log"
  "github.com/hdiniz/rtpdump/rtp"
)


var SINGLE_NAL_MODE = 0
var NON_INTERLEAVED_MODE = 1
var INTERLEAVED_MODE = 2

type H264 struct {
  packetizationMode string
  started bool
  configured bool
  timestamp uint32
}

func NewH264() Codec {
  return &H264{started: false, configured: false, timestamp: 0}
}

func (c *H264) Init() {
}

func (c *H264) SetOptions(options map[string]string) error {

  v,ok := options["packetization-mode"]
  if !ok {
    return errors.New("required codec option not present")
  }

  c.packetizationMode = v
  return nil
}

func (c H264) GetFormatMagic() []byte {
  return []byte{}
}

func (c *H264) HandleRtpPacket(packet *rtp.RtpPacket) (result []byte, err error) {
  payload := packet.Payload
  forbidden := (payload[0] & 0x80) == 0x80
  if forbidden {
    log.Warn("forbidden bit set in this payload")
    return nil, errors.New("forbidden bit set in this payload")
  }

  nri := (payload[0] & 0x60) >> 5
  nalType := payload[0] & 0x1F

  log.Sdebug("h264, seq:%d nri:%d, nalType:%d",
    packet.SequenceNumber, nri, nalType)

  switch {
    case nalType >= 1 && nalType <= 23:
      return c.handleNalUnit(payload[:])
    case nalType >= 24 && nalType <= 27:
      //aggregation packet
      log.Debug("h264, aggregation not supported")
      return nil, errors.New("h264, aggregation not supported")
    case nalType == 28:
      return c.handleFuA(payload[:])
    default:
      log.Sdebug("h264, nal type not supported")
      return nil, errors.New("h264, nal type not supported")
  }
}

func (c *H264) handleNalUnit(payload []byte) (result []byte, err error) {
  result = append(result, []byte{0x00, 0x00, 0x00, 0x01}...)
  result = append(result, payload[:]...)
  return result, nil
}
func (c *H264) handleFuA(payload []byte) (result []byte, err error) {
  isStart := payload[1] & 0x80 == 0x80
  //isEnd := payload[0] & 0x40 == 0x40

  log.Sdebug("h264, FU-A isStart:%t", isStart)
  if isStart {
    result = append(result, []byte{0x00, 0x00, 0x00, 0x01}...)
    nalUnitHeader := payload[0] & 0xE0
    nalUnitHeader = nalUnitHeader | (payload[1] & 0x1F)
    result = append(result, nalUnitHeader)
    result = append(result, payload[2:]...)
  } else {
    result = append(result, payload[2:]...)
  }
  log.Sdebug("FU-A: %#v", result)

  return
}


var H264Metadata = CodecMetadata{
  Name: "h264",
  LongName: "H.264",
  Options: []CodecOption {
    h264PacketizationModeOption,
  },
  Init: NewH264,
}

var h264PacketizationModeOption = CodecOption{
  Required: true,
  Name: "packetization-mode",
  Description: "whether this payload is octet-aligned or bandwidth-efficient",
  ValidValues: []string {"0", "1", "2"},
  ValueDescription: []string {"Single NAL Unit Mode", "Non-Interleaved Mode", "Interleaved Mode"},
  RestrictValues: true,
}
