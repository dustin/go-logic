package main

import (
	"io"
	"log"
	"os"

	"github.com/dustin/logic"
)

func main() {
	r, err := logic.NewSerialCSVReader(os.Stdin)
	if err != nil {
		log.Fatalf("Error reading csv: %v", err)
	}
	n, err := io.Copy(os.Stdout, r)
	log.Printf("copied %v bytes before %v", n, err)
}
