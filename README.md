# waygent

A Wayland desktop automation agent powered by visual LLMs.

waygent is a command-line tool that lets a vision-capable LLM control your GNOME Wayland desktop. It captures screenshots, sends them to the model along with your task description, and executes the resulting actions on your behalf.

## How It Works

waygent runs a simple loop:

1. **Capture** a screenshot of the current desktop
2. **Send** the image and task description to an OpenAI-compatible vision API
3. **Parse** the LLM's structured JSON response containing desktop actions
4. **Execute** those actions via ydotool (mouse moves, clicks, typing, key presses)
5. **Repeat** until the task is complete or the step limit is reached

## Installation

### Nix

Build directly from the flake:

```bash
nix build github:user/waygent
```

Enter the development shell:

```bash
nix develop github:user/waygent
```

### NixOS Module

Add waygent to your flake inputs and enable the module:

```nix
{
  inputs.waygent.url = "github:user/waygent";

  outputs = { nixpkgs, waygent, ... }: {
    nixosConfigurations.myhost = nixpkgs.lib.nixosSystem {
      modules = [
        waygent.nixosModules.default
        {
          services.waygent.enable = true;
        }
      ];
    };
  };
}
```

Enabling `services.waygent.enable` installs waygent and starts the ydotool daemon automatically.

### From Source

Requirements: Go 1.22+

```bash
git clone https://github.com/user/waygent.git
cd waygent
go build -o waygent ./cmd/waygent
```

## Usage

Run waygent with a task description:

```bash
waygent -task "Open Firefox and search for weather"
```

Use a different model or API endpoint:

```bash
waygent \
  -task "Open the terminal and run htop" \
  -api-url "https://api.openai.com/v1" \
  -api-key "$OPENAI_API_KEY" \
  -model "gpt-4o"
```

Enable verbose logging to see every action the agent takes:

```bash
waygent -task "Open Settings and enable dark mode" -verbose
```

## Configuration

waygent can be configured via CLI flags or environment variables. Environment variables take precedence over defaults, and CLI flags take precedence over environment variables.

| Flag | Environment Variable | Description | Default |
|------|----------------------|-------------|---------|
| `-task` | `WAYGENT_TASK` | Task description for the agent | *(required)* |
| `-api-key` | `WAYGENT_API_KEY`, `OPENAI_API_KEY` | API key for the LLM endpoint | *(required)* |
| `-api-url` | `WAYGENT_API_URL` | Base URL for the OpenAI-compatible API | `https://api.openai.com/v1` |
| `-model` | `WAYGENT_MODEL` | Model name to use | `gpt-4o` |
| `-max-steps` | `WAYGENT_MAX_STEPS` | Maximum agent loop iterations | `50` |
| `-verbose` | `WAYGENT_VERBOSE` | Enable verbose logging | `false` |

## Action Reference

The LLM returns a JSON response with `{"thinking": "...", "actions": [...]}`. waygent executes the following action types:

| Action | Parameters | Description |
|--------|------------|-------------|
| `mouse_move` | `x`, `y` | Move the cursor to the given screen coordinates |
| `click` | `button` (`left`, `right`, `middle`) | Click the specified mouse button |
| `double_click` | `button` (`left`, `right`, `middle`) | Double-click the specified mouse button |
| `type_text` | `text` | Type the given string |
| `key_press` | `keys` (array, e.g. `["ctrl", "a"]`) | Press a key combination |
| `scroll` | `direction` (`up`, `down`) | Scroll the mouse wheel |
| `wait` | `duration_ms` | Pause for the given number of milliseconds |
| `screenshot` | none | Take a fresh screenshot and continue the loop |
| `done` | none | Signal that the task is complete |

## Requirements

- A GNOME Wayland desktop, or any wlroots-based compositor with `grim` available
- `ydotool` daemon running. Start it manually with:
  ```bash
  ydotool --socket-path=/tmp/.ydotool_socket
  ```
  The NixOS module handles this automatically.
- An OpenAI-compatible API endpoint with a vision-capable model (e.g. `gpt-4o`)

## License

MIT
