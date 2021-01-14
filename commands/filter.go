package main

import (
	"flag"
	"github.com/jteutenberg/pipe-cleaner/pipeline"
	"github.com/jteutenberg/pipe-cleaner/sequencing"
	"github.com/jteutenberg/pipe-cleaner/util"
)

type lengthFilter struct {
	minLength int
	out       chan sequencing.Sequence
	input     <-chan sequencing.Sequence
}

func (r *lengthFilter) GetOutput() <-chan sequencing.Sequence {
	return r.out
}
func (r *lengthFilter) Attach(p pipeline.PipelineComponent) error {
	if producer, ok := p.(sequencing.SequenceComponent); ok {
		r.input = producer.GetOutput()
	}
	return nil
}
func (r *lengthFilter) Run(complete chan<- bool) {
	for seq := range r.input {
		if len(seq.GetContents()) >= r.minLength {
			r.out <- seq
		}
	}
	complete <- true
}

func (r *lengthFilter) GetNumRoutines() int {
	return 1
}

func (r *lengthFilter) Close() {
	close(r.out)
}

func main() {
	var inputFile = flag.String("i", "", "input fasta file name")
	var outputFile = flag.String("o", "", "output fasta file name")
	var length = flag.Int("l", 1, "minimum sequence length")
	flag.Parse()

	filter := &lengthFilter{minLength: *length, out: make(chan sequencing.Sequence, 1)}

	p := pipeline.NewPipeline()
	p.Append(util.NewLineReader(*inputFile))
	p.Append(sequencing.NewFastAReader(1))
	p.Append(filter)
	p.Append(sequencing.NewFastAWriter(*outputFile))

	p.Run()
}
