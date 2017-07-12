package main

import (
	"bytes"
	"errors"
	"github.com/jonas747/ogg"
	"io"
	"os"
	"os/exec"
	"strings"
)

/*

These functions help with audio conversion and prepare files for transmission

*/

// MakeAudioBuffer will return a [][]byte for the provided ogg/opus path
func MakeAudioBuffer(path string) (output [][]byte, err error) {
	reader, err := os.Open(path)
	defer reader.Close()
	if err != nil {
		return output, err
	}
	oggdecoder := ogg.NewDecoder(reader)
	packetdecoder := ogg.NewPacketDecoder(oggdecoder)

	for {
		packet, _, err := packetdecoder.Decode()
		if err != nil {
			return output, err
		}
		output = append(output, packet)
	}
}

// ToOpus will convert the provided file (which always comes out of "./tmp") to an opus/ogg file for buffering.
// Note that this function merely splits the pathname (file.mp4 to "file") and passes it to ConvertAudio for the rest.
func ToOpus(path string) (err error) {

	args := strings.Split(path, ".")
	if len(args) < 2 {
		return errors.New("Invalid target")
	}

	err = ConvertAudio("tmp/"+path, "tmp/"+args[0]+".opus")
	if err != nil {
		return err
	}
	return nil
}

// ConvertAudio will convert the provided file (typically in mp4 format) to opus
func ConvertAudio(from string, to string) (err error) {

	avconv := exec.Command("avconv", "-i", from, "-f", "wav", "-")
	opusenc := exec.Command("opusenc", "--bitrate", "256", "-", to)

	r, w := io.Pipe()
	avconv.Stdout = w
	opusenc.Stdin = r
	defer r.Close()

	var b2 bytes.Buffer
	opusenc.Stdout = &b2

	if err = avconv.Start(); err != nil {
		return err
	}
	if err = opusenc.Start(); err != nil {
		return err
	}
	if err = avconv.Wait(); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	if err = opusenc.Wait(); err != nil {
		return err
	}
	//io.Copy(os.Stdout, &b2)
	return nil
}
