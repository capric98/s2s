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
	bucketName  = flag.String("bucket", "yas2ttool", "GCP bucket name")
	projectID   = flag.String("projID", "", "Your GCP project ID (only used to create bucket if not exist)")
	gcpLan      = flag.String("gLCode", "ja-JP", "GCP speech to text language code")
	speakerNum  = flag.Int("num", 0, "The number of speakers. Set this to 0 to disable SpeakerDiarization")
	ydLan       = flag.String("ydCode", "ja", "Youdao API language code")
	target      = flag.String("target", "zh-CHS", "Target subtitle language code")

	appKey  = flag.String("appKey", "", "Youdao API appKey")
	appPass = flag.String("appPass", "", "Youdao API appPass")

	vout = flag.String("vout", "", "Verbose output file. Set \"\" to disable.")
	pout = flag.String("pout", "", "Plain output file. Set \"\" to disable.")
	rout = flag.String("rawout", "", "Raw output file. Set \"\" to disable.")
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

	if *vout == "" && *pout == "" && *rout == "" {
		tmp := task + ".ass"
		vout = &tmp
		log.Println("No output was set. Default set vout to:", *vout)
	}

	log.Println("Transcoding...")
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

	log.Println("Uploading...")
	defer func() { _ = buk.Delete(task) }()
	if err := buk.Upload(task, fileReader); err != nil {
		log.Panicln("Upload bucket:", err)
	}

	r, e := s2s.Recognize("gs://"+*bucketName+"/"+task, *gcpLan, *speakerNum)
	if e != nil {
		log.Panicln("Speech to Text:", e)
	}

	if *vout != "" || *pout != "" {
		s2s.SetTrans(r, *ydLan, *target, *appKey, *appPass)
	}

	if *vout != "" {
		log.Println("Write verbose out to", *vout)
		e = s2s.OutputSubtitle(r, *vout, true, true)
		if e != nil {
			log.Println("Verbose out:", e)
		}
	}
	if *pout != "" {
		log.Println("Write plain out to", *pout)
		e = s2s.OutputSubtitle(r, *pout, false, true)
		if e != nil {
			log.Println("Plain out:", e)
		}
	}
	if *rout != "" {
		log.Println("Write raw out to", *rout)
		e = s2s.OutputSubtitle(r, *rout, false, false)
		if e != nil {
			log.Println("Raw out:", e)
		}
	}
}
