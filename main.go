package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
)

var (
	flagInFolder  = flag.String("in", "./", "input folder")
	flagOutFolder = flag.String("out", "", "output folder")
	flagSize      = flag.String("size", "500x500", "Maximal image size")
)

type errorList struct {
	errs []error
}

func (e *errorList) add(err error) {
	if err != nil {
		e.errs = append(e.errs, err)
	}
}
func (e *errorList) hasError() bool {
	return len(e.errs) > 0
}
func (e *errorList) Error() string {
	if !e.hasError() {
		return ""
	}
	out := fmt.Sprintf("Number of errors: %d:", len(e.errs))
	for i, err := range e.errs {
		out = fmt.Sprintf("%s\n%d: %w", out, i, err.Error())
	}
	return out
}

type picSize struct {
	width  int
	height int
}

type resizeArgs struct {
	inPath  string
	outPath string
	size    picSize
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
	err = resizeFolderImages(*flagInFolder, outFolder, size)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func resizeFolderImages(inFolder, outFolder string, size picSize) error {
	err := os.MkdirAll(outFolder, 0755)
	if err != nil {
		return fmt.Errorf("Could not create target directory: %w", err)
	}
	dir, err := ioutil.ReadDir(*flagInFolder)
	if err != nil {
		return fmt.Errorf("Error trying to read directory: %w\n", err)
	}

	wg := &sync.WaitGroup{}
	errList := &errorList{}
	errChan := make(chan error)
	resizeChan := make(chan resizeArgs)
	wg.Add(3)
	go resizer(wg, resizeChan, errChan)
	go resizer(wg, resizeChan, errChan)
	go resizer(wg, resizeChan, errChan)
	go func(errList *errorList, errChan chan error) {
		for err := range errChan {
			errList.add(err)
		}
	}(errList, errChan)

	for _, fi := range dir {
		if fi.IsDir() || !useFile(fi.Name()) {
			continue
		}
		inPath := filepath.Join(inFolder, fi.Name())
		outPath := filepath.Join(outFolder, fi.Name())
		resizeChan <- resizeArgs{inPath, outPath, size}
	}
	close(resizeChan)
	close(errChan)
	wg.Wait()
	if errList.hasError() {
		return errList
	}
	return nil
}

func resizeClose(ps picSize, r io.ReadCloser, w io.WriteCloser) error {
	defer r.Close()
	defer w.Close()
	return resize(ps, r, w)
}

func resizer(wg *sync.WaitGroup, c chan resizeArgs, errChan chan error) {
	for a := range c {
		log.Println("Resize:", a.inPath)
		inFile, err := os.Open(a.inPath)
		if err != nil {
			errChan <- fmt.Errorf("Error opening file from %s: %s\n", a.inPath, err)
			continue
		}
		outFile, err := os.OpenFile(a.outPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			errChan <- fmt.Errorf("Error creating file for %s: %s\n", a.outPath, err)
		}
		err = resizeClose(a.size, inFile, outFile)
		if err != nil {
			errChan <- fmt.Errorf("Error resizing image from %s: %s\n", a.inPath, err)
		}
	}
	wg.Done()
}
