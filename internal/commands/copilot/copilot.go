package copilot

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

const (
	copilotChatTool               = "copilot_chat"
	copilotServiceErrorMaxRetries = 1
	copilotServiceErrorRetryDelay = time.Second
)

// CopilotCmd groups all Copilot subcommands.
type CopilotCmd struct {
	Chat CopilotChatCmd `cmd:"" help:"Ask Copilot about your M365 content"`
}

func copilotEndpoint() string {
	return config.Endpoint("copilot")
}

type copilotServiceError struct {
	message   string
	retryable bool
}

func (e *copilotServiceError) Error() string {
	return e.message
}

// CopilotChatCmd searches internal M365 content using natural language.
type CopilotChatCmd struct {
	Message        string `arg:"" help:"Natural language question about your M365 content" optional:""`
	ConversationID string `help:"Conversation ID for follow-up queries" name:"conversation-id" optional:""`
}

func (c *CopilotChatCmd) Run(ctx *commands.Context) error {
	return runChat(ctx, c.Message, c.ConversationID)
}

func runChat(ctx *commands.Context, message, conversationID string) error {
	question := strings.TrimSpace(message)
	if question == "" {
		if !ctx.CanPrompt() {
			return fmt.Errorf("question required in non-interactive mode")
		}
		return runInteractiveLoop(ctx, os.Stdin, os.Stderr, conversationID)
	}

	data, _, err := callCopilot(ctx, question, conversationID)
	if err != nil {
		return err
	}

	return printCopilotResponse(ctx, data)
}

func runInteractiveLoop(ctx *commands.Context, input io.Reader, promptWriter io.Writer, conversationID string) error {
	reader := bufio.NewReader(input)
	currentConversationID := conversationID

	for {
		fmt.Fprint(promptWriter, "> ")

		line, err := reader.ReadString('\n')
		eof := errors.Is(err, io.EOF)
		if err != nil && !eof {
			return fmt.Errorf("read question: %w", err)
		}
		if eof && line == "" {
			return nil
		}

		question := strings.TrimSpace(line)
		if question == "" {
			if eof {
				return nil
			}
			continue
		}
		if isExitCommand(question) {
			return nil
		}

		data, nextConversationID, askErr := callCopilot(ctx, question, currentConversationID)
		if askErr != nil {
			fmt.Fprintf(promptWriter, "Error: %v\n", askErr)
			if eof {
				return askErr
			}
			continue
		}

		if err := printCopilotResponse(ctx, data); err != nil {
			return err
		}

		if nextConversationID != "" {
			currentConversationID = nextConversationID
		}

		if eof {
			return nil
		}
	}
}

func callCopilot(ctx *commands.Context, message, conversationID string) (map[string]any, string, error) {
	stopSpinner := startCopilotSpinner(ctx)
	defer stopSpinner()

	client := ctx.NewMCPClient(copilotEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return nil, "", fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{
		"message": message,
	}
	if conversationID != "" {
		args["conversationId"] = conversationID
	}

	for attempt := 0; ; attempt++ {
		resp, err := client.CallTool(ctx.Ctx, copilotChatTool, args)
		if err != nil {
			return nil, "", fmt.Errorf("copilot chat: %w", err)
		}

		data, err := output.ExtractContent(resp)
		if err != nil {
			return nil, "", err
		}

		if svcErr := copilotServiceErrorFromData(data); svcErr != nil {
			if svcErr.retryable && attempt < copilotServiceErrorMaxRetries {
				if ctx.Verbose {
					fmt.Fprintf(os.Stderr, "--- Copilot returned a retryable service error; retrying (attempt %d/%d) after %v\n%s\n", attempt+1, copilotServiceErrorMaxRetries, copilotServiceErrorRetryDelay, svcErr.Error())
				}

				select {
				case <-ctx.Ctx.Done():
					return nil, "", ctx.Ctx.Err()
				case <-time.After(copilotServiceErrorRetryDelay):
				}

				continue
			}

			return nil, "", fmt.Errorf("copilot chat: %w", svcErr)
		}

		nextConversationID := findConversationID(data)
		if ctx.Output.Format != output.FormatJSON {
			data = normalizeCopilotResponse(data, nextConversationID)
		}

		return data, nextConversationID, nil
	}
}

func printCopilotResponse(ctx *commands.Context, data map[string]any) error {
	if ctx.Output.Format == output.FormatJSON {
		return ctx.Output.PrintItem(data)
	}

	isConversation := isConversationPayload(data)
	messageKey, message := extractPrimaryText(data)
	if message == "" {
		return ctx.Output.PrintItem(data)
	}

	if ctx.NoInput {
		fmt.Fprintln(ctx.Output.Writer, message)
	} else {
		fmt.Fprintln(ctx.Output.Writer, "Copilot:", message)
	}

	meta := cloneMap(data)
	if messageKey != "" {
		delete(meta, messageKey)
	}
	delete(meta, "conversationId")
	delete(meta, "conversationID")
	delete(meta, "conversation_id")
	delete(meta, "@odata.context")
	delete(meta, "createdDateTime")
	delete(meta, "displayName")
	delete(meta, "messages")
	delete(meta, "state")
	delete(meta, "turnCount")
	if isConversation {
		delete(meta, "id")
	}
	if len(meta) == 0 {
		return nil
	}

	fmt.Fprintln(ctx.Output.Writer)
	output.RenderKeyValue(ctx.Output.Writer, meta)
	return nil
}

func extractPrimaryText(data map[string]any) (string, string) {
	for _, key := range []string{"message", "response", "answer", "content", "text"} {
		value, ok := data[key].(string)
		if ok && strings.TrimSpace(value) != "" {
			return key, value
		}
	}
	if text := extractConversationMessage(data); strings.TrimSpace(text) != "" {
		return "", text
	}
	return "", ""
}

func findConversationID(data map[string]any) string {
	if isConversationPayload(data) {
		if id, ok := data["id"].(string); ok && strings.TrimSpace(id) != "" {
			return id
		}
	}
	return findStringValue(data, "conversationId", "conversationID", "conversation_id")
}

func findStringValue(v any, keys ...string) string {
	switch value := v.(type) {
	case map[string]any:
		for _, key := range keys {
			if s, ok := value[key].(string); ok && strings.TrimSpace(s) != "" {
				return s
			}
		}
		for _, nested := range value {
			if s := findStringValue(nested, keys...); s != "" {
				return s
			}
		}
	case []any:
		for _, item := range value {
			if s := findStringValue(item, keys...); s != "" {
				return s
			}
		}
	}
	return ""
}

func cloneMap(data map[string]any) map[string]any {
	cloned := make(map[string]any, len(data))
	for k, v := range data {
		cloned[k] = v
	}
	return cloned
}

func copilotServiceErrorFromData(data map[string]any) *copilotServiceError {
	_, message := extractPrimaryText(data)
	message = sanitizeCopilotServiceMessage(message)
	if message == "" {
		return nil
	}

	lower := strings.ToLower(message)
	if !strings.HasPrefix(lower, "error executing tool:") {
		return nil
	}

	return &copilotServiceError{
		message:   message,
		retryable: isRetryableCopilotServiceError(message),
	}
}

func isRetryableCopilotServiceError(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "timed out") || strings.Contains(lower, "timed-out") || strings.Contains(lower, "timeout")
}

func sanitizeCopilotServiceMessage(message string) string {
	message = strings.ReplaceAll(message, "\r\n", "\n")
	message = strings.ReplaceAll(message, "\r", "\n")

	seen := map[string]struct{}{}
	lines := strings.Split(message, "\n")
	cleaned := make([]string, 0, len(lines))

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if i == 0 {
			line = strings.TrimSpace(strings.TrimPrefix(line, "Error:"))
		}
		if line == "" {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		cleaned = append(cleaned, line)
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

func normalizeCopilotResponse(data map[string]any, conversationID string) map[string]any {
	message := extractConversationMessage(data)
	if message == "" {
		return data
	}

	normalized := map[string]any{
		"message": message,
	}
	if conversationID != "" {
		normalized["conversationId"] = conversationID
	}
	return normalized
}

func extractConversationMessage(data map[string]any) string {
	rawMessages, ok := data["messages"].([]any)
	if !ok || len(rawMessages) == 0 {
		return ""
	}

	last, ok := rawMessages[len(rawMessages)-1].(map[string]any)
	if !ok {
		return ""
	}
	text, _ := last["text"].(string)
	return text
}

func isConversationPayload(data map[string]any) bool {
	_, hasMessages := data["messages"]
	_, hasID := data["id"]
	return hasMessages && hasID
}

func isExitCommand(question string) bool {
	switch strings.ToLower(strings.TrimSpace(question)) {
	case "exit", "quit", ":q", "/exit", "/quit":
		return true
	default:
		return false
	}
}

func startCopilotSpinner(ctx *commands.Context) func() {
	if ctx.Verbose || !stderrIsTerminal() {
		return func() {}
	}

	const (
		label      = "Thinking..."
		startDelay = 200 * time.Millisecond
		frameDelay = 120 * time.Millisecond
	)

	stop := make(chan struct{})
	done := make(chan struct{})
	var once sync.Once

	go func() {
		defer close(done)

		timer := time.NewTimer(startDelay)
		defer timer.Stop()

		select {
		case <-stop:
			return
		case <-timer.C:
		}

		frames := []string{"|", "/", "-", "\\"}
		ticker := time.NewTicker(frameDelay)
		defer ticker.Stop()

		render := func(frame string) {
			fmt.Fprintf(os.Stderr, "\r%s %s", frame, label)
		}

		frameIndex := 0
		render(frames[frameIndex])

		for {
			select {
			case <-stop:
				fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", len(label)+2))
				return
			case <-ticker.C:
				frameIndex = (frameIndex + 1) % len(frames)
				render(frames[frameIndex])
			}
		}
	}()

	return func() {
		once.Do(func() {
			close(stop)
			<-done
		})
	}
}

func stderrIsTerminal() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
