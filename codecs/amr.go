package codecs
import (
  "errors"
  log "github.com/Sirupsen/logrus"
)

const AMR_NB_MAGIC string = "#!AMR\n"
const AMR_WB_MAGIC string = "#!AMR-WB\n"
var AMR_NB_FRAME_SIZE []int = []int{12, 13, 15, 17, 19, 20, 26, 31, 5, 0, 0, 0, 0, 0, 0, 0}
var AMR_WB_FRAME_SIZE []int = []int{17, 23, 32, 36, 40, 46, 50, 58, 60, 5, 5, 0, 0, 0, 0, 0}
const AMR_NB_SAMPLE_RATE = 8000
const AMR_WB_SAMPLE_RATE = 16000

type Amr struct {
  started bool
  sampleRate int
  octetAligned bool
  timestamp uint32
}

func NewAmr() Codec {
  return &Amr{}
}

func (amr *Amr) Init() {
  amr.started = false
  amr.timestamp = 0
}

func (amr *Amr) isWideBand() bool {
  return amr.sampleRate == AMR_WB_SAMPLE_RATE
}

func (amr Amr) GetFormatMagic() []byte {
  if amr.isWideBand() {
    return []byte(AMR_WB_MAGIC)
  } else {
    return []byte(AMR_NB_MAGIC)
  }
}

func (amr *Amr) invalidState() error {
  return errors.New("invalid state")
}

func (amr *Amr) SetOptions(options map[string]string) error {
  if amr.started {
    return amr.invalidState()
  }

  v,ok := options["octet-aligned"]
  if !ok {
    return errors.New("required codec option not present")
  }

  amr.octetAligned = v == "1"

  v,ok = options["sample-rate"]
  if !ok {
    return errors.New("required codec option not present")
  }

  if (v == "nb") {
    amr.sampleRate = AMR_NB_SAMPLE_RATE
  } else if (v == "wb") {
    amr.sampleRate = AMR_WB_SAMPLE_RATE
  } else {
    return errors.New("invalid codec option value")
  }

  return nil
}

func (amr *Amr) HandleRtpPacket(timestamp uint32, payload []byte) ([]byte, error) {
  if amr.octetAligned {
    return amr.handleOaMode(timestamp, payload)
  } else {
    return amr.handleBeMode(timestamp, payload)
  }
}

func (amr *Amr) handleOaMode(timestamp uint32, payload []byte) ([]byte, error) {

  var result []byte
  var lostSamplesFromPrevious uint32
  var currentTimestamp uint32

  if amr.timestamp != 0 {
    lostSamplesFromPrevious = (timestamp - amr.timestamp) / 160 -1
    for i := lostSamplesFromPrevious; i > 0; i-- {
      result = append(result, 0xFC)
    }
  }

  frame := 0
  rtpFrameHeader := payload[0:]
  // payload header := [CMR(4bit)[R(4bit)][ILL(4bit)(opt)][ILP(4bit)(opt)]
  // TOC := [F][FT(4bit)][Q][P][P]
  // storage := [0][FT(4bit)][Q][0][0]
  cmr := (rtpFrameHeader[0] & 0xF0) >> 4
  isLastFrame := (rtpFrameHeader[1] & 0x80) & 0x80 == 0x80
  frameType := (rtpFrameHeader[1] & 0x78) >> 3
  quality := (rtpFrameHeader[1] & 0x04) & 0x04 == 0x04

  speechFrameHeader := cmr << 4
  speechFrameHeader = speechFrameHeader | (rtpFrameHeader[1] & 0x40)

  var speechFrameSize int
  if amr.isWideBand() {
    speechFrameSize = AMR_WB_FRAME_SIZE[frameType]
  } else {
    speechFrameSize = AMR_NB_FRAME_SIZE[frameType]
  }

  currentTimestamp = timestamp + uint32(160*frame)
  log.WithFields(log.Fields{
    "sample-rate": amr.sampleRate,
    "rtpFrameHeader": rtpFrameHeader,
    "timestamp": timestamp,
    "currentTimestamp": currentTimestamp,
    "previousTimestamp": amr.timestamp,
    "frame": frame,
    "octet-aligned": amr.octetAligned,
    "cmr": cmr,
    "isLastFrame": isLastFrame,
    "frameType": frameType,
    "quality": quality,
    "speechFrameSize": speechFrameSize,
    "lostSamplesFromPrevious": lostSamplesFromPrevious,
  }).Debug("amr frame")

  if !isLastFrame {
    log.Warn("Amr does not suport more than one frame per payload - discarted")
    return nil, errors.New("Amr does not suport more than one frame per payload")
  }

  result = append(result, speechFrameHeader)

  if speechFrameSize != 0 {
    speechPayload := rtpFrameHeader[2:2+speechFrameSize]
    result = append(result, speechPayload...)
  }
  amr.timestamp = currentTimestamp
  return result, nil
}

func (amr *Amr) handleBeMode(timestamp uint32, payload []byte) ([]byte, error) {
  var result []byte
  var lostSamplesFromPrevious uint32
  var currentTimestamp uint32


  if amr.timestamp != 0 {
    lostSamplesFromPrevious = (timestamp - amr.timestamp) / 160 -1
    for i := lostSamplesFromPrevious; i > 0; i-- {
      result = append(result, 0xFC)
    }
  }

  frame := 0
  rtpFrameHeader := payload[0:]
  // packing frame with TOC: frame type and quality bit
  // RTP=[CMR(4bit)[F][FT(4bit)][Q][..speechFrame]] -> storage=[0][FT(4bit)][Q][0][0]
  cmr := (rtpFrameHeader[0] & 0xF0) >> 4
  isLastFrame := (rtpFrameHeader[0] & 0x08) >> 4 & 0x01 == 0x00
  frameType := (rtpFrameHeader[0] & 0x07) << 1 | (rtpFrameHeader[1] & 0x80) >> 7
  quality := (rtpFrameHeader[1] & 0x04) >> 2 & 0x01 == 0x01

  speechFrameHeader := (rtpFrameHeader[0] & 0x07)<<4 | (rtpFrameHeader[1] & 0x80)>>4
  speechFrameHeader = speechFrameHeader | (rtpFrameHeader[1] & 0x40)>>4

  var speechFrameSize int
  if amr.isWideBand() {
    speechFrameSize = AMR_WB_FRAME_SIZE[frameType]
  } else {
    speechFrameSize = AMR_NB_FRAME_SIZE[frameType]
  }

  currentTimestamp = timestamp + uint32(160*frame)
  log.WithFields(log.Fields{
    "sample-rate": amr.sampleRate,
    "rtpFrameHeader": rtpFrameHeader,
    "timestamp": timestamp,
    "currentTimestamp": currentTimestamp,
    "previousTimestamp": amr.timestamp,
    "frame": frame,
    "octet-aligned": amr.octetAligned,
    "cmr": cmr,
    "isLastFrame": isLastFrame,
    "frameType": frameType,
    "quality": quality,
    "speechFrameSize": speechFrameSize,
    "lostSamplesFromPrevious": lostSamplesFromPrevious,
  }).Debug("amrnb frame")

  if !isLastFrame {
    log.Warn("Amr does not suport more than one frame per payload - discarted")
    return nil, errors.New("Amr does not suport more than one frame per payload")
  }

  result = append(result, speechFrameHeader)

  if speechFrameSize != 0 {
    speechPayload := rtpFrameHeader[1:]
    speechFrame := make([]byte, speechFrameSize)
    // shift 2 bits left in speechFrame
    for k := 0; k < speechFrameSize; k++ {
        speechFrame[k] = (speechPayload[k] & 0x3F) << 2
        if k + 1 < speechFrameSize {
          speechFrame[k] = speechFrame[k] | (speechPayload[k+1] & 0xC0)>>6
        }
    }
    result = append(result, speechFrame...)
  }
  amr.timestamp = currentTimestamp
  return result, nil
}


var AmrMetadata = CodecMetadata{
  Name: "amr",
  LongName: "Adaptative Multi Rate",
  Options: []CodecOption {
    amrSampleRateOption,
    amrOctetAlignedOption,
  },
  Init: NewAmr,
}

var amrOctetAlignedOption = CodecOption{
  Required: true,
  Name: "octet-aligned",
  Description: "whether this payload is octet-aligned or bandwidth-efficient",
  ValidValues: []string {"0", "1"},
  ValueDescription: []string {"bandwidth-efficient", "octet-aligned"},
  RestrictValues: true,
}

var amrSampleRateOption = CodecOption{
  Required: true,
  Name: "sample-rate",
  Description: "whether this payload is narrow or wide band",
  ValidValues: []string {"nb", "wb"},
  ValueDescription: []string {"Narrow Band (8000)", "Wide Band (16000)"},
  RestrictValues: true,
}
