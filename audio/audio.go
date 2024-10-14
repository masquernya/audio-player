package audio

import (
	"github.com/google/uuid"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type Audio struct {
	path    string
	proc    *os.Process
	procId  string
	procMux sync.Mutex

	dur float32
}

func New(path string) *Audio {
	return &Audio{path: path}
}

func (a *Audio) Stop() {
	a.procMux.Lock()
	if a.proc != nil {
		a.proc.Signal(os.Kill)
		a.proc = nil
	}
	a.procMux.Unlock()
}

func (a *Audio) Duration() float32 {
	if a.dur != 0 {
		return a.dur
	}

	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", a.path)
	out, err := cmd.Output()
	if err != nil {
		log.Fatal("error getting duration:", err)
	}

	str := strings.TrimSpace(strings.ReplaceAll(string(out), "\n", ""))

	dur, err := strconv.ParseFloat(str, 32)
	if err != nil {
		log.Fatal("error parsing duration:", str, err)
	}

	a.dur = float32(dur)
	return a.dur
}

func (a *Audio) Start(positionSeconds float64) error {
	a.Stop()

	cmd := exec.Command("ffplay", "-nodisp", "-autoexit", "-ss", strconv.FormatFloat(positionSeconds, 'f', -1, 64), a.path)
	err := cmd.Start()
	if err != nil {
		return err
	}
	id := uuid.NewString()

	go (func() {
		cmd.Wait()

		a.procMux.Lock()
		if a.procId == id {
			a.proc = nil
		}
		a.procMux.Unlock()
	})()

	a.procMux.Lock()
	a.proc = cmd.Process
	a.procId = id
	a.procMux.Unlock()

	return nil
}
