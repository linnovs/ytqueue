# ytqueue

A terminal-based YouTube queue manager built with Go and Bubbletea. Download and queue YouTube videos for offline playback using mpv.

## Features

- **Terminal UI**: Intuitive interface built with Bubbletea and Lipgloss
- **YouTube Downloads**: Download videos using yt-dlp with progress tracking
- **Queue Management**: Organize downloaded videos in a queue with watched status
- **Media Playback**: Play videos using mpv media player
- **SQLite Storage**: Persistent storage of video metadata and queue state
- **Configurable**: Customizable download paths and settings via TOML config

## Prerequisites

- Go 1.25.5 or later
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - for downloading YouTube videos
- [mpv](https://mpv.io/) - for video playback
- SQLite (automatically handled via modernc.org/sqlite)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/linnovs/ytqueue.git
   cd ytqueue
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the application:
   ```bash
   go build -o ytqueue .
   ```

## Configuration

Create a configuration file at `~/.config/ytqueue/config.toml`:

```toml
[download]
path = "~/Downloads"          # Download directory (default: ~/Downloads)
temp_name = "ytqueue_temp"    # Temporary directory prefix (default: ytqueue_temp)
```

## Usage

Run the application:
```bash
./ytqueue
```

### Keybindings

- **Tab/Shift+Tab**: Navigate between sections
- **Enter**: Submit URL or play selected video
- **Space**: Update watched status of selected video
- **q**: Quit application
- **F1**: Toggle help

### Sections

1. **URL Prompt**: Enter YouTube URLs to download and queue
2. **Video Queue**: Browse and manage downloaded videos
3. **Download Status**: View active download progress

## Database

The application uses SQLite for storing video metadata. The database is automatically created and migrated on first run.

Video data includes:
- Video ID and title
- Original URL
- Local file location
- Watched status
- Queue order and creation timestamp

## Dependencies

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - YouTube downloader
- [mpv](https://mpv.io/) - Media player
- [SQLite](https://www.sqlite.org/) - Database

## Development

The project uses:
- [sqlc](https://sqlc.dev/) for type-safe SQL queries
- [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations
- [golangci-lint](https://golangci-lint.run/) for code linting

To run linting:
```bash
golangci-lint run
```

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues for bugs and feature requests.

## License

See LICENSE file for details.
