package ui

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/nsf/termbox-go"
)

const progressBarWidth = 50

func drawProgressBar(currentTime, totalTime int, width, y int) {
	percentage := float64(currentTime) / float64(totalTime)
	barWidth := (width / 2) - 2 // 50% of the screen width minus brackets
	filledWidth := int(percentage * float64(barWidth))
	startX := (width - barWidth - 2) / 2 // Center the progress bar

	// Draw the filled part of the progress bar
	for i := 0; i < filledWidth; i++ {
		termbox.SetCell(startX+i, y, '█', termbox.ColorGreen, termbox.ColorGreen)
	}

	// Draw the empty part of the progress bar
	for i := filledWidth; i < barWidth; i++ {
		termbox.SetCell(startX+i, y, ' ', termbox.ColorLightGray, termbox.ColorLightGray)
	}

	// Draw the progress percentage
	progressString := fmt.Sprintf(" %3.0f%%", percentage*100)
	for i, r := range progressString {
		termbox.SetCell(startX+barWidth+1+i, y, r, termbox.ColorWhite, termbox.ColorDefault)
	}
}

func DrawCountdownFull(total, elapsed time.Duration) {

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	remain := total - elapsed
	minutes := int(remain.Minutes())
	seconds := int(remain.Seconds()) % 60
	currentseconds := int(elapsed.Seconds())
	totalSecond := int(total.Seconds())

	countdownString := fmt.Sprintf("%02d:%02d", minutes, seconds)
	w, h := termbox.Size()

	drawProgressBar(currentseconds, totalSecond, w, h/2+6)

	x := (w - 48) / 2 // Adjusted for the new size of the big numbers and colon
	y := h/2 - 3      // Centering vertically

	for idx, r := range countdownString {
		color := termbox.ColorGreen // Default color for minutes
		if idx >= 3 {               // Change color for seconds
			color = termbox.ColorRed
		}
		for i, line := range bigNumbers[r] {
			for j, ch := range line {
				if ch != ' ' {
					termbox.SetCell(x+j, y+i, ch, color, termbox.ColorDefault)
				}
			}
		}
		x += 8 // Adjust the position for the next character, increase if needed
	}
	termbox.Flush()
}

func ClearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func DrawProgressBar(current, total float64) {
	progress := current / total
	numBlocks := int(progress * float64(progressBarWidth))
	fmt.Print("[")
	for i := 0; i < numBlocks; i++ {
		fmt.Print("█")
	}
	for i := 0; i < progressBarWidth-numBlocks; i++ {
		fmt.Print(" ")
	}
	fmt.Print("]")
	fmt.Printf(" %.2f%% %d/%d %.2f/%.f\n", progress*100, int(current), int(total), current/60.0, total/60)

}

var bigNumbers = map[rune][]string{
	':': {
		"     ",
		"  *  ",
		"     ",
		"  *  ",
		"     ",
	},
	'0': {
		" 000 ",
		"0   0",
		"0   0",
		"0   0",
		" 000 ",
	},
	'1': {
		" X ",
		"XX ",
		" X ",
		" X ",
		"XXX",
	},
	'2': {
		"XXX",
		"  X",
		"XXX",
		"X  ",
		"XXX",
	},
	'3': {
		"XXX",
		"  X",
		"XXX",
		"  X",
		"XXX",
	},
	'4': {
		"X X",
		"X X",
		"XXX",
		"  X",
		"  X",
	},
	'5': {
		"XXX",
		"X  ",
		"XXX",
		"  X",
		"XXX",
	},
	'6': {
		"XXX",
		"X  ",
		"XXX",
		"X X",
		"XXX",
	},
	'7': {
		"XXX",
		"  X",
		"  X",
		" X ",
		" X ",
	},
	'8': {
		"XXX",
		"X X",
		"XXX",
		"X X",
		"XXX",
	},
	'9': {
		"XXX",
		"X X",
		"XXX",
		"  X",
		"  X",
	},
}
