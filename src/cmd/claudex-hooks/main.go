package main

import (
	"fmt"
	"os"
	"strconv"

	"claudex/internal/doc"
	"claudex/internal/hooks/notification"
	"claudex/internal/hooks/posttooluse"
	"claudex/internal/hooks/pretooluse"
	"claudex/internal/hooks/sessionend"
	"claudex/internal/hooks/shared"
	"claudex/internal/hooks/subagent"
	"claudex/internal/notify"
	"claudex/internal/services/commander"
	"claudex/internal/services/env"

	"github.com/spf13/afero"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: claudex-hooks <command>\n")
		fmt.Fprintf(os.Stderr, "Commands: notification, pre-tool-use, post-tool-use, auto-doc, doc-update, session-end, subagent-stop\n")
		os.Exit(1)
	}

	cmd := os.Args[1]

	// Create shared dependencies
	fs := afero.NewOsFs()
	environ := env.New()
	cmdr := commander.New()

	// Create logger (hook name will be the command)
	logger := shared.NewLogger(fs, environ, cmd)

	// Create parser (reads from stdin)
	parser := shared.NewParser(os.Stdin)

	// Create builder (writes to stdout)
	builder := shared.NewBuilder(os.Stdout)

	var err error

	switch cmd {
	case "notification":
		err = handleNotification(fs, cmdr, environ, logger, parser)
	case "pre-tool-use":
		err = handlePreToolUse(fs, environ, logger, parser, builder)
	case "post-tool-use":
		err = handlePostToolUse(logger, parser, builder)
	case "auto-doc":
		err = handleAutoDoc(fs, cmdr, environ, logger, parser, builder)
	case "session-end":
		err = handleSessionEnd(fs, cmdr, environ, logger, parser, builder)
	case "subagent-stop":
		err = handleSubagentStop(fs, cmdr, environ, logger, parser, builder)
	case "doc-update":
		err = handleDocUpdate(fs, cmdr, environ, logger, parser)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		os.Exit(1)
	}

	if err != nil {
		_ = logger.LogError(fmt.Errorf("%s handler error: %w", cmd, err))
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// handleNotification processes notification hook events
func handleNotification(fs afero.Fs, cmdr commander.Commander, environ env.Environment, logger *shared.Logger, parser *shared.Parser) error {
	input, err := parser.ParseNotification()
	if err != nil {
		return err
	}

	// Create notifier with dependencies
	notifCfg := notify.DefaultConfig()
	notifCfg.NotificationsEnabled = environ.Get("CLAUDEX_NOTIFICATIONS_ENABLED") != "false"
	notifCfg.VoiceEnabled = environ.Get("CLAUDEX_VOICE_ENABLED") == "true" || environ.Get("CLAUDEX_VOICE_ENABLED") == "1"

	deps := &commanderAdapter{cmdr: cmdr}
	notifier := notify.New(notifCfg, deps)

	handler := notification.NewHandler(notifier, logger, environ)
	return handler.Handle(input)
}

// handlePreToolUse processes pre-tool-use hook events
func handlePreToolUse(fs afero.Fs, environ env.Environment, logger *shared.Logger, parser *shared.Parser, builder *shared.Builder) error {
	input, err := parser.ParsePreToolUse()
	if err != nil {
		return err
	}

	handler := pretooluse.NewHandler(fs, environ, logger)
	output, err := handler.Handle(input)
	if err != nil {
		return err
	}

	return builder.BuildCustom(*output)
}

// handlePostToolUse processes post-tool-use hook events
func handlePostToolUse(logger *shared.Logger, parser *shared.Parser, builder *shared.Builder) error {
	input, err := parser.ParsePostToolUse()
	if err != nil {
		return err
	}

	handler := posttooluse.NewHandler(logger)
	output, err := handler.Handle(input)
	if err != nil {
		return err
	}

	return builder.BuildCustom(*output)
}

// handleAutoDoc processes auto-doc hook events
func handleAutoDoc(fs afero.Fs, cmdr commander.Commander, environ env.Environment, logger *shared.Logger, parser *shared.Parser, builder *shared.Builder) error {
	input, err := parser.ParsePostToolUse()
	if err != nil {
		return err
	}

	// Create documentation updater
	updater := doc.NewUpdater(fs, cmdr, environ)

	// Read frequency from environment (default 5)
	frequency := 5
	if freqStr := environ.Get("CLAUDEX_AUTODOC_FREQUENCY"); freqStr != "" {
		if freq, err := strconv.Atoi(freqStr); err == nil && freq > 0 {
			frequency = freq
		}
	}

	handler := posttooluse.NewAutoDocHandler(fs, environ, updater, logger, frequency)
	output, err := handler.Handle(input)
	if err != nil {
		return err
	}

	return builder.BuildCustom(*output)
}

// handleSessionEnd processes session-end hook events
func handleSessionEnd(fs afero.Fs, cmdr commander.Commander, environ env.Environment, logger *shared.Logger, parser *shared.Parser, builder *shared.Builder) error {
	input, err := parser.ParseSessionEnd()
	if err != nil {
		return err
	}

	// Create documentation updater
	updater := doc.NewUpdater(fs, cmdr, environ)

	handler := sessionend.NewHandler(fs, environ, updater, logger)
	return handler.Handle(input)
}

// handleSubagentStop processes subagent-stop hook events
func handleSubagentStop(fs afero.Fs, cmdr commander.Commander, environ env.Environment, logger *shared.Logger, parser *shared.Parser, builder *shared.Builder) error {
	input, err := parser.ParseSubagentStop()
	if err != nil {
		return err
	}

	// Create notifier with dependencies
	notifCfg := notify.DefaultConfig()
	notifCfg.NotificationsEnabled = environ.Get("CLAUDEX_NOTIFICATIONS_ENABLED") != "false"

	deps := &commanderAdapter{cmdr: cmdr}
	notifier := notify.New(notifCfg, deps)

	// Create documentation updater
	updater := doc.NewUpdater(fs, cmdr, environ)

	handler := subagent.NewHandler(fs, environ, updater, notifier, logger)
	output, err := handler.Handle(input)
	if err != nil {
		return err
	}

	return builder.BuildCustom(*output)
}

// handleDocUpdate processes doc-update commands (detached subprocess for background updates)
func handleDocUpdate(fs afero.Fs, cmdr commander.Commander, environ env.Environment, logger *shared.Logger, parser *shared.Parser) error {
	input, err := parser.ParseDocUpdate()
	if err != nil {
		return err
	}

	_ = logger.LogInfo(fmt.Sprintf("Starting doc update for session: %s", input.SessionPath))

	// Create documentation updater
	updater := doc.NewUpdater(fs, cmdr, environ)

	// Convert input to UpdaterConfig
	config := doc.UpdaterConfig{
		SessionPath:    input.SessionPath,
		TranscriptPath: input.TranscriptPath,
		OutputFile:     input.OutputFile,
		PromptTemplate: input.PromptTemplate,
		SessionContext: input.SessionContext,
		Model:          input.Model,
		StartLine:      input.StartLine,
	}

	// Run synchronously - this process is detached and can take its time
	if err := updater.Run(config); err != nil {
		_ = logger.LogError(fmt.Errorf("doc update failed: %w", err))
		return err
	}

	_ = logger.LogInfo("Doc update completed successfully")
	return nil
}

// commanderAdapter adapts commander.Commander to notify.Dependencies
type commanderAdapter struct {
	cmdr commander.Commander
}

func (c *commanderAdapter) Commander() notify.Commander {
	return c.cmdr
}
