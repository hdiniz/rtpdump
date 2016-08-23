# rtpdump

Extract media files from RTP streams in pcap format

## codec support

This program is intended to support usual audio/video codecs used on IMS networks (VoLTE/VoWiFi).  
Therefore, some codecs might be limited to usual scenarios on these networks.

+ AMR - [RFC 4867](https://tools.ietf.org/html/rfc4867)  
  Supports bandwidth-efficient and octet-aligned modes.  
  Single-channel, single-frame per packet only.
+ EVS - [3GPP TS 26.445](http://www.3gpp.org/DynaReport/26445.htm)  
  *Not yet supported.*
+ H263 - [RFC 2190](https://tools.ietf.org/html/rfc2190)  
  *Not yet supported.*
+ H264 - [RFC 6184](https://tools.ietf.org/html/rfc6184)  
  *Not yet supported.*

## usage

+ rtpdump streams [pcap]  
  displays RTP streams
+ rtpdump dump -i [pcap]
  dumps a media stream.  
  `-i` options is for interactive dump. Codecs and modes are choosen via prompt.  
  **Currently only mode available**

## compiling

Checkout [gopacket](https://github.com/google/gopacket).
Linux should be straightforward.  
For Windows, make sure mingw(32/64) toolchain is on PATH for gopacket WinPcap dependency. Install WinPcap on standard location `C:\WpdPack`

## planned features

1. Support for H264
2. Include stream analisys, packets lost, jitter, etc
3. Media player directly from pcap. ffmpeg support.
4. Jitter buffer to simulate original condition, i.e. packet loss due to jitter

## contributions

Are always appreciated.
