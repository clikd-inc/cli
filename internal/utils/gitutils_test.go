package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test struct for AssignDynamicValues
type TestStruct struct {
	Name    string
	Email   string
	Age     string
	Address string
}

// Test struct for DotGet
type Author struct {
	Name  string
	Email string
}

type Commit struct {
	Hash    string
	Author  *Author
	Message string
}

func TestAssignDynamicValues(t *testing.T) {
	tests := []struct {
		name     string
		attrs    []string
		values   []string
		expected TestStruct
	}{
		{
			name:   "Basic assignment",
			attrs:  []string{"Name", "Email"},
			values: []string{"John Doe", "john@example.com"},
			expected: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
			},
		},
		{
			name:   "All fields",
			attrs:  []string{"Name", "Email", "Age", "Address"},
			values: []string{"Jane Smith", "jane@example.com", "30", "123 Main St"},
			expected: TestStruct{
				Name:    "Jane Smith",
				Email:   "jane@example.com",
				Age:     "30",
				Address: "123 Main St",
			},
		},
		{
			name:   "More values than attributes",
			attrs:  []string{"Name", "Email"},
			values: []string{"Bob", "bob@example.com", "extra", "more"},
			expected: TestStruct{
				Name:  "Bob",
				Email: "bob@example.com",
			},
		},
		{
			name:   "More attributes than values",
			attrs:  []string{"Name", "Email", "Age", "Address"},
			values: []string{"Alice", "alice@example.com"},
			expected: TestStruct{
				Name:  "Alice",
				Email: "alice@example.com",
			},
		},
		{
			name:   "Invalid field name",
			attrs:  []string{"Name", "InvalidField", "Email"},
			values: []string{"Test", "ignored", "test@example.com"},
			expected: TestStruct{
				Name:  "Test",
				Email: "test@example.com",
			},
		},
		{
			name:     "Empty arrays",
			attrs:    []string{},
			values:   []string{},
			expected: TestStruct{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := &TestStruct{}
			AssignDynamicValues(target, tt.attrs, tt.values)
			assert.Equal(t, tt.expected, *target)
		})
	}
}

func TestConvNewline(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		nlcode   string
		expected string
	}{
		{
			name:     "Convert CRLF to LF",
			input:    "line1\r\nline2\r\nline3",
			nlcode:   "\n",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "Convert CR to LF",
			input:    "line1\rline2\rline3",
			nlcode:   "\n",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "Convert LF to CRLF",
			input:    "line1\nline2\nline3",
			nlcode:   "\r\n",
			expected: "line1\r\nline2\r\nline3",
		},
		{
			name:     "Mixed line endings",
			input:    "line1\r\nline2\rline3\nline4",
			nlcode:   "\n",
			expected: "line1\nline2\nline3\nline4",
		},
		{
			name:     "No line endings",
			input:    "single line",
			nlcode:   "\n",
			expected: "single line",
		},
		{
			name:     "Empty string",
			input:    "",
			nlcode:   "\n",
			expected: "",
		},
		{
			name:     "Custom separator",
			input:    "line1\nline2\nline3",
			nlcode:   " | ",
			expected: "line1 | line2 | line3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvNewline(tt.input, tt.nlcode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDotGet(t *testing.T) {
	author := &Author{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	commit := &Commit{
		Hash:    "abc123",
		Author:  author,
		Message: "Initial commit",
	}

	tests := []struct {
		name        string
		object      interface{}
		fieldPath   string
		expected    interface{}
		shouldExist bool
	}{
		{
			name:        "Simple field access",
			object:      commit,
			fieldPath:   "Hash",
			expected:    "abc123",
			shouldExist: true,
		},
		{
			name:        "Nested field access",
			object:      commit,
			fieldPath:   "Author.Name",
			expected:    "John Doe",
			shouldExist: true,
		},
		{
			name:        "Nested field access - Email",
			object:      commit,
			fieldPath:   "Author.Email",
			expected:    "john@example.com",
			shouldExist: true,
		},
		{
			name:        "Direct struct access",
			object:      author,
			fieldPath:   "Name",
			expected:    "John Doe",
			shouldExist: true,
		},
		{
			name:        "Invalid field",
			object:      commit,
			fieldPath:   "InvalidField",
			expected:    nil,
			shouldExist: false,
		},
		{
			name:        "Invalid nested field",
			object:      commit,
			fieldPath:   "Author.InvalidField",
			expected:    nil,
			shouldExist: false,
		},
		{
			name:        "Nil object",
			object:      nil,
			fieldPath:   "Hash",
			expected:    nil,
			shouldExist: false,
		},
		{
			name:        "Empty field path",
			object:      commit,
			fieldPath:   "",
			expected:    nil,
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, exists := DotGet(tt.object, tt.fieldPath)
			assert.Equal(t, tt.shouldExist, exists)
			if tt.shouldExist {
				assert.Equal(t, tt.expected, result)
			}
		})
	}

	t.Run("Nil pointer in nested access", func(t *testing.T) {
		commitWithNilAuthor := &Commit{
			Hash:    "abc123",
			Author:  nil,
			Message: "Test commit",
		}

		result, exists := DotGet(commitWithNilAuthor, "Author.Name")
		assert.False(t, exists)
		assert.Nil(t, result)
	})

	t.Run("Non-struct object", func(t *testing.T) {
		stringObj := "test string"
		result, exists := DotGet(stringObj, "Length")
		assert.False(t, exists)
		assert.Nil(t, result)
	})
}

func TestJoinAndQuoteMeta(t *testing.T) {
	tests := []struct {
		name     string
		list     []string
		sep      string
		expected string
	}{
		{
			name:     "Basic join",
			list:     []string{"hello", "world"},
			sep:      " ",
			expected: "hello world",
		},
		{
			name:     "Join with special characters",
			list:     []string{"test.*", "pattern+", "regex?"},
			sep:      "|",
			expected: "test\\.\\*|pattern\\+|regex\\?",
		},
		{
			name:     "Empty list",
			list:     []string{},
			sep:      ",",
			expected: "",
		},
		{
			name:     "Single item",
			list:     []string{"single"},
			sep:      ",",
			expected: "single",
		},
		{
			name:     "Special regex characters",
			list:     []string{"[abc]", "(group)", "{3,5}", "^start", "end$"},
			sep:      " ",
			expected: "\\[abc\\] \\(group\\) \\{3,5\\} \\^start end\\$",
		},
		{
			name:     "Mixed content",
			list:     []string{"normal", "special.*", "more+normal"},
			sep:      ", ",
			expected: "normal, special\\.\\*, more\\+normal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinAndQuoteMeta(tt.list, tt.sep)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompare(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)

	tests := []struct {
		name        string
		a           interface{}
		operator    string
		b           interface{}
		expected    bool
		shouldError bool
	}{
		// String comparisons
		{
			name:     "String less than",
			a:        "apple",
			operator: "<",
			b:        "banana",
			expected: true,
		},
		{
			name:     "String greater than",
			a:        "zebra",
			operator: ">",
			b:        "apple",
			expected: true,
		},
		{
			name:     "String equal",
			a:        "test",
			operator: "==",
			b:        "test",
			expected: true,
		},
		{
			name:     "String not equal",
			a:        "hello",
			operator: "!=",
			b:        "world",
			expected: true,
		},

		// Integer comparisons
		{
			name:     "Int less than",
			a:        5,
			operator: "<",
			b:        10,
			expected: true,
		},
		{
			name:     "Int greater than",
			a:        15,
			operator: ">",
			b:        10,
			expected: true,
		},
		{
			name:     "Int equal",
			a:        42,
			operator: "==",
			b:        42,
			expected: true,
		},
		{
			name:     "Int not equal",
			a:        1,
			operator: "!=",
			b:        2,
			expected: true,
		},

		// Time comparisons
		{
			name:     "Time less than",
			a:        now,
			operator: "<",
			b:        later,
			expected: true,
		},
		{
			name:     "Time greater than",
			a:        later,
			operator: ">",
			b:        now,
			expected: true,
		},
		{
			name:     "Time equal",
			a:        now,
			operator: "==",
			b:        now,
			expected: true,
		},
		{
			name:     "Time not equal",
			a:        now,
			operator: "!=",
			b:        later,
			expected: true,
		},

		// Error cases
		{
			name:        "Different types",
			a:           "string",
			operator:    "==",
			b:           42,
			expected:    false,
			shouldError: true,
		},
		{
			name:     "Unsupported type",
			a:        []string{"slice"},
			operator: "==",
			b:        []string{"slice"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compare(tt.a, tt.operator, tt.b)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCompareString(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		operator string
		b        string
		expected bool
	}{
		{"Less than true", "apple", "<", "banana", true},
		{"Less than false", "zebra", "<", "apple", false},
		{"Greater than true", "zebra", ">", "apple", true},
		{"Greater than false", "apple", ">", "zebra", false},
		{"Equal true", "test", "==", "test", true},
		{"Equal false", "test", "==", "different", false},
		{"Not equal true", "hello", "!=", "world", true},
		{"Not equal false", "same", "!=", "same", false},
		{"Invalid operator", "a", "invalid", "b", false},
		{"Empty strings equal", "", "==", "", true},
		{"Empty vs non-empty", "", "<", "a", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareString(tt.a, tt.operator, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompareInt(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		operator string
		b        int
		expected bool
	}{
		{"Less than true", 5, "<", 10, true},
		{"Less than false", 15, "<", 10, false},
		{"Greater than true", 15, ">", 10, true},
		{"Greater than false", 5, ">", 10, false},
		{"Equal true", 42, "==", 42, true},
		{"Equal false", 42, "==", 24, false},
		{"Not equal true", 1, "!=", 2, true},
		{"Not equal false", 5, "!=", 5, false},
		{"Invalid operator", 1, "invalid", 2, false},
		{"Zero values", 0, "==", 0, true},
		{"Negative numbers", -5, "<", -3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareInt(tt.a, tt.operator, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompareTime(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)
	later := now.Add(time.Hour)

	tests := []struct {
		name     string
		a        time.Time
		operator string
		b        time.Time
		expected bool
	}{
		{"Less than true", earlier, "<", now, true},
		{"Less than false", later, "<", now, false},
		{"Greater than true", later, ">", now, true},
		{"Greater than false", earlier, ">", now, false},
		{"Equal true", now, "==", now, true},
		{"Equal false", now, "==", later, false},
		{"Not equal true", now, "!=", later, true},
		{"Not equal false", now, "!=", now, false},
		{"Invalid operator", now, "invalid", later, false},
		{"Same time equal", now, "==", now, true},
		{"Equal times with different references", earlier, "<", earlier.Add(time.Nanosecond), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareTime(tt.a, tt.operator, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDotGetEdgeCases(t *testing.T) {
	t.Run("Interface with nil value", func(t *testing.T) {
		var nilInterface interface{} = nil
		result, exists := DotGet(nilInterface, "field")
		assert.False(t, exists)
		assert.Nil(t, result)
	})

	t.Run("Pointer to struct", func(t *testing.T) {
		author := &Author{Name: "Test Author", Email: "test@example.com"}
		result, exists := DotGet(author, "Name")
		assert.True(t, exists)
		assert.Equal(t, "Test Author", result)
	})

	t.Run("Multiple pointer dereferences", func(t *testing.T) {
		author := &Author{Name: "Deep Author", Email: "deep@example.com"}
		authorPtr := &author
		result, exists := DotGet(authorPtr, "Name")
		assert.True(t, exists)
		assert.Equal(t, "Deep Author", result)
	})
}
