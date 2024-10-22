package gtime

import (
	"log"
	"sync"
	"time"
)

type Sampler struct {
	label   string
	samples []time.Duration
	m       sync.Mutex

	done bool
}

func NewSampler(label string) *Sampler {
	s := &Sampler{
		label: label,
		done:  false,
	}
	go s.logger()
	return s
}

func (s *Sampler) logger() {
	for !s.done {
		s.m.Lock()
		if len(s.samples) == 0 {
			s.m.Unlock()
			time.Sleep(time.Second * 1)
			continue
		}
		avg := time.Duration(0)
		for _, t := range s.samples {
			avg += t
		}
		avg = avg / time.Duration(len(s.samples))
		s.samples = make([]time.Duration, 0)
		s.m.Unlock()

		log.Println(s.label, "avg:", avg)
		time.Sleep(time.Second * 1)
	}
}

func (s *Sampler) Done() {
	s.done = true
}

func (s *Sampler) Sample(t time.Duration) {
	s.m.Lock()
	s.samples = append(s.samples, t)
	s.m.Unlock()
}
