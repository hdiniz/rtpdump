package codecs

import (
	"errors"

	"github.com/hdiniz/rtpdump/log"
	"github.com/hdiniz/rtpdump/rtp"
)

const AMR_NB_MAGIC string = "#!AMR\n"
const AMR_WB_MAGIC string = "#!AMR-WB\n"

var AMR_NB_FRAME_SIZE []int = []int{12, 13, 15, 17, 19, 20, 26, 31, 5, 0, 0, 0, 0, 0, 0, 0}
var AMR_WB_FRAME_SIZE []int = []int{17, 23, 32, 36, 40, 46, 50, 58, 60, 5, 5, 0, 0, 0, 0, 0}

const AMR_NB_SAMPLE_RATE = 8000
const AMR_WB_SAMPLE_RATE = 16000

type Amr struct {
	started      bool
	configured   bool
	sampleRate   int
	octetAligned bool
	timestamp    uint32

	lastSeq uint16
}

func NewAmr() Codec {
	return &Amr{started: false, configured: false, timestamp: 0}
}

func (amr *Amr) Init() {
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

	v, ok := options["octet-aligned"]
	if !ok {
		return errors.New("required codec option not present")
	}

	amr.octetAligned = v == "1"

	v, ok = options["sample-rate"]
	if !ok {
		return errors.New("required codec option not present")
	}

	if v == "nb" {
		amr.sampleRate = AMR_NB_SAMPLE_RATE
	} else if v == "wb" {
		amr.sampleRate = AMR_WB_SAMPLE_RATE
	} else {
		return errors.New("invalid codec option value")
	}
	amr.configured = true
	return nil
}

func (amr *Amr) HandleRtpPacket(packet *rtp.RtpPacket) (result []byte, err error) {
	if !amr.configured {
		return nil, amr.invalidState()
	}

	if packet.SequenceNumber <= amr.lastSeq {
		return nil, errors.New("Ignore out of sequence")
	}

	result = append(result, amr.handleMissingSamples(packet.Timestamp)...)

	var speechFrame []byte
	if amr.octetAligned {
		speechFrame, err = amr.handleOaMode(packet.Timestamp, packet.Payload)
	} else {
		speechFrame, err = amr.handleBeMode(packet.Timestamp, packet.Payload)
	}

	if err != nil {
		return nil, err
	}
	result = append(result, speechFrame...)
	return result, nil
}

func (amr *Amr) handleMissingSamples(timestamp uint32) (result []byte) {
	if amr.timestamp != 0 {
		lostSamplesFromPrevious := ((timestamp - amr.timestamp) / (uint32(amr.sampleRate) / 50)) - 1
		log.Sdebug("lostSamplesFromPrevious: %d, time: %d", lostSamplesFromPrevious, lostSamplesFromPrevious*20)
		for i := lostSamplesFromPrevious; i > 0; i-- {
			if amr.isWideBand() {
				result = append(result, 0xFC)
			} else {
				result = append(result, 0x7C)
			}
		}
	}
	return result
}

func (amr *Amr) getSpeechFrameByteSize(frameType int) (size int) {
	if amr.isWideBand() {
		size = AMR_WB_FRAME_SIZE[frameType]
	} else {
		size = AMR_NB_FRAME_SIZE[frameType]
	}
	return
}

func (amr *Amr) handleOaMode(timestamp uint32, payload []byte) ([]byte, error) {

	var result []byte
	var currentTimestamp uint32

	frame := 0
	rtpFrameHeader := payload[0:]
	// payload header := [CMR(4bit)[R(4bit)][ILL(4bit)(opt)][ILP(4bit)(opt)]
	// TOC := [F][FT(4bit)][Q][P][P]
	// storage := [0][FT(4bit)][Q][0][0]
	cmr := (rtpFrameHeader[0] & 0xF0) >> 4
	isLastFrame := (rtpFrameHeader[1]&0x80)&0x80 == 0x00
	frameType := (rtpFrameHeader[1] & 0x78) >> 3
	quality := (rtpFrameHeader[1]&0x04)&0x04 == 0x04

	log.Sdebug("octet-aligned, lastFrame:%t, cmr:%d, frameType:%d, quality:%t",
		isLastFrame, cmr, frameType, quality)

	speechFrameHeader := frameType << 3
	speechFrameHeader = speechFrameHeader | (rtpFrameHeader[1] & 0x04)

	speechFrameSize := amr.getSpeechFrameByteSize(int(frameType))

	currentTimestamp = timestamp + uint32((amr.sampleRate/50)*frame)

	if !isLastFrame {
		log.Warn("Amr does not suport more than one frame per payload - discarted")
		return nil, errors.New("Amr does not suport more than one frame per payload")
	}

	result = append(result, speechFrameHeader)

	if speechFrameSize != 0 {
		speechPayload := rtpFrameHeader[2 : 2+speechFrameSize]
		result = append(result, speechPayload...)
	}
	amr.timestamp = currentTimestamp
	return result, nil
}

func (amr *Amr) handleBeMode(timestamp uint32, payload []byte) ([]byte, error) {
	var result []byte
	var currentTimestamp uint32

	frame := 0
	rtpFrameHeader := payload[0:]
	// packing frame with TOC: frame type and quality bit
	// RTP=[CMR(4bit)[F][FT(4bit)][Q][..speechFrame]] -> storage=[0][FT(4bit)][Q][0][0]
	cmr := (rtpFrameHeader[0] & 0xF0) >> 4
	isLastFrame := (rtpFrameHeader[0]&0x08)>>3 == 0x00
	frameType := (rtpFrameHeader[0]&0x07)<<1 | (rtpFrameHeader[1]&0x80)>>7
	quality := (rtpFrameHeader[1] & 0x40) == 0x40

	log.Sdebug("bandwidth-efficient, lastFrame:%t, cmr:%d, frameType:%d, quality:%t",
		isLastFrame, cmr, frameType, quality)

	speechFrameHeader := (rtpFrameHeader[0]&0x07)<<4 | (rtpFrameHeader[1]&0x80)>>4
	speechFrameHeader = speechFrameHeader | (rtpFrameHeader[1]&0x40)>>4

	speechFrameSize := amr.getSpeechFrameByteSize(int(frameType))

	currentTimestamp = timestamp + uint32((amr.sampleRate/50)*frame)

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
			if k+1 < speechFrameSize {
				speechFrame[k] = speechFrame[k] | (speechPayload[k+1]&0xC0)>>6
			}
		}
		result = append(result, speechFrame...)
	}
	amr.timestamp = currentTimestamp
	return result, nil
}

var AmrMetadata = CodecMetadata{
	Name:     "amr",
	LongName: "Adaptative Multi Rate",
	Options: []CodecOption{
		amrSampleRateOption,
		amrOctetAlignedOption,
	},
	Init: NewAmr,
}

var amrOctetAlignedOption = CodecOption{
	Required:         true,
	Name:             "octet-aligned",
	Description:      "whether this payload is octet-aligned or bandwidth-efficient",
	ValidValues:      []string{"0", "1"},
	ValueDescription: []string{"bandwidth-efficient", "octet-aligned"},
	RestrictValues:   true,
}

var amrSampleRateOption = CodecOption{
	Required:         true,
	Name:             "sample-rate",
	Description:      "whether this payload is narrow or wide band",
	ValidValues:      []string{"nb", "wb"},
	ValueDescription: []string{"Narrow Band (8000)", "Wide Band (16000)"},
	RestrictValues:   true,
}
