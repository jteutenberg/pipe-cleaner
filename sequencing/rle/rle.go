package rle

import (
	"errors"
	"github.com/jteutenberg/pipe-cleaner/pipeline"
	"github.com/jteutenberg/pipe-cleaner/sequencing"
	"strings"
)

type RLESequence interface {
	sequencing.Sequence
	GetCounts() []byte
}

type rleSequence struct {
	name     string
	sequence []byte
	counts   []byte
}

func (f *rleSequence) GetName() string {
	return f.name
}
func (f *rleSequence) GetContents() []byte {
	return f.sequence
}
func (f *rleSequence) GetCounts() []byte {
	return f.counts
}

type runLengthEncoder struct {
	out   chan sequencing.Sequence
	input <-chan sequencing.Sequence
	numRoutines int
}

type rleToSequence struct {
	out   chan sequencing.Sequence
	input <-chan sequencing.Sequence
	numRoutines int
}

func RunLengthToASCII(counts []byte) string {
	cs := make([]byte, len(counts))
	for i, c := range counts {
		if c >= 93 {
			cs[i] = 126 //93+33
		} else {
			cs[i] = c + 33
		}
	}
	return string(cs)
}

func ASCIIToRunLength(countString string) []byte {
	cs := make([]byte, len(countString))
	for i, c := range countString {
		cs[i] = byte(c) - 33
	}
	return cs
}

func NewRunLengthEncoder(n int) *runLengthEncoder {
	return &runLengthEncoder{out: make(chan sequencing.Sequence, n+1), numRoutines:n}
}

func (r *runLengthEncoder) GetOutput() <-chan sequencing.Sequence {
	return r.out
}

func (r *runLengthEncoder) Attach(p pipeline.PipelineComponent) error {
	if producer, ok := p.(sequencing.SequenceComponent); ok {
		r.input = producer.GetOutput()
	} else {
		return errors.New("Run length encoder was not attached to a SequenceComponent")
	}
	return nil
}

func (r *runLengthEncoder) Run(complete chan<- bool) {
	tempOut := make([]byte, 1024)
	tempCount := make([]byte, 1024)
	for seq := range r.input {
		content := seq.GetContents()
		if len(tempOut) < len(content) {
			tempOut = make([]byte, len(content)+5)
			tempCount = make([]byte, len(content)+5)
		}
		length := RLE(content, tempOut, tempCount)
		//copy output across to fresh memory, then move on to the next sequence
		out := make([]byte, length)
		count := make([]byte, length)
		copy(out, tempOut)
		copy(count, tempCount)
		r.out <- &rleSequence{name: seq.GetName(), sequence: out, counts: count}
	}
	complete <- true
}

func (r *runLengthEncoder) GetNumRoutines() int {
	return r.numRoutines
}

func (r *runLengthEncoder) Close() {
	close(r.out)
}

//RLE does a brute comparison as it scans the input sequence
//It fills the two output slices with the RLE sequence and returns their length
func RLE(seq []byte, outputSeq, outputCounts []byte) int {
	prev := seq[0]
	index := 0
	count := byte(1)
	for _, b := range seq[1:] {
		if b != prev || count >= 127 {
			outputSeq[index] = prev
			outputCounts[index] = count
			prev = b
			count = 0
			index++
		}
		count++
	}
	outputSeq[index] = prev
	outputCounts[index] = count
	return index + 1
}

//RevertRLE is a slow inverse of RLE for test purposes
func RevertRLE(seq, counts []byte) []byte {
	output := make([]byte, 0, len(seq)*2)
	for i, b := range seq {
		for c := counts[i]; c > 0; c-- {
			output = append(output, b)
		}
	}
	return output
}

func NewRLEToSequence(n int) *rleToSequence {
	return &rleToSequence{out: make(chan sequencing.Sequence, n+1),numRoutines:n}
}

func (r *rleToSequence) GetOutput() <-chan sequencing.Sequence {
	return r.out
}

func (r *rleToSequence) Attach(p pipeline.PipelineComponent) error {
	if producer, ok := p.(sequencing.SequenceComponent); ok {
		r.input = producer.GetOutput()
	} else {
		return errors.New("RLE-to-sequence was not attached to a SequenceComponent")
	}
	return nil
}

//Run appends the run lengths to the end of sequence names
func (r *rleToSequence) Run(complete chan<- bool) {
	for baseSeq := range r.input {
		if s, ok := baseSeq.(RLESequence); ok {
			var sb strings.Builder
			sb.WriteString(s.GetName())
			sb.WriteString(" ")
			sb.WriteString(RunLengthToASCII(s.GetCounts()))
			newSeq := &rleSequence{name: sb.String(), sequence: s.GetContents(), counts: s.GetCounts()}
			r.out <- newSeq
		}
	}
	complete <- true
}

func (r *rleToSequence) GetNumRoutines() int {
	return r.numRoutines
}

func (r *rleToSequence) Close() {
	close(r.out)
}
