# termlm

> AI-powered terminal command generation for the modern shell

**termlm** is an intelligent command-line assistant that transforms natural language descriptions into executable shell commands. Simply describe what you want to accomplish, and termlm will generate the appropriate command for your system.

![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.23.4+-00ADD8.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey.svg)

## ✨ Features

- **Natural Language Processing**: Describe commands in plain English
- **Context-Aware**: Automatically gathers system information (OS, PATH, available commands, working directory) for accurate suggestions
- **Seamless Shell Integration**: Integrated zsh plugin with `Ctrl+K` hotkey
- **Real-time Generation**: Interactive TUI with immediate feedback
- **OpenAI Powered**: Leverages OpenAI's language models for intelligent command generation

## 🏗️ Architecture

termlm consists of two main components:

### 1. Core Engine (`main.go`)
- **TUI Application**: Built with [Charm's Bubbletea](https://github.com/charmbracelet/bubbletea) for an interactive terminal interface
- **Context Gathering**: Automatically collects system information including:
  - Current working directory
  - Operating system details
  - Available commands in PATH
  - Command history
- **AI Integration**: Communicates with OpenAI API to generate contextually appropriate commands
- **Response Processing**: Parses JSON responses and formats commands for execution

### 2. Shell Plugin (`llm-cmd.plugin.zsh`)
- **Zsh Integration**: Seamlessly integrates with oh-my-zsh
- **Hotkey Binding**: Bound to `Ctrl+K` for instant access
- **Command Buffer Integration**: Generated commands are inserted directly into your command line
- **Temporary File Handling**: Manages communication between the shell and the core engine

## 🚀 Installation

### Prerequisites

- Go 1.23.4 or later
- zsh with oh-my-zsh
- OpenAI API access

### Step 1: Clone and Build

```bash
git clone https://github.com/tomaspiaggio/termlm.git
cd termlm
```

### Step 2: Configure Environment

Create a `.env` file in the project directory:

```bash
OPENAI_KEY=your_openai_api_key_here
OPENAI_ENDPOINT=https://api.openai.com/v1/chat/completions
```

### Step 3: Build

```bash
go build
```

### Step 4: Install Binary

```bash
sudo cp ./termlm /usr/local/bin/termlm
```

### Step 5: Install zsh Plugin

```bash
mkdir -p ~/.oh-my-zsh/custom/plugins/llm-cmd
cp llm-cmd.plugin.zsh ~/.oh-my-zsh/custom/plugins/llm-cmd
```

### Step 6: Enable Plugin

Add `llm-cmd` to your plugins in `~/.zshrc`:

```bash
plugins=(... llm-cmd)
```

### Step 6: Restart Terminal

```bash
# Reload your shell configuration
source ~/.zshrc
```

## 🎯 Usage

### Interactive Mode

Run termlm directly for interactive command generation:

```bash
termlm
```

Type your natural language description and press Enter to generate a command.

**Example interactions:**

- "find all python files in this directory" → `find . -name "*.py"`
- "show disk usage of current folder" → `du -sh .`
- "kill process running on port 3000" → `lsof -ti:3000 | xargs kill`

### Shell Integration (Recommended)

With the zsh plugin installed:

1. Press `Ctrl+K` in your terminal
2. Type your command description
3. Press Enter to generate
4. Press Enter again to execute, or Esc to cancel

The generated command will appear in your command buffer, ready to run or modify.

## ⚙️ Configuration

### Environment Variables

- `OPENAI_KEY`: Your OpenAI API key (required)
- `OPENAI_ENDPOINT`: OpenAI API endpoint (required)

### Customizing Key Binding

To change the default `Ctrl+K` binding, edit `llm-cmd.plugin.zsh`:

```bash
# Change the last line from:
bindkey '^K' llm_generate_widget

# To your preferred binding, e.g., Ctrl+L:
bindkey '^L' llm_generate_widget
```

## 🔧 Development

### Dependencies

- [Bubbletea](https://github.com/charmbracelet/bubbletea): Terminal UI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss): Terminal styling

### Building from Source

```bash
git clone https://github.com/tomaspiaggio/termlm.git
cd termlm
go mod download
go build
```

### Running Tests

```bash
go test ./...
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 👨‍💻 Author

**Tom Piaggio**
- Twitter: [@tomaspiaggio](https://twitter.com/tomaspiaggio)
- GitHub: [@tomaspiaggio](https://github.com/tomaspiaggio)

## 🙏 Acknowledgments

- Built with [Charm](https://charm.sh/) terminal libraries
- Powered by OpenAI's language models
- Inspired by the need for more intuitive command-line interactions

---

**Made with ❤️ for developers who love the terminal but sometimes forget the syntax**
