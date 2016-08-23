package codecs

import (
  "fmt"
  "github.com/hdiniz/rtpdump/rtp"
)

type Codec interface {
  Init()
  SetOptions(options map[string]string) error
  HandleRtpPacket(packet *rtp.RtpPacket) ([]byte, error)
  GetFormatMagic() []byte
}

type CodecMetadata struct {
  Name string
  LongName string
  Options []CodecOption
  Init func()Codec
}

type CodecOption struct {
  Required bool
  Name string
  Description string
  ValidValues []string
  ValueDescription []string
  RestrictValues bool
}

func (m CodecMetadata) Describe() string {
  options := ""
  if len(m.Options) > 0 {
    options = "\tOptions:"
    for _, v := range m.Options {
      options += fmt.Sprintf(
        "\n\t\t%s\n\n\t\tRequired: %t\n\t\t%s\n\t\t",
        v.Name, v.Required, v.Description)
      if v.RestrictValues {
        options += "Valid values:\n"
        for i, rv := range v.ValidValues {
          options += fmt.Sprintf("\t\t\t(%s) - %s\n", rv, v.ValueDescription[i])
        }
      }
    }
  }

  return fmt.Sprintf(
    "%s\n\t%s\n%s",
    m.Name, m.LongName, options)
}
