package notify

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCommander provides a mock implementation of Commander for testing
type MockCommander struct {
	RunFunc func(name string, args ...string) ([]byte, error)
	calls   []CommandCall
}

type CommandCall struct {
	Name string
	Args []string
}

func NewMockCommander() *MockCommander {
	return &MockCommander{
		calls: make([]CommandCall, 0),
	}
}

func (m *MockCommander) Run(name string, args ...string) ([]byte, error) {
	m.calls = append(m.calls, CommandCall{Name: name, Args: args})
	if m.RunFunc != nil {
		return m.RunFunc(name, args...)
	}
	return []byte{}, nil
}

func (m *MockCommander) GetCalls() []CommandCall {
	return m.calls
}

func (m *MockCommander) GetLastCall() *CommandCall {
	if len(m.calls) == 0 {
		return nil
	}
	return &m.calls[len(m.calls)-1]
}

// MockDependencies provides mock dependencies for testing
type MockDependencies struct {
	commander Commander
}

func NewMockDependencies(commander Commander) *MockDependencies {
	return &MockDependencies{
		commander: commander,
	}
}

func (d *MockDependencies) Commander() Commander {
	return d.commander
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.True(t, cfg.NotificationsEnabled, "notifications should be enabled by default")
	assert.False(t, cfg.VoiceEnabled, "voice should be disabled by default")
	assert.Equal(t, "default", cfg.DefaultSound)
	assert.Equal(t, "Samantha", cfg.DefaultVoice)
}

func TestNew_MacOS(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-macOS platform")
	}

	cfg := DefaultConfig()
	mockCommander := NewMockCommander()
	deps := NewMockDependencies(mockCommander)

	notifier := New(cfg, deps)

	assert.NotNil(t, notifier)
	assert.IsType(t, &macOSNotifier{}, notifier)
	assert.True(t, notifier.IsAvailable())
}

func TestNew_NonMacOS(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("Skipping non-macOS test on macOS platform")
	}

	cfg := DefaultConfig()
	mockCommander := NewMockCommander()
	deps := NewMockDependencies(mockCommander)

	notifier := New(cfg, deps)

	assert.NotNil(t, notifier)
	assert.IsType(t, &noopNotifier{}, notifier)
	assert.False(t, notifier.IsAvailable())
}

func TestGetNotificationConfig(t *testing.T) {
	tests := []struct {
		name             string
		notificationType string
		expectedTitle    string
		expectedSound    string
	}{
		{
			name:             "permission_prompt",
			notificationType: "permission_prompt",
			expectedTitle:    "Permission Required",
			expectedSound:    "Blow",
		},
		{
			name:             "idle_timeout",
			notificationType: "idle_timeout",
			expectedTitle:    "Claudex Idle",
			expectedSound:    "Ping",
		},
		{
			name:             "agent_complete",
			notificationType: "agent_complete",
			expectedTitle:    "Agent Complete",
			expectedSound:    "Glass",
		},
		{
			name:             "session_end",
			notificationType: "session_end",
			expectedTitle:    "Session Ended",
			expectedSound:    "Tink",
		},
		{
			name:             "error",
			notificationType: "error",
			expectedTitle:    "Claudex Error",
			expectedSound:    "Basso",
		},
		{
			name:             "unknown_type",
			notificationType: "unknown",
			expectedTitle:    "Claudex",
			expectedSound:    "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GetNotificationConfig(tt.notificationType)
			assert.Equal(t, tt.expectedTitle, cfg.Title)
			assert.Equal(t, tt.expectedSound, cfg.Sound)
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "message",
		Message: "cannot be empty",
	}

	expected := "notification validation error: message - cannot be empty"
	assert.Equal(t, expected, err.Error())
}

func TestEscapeAppleScript(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no_special_chars",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "quotes",
			input:    `He said "Hello"`,
			expected: `He said \"Hello\"`,
		},
		{
			name:     "backslashes",
			input:    `C:\Users\test`,
			expected: `C:\\Users\\test`,
		},
		{
			name:     "quotes_and_backslashes",
			input:    `Path: "C:\Users\test"`,
			expected: `Path: \"C:\\Users\\test\"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeAppleScript(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMacOSNotifier_Send(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		title         string
		message       string
		sound         string
		mockFunc      func(name string, args ...string) ([]byte, error)
		expectError   bool
		errorContains string
		validateCall  func(t *testing.T, call *CommandCall)
	}{
		{
			name: "successful_notification",
			config: Config{
				NotificationsEnabled: true,
				DefaultSound:         "default",
			},
			title:   "Test Title",
			message: "Test Message",
			sound:   "Blow",
			mockFunc: func(name string, args ...string) ([]byte, error) {
				return []byte{}, nil
			},
			expectError: false,
			validateCall: func(t *testing.T, call *CommandCall) {
				require.NotNil(t, call)
				assert.Equal(t, "osascript", call.Name)
				assert.Len(t, call.Args, 2)
				assert.Equal(t, "-e", call.Args[0])
				assert.Contains(t, call.Args[1], "Test Message")
				assert.Contains(t, call.Args[1], "Test Title")
				assert.Contains(t, call.Args[1], "Blow")
			},
		},
		{
			name: "use_default_sound",
			config: Config{
				NotificationsEnabled: true,
				DefaultSound:         "Glass",
			},
			title:   "Title",
			message: "Message",
			sound:   "", // Empty sound should use default
			mockFunc: func(name string, args ...string) ([]byte, error) {
				return []byte{}, nil
			},
			expectError: false,
			validateCall: func(t *testing.T, call *CommandCall) {
				require.NotNil(t, call)
				assert.Contains(t, call.Args[1], "Glass")
			},
		},
		{
			name: "notifications_disabled",
			config: Config{
				NotificationsEnabled: false,
			},
			title:       "Title",
			message:     "Message",
			sound:       "Blow",
			mockFunc:    nil, // Should not be called
			expectError: false,
			validateCall: func(t *testing.T, call *CommandCall) {
				assert.Nil(t, call, "should not call osascript when disabled")
			},
		},
		{
			name: "empty_message",
			config: Config{
				NotificationsEnabled: true,
			},
			title:         "Title",
			message:       "",
			sound:         "Blow",
			expectError:   true,
			errorContains: "message cannot be empty",
		},
		{
			name: "osascript_not_found",
			config: Config{
				NotificationsEnabled: true,
			},
			title:   "Title",
			message: "Message",
			sound:   "Blow",
			mockFunc: func(name string, args ...string) ([]byte, error) {
				return nil, errors.New("executable file not found")
			},
			expectError: false, // Should silently ignore missing osascript
		},
		{
			name: "osascript_error",
			config: Config{
				NotificationsEnabled: true,
			},
			title:   "Title",
			message: "Message",
			sound:   "Blow",
			mockFunc: func(name string, args ...string) ([]byte, error) {
				return []byte("error output"), errors.New("command failed")
			},
			expectError:   true,
			errorContains: "osascript failed",
		},
		{
			name: "escape_special_characters",
			config: Config{
				NotificationsEnabled: true,
				DefaultSound:         "default",
			},
			title:   `Title with "quotes"`,
			message: `Message with "quotes" and \backslashes`,
			sound:   "Blow",
			mockFunc: func(name string, args ...string) ([]byte, error) {
				return []byte{}, nil
			},
			expectError: false,
			validateCall: func(t *testing.T, call *CommandCall) {
				require.NotNil(t, call)
				script := call.Args[1]
				// Should contain escaped quotes and backslashes
				assert.Contains(t, script, `\"quotes\"`)
				assert.Contains(t, script, `\\backslashes`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			mockCommander.RunFunc = tt.mockFunc
			deps := NewMockDependencies(mockCommander)

			notifier := &macOSNotifier{
				config: tt.config,
				deps:   deps,
			}

			err := notifier.Send(tt.title, tt.message, tt.sound)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}

			if tt.validateCall != nil {
				tt.validateCall(t, mockCommander.GetLastCall())
			}
		})
	}
}

func TestMacOSNotifier_Speak(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		message       string
		mockFunc      func(name string, args ...string) ([]byte, error)
		expectError   bool
		errorContains string
		validateCall  func(t *testing.T, call *CommandCall)
	}{
		{
			name: "successful_speech",
			config: Config{
				VoiceEnabled: true,
				DefaultVoice: "Samantha",
			},
			message: "Hello World",
			mockFunc: func(name string, args ...string) ([]byte, error) {
				return []byte{}, nil
			},
			expectError: false,
			validateCall: func(t *testing.T, call *CommandCall) {
				require.NotNil(t, call)
				assert.Equal(t, "say", call.Name)
				assert.Len(t, call.Args, 3)
				assert.Equal(t, "-v", call.Args[0])
				assert.Equal(t, "Samantha", call.Args[1])
				assert.Equal(t, "Hello World", call.Args[2])
			},
		},
		{
			name: "voice_disabled",
			config: Config{
				VoiceEnabled: false,
			},
			message:     "Hello World",
			mockFunc:    nil, // Should not be called
			expectError: false,
			validateCall: func(t *testing.T, call *CommandCall) {
				assert.Nil(t, call, "should not call say when disabled")
			},
		},
		{
			name: "empty_message",
			config: Config{
				VoiceEnabled: true,
			},
			message:       "",
			expectError:   true,
			errorContains: "message cannot be empty",
		},
		{
			name: "say_not_found",
			config: Config{
				VoiceEnabled: true,
				DefaultVoice: "Samantha",
			},
			message: "Hello",
			mockFunc: func(name string, args ...string) ([]byte, error) {
				return nil, errors.New("executable file not found")
			},
			expectError: false, // Should silently ignore missing say
		},
		{
			name: "say_error",
			config: Config{
				VoiceEnabled: true,
				DefaultVoice: "Samantha",
			},
			message: "Hello",
			mockFunc: func(name string, args ...string) ([]byte, error) {
				return []byte("error"), errors.New("command failed")
			},
			expectError:   true,
			errorContains: "say command failed",
		},
		{
			name: "custom_voice",
			config: Config{
				VoiceEnabled: true,
				DefaultVoice: "Alex",
			},
			message: "Testing",
			mockFunc: func(name string, args ...string) ([]byte, error) {
				return []byte{}, nil
			},
			expectError: false,
			validateCall: func(t *testing.T, call *CommandCall) {
				require.NotNil(t, call)
				assert.Equal(t, "Alex", call.Args[1])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			mockCommander.RunFunc = tt.mockFunc
			deps := NewMockDependencies(mockCommander)

			notifier := &macOSNotifier{
				config: tt.config,
				deps:   deps,
			}

			err := notifier.Speak(tt.message)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}

			if tt.validateCall != nil {
				tt.validateCall(t, mockCommander.GetLastCall())
			}
		})
	}
}

func TestMacOSNotifier_IsAvailable(t *testing.T) {
	mockCommander := NewMockCommander()
	deps := NewMockDependencies(mockCommander)

	notifier := &macOSNotifier{
		config: DefaultConfig(),
		deps:   deps,
	}

	assert.True(t, notifier.IsAvailable())
}

func TestNoopNotifier(t *testing.T) {
	notifier := NewNoop()

	t.Run("Send", func(t *testing.T) {
		err := notifier.Send("Title", "Message", "Sound")
		assert.NoError(t, err)
	})

	t.Run("Speak", func(t *testing.T) {
		err := notifier.Speak("Message")
		assert.NoError(t, err)
	})

	t.Run("IsAvailable", func(t *testing.T) {
		assert.False(t, notifier.IsAvailable())
	})
}

func TestIntegration_NotificationWorkflow(t *testing.T) {
	// This test simulates a complete notification workflow
	mockCommander := NewMockCommander()
	var capturedCommands []string

	mockCommander.RunFunc = func(name string, args ...string) ([]byte, error) {
		capturedCommands = append(capturedCommands, fmt.Sprintf("%s %s", name, strings.Join(args, " ")))
		return []byte{}, nil
	}

	deps := NewMockDependencies(mockCommander)
	cfg := Config{
		NotificationsEnabled: true,
		VoiceEnabled:         true,
		DefaultSound:         "default",
		DefaultVoice:         "Samantha",
	}

	notifier := &macOSNotifier{
		config: cfg,
		deps:   deps,
	}

	// Simulate permission prompt notification
	notifCfg := GetNotificationConfig("permission_prompt")
	err := notifier.Send(notifCfg.Title, "Claude needs permission to use Bash", notifCfg.Sound)
	require.NoError(t, err)

	// Optionally speak the message
	err = notifier.Speak("Permission required")
	require.NoError(t, err)

	// Verify both commands were executed
	assert.Len(t, capturedCommands, 2)
	assert.Contains(t, capturedCommands[0], "osascript")
	assert.Contains(t, capturedCommands[1], "say")
}
