package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//go:embed .env
var envFile string

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type Command struct {
	Command string `json:"command"`
}

type model struct {
	textInput string
	cursor    int
	loading   bool
	result    string
	error     string
	context   string
	done      bool
}

type generateMsg struct {
	result string
	err    error
}

var (
	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF75B7"))

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	resultStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)
)

func initialModel() model {
	ctx := gatherContext()

	return model{
		textInput: "",
		cursor:    0,
		loading:   false,
		context:   ctx,
		done:      false,
	}
}

func gatherContext() string {
	var ctx strings.Builder

	// PWD
	if pwd, err := os.Getwd(); err == nil {
		ctx.WriteString(fmt.Sprintf("PWD: %s\n", pwd))
	}

	// OS (uname -a equivalent)
	if cmd := exec.Command("uname", "-a"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			ctx.WriteString(fmt.Sprintf("OS: %s\n", strings.TrimSpace(string(output))))
		}
	}

	// PATH
	ctx.WriteString(fmt.Sprintf("PATH: %s\n", os.Getenv("PATH")))

	// COMMANDS (compgen -c equivalent - get available commands)
	if cmd := exec.Command("bash", "-c", "compgen -c | sort | uniq | tr '\\n' ' '"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			ctx.WriteString(fmt.Sprintf("COMMANDS: %s\n", strings.TrimSpace(string(output))))
		}
	}

	// HISTORY (last command)
	if cmd := exec.Command("bash", "-c", "history 1"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			ctx.WriteString(fmt.Sprintf("HISTORY: %s\n", strings.TrimSpace(string(output))))
		}
	}

	return ctx.String()
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "enter":
			if m.result != "" {
				// Command is ready, exit and print result
				m.done = true
				return m, tea.Quit
			}
			if m.textInput != "" {
				// Start generating
				m.loading = true
				return m, m.generateCommand()
			}

		case "backspace":
			if len(m.textInput) > 0 {
				m.textInput = m.textInput[:len(m.textInput)-1]
				if m.cursor > 0 {
					m.cursor--
				}
			}

		default:
			// Regular character input
			if !m.loading && m.result == "" {
				m.textInput += msg.String()
				m.cursor = len(m.textInput)
			}
		}

	case generateMsg:
		m.loading = false
		if msg.err != nil {
			m.error = msg.err.Error()
		} else {
			m.result = msg.result
		}
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	if m.loading {
		s.WriteString(loadingStyle.Render("🔄 Generating command..."))
		s.WriteString("\n")
	} else if m.result != "" {
		s.WriteString("Generated: ")
		s.WriteString(resultStyle.Render(m.result))
		s.WriteString("\n")
		s.WriteString("Press Enter to use, ESC to cancel")
	} else if m.error != "" {
		s.WriteString(errorStyle.Render("Error: " + m.error))
		s.WriteString("\n")
		s.WriteString("Press ESC to exit")
	} else {
		s.WriteString("What command do you need? ")
		s.WriteString(inputStyle.Render(m.textInput))

		// Show cursor
		if m.cursor == len(m.textInput) {
			s.WriteString(inputStyle.Render("_"))
		}
	}

	return s.String()
}

func (m model) generateCommand() tea.Cmd {
	return func() tea.Msg {
		fullPrompt := m.context + "---" + m.textInput
		result, err := GetCommand(
			"You are an expert shell assistant. Output only a single valid shell command.\n\nRETURN THE RESULT ONLY IN THE FOLLOWING JSON FORMAT: {\"command\": \"<command>\"}",
			fullPrompt,
		)
		return generateMsg{result: result, err: err}
	}
}

func buildJsonBody(systemPrompt string, prompt string) ([]byte, error) {
	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	type RequestBody struct {
		Messages       []Message `json:"messages"`
		ResponseFormat struct {
			Type string `json:"type"`
		} `json:"response_format"`
	}

	body := RequestBody{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
	}
	body.ResponseFormat.Type = "json_object"

	return json.Marshal(body)
}

func loadEmbeddedEnv() {
	lines := strings.Split(envFile, "\n")
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
				fixed := strings.ReplaceAll(parts[1], `"`, "")
				os.Setenv(parts[0], fixed)
			}
		}
	}
}

func GetCommand(systemPrompt string, prompt string) (string, error) {
	openaiKey := os.Getenv("OPENAI_KEY")
	if openaiKey == "" {
		return "", fmt.Errorf("OPENAI_KEY environment variable is not set")
	}

	openaiEndpoint := os.Getenv("OPENAI_ENDPOINT")
	if openaiEndpoint == "" {
		return "", fmt.Errorf("OPENAI_ENDPOINT environment variable is not set")
	}

	body, err := buildJsonBody(systemPrompt, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to build JSON body: %w", err)
	}

	req, err := http.NewRequest("POST", openaiEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", openaiKey)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var openAIResponse OpenAIResponse
	err = json.Unmarshal(responseData, &openAIResponse)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal OpenAI response: %w", err)
	}

	if len(openAIResponse.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from OpenAI: %s", string(responseData))
	}

	var command Command
	err = json.Unmarshal([]byte(openAIResponse.Choices[0].Message.Content), &command)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal command: %w", err)
	}

	return command.Command, nil
}

func main() {
	loadEmbeddedEnv()

	// Check if we have a temp file argument
	var tempFile string
	if len(os.Args) > 1 {
		tempFile = os.Args[1]
	}

	// Run the TUI
	p := tea.NewProgram(initialModel())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}

	// Output the result
	if m, ok := finalModel.(model); ok && m.result != "" && m.done {
		if tempFile != "" {
			// Write to temp file for zsh plugin
			if err := os.WriteFile(tempFile, []byte(m.result), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing result: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Write to stdout for direct usage
			fmt.Print(m.result)
		}
	}
}
