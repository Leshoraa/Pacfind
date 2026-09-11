package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"pacfind/internal/aur"
	"pacfind/internal/models"
	"pacfind/internal/pacman"
	"pacfind/internal/ui"
)

const Version = "1.3.0"

func main() {
	var (
		flagOfficial    bool
		flagAUR         bool
		flagLimit       int
		flagMore        bool
		flagInteractive bool
		flagTopDown     bool
		flagVersion     bool
	)

	flag.BoolVar(&flagOfficial, "o", false, "Search official repositories only")
	flag.BoolVar(&flagOfficial, "official", false, "Search official repositories only")
	flag.BoolVar(&flagAUR, "a", false, "Search AUR only")
	flag.BoolVar(&flagAUR, "aur", false, "Search AUR only")
	flag.IntVar(&flagLimit, "l", 6, "Limit number of results per category (default 6)")
	flag.IntVar(&flagLimit, "limit", 6, "Limit number of results per category (default 6)")
	flag.BoolVar(&flagMore, "m", false, "Show more/all results (unlimited)")
	flag.BoolVar(&flagMore, "more", false, "Show more/all results (unlimited)")
	flag.BoolVar(&flagInteractive, "i", false, "Interactive selection and installation via fzf")
	flag.BoolVar(&flagInteractive, "interactive", false, "Interactive selection and installation via fzf")
	flag.BoolVar(&flagTopDown, "t", false, "Render results top-down instead of default bottom-up")
	flag.BoolVar(&flagTopDown, "topdown", false, "Render results top-down instead of default bottom-up")
	flag.BoolVar(&flagVersion, "v", false, "Show version")
	flag.BoolVar(&flagVersion, "version", false, "Show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: pacfind [options] <query>\n\n")
		fmt.Fprintf(os.Stderr, "A clean, fast, and minimalist Arch Linux & AUR package search tool.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -m, --more         Show all results (overrides default limit of 6)\n")
		fmt.Fprintf(os.Stderr, "  -l, --limit <num>  Custom limit per category (default 6)\n")
		fmt.Fprintf(os.Stderr, "  -o, --official     Search official repositories only\n")
		fmt.Fprintf(os.Stderr, "  -a, --aur          Search AUR only\n")
		fmt.Fprintf(os.Stderr, "  -t, --topdown      Render results top-down (default is bottom-up / descending)\n")
		fmt.Fprintf(os.Stderr, "  -i, --interactive  Select and install package interactively using fzf\n")
		fmt.Fprintf(os.Stderr, "  -v, --version      Show pacfind version\n")
		fmt.Fprintf(os.Stderr, "  -h, --help         Show this help message\n")
	}

	flag.Parse()

	if flagVersion {
		fmt.Printf("pacfind %s (Go build)\n", Version)
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	query := strings.TrimSpace(strings.Join(args, " "))
	if query == "" {
		flag.Usage()
		os.Exit(1)
	}

	effectiveLimit := flagLimit
	if flagMore {
		effectiveLimit = 0 // unlimited
	}

	showOfficial := true
	showAUR := true
	if flagOfficial && !flagAUR {
		showAUR = false
	} else if flagAUR && !flagOfficial {
		showOfficial = false
	}

	// Concurrency via Goroutines
	var (
		wg           sync.WaitGroup
		installedMap map[string]bool
		officialPkgs []models.Package
		aurPkgs      []models.Package
		aurErr       error
	)

	// 1. Fetch installed packages
	wg.Add(1)
	go func() {
		defer wg.Done()
		installedMap = pacman.GetInstalled()
	}()

	// 2. Fetch official packages
	if showOfficial {
		wg.Add(1)
		go func() {
			defer wg.Done()
			officialPkgs, _ = pacman.SearchOfficial(query)
		}()
	}

	wg.Wait() // wait for installedMap before searching AUR so we can tag installed packages

	// 3. Fetch AUR packages
	if showAUR {
		wg.Add(1)
		go func() {
			defer wg.Done()
			aurPkgs, aurErr = aur.SearchAUR(query, installedMap)
		}()
		wg.Wait()
	}

	// Handle Interactive Mode (-i)
	if flagInteractive {
		runInteractive(officialPkgs, aurPkgs)
		return
	}

	// Render Clean Card UI in bottom-up (or top-down) mode
	ui.RenderResults(officialPkgs, aurPkgs, showOfficial, showAUR, effectiveLimit, aurErr, flagTopDown)
}

func runInteractive(official []models.Package, aurPkgs []models.Package) {
	var items []string

	for _, p := range official {
		st := ""
		if p.IsInstalled {
			st = " [installed]"
		}
		items = append(items, fmt.Sprintf("[official] %s/%s (%s)%s - %s",
			p.Repo, p.Name, p.Version, st, p.Description))
	}

	for _, p := range aurPkgs {
		st := ""
		if p.IsInstalled {
			st = " [installed]"
		}
		items = append(items, fmt.Sprintf("[aur]      aur/%s (%s)%s - %s",
			p.Name, p.Version, st, p.Description))
	}

	if len(items) == 0 {
		fmt.Printf("%sNo packages found to select.%s\n", ui.Yellow, ui.Reset)
		return
	}

	cmd := exec.Command("fzf",
		"--ansi",
		"--header=Select a package to install (ENTER: Install, ESC: Cancel)",
		"--prompt=pacfind > ",
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating fzf pipe: %v\n", err)
		return
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, strings.Join(items, "\n"))
	}()

	out, err := cmd.Output()
	if err != nil {
		return
	}

	selected := strings.TrimSpace(string(out))
	if selected == "" {
		return
	}

	parts := strings.Fields(selected)
	if len(parts) < 2 {
		return
	}

	isAUR := parts[0] == "[aur]"
	target := parts[1]
	if slashIdx := strings.LastIndex(target, "/"); slashIdx != -1 {
		target = target[slashIdx+1:]
	}

	var installCmd *exec.Cmd
	if isAUR {
		installCmd = exec.Command("yay", "-S", target)
	} else {
		installCmd = exec.Command("sudo", "pacman", "-S", target)
	}

	installCmd.Stdin = os.Stdin
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	fmt.Printf("\n%sExecuting: %s%s\n\n", ui.Bold, strings.Join(installCmd.Args, " "), ui.Reset)
	_ = installCmd.Run()
}
