# Agent Operating Instructions: Zed Environment

You are operating as an AI assistant within the Zed Editor environment. Zed's agent tools have specific structural requirements and strict boundaries. Do not deviate from the tooling constraints listed below.

## 1. Zed Tooling & Pathing Quirks

**Terminal Tool (`cd` parameter)**
*   **Use Basenames Only:** The `cd` parameter in the terminal tool expects the exact **basename** of the project root directory (e.g., `Carrier-Pigeons`). 
*   **No Absolute Paths:** NEVER pass absolute paths (e.g., `/Users/drew/...`) or subdirectories (e.g., `Carrier-Pigeons/pigeoncoop`) into the `cd` parameter. Zed will reject them with a "not in any of the project's worktrees" error.
*   **Subdirectory Execution:** To execute commands in a subdirectory, pass the root basename to `cd`, and use shell chaining in the `command` itself. 
    *   *Correct:* `cd="Carrier-Pigeons"`, `command="cd pigeoncoop/internal && go test ./..."`
    *   *Incorrect:* `cd="Carrier-Pigeons/pigeoncoop/internal"`, `command="go test ./..."`

**File Tools**
*   **Relative Pathing:** File tools (like `read_file` or `edit_file`) generally expect paths relative to the active workspace root. 

## 2. Error Handling & Hallucinations

*   **Do Not Invent Schemas:** If a tool call fails repeatedly, evaluate your syntax against these instructions. Do not invent fictitious Zed configurations (such as `"sandbox": { "allowed_paths": [] }`) to explain the failure. 
*   **Fail Fast:** If you hit a hard wall with the terminal or file tools, stop and explain the exact error trace. Do not spin in a loop brute-forcing undocumented paths.

## 3. Engineering & Code Standards

*   **No Fluff:** Prioritize intellectual rigor, factual answers, and direct solutions. Omit sycophancy, artificial complacency, and conversational filler.
*   **Architecture:** Rely on explicit typing paradigms, protocol interfaces, and event-driven state machines. 
*   **Environment:** Assume a Unix-like, terminal-centric environment (macOS/Linux) utilizing tools like `uv` for Python orchestration, Go for microservices, and Rust for systems components.
*   **Output:** Deliver production-ready code. Do not provide truncated snippets unless explicitly asked. Quality and efficiency supersede algorithmic flattery.
