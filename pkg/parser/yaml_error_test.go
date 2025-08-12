package parser

import (
	"errors"
	"testing"
)

func TestExtractYAMLError(t *testing.T) {
	tests := []struct {
		name                 string
		err                  error
		frontmatterStartLine int
		expectedLine         int
		expectedColumn       int
		expectedMessage      string
	}{
		{
			name:                 "yaml line error",
			err:                  errors.New("yaml: line 7: mapping values are not allowed in this context"),
			frontmatterStartLine: 1,
			expectedLine:         8, // 7 + 1
			expectedColumn:       1,
			expectedMessage:      "mapping values are not allowed in this context",
		},
		{
			name:                 "yaml line error with frontmatter offset",
			err:                  errors.New("yaml: line 3: found character that cannot start any token"),
			frontmatterStartLine: 5,
			expectedLine:         8, // 3 + 5
			expectedColumn:       1,
			expectedMessage:      "found character that cannot start any token",
		},
		{
			name:                 "non-yaml error",
			err:                  errors.New("some other error"),
			frontmatterStartLine: 1,
			expectedLine:         0,
			expectedColumn:       0,
			expectedMessage:      "some other error",
		},
		{
			name:                 "yaml error with different message format",
			err:                  errors.New("yaml: line 15: found unexpected end of stream"),
			frontmatterStartLine: 2,
			expectedLine:         17, // 15 + 2
			expectedColumn:       1,
			expectedMessage:      "found unexpected end of stream",
		},
		{
			name:                 "yaml error with indentation issue",
			err:                  errors.New("yaml: line 4: bad indentation of a mapping entry"),
			frontmatterStartLine: 1,
			expectedLine:         5, // 4 + 1
			expectedColumn:       1,
			expectedMessage:      "bad indentation of a mapping entry",
		},
		{
			name:                 "yaml error with duplicate key",
			err:                  errors.New("yaml: line 6: found duplicate key"),
			frontmatterStartLine: 3,
			expectedLine:         9, // 6 + 3
			expectedColumn:       1,
			expectedMessage:      "found duplicate key",
		},
		{
			name:                 "yaml error with complex format",
			err:                  errors.New("yaml: line 12: did not find expected ',' or ']'"),
			frontmatterStartLine: 0,
			expectedLine:         12, // 12 + 0
			expectedColumn:       1,
			expectedMessage:      "did not find expected ',' or ']'",
		},
		{
			name:                 "yaml unmarshal error multiline",
			err:                  errors.New("yaml: unmarshal errors:\n  line 4: mapping key \"permissions\" already defined at line 2"),
			frontmatterStartLine: 1,
			expectedLine:         5, // 4 + 1
			expectedColumn:       1,
			expectedMessage:      "mapping key \"permissions\" already defined at line 2",
		},
		{
			name:                 "yaml error with flow mapping",
			err:                  errors.New("yaml: line 8: did not find expected ',' or '}'"),
			frontmatterStartLine: 1,
			expectedLine:         9, // 8 + 1
			expectedColumn:       1,
			expectedMessage:      "did not find expected ',' or '}'",
		},
		{
			name:                 "yaml error with invalid character",
			err:                  errors.New("yaml: line 5: found character that cannot start any token"),
			frontmatterStartLine: 0,
			expectedLine:         5, // 5 + 0
			expectedColumn:       1,
			expectedMessage:      "found character that cannot start any token",
		},
		{
			name:                 "yaml error with unmarshal type issue",
			err:                  errors.New("yaml: line 3: cannot unmarshal !!str `yes_please` into bool"),
			frontmatterStartLine: 2,
			expectedLine:         5, // 3 + 2
			expectedColumn:       1,
			expectedMessage:      "cannot unmarshal !!str `yes_please` into bool",
		},
		{
			name:                 "yaml complex unmarshal error with nested line info",
			err:                  errors.New("yaml: unmarshal errors:\n  line 7: found unexpected end of stream\n  line 9: mapping values are not allowed in this context"),
			frontmatterStartLine: 1,
			expectedLine:         8, // First line 7 + 1
			expectedColumn:       1,
			expectedMessage:      "found unexpected end of stream",
		},
		{
			name:                 "yaml error with column information greater than 1",
			err:                  errors.New("yaml: line 5: column 12: invalid character at position"),
			frontmatterStartLine: 1,
			expectedLine:         6, // 5 + 1
			expectedColumn:       12,
			expectedMessage:      "invalid character at position",
		},
		{
			name:                 "yaml error with high column number",
			err:                  errors.New("yaml: line 3: column 45: unexpected token found"),
			frontmatterStartLine: 2,
			expectedLine:         5, // 3 + 2
			expectedColumn:       45,
			expectedMessage:      "unexpected token found",
		},
		{
			name:                 "yaml error with column 1 explicitly specified",
			err:                  errors.New("yaml: line 8: column 1: mapping values not allowed in this context"),
			frontmatterStartLine: 0,
			expectedLine:         8, // 8 + 0
			expectedColumn:       1,
			expectedMessage:      "mapping values not allowed in this context",
		},
		{
			name:                 "yaml error with medium column position",
			err:                  errors.New("yaml: line 2: column 23: found character that cannot start any token"),
			frontmatterStartLine: 3,
			expectedLine:         5, // 2 + 3
			expectedColumn:       23,
			expectedMessage:      "found character that cannot start any token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, column, message := ExtractYAMLError(tt.err, tt.frontmatterStartLine)

			if line != tt.expectedLine {
				t.Errorf("Expected line %d, got %d", tt.expectedLine, line)
			}
			if column != tt.expectedColumn {
				t.Errorf("Expected column %d, got %d", tt.expectedColumn, column)
			}
			if message != tt.expectedMessage {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedMessage, message)
			}
		})
	}
}
