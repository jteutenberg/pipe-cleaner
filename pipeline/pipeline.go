package pipeline

import (
	"fmt"
	"io"
	"os"
)

type PipelineComponent interface {
	//Attach should pass a producer/consumer channel across between the components
	Attach(PipelineComponent) error
	//Run starts the pipeline component which must return a bool on the channel when complete
	Run(chan<- bool)
	//GetNumRoutines retrieves the number of threads this component expects to be run across. Approximate values won't break things.
	GetNumRoutines() int
	//Close will be called on this component after all of its Run calls have completed. It can then close any output channels.
	Close()
}

type StringComponent interface {
	GetOutput() <-chan string
}

type Pipeline struct {
	components []PipelineComponent
	nRoutines  []int
	complete   []chan bool
}

func NewPipeline() *Pipeline {
	return &Pipeline{components: make([]PipelineComponent, 0, 10), nRoutines: make([]int, 0, 10)}
}

func (p *Pipeline) Append(c PipelineComponent) {
	numRoutines := c.GetNumRoutines()
	p.components = append(p.components, c)
	p.nRoutines = append(p.nRoutines, numRoutines)
	if len(p.components) > 1 {
		if err := c.Attach(p.components[len(p.components)-2]); err != nil {
			io.WriteString(os.Stderr, fmt.Sprintln(err))
		}
	}
}

func (p *Pipeline) Run() {
	p.complete = make([]chan bool, len(p.components))
	//run in reverse order
	for i := len(p.nRoutines) - 1; i >= 0; i-- {
		complete := make(chan bool, p.nRoutines[i])
		p.complete[i] = complete
		for j := 0; j < p.nRoutines[i]; j++ {
			go p.components[i].Run(complete)
		}
	}
	//wait for completion in-order
	for i := 0; i < len(p.nRoutines); i++ {
		for j := 0; j < p.nRoutines[i]; j++ {
			<-p.complete[i]
		}
		//processing of all routines complete, close its output channel
		p.components[i].Close()
	}
}
