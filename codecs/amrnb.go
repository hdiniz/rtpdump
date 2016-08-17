package codecs
import (
  "github.com/hdiniz/rtpdump/rtp"
)

var AmrNbMagic []byte = []byte("#!AMR\n")

//{DecoderOption{Required: true,Name: "format",Description: "Payload format: (1) bandwidth-efficient, (2) octet-aligned",Value: "1",Default: "1",}}

type AmrNbDecoder struct {
  options []*DecoderOption
  storage []byte
}

func (d *AmrNbDecoder) Name() string {
  return "AMR NB"
}

func (d *AmrNbDecoder) Description() string {
  return "Adaptive Multi Rate Narrow Band"
}

func NewAmrNbDecoder() *AmrNbDecoder {
  return &AmrNbDecoder{}
}

func (d *AmrNbDecoder) GetOptions() []*DecoderOption {
  var options []*DecoderOption
  var payload = &DecoderOption{
    Required: true,
    Name: "format",
    Description: "Payload format: (1) bandwidth-efficient, (2) octet-aligned",Value: "1",
    Default: "1",
  }
  return append(options, payload)
}

func (d* AmrNbDecoder) SetOptions(options []*DecoderOption) {
  d.options = options
}

func (d* AmrNbDecoder) GetStorageFormat() []byte {
  var file []byte
  file = append(file, AmrNbMagic...)
  return append(file, d.storage...)
}

func (d* AmrNbDecoder) isOctetAligned() bool {
  for _,v := range d.options {
    if v.Name == "format" {
      return v.Value == "2"
    }
  }
  return false
}
// improve frame validation, implement missing packet NO_DATA and jitter buffer
func (d* AmrNbDecoder) ProcessRtpPacket(packet *rtp.RtpPacket) error {
  if d.isOctetAligned() {
    panic("not supported")
  } else {
    return d.processBeMode(packet)
  }
  return nil
}

func (d* AmrNbDecoder) processBeMode(p *rtp.RtpPacket) error {
  var frameHeader byte
  // packing frame with TOC: frame type and quality bit
  // RTP=[CMR(4bit)[F][FT(4bit)][Q][..speechFrame]] -> storage=[0][FT(4bit)][Q][0][0]
  frameHeader = (p.Payload[0] & 0x07)<<4 | (p.Payload[1] & 0x80)>>4
  frameHeader = frameHeader | (p.Payload[1] & 0x40)>>4

  d.storage = append(d.storage, frameHeader)


  speechPayload := p.Payload[1:]
  // match frame type with resulting bytes to check validity
  speechFrame := make([]byte, len(speechPayload))
  // shift 2 bits left in speechFrame
  for k := 0; k < len(speechPayload)-2; k++ {
      speechFrame[k] = (speechPayload[k] & 0x3F) << 2 | (speechPayload[k+1] & 0xC0)>>6
  }
  d.storage = append(d.storage, speechFrame...)

  return nil
}
