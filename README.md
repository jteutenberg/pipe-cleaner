# pipe-cleaner
Golang framework for programming linear pipelines, with simple examples for DNA sequences

The purpose of this framework is to hide away most of the boiler plate code for typical pipeline-style programs with variable multi-threading, so programs can be constructed as:

```
p := pipeline.NewPipeline()

p.Append(util.NewLineReader(*inputFile), 1)
p.Append(sequencing.NewFastAReader(1), 1)
p.Append(rle.NewRunLengthEncoder(threads), threads)
p.Append(rle.NewRLEToSequence(threads), threads)
p.Append(sequencing.NewFastAWriter(*outputFile), 1)

p.Run()
```
Which accepts sequences in the fasta format from either stdin or a file (using a single thread); performs a run length encoding using multiple threads; annotates the fasta names with run lengths using multiple threads; then writes the output fasta on a single thread.

The pipeline handles the creation of goroutines and channels, and does all the work that would usually require WaitGroups for ensuring they complete in order.
