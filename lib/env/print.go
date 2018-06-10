package env

import (
	"log"
	"os"
)

func PrintEnvVariables() {
	log.Println(">>> Environment variables <<<")
	for _, e := range os.Environ() {
		log.Println(e)
	}
	log.Println(">>> end <<<")
}
