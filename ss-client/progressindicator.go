package main

import (
	"fmt"
	"io"
	"strconv"
)

func copyWithProgress(filename string, size uint64, src io.Reader, dst io.Writer) error {
	written := uint64(0)
	buf := make([]byte, 65536) // There is nothing wrong with using big buffers.

	eof := false
	for !eof {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += uint64(nw)
			}
			if ew != nil {
				return ew
			}
			if nr != nw {
				return io.ErrShortWrite
			}
		}
		if er != nil {
			if er == io.EOF {
				eof = true
			} else {
				return er
			}
		}

		fmt.Printf("\rReceiving %s (%s of %s, %v%%)...",
			filename, humanReadableSize(written), humanReadableSize(size),
			int(float64(written)/float64(size)*100))
	}

	// This whitespace should override indicator left on line.
	fmt.Printf("\rReceived %s						\n", filename)

	return nil
}

func humanReadableSize(bytes uint64) string {
	suffix := " B"
	val := float64(bytes)
	if val > 1024 {
		val /= 1024
		suffix = " KB"
	}
	if val > 1024 {
		val /= 1024
		suffix = " MB"
	}
	if val > 1024 {
		val /= 1024
		suffix = " GB"
	}

	return strconv.FormatFloat(val, 'f', 2, 64) + suffix
}
