package ui

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"pacfind/internal/models"
)

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// Colors holds ANSI escape codes
var (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Cyan      = "\033[1;36m"
	Blue      = "\033[1;34m"
	Magenta   = "\033[1;35m"
	Green     = "\033[0;32m"
	BoldGreen = "\033[1;32m"
	Yellow    = "\033[0;33m"
	White     = "\033[1;37m"
	Red       = "\033[1;31m"
	BoxBorder = "\033[2m" // Dim clean box borders
)

func init() {
	if os.Getenv("NO_COLOR") != "" {
		disableColors()
	}
}

func disableColors() {
	Reset = ""
	Bold = ""
	Dim = ""
	Cyan = ""
	Blue = ""
	Magenta = ""
	Green = ""
	BoldGreen = ""
	Yellow = ""
	White = ""
	Red = ""
	BoxBorder = ""
}

// GetTerminalWidth returns the column count of the current terminal
func GetTerminalWidth() int {
	ws := &winsize{}
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))
	if int(retCode) >= 0 && ws.Col > 0 {
		return int(ws.Col)
	}
	return 80
}

// RenderCard prints a package in the user's requested box-drawing card style
func RenderCard(pkg models.Package, width int) {
	cardWidth := width
	if cardWidth > 80 {
		cardWidth = 80
	}
	if cardWidth < 40 {
		cardWidth = 40
	}

	repoColor := Blue
	if pkg.Repo == "aur" {
		repoColor = Magenta
	}

	status := ""
	if pkg.IsInstalled {
		status = fmt.Sprintf(" %s[installed]%s", BoldGreen, Reset)
	}

	metrics := ""
	if pkg.Repo == "aur" {
		metrics = fmt.Sprintf(" %s(+%d, %.1f)%s", Dim, pkg.Votes, pkg.Popularity, Reset)
	}

	ood := ""
	if pkg.OutOfDate {
		ood = fmt.Sprintf(" %s[out-of-date]%s", Red, Reset)
	}

	// Top line: ┌─ name - repo/name version [metrics] [ood] [installed]
	fmt.Printf("%s┌─%s %s%s%s %s-%s %s%s%s%s/%s%s%s%s %s%s%s%s%s%s\n",
		BoxBorder, Reset,
		White, pkg.Name, Reset,
		Dim, Reset,
		repoColor, pkg.Repo, Reset,
		Dim, Reset,
		White, pkg.Name, Reset,
		Green, pkg.Version, Reset,
		metrics, ood, status,
	)

	// Middle line(s): │ description (wrapped gracefully)
	desc := pkg.Description
	if desc == "" {
		desc = "No description available"
	}

	maxTextWidth := cardWidth - 4
	wrappedLines := wrapText(desc, maxTextWidth)
	for _, l := range wrappedLines {
		fmt.Printf("%s│%s %s\n", BoxBorder, Reset, l)
	}

	// Bottom line: └────────────────────────────
	bottomLen := cardWidth - 1
	if bottomLen < 10 {
		bottomLen = 10
	}
	fmt.Printf("%s└%s%s\n", BoxBorder, strings.Repeat("─", bottomLen), Reset)
	fmt.Println() // space between cards
}

// wrapText breaks long descriptions into clean lines
func wrapText(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	var lines []string
	var current strings.Builder

	for _, w := range words {
		if current.Len()+len(w)+1 > maxLen {
			if current.Len() > 0 {
				lines = append(lines, current.String())
				current.Reset()
			}
		}
		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(w)
	}

	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return lines
}

// RenderResults prints the full search view.
// By default (topDown = false), results are rendered in bottom-up order (AUR first, Official second,
// with items descending so the #1 best match is closest to the bottom / prompt).
func RenderResults(
	official []models.Package,
	aurPkgs []models.Package,
	showOfficial bool,
	showAUR bool,
	limit int,
	aurErr error,
	topDown bool,
) {
	width := GetTerminalWidth()

	if topDown {
		renderTopDown(official, aurPkgs, showOfficial, showAUR, limit, aurErr, width)
	} else {
		renderBottomUp(official, aurPkgs, showOfficial, showAUR, limit, aurErr, width)
	}
}

// renderBottomUp renders AUR first and Official second, with items reversed
// so the most relevant/exact match is closest to the prompt at the bottom!
func renderBottomUp(
	official []models.Package,
	aurPkgs []models.Package,
	showOfficial bool,
	showAUR bool,
	limit int,
	aurErr error,
	width int,
) {
	// 1. Arch User Repository (AUR) rendered first (scrolled up)
	if showAUR {
		fmt.Printf("%s%sARCH USER REPOSITORY (AUR)%s\n", Bold, Magenta, Reset)
		fmt.Println()

		if aurErr != nil {
			fmt.Printf("  %sNotice: %v%s\n\n", Yellow, aurErr, Reset)
		} else if len(aurPkgs) == 0 {
			fmt.Printf("  %sNo packages found in AUR.%s\n\n", Dim, Reset)
		} else {
			displayList := aurPkgs
			if limit > 0 && len(displayList) > limit {
				displayList = displayList[:limit]
			}

			// In bottom-up, show overflow notice at top
			if limit > 0 && len(aurPkgs) > limit {
				fmt.Printf("  %s... (%d more packages not shown, use -l or omit to see all)%s\n\n",
					Dim, len(aurPkgs)-limit, Reset)
			}

			// Render items in reverse order (least relevant -> most relevant at bottom)
			for i := len(displayList) - 1; i >= 0; i-- {
				RenderCard(displayList[i], width)
			}
		}
	}

	// 2. Official Repositories rendered second (closest to bottom / visible on screen)
	if showOfficial {
		fmt.Printf("%s%sOFFICIAL REPOSITORIES%s\n", Bold, Cyan, Reset)
		fmt.Println()

		if len(official) == 0 {
			fmt.Printf("  %sNo packages found in official repositories.%s\n\n", Dim, Reset)
		} else {
			displayList := official
			if limit > 0 && len(displayList) > limit {
				displayList = displayList[:limit]
			}

			// In bottom-up, show overflow notice at top
			if limit > 0 && len(official) > limit {
				fmt.Printf("  %s... (%d more packages not shown, use -l or omit to see all)%s\n\n",
					Dim, len(official)-limit, Reset)
			}

			// Render items in reverse order (partial matches -> exact match at bottom)
			for i := len(displayList) - 1; i >= 0; i-- {
				RenderCard(displayList[i], width)
			}
		}
	}

	// 3. Summary Footer
	dividerLen := width
	if dividerLen > 80 {
		dividerLen = 80
	}
	fmt.Printf("%s%s%s\n", Dim, strings.Repeat("─", dividerLen), Reset)

	totalOff := len(official)
	totalAur := len(aurPkgs)

	if showOfficial && showAUR {
		fmt.Printf("%sSummary:%s %d found in official repositories, %d found in AUR\n",
			Bold, Reset, totalOff, totalAur)
	} else if showOfficial {
		fmt.Printf("%sSummary:%s %d found in official repositories\n",
			Bold, Reset, totalOff)
	} else if showAUR {
		fmt.Printf("%sSummary:%s %d found in AUR\n",
			Bold, Reset, totalAur)
	}
	fmt.Printf("%s%s%s\n", Dim, strings.Repeat("─", dividerLen), Reset)
}

// renderTopDown renders Official first and AUR second in standard top-to-bottom order
func renderTopDown(
	official []models.Package,
	aurPkgs []models.Package,
	showOfficial bool,
	showAUR bool,
	limit int,
	aurErr error,
	width int,
) {
	if showOfficial {
		fmt.Printf("%s%sOFFICIAL REPOSITORIES%s\n", Bold, Cyan, Reset)
		fmt.Println()

		if len(official) == 0 {
			fmt.Printf("  %sNo packages found in official repositories.%s\n\n", Dim, Reset)
		} else {
			displayList := official
			if limit > 0 && len(displayList) > limit {
				displayList = displayList[:limit]
			}

			for _, p := range displayList {
				RenderCard(p, width)
			}

			if limit > 0 && len(official) > limit {
				fmt.Printf("  %s... (%d more packages not shown, use -l or omit to see all)%s\n\n",
					Dim, len(official)-limit, Reset)
			}
		}
	}

	if showAUR {
		fmt.Printf("%s%sARCH USER REPOSITORY (AUR)%s\n", Bold, Magenta, Reset)
		fmt.Println()

		if aurErr != nil {
			fmt.Printf("  %sNotice: %v%s\n\n", Yellow, aurErr, Reset)
		} else if len(aurPkgs) == 0 {
			fmt.Printf("  %sNo packages found in AUR.%s\n\n", Dim, Reset)
		} else {
			displayList := aurPkgs
			if limit > 0 && len(displayList) > limit {
				displayList = displayList[:limit]
			}

			for _, p := range displayList {
				RenderCard(p, width)
			}

			if limit > 0 && len(aurPkgs) > limit {
				fmt.Printf("  %s... (%d more packages not shown, use -l or omit to see all)%s\n\n",
					Dim, len(aurPkgs)-limit, Reset)
			}
		}
	}

	dividerLen := width
	if dividerLen > 80 {
		dividerLen = 80
	}
	fmt.Printf("%s%s%s\n", Dim, strings.Repeat("─", dividerLen), Reset)

	totalOff := len(official)
	totalAur := len(aurPkgs)

	if showOfficial && showAUR {
		fmt.Printf("%sSummary:%s %d found in official repositories, %d found in AUR\n",
			Bold, Reset, totalOff, totalAur)
	} else if showOfficial {
		fmt.Printf("%sSummary:%s %d found in official repositories\n",
			Bold, Reset, totalOff)
	} else if showAUR {
		fmt.Printf("%sSummary:%s %d found in AUR\n",
			Bold, Reset, totalAur)
	}
	fmt.Printf("%s%s%s\n", Dim, strings.Repeat("─", dividerLen), Reset)
}
