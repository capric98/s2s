package s2s

import (
	"log"
	"testing"
)

func TestEncodeToFLAC(t *testing.T) {
	if e := EncodeToFLAC("test/media.mp4", "test/media-encode.flac"); e != nil {
		log.Println(e)
		t.Fail()
	}
}
