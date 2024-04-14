package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
)

const (
	rootDirDefault       = "/"
	sizeThresholdDefault = "100MB"
)

func main() {
	rootDir := flag.String("d", rootDirDefault, "directory to search")
	sizeThreshold := flag.String("s", sizeThresholdDefault, "print directories and files exceeding this threshold (example: 100MB)")
	flag.Parse()

	sizeThresholdParsed, err := humanize.ParseBigBytes(*sizeThreshold)
	if err != nil {
		log.Fatalf("invalid size threshold '%v': %v", *sizeThreshold, err)
	}

	visualiser := newVisualiser(sizeThresholdParsed.Int64())
	visualiser.visualise(*rootDir)

	return
}

type visualiser struct {
	sizeThreshold int64
}

func newVisualiser(sizeThreshold int64) *visualiser {
	return &visualiser{
		sizeThreshold: sizeThreshold,
	}
}

func (v *visualiser) visualise(dir string) {
	dirSize, _, err := getDirSize(dir, v.sizeThreshold)
	if err != nil {
		log.Printf("error: could not visualise directory %v: %v", dir, err)
		return
	}

	if dirSize > v.sizeThreshold {
		fmt.Printf("%v: %v\n", dir, humanize.BigBytes(big.NewInt(dirSize)))
		fmt.Println()
	}
}

// getDirSize calculates size for the given directory recursively. It prints size of entries
// exceeding the sizeThreshold.
func getDirSize(dir string, sizeThreshold int64) (int64, int, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("error: could not read contents of directory %v: %v", dir, err)
		log.Printf("warning: will skip directory %v in calculations", dir)

		return 0, 0, nil
	}

	dirSize := int64(0)
	filesPrintedInThisDir := 0
	shouldPrintAClosingNewLine := false

	for _, entry := range dirEntries {
		fullPath := filepath.Join(dir, entry.Name())
		entrySize := int64(0)

		switch {
		case entry.Type().IsRegular():
			info, err := entry.Info()
			entrySize = info.Size()

			if err != nil {
				log.Printf("error: could not get info for file %v: %v", fullPath, err)
				log.Printf("warning: file %v will not be included in calculations", fullPath)

				continue
			}

		case entry.Type().IsDir():
			var filesPrintedInThisEntry int

			entrySize, filesPrintedInThisEntry, err = getDirSize(fullPath, sizeThreshold)
			if err != nil {
				log.Printf("error: could not read contents of directory %v: %v", dir, err)
				log.Printf("warning: will skip directory %v in calculations", dir)

				continue
			}

			if filesPrintedInThisEntry > 0 {
				shouldPrintAClosingNewLine = true
			}
		}

		if entrySize > sizeThreshold {
			if entry.Type().IsRegular() && filesPrintedInThisDir == 0 {
				// create an empty line before a group of files in one directory
				fmt.Println()
			}

			fmt.Printf("%v: %v\n", fullPath, humanize.BigBytes(big.NewInt(entrySize)))

			if shouldPrintAClosingNewLine {
				// create an empty line after a group of files in one directory
				fmt.Println()
			}

			if entry.Type().IsRegular() {
				filesPrintedInThisDir++
			}
		}

		dirSize += entrySize
	}

	return dirSize, filesPrintedInThisDir, nil
}
