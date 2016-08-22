package main

import (
    "fmt"
    log "github.com/Sirupsen/logrus"
    "os"
    "github.com/urfave/cli"
    "github.com/hdiniz/rtpdump/codecs"
    "github.com/hdiniz/rtpdump/console"
    "github.com/hdiniz/rtpdump/rtp"
)


var streamsCmd = func (c *cli.Context) error {
  inputFile := c.Args().First()

  if len(c.Args()) <= 0 {
    log.Error("wrong usage for streams")
    cli.ShowCommandHelp(c, "streams");
    return cli.NewExitError("wrong usage for streams", 1)
  }

  log.WithFields(log.Fields{
      "input-file": inputFile,
  }).Debug("streams cmd invoked")

  rtpReader, err := rtp.NewRtpReader(inputFile)

  if err != nil {
      return cli.NewMultiError(cli.NewExitError("failed to open file", 1), err)
  }

  defer rtpReader.Close()

  rtpStreams := rtpReader.GetStreams()

  for _,v := range rtpStreams {
    fmt.Printf("%s\n", v)
  }

  return nil
}

var dumpCmd = func (c *cli.Context) error {
  inputFile := c.Args().First()
  outputFile := c.Args().Get(1)

  interactive := c.Bool("interactive")
  codec := c.String("codec")
  codecOptions := c.String("codec-options")

  if inputFile == "" ||
    (!interactive &&
    (codec == "" || codecOptions == "" || outputFile == "")) {

    log.Error("wrong usage for dump")
    cli.ShowCommandHelp(c, "dump");
    return cli.NewExitError("wrong usage for dump", 1)
  }

  log.WithFields(log.Fields{
      "input-file": inputFile,
      "output-file": outputFile,
      "interactive": interactive,
      "codec": codec,
      "codec-options": codecOptions,
  }).Debug("dump cmd invoked")

  rtpReader, err := rtp.NewRtpReader(inputFile)

  if err != nil {
      return cli.NewMultiError(cli.NewExitError("failed to open file", 1), err)
  }

  defer rtpReader.Close()

  if interactive {
    return doInteractiveDump(c, rtpReader)
  } else {
    log.Error("non interactive not supported at the moment")
    return cli.NewExitError("not supported", 1)
  }
}

func doInteractiveDump(c *cli.Context, rtpReader *rtp.RtpReader) error {
  rtpStreams := rtpReader.GetStreams()

  var chooseRtpStream = func(attempts int) error {
    fmt.Println("Choose RTP Stream:")
    for i,v := range rtpStreams {
      fmt.Printf("(%03d) %s\n", i+1, v)
    }
    fmt.Printf("[%d-%d]: ", 1, len(rtpStreams))
    return nil
  }
  streamIndex, err := console.ExpectIntRange(1, len(rtpStreams), chooseRtpStream)
  if err != nil {
    return cli.NewMultiError(cli.NewExitError("invalid input", 1), err)
  }
  fmt.Printf("(%-3d) %s\n\n", streamIndex, rtpStreams[streamIndex-1])

  var chooseRtpCodec = func(attempts int) error {
    fmt.Println("Choose codec:")
    for i,v := range codecs.CodecList {
      fmt.Printf("(%03d) %s\n", i+1, v.Name)
    }
    fmt.Printf("[%d-%d]: ", 1, len(codecs.CodecList))
    return nil
  }

  codecIndex, err := console.ExpectIntRange(1, len(codecs.CodecList), chooseRtpCodec)
  if err != nil {
    return cli.NewMultiError(cli.NewExitError("invalid input", 1), err)
  }
  fmt.Printf("(%-3d) %s\n\n", codecIndex, codecs.CodecList[codecIndex-1].Name)


  codecMetadata := codecs.CodecList[codecIndex-1]

  optionsMap := make(map[string]string)
  for _,v := range codecMetadata.Options {
    var chooseCodecOption = func(attempts int) error {
      fmt.Printf("%s - %s\n", v.Name, v.Description)
      if v.RestrictValues {
        for k,rv := range v.ValidValues {
          fmt.Printf("(%s) %s\n", rv, v.ValueDescription[k])
        }
      }
      return nil
    }
    var optionValue string
    if v.RestrictValues {
      optionValue, err = console.ExpectRestrictedString(v.ValidValues, chooseCodecOption)
    } else {
      optionValue, err = console.ExpectAnyString(chooseCodecOption)
    }

    if err != nil {
      return cli.NewMultiError(cli.NewExitError("invalid input", 1), err)
    }
    optionsMap[v.Name] = optionValue
  }


  outputFile, err := console.ExpectAnyString(console.Prompt("Output file: "))

  if err != nil {
    return cli.NewMultiError(cli.NewExitError("invalid input", 1), err)
  }

  fmt.Printf("%s\n", outputFile)


  codec := codecMetadata.Init()
  err = codec.SetOptions(optionsMap)

  if err != nil {
    return err
  }

  codec.Init()

  f, err := os.Create(outputFile)
  defer f.Close()
  f.Write(codec.GetFormatMagic())
  for _,r := range rtpStreams[streamIndex-1].RtpPackets {
    frames, err := codec.HandleRtpPacket(r.Timestamp, r.Payload)
    if err == nil {
      f.Write(frames)
    }
  }
  f.Sync()


  return nil
}

func codecsList(c *cli.Context) error {
  codec := c.Args().First()
  found := codec == ""

  for _,v := range codecs.CodecList {
    if found || codec == v.Name {
      fmt.Printf("%s\n", v.Describe())
      found = true
    }
  }

  if !found {
    fmt.Printf("Codec %s not available\n", codec)
  }
  return nil
}


func main() {

    app := cli.NewApp()
    app.Name = "rtpdump"
    app.Version = "0.1.0"
    cli.AppHelpTemplate += `
     /\_/\
    ( o.o )
     > ^ <
    `

    app.Before = func(c *cli.Context) error {
      log.SetOutput(os.Stdout)
      if c.GlobalBool("debug") {
        log.SetLevel(log.DebugLevel)
      } else {
        log.SetLevel(log.WarnLevel)
      }
      return nil
    }

    app.Flags = []cli.Flag{
      cli.BoolFlag{
        Name: "debug, d",
        Usage: "use for debug logs",
      },
    }

    app.Commands = []cli.Command{
      {
          Name: "streams",
          Aliases: []string{"s"},
          Usage: "display rtp streams in pcap file",
          ArgsUsage: "[pcap file]",
          Action: streamsCmd,
      },
      {
          Name: "dump",
          Aliases: []string{"d"},
          Usage: "dumps rtp payload to file",
          ArgsUsage: "[pcap file] [output file]",
          Action: dumpCmd,
          Flags: []cli.Flag {
            cli.BoolFlag{
              Name: "interactive, i",
              Usage: "enables interactive prompt to choose stream",
            },
            cli.StringFlag{
              Name: "ssrc, s",
              Usage: "SSRC of the stream to be decoded",
            },
            cli.StringFlag{
              Name: "codec, c",
              Usage: "codec to be used for stream dump",
            },
            cli.StringFlag{
              Name: "codec-options, co",
              Usage: "options for this codec dump",
            },
          },
      },
      {
        Name: "codecs",
        Aliases: []string{"c"},
        Usage: "lists supported codecs information",
        Subcommands: cli.Commands{
          cli.Command{
            Name: "list",
            Action: codecsList,
            ArgsUsage: "[codec name or empty for all]",
          },
        },
      },
    }

    app.Run(os.Args)
}
