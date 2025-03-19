# LazyHTTP - A TUI HTTP Client

![LazyHTTP Client](https://img.shields.io/badge/LazyHTTP-TUI%20HTTP%20Client-7D56F4)
![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

A beautiful terminal-based HTTP client built with Go, featuring syntax highlighting and a user-friendly interface for exploring web content.

![Screenshot placeholder - add your own screenshot here]()

## Features

- **User-friendly Terminal UI** - Intuitive interface for making HTTP requests
- **Automatic Content Detection** - Identifies JSON, HTML, XML, CSS, and JavaScript
- **Syntax Highlighting** - Beautiful syntax coloring for better readability
- **Response Metadata** - Displays status codes, content types, and server information
- **Pretty Printing** - Formats JSON and HTML for improved readability
- **Keyboard Navigation** - Easy scrolling through large responses

## Installation

### Prerequisites

- Go 1.23 or higher

### From Source

1. Clone the repository:
   ```bash
   git clone https://github.com/lafarr/lazyhttp.git
   cd lazyhttp
   ```

2. Build the application:
   ```bash
   go build
   ```

3. Run the executable:
   ```bash
   ./lazyhttp
   ```

### Using Go Install

```bash
go install github.com/lafarr/lazyhttp@latest
```

## Usage

1. Launch the application:
   ```bash
   ./lazyhttp
   ```

2. Enter a URL in the input field (e.g., `https://example.com` or just `example.com`)

3. Press `Enter` to fetch the content

4. Use the up/down arrow keys to scroll through the response

5. Press `Esc` or `Ctrl+C` to exit

## Key Controls

- **↑/↓**: Scroll through content
- **Enter**: Fetch URL
- **Ctrl+C/Esc**: Quit application

## Dependencies

This project uses the following Go packages:

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - TUI framework
- **[Bubbles](https://github.com/charmbracelet/bubbles)** - TUI components
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** - Style definitions
- **[Chroma](https://github.com/alecthomas/chroma)** - Syntax highlighting
- **[GoHTML](https://github.com/yosssi/gohtml)** - HTML formatting

## Building from Source

To build from source, you need Go 1.23 or later installed on your system.

```bash
# Clone the repository
git clone https://github.com/lafarr/lazyhttp.git
cd lazyhttp

# Install dependencies
go mod download

# Build the application
go build -o lazyhttp

# Run the application
./lazyhttp
```

## Contributing

Contributions are welcome! Feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 