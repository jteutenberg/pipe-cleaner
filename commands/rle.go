package main

import (
	"flag"
	"github.com/jteutenberg/pipe-cleaner/pipeline"
	"github.com/jteutenberg/pipe-cleaner/sequencing"
	"github.com/jteutenberg/pipe-cleaner/sequencing/rle"
	"github.com/jteutenberg/pipe-cleaner/util"
)

func main() {
	var inputFile = flag.String("i", "", "input fasta file name")
	var outputFile = flag.String("o", "", "output fasta file name")
	flag.Parse()

	threads := 4
	p := pipeline.NewPipeline()
	p.Append(util.NewLineReader(*inputFile))
	p.Append(sequencing.NewFastAReader(1))
	p.Append(rle.NewRunLengthEncoder(threads))
	p.Append(rle.NewRLEToSequence(threads))
	p.Append(sequencing.NewFastAWriter(*outputFile))

	p.Run()
}
