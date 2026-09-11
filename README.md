# Pacfind

A minimalist, clean, and fast command-line search tool for Arch Linux and the Arch User Repository (AUR).

Designed for clarity and speed: no emojis, clean ANSI typography, categorized sections, and parallel query execution.

## Features

- **Categorized Sections**: Clearly separates Official Repositories (`core`, `extra`, `multilib`) from the Arch User Repository (`aur`).
- **Parallel Search**: Executes `pacman -Ss` (local sync DB) and AUR RPC v5 queries concurrently.
- **Clean Layout**: Zero emojis, consistent indentation, terminal-adaptive dividers, and subtle color highlights.
- **Installation Detection**: Automatically detects and highlights packages already installed on your system with `[installed]`.
- **AUR Metrics**: Shows votes, popularity, and `[out-of-date]` flags for AUR packages.
- **Smart Sorting**: Exact matches first, followed by prefix matches and popularity metrics.
- **Interactive Mode**: Optional `-i` / `--interactive` flag with `fzf` to pick and install packages directly via `pacman` or `yay`.

## Installation

Clone the repository and run the install script:

```bash
git clone https://github.com/<your-username>/Pacfind.git ~/Projects/Pacfind
cd ~/Projects/Pacfind
./install.sh
```

Ensure `~/.local/bin` is in your `$PATH`.

## Usage

### Basic Search
```bash
pacfind firefox
```

### Search Official Repositories Only
```bash
pacfind -o neovim
```

### Search AUR Only
```bash
pacfind -a zen-browser
```

### Limit Results
```bash
pacfind -l 5 hyprland
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

## Requirements

- Python 3.8+
- Arch Linux (`pacman`, `yay`)
- `fzf` (optional, required only for interactive mode)

## License

MIT License
