package s2s

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	speech "cloud.google.com/go/speech/apiv1p1beta1"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1p1beta1"
)

var (
	threshold = 650 * time.Second
)

func Recognize(gsUri string, language string) (e error) {
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
			EnableSpeakerDiarization: true,
			DiarizationSpeakerCount:  3,
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Uri{Uri: gsUri},
		},
	}

	op, e := client.LongRunningRecognize(ctx, req)
	if e != nil {
		fmt.Println(e)
		return
	}

	//start:=time.Now()
	log.Println("Processing...")

	resp, _ := op.Wait(ctx)
	var lst time.Duration
	for _, result := range resp.Results {
		speaker := make([]string, 5)
		for _, alt := range result.Alternatives {
			//fmt.Printf("\"%v\" (confidence=%3f)\n", alt.Transcript, alt.Confidence)
			words := alt.GetWords()
			for _, word := range words {
				wordsplit := strings.Split(word.GetWord(), "|")
				st := word.GetStartTime()
				et := word.GetEndTime()
				etd := time.Duration(et.GetSeconds())*time.Second + time.Duration(et.GetNanos())

				if etd-lst >= threshold {
					fmt.Println("++++++++++++", etd-lst)
				}
				fmt.Printf("%v: \"%v\", start:%vs%vms, end:%vs%vms\n", word.GetSpeakerTag(), wordsplit[0], st.GetSeconds(), st.GetNanos()/1000000, et.GetSeconds(), et.GetNanos()/1000000)
				speaker[word.GetSpeakerTag()] += wordsplit[0]
				lst = etd
			}
			// for _, v := range speaker {
			// 	fmt.Println("===============================")
			// 	fmt.Println(v)
			// }

		}
	}

	log.Println("Finish")

	return
}
