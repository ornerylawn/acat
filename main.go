// acat outputs audio files to stdout as interleaved stereo 32-bit
// float pcm. It is intended to be used with github.com/rynlbrwn/spkr.
//
//   $ acat foo.wav bar.wav | spkr
//
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"

	"unsafe"

	"github.com/rynlbrwn/acat/Godeps/_workspace/src/github.com/mkb218/gosndfile/sndfile"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s <wav_files...>\n", os.Args[0])
		os.Exit(1)
	}
	filenames := os.Args[1:]
	bw := bufio.NewWriterSize(os.Stdout, os.Getpagesize()*8)
	for _, filename := range filenames {
		if err := cat(filename, bw); err != nil {
			log.Fatal(err)
		}
	}
	if err := bw.Flush(); err != nil {
		log.Fatal(err)
	}
}

func cat(filename string, w io.Writer) error {
	var info sndfile.Info
	f, err := sndfile.Open(filename, sndfile.Read, &info)
	if err != nil {
		return err
	}
	defer f.Close()
	channels := int(info.Channels)
	nframes := os.Getpagesize() / (channels * 4)
	chunk := make([]float32, nframes*channels)
	bslice := reflect.SliceHeader{
		Data: reflect.ValueOf(chunk).Index(0).Addr().Pointer(),
		Len:  len(chunk) * 4,
		Cap:  cap(chunk) * 4,
	}
	bytes := *(*[]byte)(unsafe.Pointer(&bslice)) // :)
	for {
		rframes, err := f.ReadFrames(chunk)
		if err != nil {
			return err
		}
		for i := int(rframes) * channels; i < len(chunk); i++ {
			chunk[i] = 0.0
		}
		// TODO: convert to stereo if input is not stereo.
		if _, err := w.Write(bytes); err != nil {
			return err
		}
		if int(rframes) < nframes {
			return nil // no more data
		}
	}
}
