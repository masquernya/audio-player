package gtime

import (
	"log"
	"time"
)

type GTime struct {
	labels map[string]time.Time
}

func New() *GTime {
	g := &GTime{
		labels: make(map[string]time.Time),
	}
	return g
}

func (g *GTime) Start(label string) {
	g.labels[label] = time.Now()
}

func (g *GTime) End(label string) {
	start, exists := g.labels[label]
	if !exists {
		log.Println("[warning] label", label, "does not exist")
		return
	}
	delete(g.labels, label)
	s := time.Since(start)
	log.Println(label, "took:", s)
}
