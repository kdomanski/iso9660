package main

import (
	"log"
	"os"

	"github.com/kdomanski/iso9660/util"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s ISOFILE TARGET_DIR", os.Args[0])
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("failed to open %s: %s", os.Args[1], err)
	}
	defer f.Close()

	if err = util.ExtractImageToDirectory(f, os.Args[2]); err != nil {
		log.Fatalf("failed to extract image: %s", err)
	}
}
