package kmer

import (
	"errors"
	"github.com/jteutenberg/pipe-cleaner/pipeline"
	"github.com/jteutenberg/pipe-cleaner/sequencing"
	"io"
	"os"
	"sort"
	"strconv"
	"sync"
)

//KmerSequence holds an integer representation of substrings of A,C,G and T.
//Other alphabets will go through but with ambiguous values, essentially cause hash collisions.
type KmerSequence struct {
	kmers []uint64
	k     int
}
type KmerSequenceComponent interface {
	GetOutput() <-chan *KmerSequence
}

type kmerComponent struct {
	output chan *KmerSequence
	input  <-chan sequencing.Sequence
	k      int
	numRoutines int
}

type kmerCounter struct {
	outputName string
	input      <-chan *KmerSequence
	k          int
	lock       sync.Mutex
	counts     map[uint64]int
}

func NewKmerComponent(k int, n int) *kmerComponent {
	return &kmerComponent{output: make(chan *KmerSequence, n+1), k: k, numRoutines: n}
}

func (r *kmerComponent) GetOutput() <-chan *KmerSequence {
	return r.output
}

func (r *kmerComponent) Attach(p pipeline.PipelineComponent) error {
	if producer, ok := p.(sequencing.SequenceComponent); ok {
		r.input = producer.GetOutput()
	} else {
		return errors.New("K-mer component was not attached to a SequenceComponent")
	}
	return nil
}

func (r *kmerComponent) Run(complete chan<- bool) {
	toShift := uint((r.k-1)*2) //how far to shift a base to get to the reverse position
	mask := (^uint64(0)) >> uint(64-r.k*2) //1s for all bits that will be used in a k-mer
	n := 0
	for seq := range r.input {
		content := seq.GetContents()
		if len(content) < r.k {
			continue //too short, no k-mers
		}
		//NOTE: fresh memory for each sequence. If efficiency matters, here's a good place to start.
		kmers := make([]uint64, len(content)-r.k+1)
		var next uint64 //slide across bases as a k-mer
		var nextRC uint64 //the reverse-complement version
		//NOTE: we ignore the reverse-complement here. Not suitable for many applications
		for i, b := range []byte(content) {
			//convert the character into a 0-3 value
			b = ((b >> 1) ^ ((b & 4) >> 2)) & 3
			bb := uint64(b)
			//shuffle it in to make the next k-mer
			next = ((next << 2) | bb) & mask
			nextRC = ((nextRC >> 2) | ((^bb) << toShift) & mask) //XOR to complement
			if i < r.k-1 {
				//haven't seen enough bases to make the first k-mer yet
				continue
			}
			//select the minimum of the two k-mers. Note this can be made branchless pretty easily.
			minKmer := next
			if minKmer > nextRC {
				minKmer = nextRC
			}
			kmers[n] = minKmer
			n++
		}
		r.output <- &KmerSequence{kmers: kmers, k: r.k}
		n = 0
	}
	complete <- true
}

func (r *kmerComponent) GetNumRoutines() int {
	return r.numRoutines
}

func (r *kmerComponent) Close() {
	close(r.output)
}

func NewKmerCounter(outputName string, k int, capacity int) *kmerCounter {
	counts := make(map[uint64]int, capacity)
	return &kmerCounter{outputName: outputName, k: k, counts: counts}
}

func (r *kmerCounter) Attach(p pipeline.PipelineComponent) error {
	if producer, ok := p.(KmerSequenceComponent); ok {
		r.input = producer.GetOutput()
	} else {
		return errors.New("K-mer counter was not attached to a KmerSequenceComponent")
	}
	return nil
}

func (r *kmerCounter) Run(complete chan<- bool) {
	for ks := range r.input {
		//lock the count and enter all k-mers
		r.lock.Lock() //NOTE: not actually required since this component is fixed to one thread
		for _, kmer := range ks.kmers {
			if count, exists := r.counts[kmer]; exists {
				r.counts[kmer] = count + 1
			} else {
				r.counts[kmer] = 1
			}
		}
		r.lock.Unlock()
	}
	complete <- true
}

func (r *kmerCounter) GetNumRoutines() int {
	return 1
}

type outputRow struct {
	kmer  uint64
	count int
}

func KmerToString(value uint64, k int) string {
	bs := make([]byte, k)
	for i := k - 1; i >= 0; i-- {
		base := value & 3
		switch base {
		case 0:
			bs[i] = 'A'
		case 1:
			bs[i] = 'C'
		case 2:
			bs[i] = 'G'
		case 3:
			bs[i] = 'T'
		}
		value = value >> 2
	}
	return string(bs)
}

func (r *kmerCounter) Close() {
	//ready to write all output
	var out io.Writer
	var outFile *os.File
	if len(r.outputName) == 0 {
		out = os.Stdout
	} else {
		var err error
		if outFile, err = os.Create(r.outputName); err == nil {
			defer outFile.Close()
			out = outFile
		} else {
			return
		}
	}
	values := make([]outputRow, 0, len(r.counts))
	for k, v := range r.counts {
		values = append(values, outputRow{kmer: k, count: v})
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].count > values[j].count
	})
	for i, row := range values {
		if i > 100 {
			break
		}
		io.WriteString(out, KmerToString(row.kmer, r.k))
		io.WriteString(out, " ")
		io.WriteString(out, strconv.Itoa(row.count))
		io.WriteString(out, "\n")
	}
}
