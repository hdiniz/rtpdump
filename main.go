package main

import (
    "fmt"
    "os"
    "github.com/urfave/cli"
    "github.com/hdiniz/rtpdump/codecs"
    "github.com/hdiniz/rtpdump/console"
    "github.com/hdiniz/rtpdump/log"
    "github.com/hdiniz/rtpdump/rtp"
)


var streamsCmd = func (c *cli.Context) error {
  inputFile := c.Args().First()

  if len(c.Args()) <= 0 {
    cli.ShowCommandHelp(c, "streams");
    return cli.NewExitError("wrong usage for streams", 1)
  }

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
    cli.ShowCommandHelp(c, "dump");
    return cli.NewExitError("wrong usage for dump", 1)
  }

  rtpReader, err := rtp.NewRtpReader(inputFile)

  if err != nil {
      return cli.NewMultiError(cli.NewExitError("failed to open file", 1), err)
  }

  defer rtpReader.Close()

  if interactive {
    return doInteractiveDump(c, rtpReader)
  } else {
    return cli.NewExitError("non interactive not supported", 1)
  }
}

func doInteractiveDump(c *cli.Context, rtpReader *rtp.RtpReader) error {
  rtpStreams := rtpReader.GetStreams()

  var rtpStreamsOptions []string
  for _,v := range rtpStreams {
    rtpStreamsOptions = append(rtpStreamsOptions, v.String())
  }

  streamIndex, err := console.ExpectIntRange(
    1,
    len(rtpStreams),
    console.ListPrompt("Choose RTP Stream", rtpStreamsOptions...))

  if err != nil {
    return cli.NewMultiError(cli.NewExitError("invalid input", 1), err)
  }
  fmt.Printf("(%-3d) %s\n\n", streamIndex, rtpStreams[streamIndex-1])

  var codecList []string
  for _,v := range codecs.CodecList {
    codecList = append(codecList, v.Name)
  }
  codecIndex, err := console.ExpectIntRange(
    1,
    len(codecs.CodecList),
    console.ListPrompt("Choose codec:", codecList...))

  if err != nil {
    return cli.NewMultiError(cli.NewExitError("invalid input", 1), err)
  }
  fmt.Printf("(%-3d) %s\n\n", codecIndex, codecs.CodecList[codecIndex-1].Name)

  codecMetadata := codecs.CodecList[codecIndex-1]

  optionsMap := make(map[string]string)
  for _,v := range codecMetadata.Options {
    var optionValue string
    if v.RestrictValues {
      optionValue, err = console.ExpectRestrictedString(
        v.ValidValues,
        console.KeyValuePrompt(fmt.Sprintf("%s - %s", v.Name, v.Description),
          v.ValidValues, v.ValueDescription))
    } else {
      optionValue, err = console.ExpectAnyString(
        console.Prompt(fmt.Sprintf("%s - %s: ", v.Name, v.Description)))
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
    frames, err := codec.HandleRtpPacket(r)
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
    app.Version = "0.5.0"
    cli.AppHelpTemplate += `
     /\_/\
    ( o.o )
     > ^ <
    `

    app.Before = func(c *cli.Context) error {
      if c.GlobalBool("debug") {
        log.SetLevel(log.TRACE)
      } else {
        log.SetLevel(log.INFO)
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
