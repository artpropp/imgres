package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

var (
	flagInFolder  = flag.String("in", "./", "input folder")
	flagOutFolder = flag.String("out", "", "output folder")
	flagSize      = flag.String("size", "500x500", "Maximal image size")
)

type picSize struct {
	width  int
	height int
}

func parseSize(s string) (picSize, error) {
	var ps picSize

	parts := strings.Split(s, "x")
	if len(parts) != 2 {
		return ps, fmt.Errorf("%s is not in the expected form", s)
	}
	var err error
	ps.width, err = strconv.Atoi(parts[0])
	if err != nil {
		return ps, fmt.Errorf("parseSize: ps.x: %w", err)
	}
	ps.height, err = strconv.Atoi(parts[0])
	if err != nil {
		return ps, fmt.Errorf("parseSize: ps.y: %w", err)
	}

	return ps, nil
}

func resize(ps picSize, r io.Reader, w io.Writer) error {
	img, format, err := image.Decode(r)
	if err != nil {
		return fmt.Errorf("Error during decoding: %w", err)
	}
	if format != "jpeg" {
		return fmt.Errorf("Only jpeg is supported")
	}

	resized := imaging.Fit(img, ps.width, ps.height, imaging.Lanczos)
	return jpeg.Encode(w, resized, nil)
}

func useFile(fileName string) bool {
	allowed := []string{".jpg", ".jpeg"}
	ext := filepath.Ext(fileName)
	for _, e := range allowed {
		if strings.EqualFold(ext, e) {
			return true
		}
	}
	return false
}

func main() {
	flag.Parse()
	size, err := parseSize(*flagSize)
	if err != nil {
		fmt.Println("Could not create size:", err)
		os.Exit(1)
	}

	outFolder := *flagSize
	if *flagOutFolder != "" {
		outFolder = *flagOutFolder
	}
	err = os.MkdirAll(outFolder, 0777)
	if err != nil {
		fmt.Println("Could not create target directory:", err)
	}

	dir, err := ioutil.ReadDir(*flagInFolder)
	if err != nil {
		fmt.Println("Error trying to read directory:")
		fmt.Println(err)
		os.Exit(1)
	}

	for _, fi := range dir {
		if fi.IsDir() || !useFile(fi.Name()) {
			continue
		}
		inPath := filepath.Join(*flagInFolder, fi.Name())
		inFile, err := os.Open(inPath)
		if err != nil {
			fmt.Printf("Error opening file from %s: %s\n", inPath, err)
			continue
		}
		outPath := filepath.Join(outFolder, fi.Name())
		outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			fmt.Printf("Error creating file for %s: %s\n", outPath, err)
		}
		err = resize(size, inFile, outFile)
		if err != nil {
			fmt.Printf("Error resizing image from %s: %s\n", inPath, err)
		}
		inFile.Close()
		outFile.Close()
	}
}
