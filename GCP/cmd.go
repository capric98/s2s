package main

import (
	"flag"
	"os"
)

var (
	credentials = flag.String("cred", "credentials.json", "Google Application credentials file")
	projectID   = flag.String("projID", "", "Your GCP project ID.")
)

func main() {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", *credentials)
}
