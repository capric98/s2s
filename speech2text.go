package s2s

import (
	"context"
	"log"
	"strings"
	"time"

	speech "cloud.google.com/go/speech/apiv1p1beta1"
	"github.com/golang/protobuf/ptypes/duration"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1p1beta1"
)

type Sentence struct {
	w    []Word
	cont string
	// position
	p, wn int
	Trans string
}

type Word struct {
	Start, End time.Duration
	// Content
	C string
}

var (
	threshold      = 650 * time.Millisecond
	splitThreshold = 5
)

func Recognize(gsUri string, language string, speakerN int, moji bool) (result [][]Sentence, e error) {
	log.Println("Start recognizing...")

	var client *speech.Client
	ctx := context.Background()
	client, e = speech.NewClient(ctx)
	if e != nil {
		return
	}
	defer client.Close()
	req := &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:                 speechpb.RecognitionConfig_FLAC,
			SampleRateHertz:          SAMPLERATE,
			LanguageCode:             language,
			EnableWordTimeOffsets:    true,
			AudioChannelCount:        2,
			EnableSpeakerDiarization: speakerN != 0,
			DiarizationSpeakerCount:  int32(speakerN),
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Uri{Uri: gsUri},
		},
	}

	op, e := client.LongRunningRecognize(ctx, req)
	if e != nil {
		log.Println(e)
		return
	}

	log.Println("Processing...")
	pstart := time.Now()

	resp, _ := op.Wait(ctx)
	log.Printf("Speech to Text finished in %v.\n\n", time.Since(pstart))

	var tagThreshold = 1
	var rcount []int
	var lst []time.Duration
	if speakerN == 0 {
		tagThreshold = 0
		result = make([][]Sentence, 1)
		result[0] = make([]Sentence, 1)
		result[0][0].w = make([]Word, 0)
		rcount = make([]int, 1)
		lst = make([]time.Duration, 1)
	} else {
		result = make([][]Sentence, speakerN)
		for i := 0; i < speakerN; i++ {
			result[i] = make([]Sentence, 1)
			result[i][0].w = make([]Word, 0)
		}
		rcount = make([]int, speakerN)
		lst = make([]time.Duration, speakerN)
	}

	for _, res := range resp.Results {
		for _, alt := range res.Alternatives {
			for _, ws := range alt.GetWords() {
				tag := int(ws.GetSpeakerTag()) - tagThreshold
				if tag < 0 {
					continue
				}
				word := strings.Split(ws.GetWord(), "|")[0]
				start := toDuration(ws.StartTime)
				end := toDuration(ws.EndTime)

				// not start-lst
				// Google's magic!
				if lst[tag] != 0 && end-lst[tag] >= threshold {
					if end-lst[tag] > 3*threshold || len([]rune(result[tag][rcount[tag]].cont)) > splitThreshold {
						result[tag] = append(result[tag], Sentence{
							w: make([]Word, 0),
						})
						//log.Println(tag, ":", result[tag][rcount[tag]].cont)
						rcount[tag]++
					}
				}
				lst[tag] = end

				//log.Println(word)
				if !moji {
					result[tag][rcount[tag]].cont += " "
				}
				result[tag][rcount[tag]].cont += word
				result[tag][rcount[tag]].wn++
				result[tag][rcount[tag]].w = append(result[tag][rcount[tag]].w,
					Word{
						C:     word,
						Start: start,
						End:   end,
					})
			}
		}
	}
	// for i := 0; i < len(result); i++ {
	// 	if len(result[i][rcount[i]].cont) != 0 {
	// 		log.Println(i, ":", result[i][rcount[i]].cont)
	// 	}
	// }

	return
}

func (s *Sentence) IsEnd() bool {
	return s.p == s.wn
}

func (s *Sentence) Pop() (w Word) {
	w = s.w[s.p]
	s.p++
	return
}

func (s *Sentence) Content() string {
	return s.cont
}

func (s *Sentence) Reset() {
	s.p = 0
}

func toDuration(t *duration.Duration) time.Duration {
	return time.Duration(t.GetSeconds())*time.Second + time.Duration(t.GetNanos())
}
