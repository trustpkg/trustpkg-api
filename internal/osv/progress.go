package osv

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const progressLogInterval = 100 * time.Millisecond

type statusReporter struct {
	mu        sync.Mutex
	phase     string
	current   int64
	total     int64
	divisor   int64
	unit      string
	started   time.Time
	stop      chan struct{}
	done      chan struct{}
	lastLine  int
	lastPrint time.Time
	printMu   sync.Mutex
	finished  atomic.Bool
}

func newStatusReporter(phase, unit string, total, divisor int64) *statusReporter {
	reporter := &statusReporter{
		phase: phase, unit: unit, total: total, divisor: divisor, started: time.Now(),
		stop: make(chan struct{}), done: make(chan struct{}),
	}
	reporter.print(false)
	go reporter.run()
	return reporter
}

func (reporter *statusReporter) run() {
	ticker := time.NewTicker(progressLogInterval)
	defer ticker.Stop()
	defer close(reporter.done)
	for {
		select {
		case <-ticker.C:
			reporter.print(false)
		case <-reporter.stop:
			return
		}
	}
}

func (reporter *statusReporter) update(current int64) {
	reporter.mu.Lock()
	reporter.current = current
	shouldPrint := time.Since(reporter.lastPrint) >= progressLogInterval
	reporter.mu.Unlock()
	if shouldPrint {
		reporter.print(false)
	}
}

func (reporter *statusReporter) finish(message string) {
	if !reporter.finished.CompareAndSwap(false, true) {
		return
	}
	close(reporter.stop)
	<-reporter.done
	reporter.print(true)
	log.Printf("%s", message)
}

func (reporter *statusReporter) print(final bool) {
	reporter.mu.Lock()
	if !final && time.Since(reporter.lastPrint) < progressLogInterval {
		reporter.mu.Unlock()
		return
	}
	current, total, phase, unit, started := reporter.current, reporter.total, reporter.phase, reporter.unit, reporter.started
	reporter.lastPrint = time.Now()
	reporter.mu.Unlock()

	percent := float64(0)
	if total > 0 {
		percent = float64(current) / float64(total) * 100
	}
	perSecond := float64(current) / time.Since(started).Seconds()
	displayedCurrent := float64(current) / float64(reporter.divisor)
	displayedTotal := float64(total) / float64(reporter.divisor)
	line := fmt.Sprintf("%s: %.1f%% | %.2f/%.2f %s | %.2f %s/s", phase, percent, displayedCurrent, displayedTotal, unit, perSecond/float64(reporter.divisor), unit)
	if total <= 0 {
		line = fmt.Sprintf("%s: %.2f %s | %.2f %s/s", phase, displayedCurrent, unit, perSecond/float64(reporter.divisor), unit)
	}

	reporter.printMu.Lock()
	defer reporter.printMu.Unlock()
	// Clear the previous line before writing the next state.
	padding := ""
	if len(line) < reporter.lastLine {
		padding = fmt.Sprintf("%*s", reporter.lastLine-len(line), "")
	}
	fmt.Fprintf(os.Stdout, "\r\033[2K%s%s", line, padding)
	reporter.lastLine = len(line)
	if final {
		fmt.Fprintln(os.Stdout)
	}
}

type progressReader struct {
	reader     io.Reader
	total      int64
	downloaded int64
	start      time.Time
	reporter   *statusReporter
	closer     io.Closer
}

func (reader *progressReader) Read(p []byte) (int, error) {
	n, err := reader.reader.Read(p)
	reader.downloaded += int64(n)
	if n > 0 {
		reader.reporter.update(reader.downloaded)
	}
	return n, err
}

func (reader *progressReader) Close() error {
	reader.reporter.finish(fmt.Sprintf("OSV download completed: %.2f MB", float64(reader.downloaded)/(1024*1024)))
	return reader.closer.Close()
}
