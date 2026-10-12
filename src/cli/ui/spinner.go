package ui

import (
	"fmt"
	"sync"
	"time"
)

type Spinner struct {
	label     string
	stopChan  chan struct{}
	wg        sync.WaitGroup
	startTime time.Time
	mu        sync.Mutex
	active    bool
}

// Crée un nouveau spinner avec le libellé indiqué.
func NewSpinner(label string) *Spinner {
	return &Spinner{
		label:    label,
		stopChan: make(chan struct{}),
	}
}

// Démarre l'animation du spinner dans une goroutine.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.startTime = time.Now()
	s.mu.Unlock()

	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	fmt.Print(HideCursor)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer fmt.Print(ShowCursor)

		i := 0
		for {
			select {
			case <-s.stopChan:
				return
			default:
				s.mu.Lock()
				currentLabel := s.label
				s.mu.Unlock()

				elapsed := time.Since(s.startTime).Seconds()
				frame := frames[i%len(frames)]

				fmt.Printf("\r\033[K  %s%s%s %s %s(%.1fs)%s",
					Cyan+Bold, frame, Reset,
					currentLabel,
					Dim, elapsed, Reset,
				)

				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

// Met à jour le libellé affiché par le spinner.
func (s *Spinner) UpdateLabel(newLabel string) {
	s.mu.Lock()
	s.label = newLabel
	s.mu.Unlock()
}

// Arrête le spinner et affiche éventuellement un message final.
func (s *Spinner) Stop(finalMsg string) {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	close(s.stopChan)
	s.mu.Unlock()

	s.wg.Wait()
	fmt.Print(ShowCursor)

	if finalMsg != "" {
		fmt.Printf("\r\033[K  %s\n", finalMsg)
	} else {
		fmt.Printf("\r\033[K")
	}
}
