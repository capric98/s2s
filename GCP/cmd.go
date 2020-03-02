package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/capric98/s2s"
)

var (
	credentials = flag.String("cred", "credentials.json", "Google Application credentials file")
	bucketName  = flag.String("bucket", "speechtotextbtest", "GCP bucket name")
	projectID   = flag.String("projID", "", "Your GCP project ID")
	gcpLan      = flag.String("gLCode", "ja-JP", "GCP speech to text language code")
	// speakerNum  = flag.Int("num", 0, "The number of speakers. Set this to 0 to disable SpeakerDiarization.")
	// ydLan       = flag.String("ydCode", "ja", "Youdao API language code")
	// target      = flag.String("target", "zh-CHS", "Target subtitle language code")
)

func main() {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", *credentials)
	flag.Parse()
	resArgs := flag.Args()
	if len(resArgs) == 0 {
		log.Panicln("Please input a media file!")
	}
	if len(resArgs) > 1 {
		log.Println("Got more than one input files, emit the latters:", resArgs[1:])
		time.Sleep(2 * time.Second)
	}

	file := resArgs[0]
	if info, err := os.Stat(file); os.IsNotExist(err) {
		info.Name()
		log.Panicln(file, "does not exists!")
	}
	task := time.Now().Format("20060102-150405-") + filepath.Base(file) + ".flac"

	if err := s2s.EncodeToFLAC(file, task); err != nil {
		_ = os.Remove(task)
		log.Panicln(err)
	}
	defer os.Remove(task)

	buk, err := s2s.OpenBuk(*bucketName, *projectID)
	if err != nil {
		log.Panicln("Open bucket:", err)
	}
	defer buk.Close()

	fileReader, err := os.Open(task)
	if err != nil {
		os.Remove(task)
		log.Panicln("Open encoded file:", err)
	}
	defer fileReader.Close()

	defer func() { _ = buk.Delete(task) }()
	if err := buk.Upload(task, fileReader); err != nil {
		log.Panicln("Upload bucket:", err)
	}

	r, _ := s2s.Recognize("gs://"+*bucketName+"/"+task, *gcpLan, 0)
	log.Println(s2s.OutputSubtitle(r, "test.ass", true, false))
}
