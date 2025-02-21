package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	BLUETOOTHCTL = "bluetoothctl"
)

var (
	listen        = flag.String("listen", "[::]:8500", "Listen to this IP:port")
	rfcommChannel = flag.Int("rfcomm-channel", 2, "RFCOMM channel (defaults to 2)")
	hwaddr        = flag.String("hwaddr", "", "Hardware MAC address for bluetooth connection and rfcomm bind")
	pingInterval  = flag.Duration("ping-interval", 0, "Interval between ping packets (set to 0 to disable)")
	//timeout       = flag.Duration("timeout", time.Second, "Timeout for response")

	ser   RFCommSocket
	mutex sync.Mutex
)

func fatalError(f string, args ...any) {
	slog.Error("fatal error", "err", fmt.Sprintf(f, args...))
	os.Exit(99)
}

func openPort(reopen bool) {
	mutex.Lock()
	defer mutex.Unlock()

	var err error

	if ser.IsConnected() {
		if !reopen {
			return
		}
		err := ser.Close()
		if err != nil {
			slog.Error("error closing port", "err", err)
		} else {
			slog.Info("existing port closed")
		}
		time.Sleep(3 * time.Second)
	}

	if *hwaddr == "" {
		fatalError("-hwaddr parameter must be set")
	}

	for {
		err = ser.Connect(*hwaddr, uint8(*rfcommChannel))
		if err == nil {
			break
		}
		slog.Error("error opening port", "err", err)
		time.Sleep(1 * time.Second)
	}

	/*err = ser.SetNonBlocking()
	if err != nil {
		slog.Error("error setting the socket non-blocking", "err", err)
	}*/

	slog.Info("port opened")
}

func dumpLines(data []byte) []string {
	lines := strings.Split(hex.Dump(data), "|\n")
	if len(lines) > 0 {
		return lines[0 : len(lines)-1]
	}
	return lines
}

func sendHex(hexs string) {
	hexs = removeSpaces(hexs)

	data, err := hex.DecodeString(hexs)
	if err != nil {
		slog.Error("error decoding hex string", "string", hexs, "err", err)
		return
	}

	retry := 3

	for {
		retry -= 1
		if retry == -1 {
			break
		}

		mutex.Lock()
		_, err = ser.Write(data)
		mutex.Unlock()

		if err != nil {
			slog.Error("error writing to port", "data", hex.EncodeToString(data), "err", err, "will_retry", retry > 0)
			openPort(true)
			time.Sleep(time.Second)
			continue // retry
		}

		for _, dumpline := range dumpLines(data) {
			slog.Info("wrote data", "data", dumpline)
		}

		break
	}
}

func removeSpaces(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' {
			result = append(result, s[i])
		}
	}
	return string(result)
}

func respErr(ctx echo.Context, err string) error {
	return ctx.JSON(http.StatusBadRequest, map[string]any{
		"error": err,
	})
}

func respOK(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]any{})
}

func handleQuick(ctx echo.Context) error {
	switch ctx.Param("quick") {
	case "1":
		sendHex("4154 00 08 02 00 01 fd")
	case "2":
		sendHex("4154 00 08 02 00 02 fc")
	case "3":
		sendHex("4154 00 08 02 00 03 fb")
	case "4":
		sendHex("4154 00 08 02 00 04 fa")
	default:
		return respErr(ctx, "quick numbers must be 1-4")
	}
	return respOK(ctx)
}

func handleVolumeUp(ctx echo.Context) error {
	sendHex("4154 07 00 00 00")
	return respOK(ctx)
}

func handleVolumeDown(ctx echo.Context) error {
	sendHex("4154 07 01 00 00")
	return respOK(ctx)
}

func handleMute(ctx echo.Context) error {
	sendHex("4154 07 1d 01 01 fe")
	return respOK(ctx)
}

func handleUnmute(ctx echo.Context) error {
	sendHex("4154 07 1d 01 00 ff")
	return respOK(ctx)
}

func handleInput(ctx echo.Context) error {
	switch strings.ToLower(ctx.Param("input")) {
	case "cbl", "cbl/sat":
		sendHex("4154 00 01 01 52 ad")
	case "dvd", "dvd/blueray", "dvd/blue-ray":
		sendHex("4154 00 01 01 53 ac")
	case "mp", "mediaplayer":
		sendHex("4154 00 01 01 54 ab")
	case "br", "blueray", "blue-ray":
		sendHex("4154 00 01 01 55 aa")
	case "g", "game":
		sendHex("4154 00 01 01 56 a9")
	case "tv", "tvaudio", "tv-audio":
		sendHex("4154 00 01 01 57 a8")
	default:
		return respErr(ctx, "available inputs: cbl, dvd, blueray, game, mediaplayer, tvaudio")
	}
	return respOK(ctx)
}

func handlePower(ctx echo.Context) error {
	switch strings.ToLower(ctx.Param("onoff")) {
	case "off", "0":
		sendHex("4154 00 0a 01 00 ff")
	case "on", "1":
		sendHex("4154 00 0a 01 01 fe")
	default:
		return respErr(ctx, "available power commands: on, off")
	}
	return respOK(ctx)
}

func handleSendHex(ctx echo.Context) error {
	body := ctx.Request().Body
	defer body.Close()

	bts, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	sendHex(string(bts))

	return respOK(ctx)
}

func pinger() {
	for {
		if ser.IsConnected() {
			sendHex("4154 00 0b 00 00") // unknown
		}
		time.Sleep(*pingInterval)
	}
}

func bgReader() {
	for {
		var buf [512]byte

		n, err := ser.Read(buf[:])

		if err != nil {
			time.Sleep(time.Second)
		} else {
			for _, dumpline := range dumpLines(buf[:n]) {
				slog.Info("read data", "data", dumpline)
			}
		}
	}
}

func main() {
	flag.Parse()

	slog.Info("denonbt v0.0.4 started")

	openPort(false)

	// something to kickstart comm with (I don't know what it is but the app started with it)
	//sendHex("4154 00 0a 01 02 fd")
	sendHex("4154 00 0b 00 00") // unknown

	// pinger to keepalive
	// probably not necessary but will ensure the connection is always ready (by trying to re-connect eventually)
	if *pingInterval > 0 {
		go pinger()
	}
	go bgReader()

	e := echo.New()
	e.HideBanner = true

	e.POST("/quick/:quick", handleQuick)
	e.POST("/volup", handleVolumeUp)
	e.POST("/voldn", handleVolumeDown)
	e.POST("/mute", handleMute)
	e.POST("/unmute", handleUnmute)
	e.POST("/input/:input", handleInput)
	e.POST("/power/:onoff", handlePower)
	e.POST("/sendhex", handleSendHex)

	err := e.Start(*listen)
	slog.Error("error listening", "err", err)
}
