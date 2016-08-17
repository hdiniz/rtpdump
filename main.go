package main

import (
    "fmt"
    "log"
    "os"
    "github.com/hdiniz/rtpdump/rtp"
    "github.com/hdiniz/rtpdump/codecs"
  )

func main() {

    if len(os.Args) <= 1 {
        fmt.Println("usage: pcap-rtp-extractor <pcap>")
        return
    }
    pcapFile := os.Args[1]

    rtpReader, err := rtp.NewRtpReader(pcapFile)

    if err != nil {
        log.Fatal(err)
    }

    defer rtpReader.Close()

    rtpStreams := rtpReader.GetStreams()

    fmt.Println("Choose RTP Stream:")

    for i,v := range rtpStreams {
        fmt.Printf("\t(%2d) - %s\n",i+1, v)
    }

    var n int = 0
    var stream *rtp.RtpStream
    for ;n <= 0 || n >= len(rtpStreams); {
      fmt.Print("[n]: " )
      fmt.Scanf("%d", &n)
      fmt.Printf("%s\n", rtpStreams[n-1])
      stream = rtpStreams[n-1]
    }

    fmt.Printf("\nMedia type:" +
        "\n\t(1) Audio" +
        "\n\t(2) Video\n")

    fmt.Print("[n]: " )
    fmt.Scanf("%d", &n)

    if n != 1 {
      fmt.Println("Video not supported... yet.")
      return
    }

    n = 0
    audioCodecs := codecs.GetAudioCodecs()
    for ;n <= 0 || n > len(audioCodecs); {
      fmt.Printf("\nChoose codec:")
      for i,v := range audioCodecs {
        fmt.Printf("\n\t(%d) - %s", i+1, v.Name())
      }
      fmt.Print("\n[n]: " )
      fmt.Scanf("%d", &n)
    }
    decoder := audioCodecs[n-1]

    options := decoder.GetOptions()
    fmt.Printf("\nCodec options:\n")
    for _,v := range options {
      var value string
      fmt.Printf("%s\nrequired: %t\ndefault: %s\n[value]: ",
        v.Description, v.Required, v.Default)
      fmt.Scanln(&value)
      if (value == "") {
        value = v.Default
      }
      v.Value = value
      fmt.Printf("\nset: %s\n", value)
    }
    decoder.SetOptions(options)

    fmt.Printf("\nOutput file: ")
    var outputFile string
    fmt.Scanln(&outputFile)

    for _,p := range stream.RtpPackets {
      decoder.ProcessRtpPacket(p)
    }

    f, err := os.Create(outputFile)
    defer f.Close()
    f.Write(decoder.GetStorageFormat())
    f.Sync()

/*
    f, err := os.Create(outputFile)
    defer f.Close()
    if err != nil {
        log.Panic(err)
    }
    amrMagic := "#!AMR\n"
    switch n {
        case 1:
        case 2:
        case 3:
        case 4:
            f.Write([]byte(amrMagic))
            f.Sync()
            fmt.Println("Wrote file")
            for _,p := range stream.RtpPackets {
                amrRtpPayload := p.Payload
                frameHeader := []byte{0x00}
                frameHeader[0] = (amrRtpPayload[0] & 0x07)<<4 | (amrRtpPayload[1] & 0x80)>>4
                frameHeader[0] = frameHeader[0] | (amrRtpPayload[1] & 0x40)>>4
                f.Write(frameHeader)

                speechPayload := amrRtpPayload[1:]
                speechFrame := make([]byte, len(speechPayload))
                for k := 0; k < len(speechPayload)-2; k++ {
                    speechFrame[k] = (speechPayload[k] & 0x3F) << 2 | (speechPayload[k+1] & 0xC0)>>6
                }
                f.Write(speechFrame)
            }
            f.Sync()
    }*/

}
