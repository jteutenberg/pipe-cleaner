package sequencing

import (
	"errors"
	"github.com/jteutenberg/pipe-cleaner/pipeline"
	"io"
	"os"
)

type FastA struct {
	name     string
	sequence []byte
}

func (f *FastA) GetName() string {
	return f.name
}
func (f *FastA) GetContents() []byte {
	return f.sequence
}

type fastAReader struct {
	out   chan Sequence
	input <-chan string
	numRoutines int
}

type fastAWriter struct {
	outputName string
	input      <-chan Sequence
}

func NewFastAReader(n int) *fastAReader {
	return &fastAReader{out: make(chan Sequence, n+1), numRoutines:n}
}

func (r *fastAReader) GetOutput() <-chan Sequence {
	return r.out
}

func (r *fastAReader) GetNumRoutines() int {
	return r.numRoutines
}

//Attach expects a StringComponent
func (r *fastAReader) Attach(p pipeline.PipelineComponent) error {
	if producer, ok := p.(pipeline.StringComponent); ok {
		r.input = producer.GetOutput()
	} else {
		return errors.New("Fasta reader was not attached to a StringComponent")
	}
	return nil
}

func (r *fastAReader) Run(complete chan<- bool) {
	header := true //next line is a name
	var name string
	for line := range r.input {
		if header {
			if len(line) == 0 || line[0] != '>' {
				//look for a header in a later line
				continue
			}
			name = line[1:]
		} else {
			fa := FastA{name: name, sequence: []byte(line)}
			r.out <- &fa
		}
		header = !header
	}
	complete <- true
}

func (r *fastAReader) Close() {
	close(r.out)
}

func NewFastAWriter(outputName string) *fastAWriter {
	return &fastAWriter{outputName: outputName}
}

func (r *fastAWriter) Attach(p pipeline.PipelineComponent) error {
	if producer, ok := p.(SequenceComponent); ok {
		r.input = producer.GetOutput()
	} else {
		return errors.New("Fasta writer was not attached to a SequenceComponent")
	}
	return nil
}

func (r *fastAWriter) Run(complete chan<- bool) {
	var out io.Writer
	if len(r.outputName) == 0 {
		out = os.Stdout
	} else {
		if f, err := os.Create(r.outputName); err == nil {
			out = f
			defer f.Close()
		}
	}
	for s := range r.input {
		io.WriteString(out, ">")
		io.WriteString(out, s.GetName())
		io.WriteString(out, "\n")
		out.Write(s.GetContents())
		io.WriteString(out, "\n")
	}
	complete <- true
}
func (r *fastAWriter) GetNumRoutines() int {
	return 1
}
func (r *fastAWriter) Close() {}
