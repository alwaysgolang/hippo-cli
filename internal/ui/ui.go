package ui

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

type Loader struct {
	s *spinner.Spinner
}

func New(message string) *Loader {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Start()

	return &Loader{s: s}
}

func (l *Loader) StopSuccess(message string) {
	l.s.Stop()
	fmt.Println("✔", message)
}

func (l *Loader) StopError(message string) {
	l.s.Stop()
	fmt.Println("✖", message)
}

func Banner() {
	fmt.Println(`
🦛 HIPPO CLI
──────────────────────────────
  Building strong backends.
──────────────────────────────
`)
}

type HippoSpinner struct {
	out     io.Writer
	stop    chan struct{}
	running bool
	wg      sync.WaitGroup
}

var frames = []string{
	"🦛 ᕕ(ᐛ)ᕗ",
	"🦛 ᕗ(ᐛ)ᕕ",
	"🦛 ᕕ(•‿•)ᕗ",
	"🦛 ᕗ(•‿•)ᕕ",
}

func NewHippoSpinner(out io.Writer) *HippoSpinner {
	return &HippoSpinner{
		out:  out,
		stop: make(chan struct{}),
	}
}

func (h *HippoSpinner) Start(message string) {
	if h.running {
		return
	}
	h.running = true

	h.wg.Add(1)

	go func() {
		defer h.wg.Done()

		i := 0
		for {
			select {
			case <-h.stop:
				return
			default:
				fmt.Fprintf(h.out, "\r%s  %s", frames[i%len(frames)], message)
				time.Sleep(120 * time.Millisecond)
				i++
			}
		}
	}()
}

func (h *HippoSpinner) Stop() {
	if !h.running {
		return
	}
	close(h.stop)
	h.wg.Wait()

	fmt.Fprint(h.out, "\r\033[K")
}
