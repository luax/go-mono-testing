package print

import (
	"log"
	"os"
	"net/http"
	"net/http/httputil"
)

func PrintEnvVariables() {
	log.Println(">>> Environment variables <<<")
	for _, e := range os.Environ() {
		log.Println(e)
	}
	log.Println(">>> end <<<")
}

func PrintRequest(req *http.Request) {
	// Save a copy of this request for debugging.
	log.Println(">>> Request <<<")
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(requestDump))
	log.Println(">>> end <<<")
}
