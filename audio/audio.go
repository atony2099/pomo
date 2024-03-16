package audio

import (
	"bytes"
	"io"
	"math/rand"
	"time"

	_ "embed"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

// enum for the audio files
type SoundType int

const (
	Finish SoundType = iota
	Interrupt
	Breaks
)

// Embed the MP3 files from the assets directory
//
//go:embed assets/happy_1.mp3
var happy1MP3 []byte

//go:embed assets/happy_2.mp3
var happy2MP3 []byte

//go:embed assets/happy_3.mp3
var happy3MP3 []byte

//go:embed assets/good_job.mp3
var goodJob []byte

//go:embed  assets/interrupt.mp3
var interruptMP3 []byte

func PlaySound(types SoundType) {

	var data []byte
	var err error

	var succesList = [][]byte{happy1MP3, happy2MP3, happy3MP3, goodJob}
	randomIndex := rand.Intn(len(succesList))

	switch types {
	case Finish:
		data = succesList[randomIndex]
	case Interrupt:
		data = interruptMP3
	case Breaks:
		data = interruptMP3
	}

	r := bytes.NewReader(data)
	readCloser := io.NopCloser(r)

	streamer, format, err := mp3.Decode(readCloser)
	if err != nil {
		panic(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done // Wait for the callback to signal that the sound has finished playing
}
