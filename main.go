package main

import "flag"

var (
	flagInFolder  = flag.String("in", "./", "input folder")
	flagOutFolder = flag.String("out", "", "output folder")
	flagSize      = flag.String("size", "500x500", "Maximal image size")
)

func main() {
	flag.Parse()
}
