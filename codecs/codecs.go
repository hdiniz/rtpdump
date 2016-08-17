package codecs

import (
  "github.com/hdiniz/rtpdump/rtp"
)

type DecoderOption struct {
  Required bool
  Name string
  Description string
  Option string
  Value string
  Default string
}

type Decoder interface {
  Name() string
  Description() string
  GetOptions() []*DecoderOption
  SetOptions(options []*DecoderOption)
  ProcessRtpPacket(packet *rtp.RtpPacket) error
  GetStorageFormat() []byte
}


func GetAudioCodecs() []Decoder {
  return []Decoder{NewAmrNbDecoder()}
}
