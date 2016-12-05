# rtpdump

Extract media files from RTP streams in pcap format

## codec support

This program is intended to support usual audio/video codecs used on IMS networks (VoLTE/VoWiFi).  
Therefore, some codecs might be limited to usual scenarios on these networks.

+ AMR - [RFC 4867](https://tools.ietf.org/html/rfc4867)  
  Supports bandwidth-efficient and octet-aligned modes.  
  Single-channel, single-frame per packet only.
+ H264 - [RFC 6184](https://tools.ietf.org/html/rfc6184)  
Supports Single NAL Mode and some Non-Interleaved Mode streams, due to current lack of STAP-A support  

| Payload Type  	| Support      	|
|---------------	|--------------	|
| 1-23 NAL Unit 	| Yes          	|
| 24 STAP-A     	| No - planned 	|
| 25 STAP-B     	| No           	|
| 26 MTAP16     	| No           	|
| 27 MTAP24     	| Yes          	|
| 28 FU-A       	| Yes          	|
| 29 FU-B       	| No           	|

+ EVS - [3GPP TS 26.445](http://www.3gpp.org/DynaReport/26445.htm)  
  *Not yet supported.*
+ H263 - [RFC 2190](https://tools.ietf.org/html/rfc2190)  
  *Not yet supported.*

## ipsec support

In order to support dumping VoWiFi media some support for ESP (Encapsulating Security Payload) decryption is present.

| Encryption Algorithm | Support       |
|--------------------- |-------------- |
| 3DES CBC             | Yes           |
| DES CBC              | No - Planned  |
| AES CBC              | No - Planned  |

Keys are read from file 'esp-keys.txt' on the current directory. One key per file, for example:
[SPI] [Encryption Algorithm] [Key]
0x00d40016 des3_cbc 0x091199869ec18afd8e38f77eb1252685924937d3921a178e
0xcb97da43 des3_cbc 0xaaa316cd3fa41daa9afe6e8f42a9ae0ce2bd5128cef5a60f

## usage

+ rtpdump streams [pcap]  
  displays RTP streams
+ rtpdump dump [pcap]
  dumps a media stream.

## compiling

Checkout [gopacket](https://github.com/google/gopacket).
Linux should be straightforward.  
For Windows, make sure mingw(32/64) toolchain is on PATH for gopacket WinPcap dependency. Install WinPcap on standard location `C:\WpdPack`

## planned features

1. Support for VoWiFi streams - i.e decoding IPSec packets
2. Include stream analisys, packets lost, jitter, etc
3. Media player directly from pcap. ffmpeg support.
4. Jitter buffer to simulate original condition, i.e. packet loss due to jitter
5. Support multiple speach frames in audio packet

## contributions

Are always appreciated.
