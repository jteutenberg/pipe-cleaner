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
	p.Append(util.NewLineReader(*inputFile), 1)
	p.Append(sequencing.NewFastAReader(1), 1)
	p.Append(rle.NewRunLengthEncoder(threads), threads)
	p.Append(rle.NewRLEToSequence(threads), threads)
	p.Append(sequencing.NewFastAWriter(*outputFile), 1)

	p.Run()
}
