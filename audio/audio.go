package audio

import (
	"audio-player/gtime"
	"bytes"
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

	dur    float32
	durMux sync.Mutex

	peak    float32
	peakSet bool
	peakMux sync.Mutex
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

func parseMaxVolume(s string) float32 {
	// string will look like:
	// [Parsed_volumedetect_0 @ 0x76acac004700] max_volume: -13.5 dB
	// we want to extract the -13.5
	mv := "max_volume: "
	sp := strings.Index(s, mv)
	if sp == -1 {
		log.Println("error parsing max volume:", s)
		return 0
	}
	sp += len(mv)

	dbStrIdx := strings.Index(s[sp:], " dB")
	if dbStrIdx == -1 {
		log.Println("error parsing max volume:", s)
		return 0
	}
	dbStrIdx += sp

	peak, err := strconv.ParseFloat(s[sp:dbStrIdx], 32)
	if err != nil {
		log.Println("error parsing max volume:", s, err)
		return 0
	}
	return float32(peak)
}

func (a *Audio) Peak() float32 {
	a.peakMux.Lock()
	defer a.peakMux.Unlock()

	if a.peakSet {
		return a.peak
	}
	// ffmpeg -i video.avi -af "volumedetect" -vn -sn -dn -f null /dev/null
	cmd := exec.Command("ffmpeg", "-i", a.path, "-hide_banner", "-loglevel", "info", "-af", "volumedetect", "-vn", "-sn", "-dn", "-f", "null", "/dev/null")
	buf := &bytes.Buffer{}
	cmd.Stderr = buf
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		log.Println("error getting peak:", err)
	}
	str := buf.String()
	for _, str := range strings.Split(str, "\n") {
		if strings.Contains(str, "max_volume") {
			a.peak = parseMaxVolume(str)
			a.peakSet = true
			return a.peak
		}
	}

	log.Println("error getting peak: no max_volume found in", str)

	return 0
}

func (a *Audio) Duration() float32 {
	a.durMux.Lock()
	defer a.durMux.Unlock()

	if a.dur != 0 {
		return a.dur
	}

	gtime.Start("ffprobe")
	defer gtime.End("ffprobe")

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
	peak := a.Peak()
	increase := float32(0)
	if peak < 0 {
		increase = -peak
	}

	cmd := exec.Command("ffplay", "-nodisp", "-autoexit", "-ss", strconv.FormatFloat(positionSeconds, 'f', -1, 64), "-af", "volume="+strconv.FormatFloat(float64(increase), 'f', -1, 32)+"dB", a.path)
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
