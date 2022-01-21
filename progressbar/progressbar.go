package progressbar

import (
	"fmt"
	"strings"
)

type Bar struct {
	percent int64  // progress percentage
	cur     int64  // current progress
	total   int64  // total value for progress
	rate    string // the actual progress bar to be printed
	graph   string // the fill value for progress bar
}

func (bar *Bar) NewOption(start, total int64) {
	bar.cur = start
	bar.total = total
	bar.graph = "█"
	bar.percent = bar.getPercent()
	for i := 0; i < int(bar.percent); i += 2 {
		bar.rate += bar.graph // initial progress position
	}
}

func (bar *Bar) getPercent() int64 {
	return int64((float32(bar.cur) / float32(bar.total)) * 100)
}

func (bar *Bar) Play(cur int64) {
	if cur <= bar.total {
		bar.cur = cur
	} else {
		bar.cur = bar.total
	}
	// last := bar.percent
	bar.percent = bar.getPercent()
	if bar.percent < 0 {
		return
	}
	bar.rate = strings.Repeat(bar.graph, int(bar.percent)/2)
	fmt.Printf("\r[%-50s]%3d%% %8d/%d", bar.rate, bar.percent, bar.cur, bar.total)
}

func (bar *Bar) Finish() {
	bar.Play(bar.total)
	// fmt.Println()
}
