# Pacfind

A minimalist, clean, and blazingly fast command-line search tool for Arch Linux and the Arch User Repository (AUR), written in **Go**.

Designed for clarity and speed: no emojis, native binary startup (~2 ms), parallel Goroutines, bottom-up rendering, and a clean box-drawing card interface.

## Preview

```text
ARCH USER REPOSITORY (AUR)

  ... (111 more packages not shown, use -m to see all)

┌─ waybar-module-music-git - aur/waybar-module-music-git 0.4.1_r235.39e4371-1 (+2, 0.4)
│ A Waybar module to show & control the current MPRIS media players state
└──────────────────────────────────────────────────────────

┌─ waybar-module-pacman-updates-git - aur/waybar-module-pacman-updates-git 0.2.14-1 (+9, 0.5)
│ Waybar module for Arch to show system updates available
└──────────────────────────────────────────────────────────

OFFICIAL REPOSITORIES

┌─ waybar - extra/waybar 0.15.0-3 [installed]
│ Highly customizable Wayland bar for Sway and Wlroots based compositors
└──────────────────────────────────────────────────────────

Summary: 1 found in official repositories, 117 found in AUR
```

## Features

- **Blazingly Fast**: Written in pure Go with zero runtime dependencies. Startup latency is ~2 milliseconds.
- **Bottom-Up Rendering**: Auto-scroll friendly! The most relevant and official packages are printed at the bottom right above your prompt, so you never have to scroll up to find the best match.
- **Sensible Defaults**: Automatically shows the top 6 most relevant packages. Use `-m` to reveal all results.
- **Parallel Search**: Concurrently queries local pacman databases (`pacman -Ss`) and the AUR RPC v5 REST API using lightweight goroutines.
- **Card-Style Interface**: Elegant box-drawing borders (`┌─`, `│`, `└─`), adaptive width wrapping, and subtle color highlights with zero emojis.
- **Installation Detection**: Highlights packages already installed on your system with `[installed]`.
- **AUR Community Metrics**: Displays votes, popularity scores, and `[out-of-date]` warning badges.
- **Interactive Mode**: Built-in `-i` / `--interactive` flag integrating with `fzf` to pick and install packages via `yay` or `pacman`.

## Installation

### Prerequisites
- [Go](https://go.dev/) 1.22+ (for building)
- `pacman` and `yay` (Arch Linux)
- `fzf` (optional, for interactive install mode)

### Build & Install

```bash
git clone https://github.com/<your-username>/Pacfind.git ~/Projects/CLI/Pacfind
cd ~/Projects/CLI/Pacfind
./install.sh
```

Or manually:
```bash
go build -ldflags="-s -w" -o pacfind .
install -m 755 pacfind ~/.local/bin/
```

Make sure `~/.local/bin` is in your `$PATH`.

## Usage

### Basic Search (Default top 6)
```bash
pacfind waybar
```

### Show More / All Results
```bash
pacfind -m waybar
```

### Custom Limit
```bash
pacfind -l 10 waybar
```

### Search Official Repositories Only
```bash
pacfind -o neovim
```

### Search AUR Only
```bash
pacfind -a zen-browser
```

### Top-Down Render (Old order)
```bash
pacfind -t waybar
```

### Interactive Install Mode (fzf)
```bash
pacfind -i rofi
```

### Help & Version
```bash
pacfind --help
pacfind --version
```

## License

MIT License
