package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yosssi/gohtml"
)

const (
	inputHeight = 1
	padding     = 2
)

var (
	// UI Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Width(20).
			Align(lipgloss.Center)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#3C3C3C")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#61AFEF"))
)

type fetchMsg struct {
	response string
	err      error
}

// Model represents the application state
type model struct {
	textInput textinput.Model
	viewport  viewport.Model
	response  string
	err       error
	fetching  bool
	width     int
	height    int
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter URL (e.g. https://example.com)"
	ti.Focus()
	ti.Width = 40
	ti.Prompt = "URL: "

	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2)

	return model{
		textInput: ti,
		viewport:  vp,
		response:  "Response will appear here",
		fetching:  false,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// prettyPrintJSON formats JSON with syntax highlighting using chroma
func prettyPrintJSON(input []byte) (string, error) {
	var data interface{}

	// Try to unmarshal as JSON
	if err := json.Unmarshal(input, &data); err != nil {
		return "", err
	}

	// Pretty print with indentation
	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	// Use chroma for syntax highlighting
	lexer := lexers.Get("json")
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Use a theme that works well in the terminal
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, string(prettyJSON))
	if err != nil {
		return string(prettyJSON), nil // Fall back to uncolored JSON
	}

	var buf strings.Builder
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return string(prettyJSON), nil
	}

	return buf.String(), nil
}

// prettyPrintHTML indents HTML using gohtml and adds syntax highlighting with chroma
func prettyPrintHTML(input []byte) (string, error) {
	// Format HTML using gohtml (handles proper indentation)
	formatted := gohtml.Format(string(input))

	// Use chroma for syntax highlighting
	lexer := lexers.Get("html")
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Use a theme that works well in the terminal
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, formatted)
	if err != nil {
		return formatted, nil // Fall back to uncolored but formatted HTML
	}

	var buf strings.Builder
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return formatted, nil
	}

	return buf.String(), nil
}

// detectContentType tries to determine the content type from response body
func detectContentType(body []byte, contentType string) string {
	// First, try to use the provided content type
	contentTypeLower := strings.ToLower(contentType)

	if strings.Contains(contentTypeLower, "application/json") {
		return "json"
	} else if strings.Contains(contentTypeLower, "text/html") {
		return "html"
	} else if strings.Contains(contentTypeLower, "text/css") {
		return "css"
	} else if strings.Contains(contentTypeLower, "application/javascript") ||
		strings.Contains(contentTypeLower, "text/javascript") {
		return "javascript"
	} else if strings.Contains(contentTypeLower, "text/xml") ||
		strings.Contains(contentTypeLower, "application/xml") {
		return "xml"
	} else if strings.Contains(contentTypeLower, "text/plain") {
		// For plain text, try to guess the format from content
		return detectTextFormat(body)
	}

	// If content-type is not helpful, try to guess from content
	return detectTextFormat(body)
}

// detectTextFormat tries to guess the format of text content
func detectTextFormat(body []byte) string {
	content := string(body)

	// Check for JSON
	if len(content) > 0 && (content[0] == '{' || content[0] == '[') {
		var js interface{}
		if json.Unmarshal(body, &js) == nil {
			return "json"
		}
	}

	// Check for HTML
	if strings.Contains(content, "<!DOCTYPE html>") ||
		strings.Contains(content, "<html") ||
		(strings.Contains(content, "<head") && strings.Contains(content, "<body")) {
		return "html"
	}

	// Check for XML
	if strings.HasPrefix(strings.TrimSpace(content), "<?xml") ||
		strings.HasPrefix(strings.TrimSpace(content), "<") &&
			regexp.MustCompile(`<[a-zA-Z0-9]+( [^>]*)?>.*</[a-zA-Z0-9]+>`).MatchString(content) {
		return "xml"
	}

	// Check for CSS
	if regexp.MustCompile(`[a-z0-9\-_\.#]+ {[^}]*}`).MatchString(content) {
		return "css"
	}

	// Check for JavaScript
	if regexp.MustCompile(`function [a-zA-Z0-9_]+ *\(`).MatchString(content) ||
		regexp.MustCompile(`var [a-zA-Z0-9_]+ *=`).MatchString(content) ||
		regexp.MustCompile(`const [a-zA-Z0-9_]+ *=`).MatchString(content) {
		return "javascript"
	}

	// Default to plain text
	return "text"
}

// prettyPrintContent applies syntax highlighting based on content type
func prettyPrintContent(body []byte, detectedType string) string {
	// Get lexer based on detected type
	var lexer chroma.Lexer

	switch detectedType {
	case "json":
		lexer = lexers.Get("json")
	case "html":
		// For HTML, use gohtml first for proper indentation
		formatted := gohtml.Format(string(body))
		lexer = lexers.Get("html")
		body = []byte(formatted)
	case "xml":
		lexer = lexers.Get("xml")
	case "css":
		lexer = lexers.Get("css")
	case "javascript":
		lexer = lexers.Get("javascript")
	default:
		// Try to detect by content
		lexer = lexers.Analyse(string(body))
	}

	// Fallback if no lexer found
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Use a theme that works well in terminals
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	// Use terminal formatter
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Apply highlighting
	iterator, err := lexer.Tokenise(nil, string(body))
	if err != nil {
		return string(body) // Fall back to raw content
	}

	var buf strings.Builder
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return string(body)
	}

	return buf.String()
}

func fetchURL(url string) tea.Cmd {
	return func() tea.Msg {
		// Create a request with custom User-Agent to avoid some blocks
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fetchMsg{err: err}
		}

		// Add a common user agent
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

		// Send the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fetchMsg{err: err}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fetchMsg{err: err}
		}

		// Get content type from header
		contentType := resp.Header.Get("Content-Type")

		// Create a header with response information
		headerInfo := &strings.Builder{}
		fmt.Fprintf(headerInfo, "%s %s\n",
			headerStyle.Render("Status:"),
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#56B6C2")).Render(resp.Status))

		fmt.Fprintf(headerInfo, "%s %s\n",
			headerStyle.Render("Content-Type:"),
			lipgloss.NewStyle().Italic(true).Render(contentType))

		if len(resp.Header.Get("Server")) > 0 {
			fmt.Fprintf(headerInfo, "%s %s\n",
				headerStyle.Render("Server:"),
				resp.Header.Get("Server"))
		}

		// Detect the actual content type from the body
		detectedType := detectContentType(body, contentType)

		// Add the detected type if it differs from content-type header
		if !strings.Contains(strings.ToLower(contentType), detectedType) {
			fmt.Fprintf(headerInfo, "%s %s\n",
				headerStyle.Render("Detected Format:"),
				lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFCC00")).
					Render(strings.ToUpper(detectedType)))
		}

		headerInfo.WriteString("\n")

		// Apply syntax highlighting based on detected type
		formattedContent := prettyPrintContent(body, detectedType)

		// Combine header and formatted content
		formattedResponse := headerInfo.String() + formattedContent

		return fetchMsg{response: formattedResponse}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if !m.fetching && m.textInput.Value() != "" {
				url := m.textInput.Value()
				if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
					url = "https://" + url
				}
				m.fetching = true
				m.response = "Fetching..."
				m.err = nil
				return m, fetchURL(url)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = m.width - padding*2
		m.viewport.Height = m.height - inputHeight - padding*3
		m.textInput.Width = m.width - padding*2 - len(m.textInput.Prompt)
		m.viewport.SetContent(m.response)

	case fetchMsg:
		m.fetching = false
		if msg.err != nil {
			m.err = msg.err
			m.response = ""
		} else {
			m.err = nil
			m.response = msg.response
		}
		m.viewport.SetContent(m.response)
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	title := titleStyle.Render("URL Fetcher")
	input := m.textInput.View()
	if m.fetching {
		input += " " + lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC00")).Render("Loading...")
	}
	inputBox := inputStyle.Render(input)

	var responseView string
	if m.err != nil {
		responseView = errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	} else {
		responseView = m.viewport.View()
	}

	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Render("\n↑/↓: Scroll • Enter: Fetch URL • Ctrl+C/Esc: Quit")

	// Create a border around everything
	container := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#336699")).
		Padding(1, 2).
		Render(fmt.Sprintf("%s\n\n%s\n\n%s", title, inputBox, responseView))

	// Lay out the components
	return container + helpText
}

func main() {
	fmt.Println("Starting URL Fetcher TUI...")

	// Set up the program with mouse support
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
	}
}
