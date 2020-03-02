package s2s

import (
	"os/exec"
	"strconv"
)

const (
	// Default 44.1kHz sample rate.
	SAMPLERATE = 44100
)

var (
	FFmpeg, ffmpegOK = exec.LookPath("ffmpeg")
)

func EncodeToFLAC(file, task string) (e error) {
	if ffmpegOK != nil {
		e = ffmpegOK
		return
	}

	cmd := exec.Command(FFmpeg, "-i", file,
		"-ar", strconv.Itoa(SAMPLERATE), "-ac", "2", "-sample_fmt", "s16",
		"-vn", "-hide_banner", "-y",
		task)
	//cmd.Stderr = os.Stderr
	e = cmd.Run()
	//_, _ = os.Stderr.Write([]byte("\n\n"))
	return
}
