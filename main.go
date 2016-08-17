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
      fmt.Scanf("%d\n", &n)
      fmt.Printf("%s\n", rtpStreams[n-1])
      stream = rtpStreams[n-1]
    }

    fmt.Printf("\nMedia type:" +
        "\n\t(1) Audio" +
        "\n\t(2) Video\n")

    fmt.Print("[n]: " )
    fmt.Scanf("%d\n", &n)

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
      fmt.Scanf("%d\n", &n)
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
}
