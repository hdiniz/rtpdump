package esp

import (
	"bufio"
	"crypto/cipher"
	"crypto/des"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/hdiniz/rtpdump/log"
)

//EncKey describes IPSec Enc keys
type EncKey struct {
	algorithm string
	spi       uint32
	key       []byte
}

var keyList map[uint32]*EncKey

//LoadKeyFile load key file
func LoadKeyFile(path string) error {
	keyList = make(map[uint32]*EncKey)
	// open a file
	if file, err := os.Open(path); err == nil {

		// make sure it gets closed
		defer file.Close()

		// create a new scanner and read the file line by line
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var spiStr, algStr, keyStr string
			n, err2 := fmt.Sscanf(scanner.Text(), "%s %s %s", &spiStr, &algStr, &keyStr)
			if err2 != nil || n != 3 {
				continue
			}
			spi, err3 := spiHexToInt(spiStr)
			if err3 != nil {
				continue
			}
			key := bytesFromHex(keyStr)
			if key == nil {
				continue
			}
			keyList[spi] = &EncKey{
				algorithm: algStr,
				key:       key,
			}
		}

	} else {
		return err
	}
	return nil
}

//DecodeESPLayer decrypts and returns payload
func DecodeESPLayer(packet gopacket.Packet, esp *layers.IPSecESP) gopacket.Packet {

	entry := getKeyEntry(esp.SPI)
	if entry == nil {
		return nil
	}
	key := entry.key
	algorithm := entry.algorithm

	var clearData []byte

	if algorithm == "des3_cbc" {
		payloadLen := len(esp.Encrypted[8:])
		cipherData := make([]byte, payloadLen+payloadLen%8)
		copy(cipherData, esp.Encrypted[8:])

		iv := esp.Encrypted[:8]
		blockCipher, _ := des.NewTripleDESCipher(key)
		mode := cipher.NewCBCDecrypter(blockCipher, iv)

		clearData = make([]byte, len(cipherData)+1)
		mode.CryptBlocks(clearData, cipherData)
	} else {
		return nil
	}
	return makePacket(packet, clearData)
}

func makePacket(packet gopacket.Packet, d []byte) gopacket.Packet {
	espPacket := gopacket.NewPacket(d, layers.LayerTypeIPv4, gopacket.Default)
	if espPacket.ErrorLayer() != nil {
		espPacket = gopacket.NewPacket(d, layers.LayerTypeIPv6, gopacket.Default)
	}
	return espPacket
}

func bytesFromHex(s string) []byte {
	hexString := strings.TrimPrefix(s, "0x")
	b, err := hex.DecodeString(hexString)
	if err != nil {
		log.Debug("failed to convert hex string")
		return nil
	}
	return b
}

func spiHexToInt(s string) (uint32, error) {
	n, err := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 32)
	return uint32(n), err
}

func getKeyEntry(spi uint32) *EncKey {
	return keyList[spi]
}

//0x10e142e956c808bce1d763f369062e53d6bffc40eaf47ff7
