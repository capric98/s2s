package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	speech "cloud.google.com/go/speech/apiv1p1beta1"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1p1beta1"
)

const (
	ProjectID   = ""
	Credentials = "credentials.json"
	BucketName  = ""
)

func bukOpt(bukname string, objname string, opt string, localfile string) {
	bucket, object := bukname, objname

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	switch opt {
	case "upload":
		if err := bukwrite(client, bucket, object, localfile); err != nil {
			log.Fatalf("Cannot write object: %v", err)
		}
	case "delete":
		if err := bukdelete(client, bucket, object); err != nil {
			log.Fatalf("Cannot to delete object: %v", err)
		}
	}
}

func checkBucket() error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	it := client.Bucket(BucketName).Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		log.Println(attrs.Name)
	}
	return nil
}
func createBucket() {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	bucket := client.Bucket(BucketName)
	if err := bucket.Create(ctx, ProjectID, nil); err != nil {
		log.Fatalf("Failed to create bucket: %v", err)
	}
	fmt.Printf("Bucket %v created.\n", BucketName)
}

func main() {
	ctx := context.Background()

	// Creates a client.
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", Credentials)
	client, err := speech.NewClient(ctx)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Sets the name of the audio file to transcribe.
	filename := "gangan_part.flac"
	if checkBucket() == storage.ErrBucketNotExist {
		createBucket()
	}
	//upload(filename)
	t := time.Now()
	objname := fmt.Sprintf("%d%02d%02d-%02d%02d%02d-", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()) + filepath.Base(filename)
	bukOpt(BucketName, objname, "upload", filename)
	defer bukOpt(BucketName, objname, "delete", "")

	req := &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:              speechpb.RecognitionConfig_FLAC,
			SampleRateHertz:       16000,
			LanguageCode:          "ja-JP",
			EnableWordTimeOffsets: true,
			//EnableSpeakerDiarization:   true,
			//EnableAutomaticPunctuation: true,
			//DiarizationSpeakerCount:    2,
			//SpeechContexts: []*speechpb.SpeechContext{&speechpb.SpeechContext{
			//	Phrases: []string{"松岡", "禎丞", "小林", "裕介"},
			//	Boost:   10,
			//}},
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Uri{Uri: "gs://" + BucketName + "/" + objname},
		},
	}
	op, err := client.LongRunningRecognize(ctx, req)
	if err != nil {
		log.Fatalf("failed to start longrun: %v", err)
	}
	fmt.Println(time.Since(t))
	fmt.Println("Waiting...")
	resp, err := op.Wait(ctx)
	if err != nil {
		log.Fatalf("failed to recognize: %v", err)
	}
	fmt.Println(time.Since(t))

	// Prints the results.
	fmt.Println("========================")

	for _, result := range resp.Results {
		//fmt.Println(result.String())
		speaker := make([]string, 3)
		for _, alt := range result.Alternatives {
			//fmt.Printf("\"%v\" (confidence=%3f)\n", alt.Transcript, alt.Confidence)
			words := alt.GetWords()
			for _, word := range words {
				wordsplit := strings.Split(word.GetWord(), "|")
				fmt.Printf("%v: \"%v\", start:%v, end:%v\n", word.GetSpeakerTag(), wordsplit[0], word.GetStartTime(), word.GetEndTime())
				speaker[word.GetSpeakerTag()] += wordsplit[0]
			}
			for _, v := range speaker {
				fmt.Println("===============================")
				fmt.Println(v)
			}

		}
	}

	fmt.Println(time.Since(t))
}

func bukwrite(client *storage.Client, bucket, object string, local string) error {
	ctx := context.Background()
	// [START upload_file]
	f, err := os.Open(local)
	if err != nil {
		return err
	}
	defer f.Close()

	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	// [END upload_file]
	return nil
}

func bukdelete(client *storage.Client, bucket, object string) error {
	ctx := context.Background()
	// [START delete_file]
	o := client.Bucket(bucket).Object(object)
	if err := o.Delete(ctx); err != nil {
		return err
	}
	// [END delete_file]
	return nil
}
