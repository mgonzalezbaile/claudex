package ui

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockInputReader is a test implementation of InputReader that returns preconfigured
// values. It allows testing PromptDescriptionWithReader without requiring actual
// terminal interaction.
type MockInputReader struct {
	// Input is the string returned by Readline.
	Input string
	// Err is the error returned by Readline. If non-nil, Input is ignored.
	Err error
}

// Readline returns the preconfigured Input and Err values.
// This allows tests to simulate various input scenarios including errors.
func (m *MockInputReader) Readline() (string, error) {
	return m.Input, m.Err
}

// Close is a no-op for the mock reader. Always returns nil.
func (m *MockInputReader) Close() error {
	return nil
}

// TestPromptDescriptionWithReader verifies PromptDescriptionWithReader behavior using table-driven tests.
func TestPromptDescriptionWithReader(t *testing.T) {
	tests := []struct {
		name            string
		title           string
		originalSession string
		mockInput       string
		mockErr         error
		expectedResult  string
		expectedError   string
	}{
		{
			name:            "successful input returns trimmed description",
			title:           "Create New Session",
			originalSession: "",
			mockInput:       "  test session  ",
			mockErr:         nil,
			expectedResult:  "test session",
			expectedError:   "",
		},
		{
			name:            "empty input returns error",
			title:           "Create New Session",
			originalSession: "",
			mockInput:       "",
			mockErr:         nil,
			expectedResult:  "",
			expectedError:   "description cannot be empty",
		},
		{
			name:            "reader error propagates correctly",
			title:           "Create New Session",
			originalSession: "",
			mockInput:       "",
			mockErr:         errors.New("read error"),
			expectedResult:  "",
			expectedError:   "read error",
		},
		{
			name:            "whitespace-only input returns empty error",
			title:           "Create New Session",
			originalSession: "",
			mockInput:       "   ",
			mockErr:         nil,
			expectedResult:  "",
			expectedError:   "description cannot be empty",
		},
		{
			name:            "fork context with valid input",
			title:           "Fork Session",
			originalSession: "original-session",
			mockInput:       "forked session description",
			mockErr:         nil,
			expectedResult:  "forked session description",
			expectedError:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := &MockInputReader{
				Input: tt.mockInput,
				Err:   tt.mockErr,
			}

			result, err := PromptDescriptionWithReader(tt.title, tt.originalSession, mockReader)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

// TestPromptDescriptionWithReader_NilReader verifies that passing a nil reader returns an error
// instead of panicking.
func TestPromptDescriptionWithReader_NilReader(t *testing.T) {
	result, err := PromptDescriptionWithReader("Test", "", nil)

	assert.Error(t, err)
	assert.Equal(t, "reader cannot be nil", err.Error())
	assert.Equal(t, "", result)
}
