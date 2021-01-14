package main

import (
	"flag"
	"github.com/jteutenberg/pipe-cleaner/pipeline"
	"github.com/jteutenberg/pipe-cleaner/sequencing"
	"github.com/jteutenberg/pipe-cleaner/sequencing/kmer"
	"github.com/jteutenberg/pipe-cleaner/sequencing/rle"
	"github.com/jteutenberg/pipe-cleaner/util"
)

func main() {
	var inputFile = flag.String("i", "", "input fasta file name")
	var outputFile = flag.String("o", "", "output fasta file name")
	var k = flag.Int("k", 5, "kmer size")
	var h = flag.Bool("h", false, "homopolymer collapse")
	flag.Parse()

	threads := 4

	p := pipeline.NewPipeline()
	p.Append(util.NewLineReader(*inputFile), 1)
	p.Append(sequencing.NewFastAReader(1), 1)
	if *h {
		p.Append(rle.NewRunLengthEncoder(threads), threads)
	}
	p.Append(kmer.NewKmerComponent(*k, threads), threads)
	p.Append(kmer.NewKmerCounter(*outputFile, *k, 1000000), 1) //bottlenecked at the map

	p.Run()
}
