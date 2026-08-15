package lib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"time"
)

type WorldType struct {
	mask      uint32
	worldType string
}

// https://github.com/runelite/api.runelite.net/blob/99b746814edd84b111e73aa3fd2a0d6663e093c8/http-service/src/main/java/net/runelite/http/service/worlds/ServiceWorldType.java
var excludedWorldTypes = []WorldType{
	{mask: 1 << 2, worldType: "PvP"},
	{mask: 1 << 6, worldType: "PvP Arena"},
	{mask: 1 << 8, worldType: "Quest Speedrunning"},
	{mask: 1 << 16, worldType: "Beta"},
	{mask: 1 << 22, worldType: "Legacy only"},
	{mask: 1 << 26, worldType: "Tournament"},
	{mask: 1 << 27, worldType: "Fresh start world"},
	{mask: 1 << 29, worldType: "Deadman"},
	{mask: 1 << 30, worldType: "Seasonal"},
}

func GetExcludedWorlds() (*[]uint16, error) {
	client := http.Client{
		Timeout: time.Second * time.Duration(ApiTimeout),
	}
	req, err := http.NewRequest(http.MethodGet, WorldsUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %f", err)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get stars: %f", err)
	}

	if res.Body != nil {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				panic(err)
			}
		}(res.Body)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %f", err)
	}

	buf := bytes.NewBuffer(body)

	readInt(buf) // length
	num := readShort(buf)

	var excluded []uint16

	for i := 0; i < int(num); i++ {
		id := readShort(buf)
		worldType := readInt(buf)
		_, _ = readString(buf) // address
		_, _ = readString(buf) // activity
		_, _ = buf.ReadByte()  // location
		_ = readShort(buf)     // players

		if isExcluded(uint32(worldType)) {
			excluded = append(excluded, id)
		}
	}

	return &excluded, nil
}

func isExcluded(mask uint32) bool {
	for _, worldType := range excludedWorldTypes {
		if mask == 0 || (mask&worldType.mask) != 0 {
			return true
		}
	}
	return false
}

func readInt(buf *bytes.Buffer) int32 {
	var result int32
	_ = binary.Read(buf, binary.BigEndian, &result)
	return result
}

func readShort(buf *bytes.Buffer) uint16 {
	var result uint16
	_ = binary.Read(buf, binary.BigEndian, &result)
	return result
}

func readString(buf *bytes.Buffer) (string, error) {
	line, err := buf.ReadBytes(0)
	if err != nil && err != io.EOF {
		return "", err
	}
	if len(line) == 0 {
		return "", nil
	}
	if line[len(line)-1] == 0 {
		line = line[:len(line)-1]
	}
	return string(line), nil
}
