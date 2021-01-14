package util

import (
	"bufio"
	"github.com/jteutenberg/pipe-cleaner/pipeline"
	"os"
)

type lineReader struct {
	inputName string
	out       chan string
}

func NewLineReader(inputName string) *lineReader {
	return &lineReader{inputName: inputName, out: make(chan string, 2)}
}

func (r *lineReader) GetOutput() <-chan string {
	return r.out
}

func (r *lineReader) Attach(p pipeline.PipelineComponent) error {
	return nil
}

//Run will either read lines from a file with the given
//name, or if this is the empty string it will read from stdin
func (r *lineReader) Run(complete chan<- bool) {
	var scanner *bufio.Scanner
	var infile *os.File
	if len(r.inputName) == 0 {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		var err error
		if infile, err = os.Open(r.inputName); err != nil {
			complete <- true
			return
		}
		scanner = bufio.NewScanner(infile)
	}
	// send output line by line as it is read
	go func() {
		for scanner.Scan() {
			r.out <- scanner.Text()
		}
		if infile != nil {
			infile.Close()
		}
		complete <- true
	}()
}

func (r *lineReader) Close() {
	close(r.out)
}
