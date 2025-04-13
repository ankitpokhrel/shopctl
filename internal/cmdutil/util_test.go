package cmdutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ankitpokhrel/shopctl/internal/config"
)

func TestFormatDateTime(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		format   func() string
		dt       string
		tz       string
		expected string
	}{
		{
			name: "returns original input on invalid datetime input",
			format: func() string {
				return FormatDateTime("not-a-date", "")
			},
			expected: "",
		},
		{
			name: "returns original input on valid dt with invalid timezone",
			format: func() string {
				return FormatDateTime("2025-01-01T11:00:00Z", "Invalid/Timezone")
			},
			expected: "2025-01-01T11:00:00Z",
		},
		{
			name: "formats valid dt with no timezone specified",
			format: func() string {
				return FormatDateTime("2025-01-01T11:00:00Z", "")
			},
			expected: "2025-01-01 11:00:00",
		},
		{
			name: "formats valid dt with valid timezone",
			format: func() string {
				return FormatDateTime("2025-01-01T11:00:00Z", "Asia/Kathmandu")
			},
			expected: "2025-01-01 16:45:00",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.format())
		})
	}
}

func TestFormatDateTimeHuman(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		format   func() string
		expected string
	}{
		{
			name: "it returns input date for invalid date input",
			format: func() string {
				return FormatDateTimeHuman("2024-12-03 10:00:00", time.RFC3339)
			},
			expected: "2024-12-03 10:00:00",
		},
		{
			name: "it returns input date for invalid input format",
			format: func() string {
				return FormatDateTimeHuman("2025-01-10 10:00:00", "invalid")
			},
			expected: "2025-01-10 10:00:00",
		},
		{
			name: "it format input date from RFC3339 date format",
			format: func() string {
				return FormatDateTimeHuman("2025-01-10T16:12:00.000Z", time.RFC3339)
			},
			expected: "Fri, 10 Jan 25",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.format())
		})
	}
}

func TestGetStoreSlug(t *testing.T) {
	tests := []struct {
		name  string
		store string
		want  string
	}{
		{
			name:  "empty store url",
			store: "",
			want:  "",
		},
		{
			name:  "valid store url without protocol",
			store: "store1.myshopify.com",
			want:  "store1",
		},
		{
			name:  "valid store url with http protocol",
			store: "http://store2.myshopify.com",
			want:  "store2",
		},
		{
			name:  "valid store url with https protocol",
			store: "https://store3.myshopify.com",
			want:  "store3",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, GetStoreSlug(tc.store))
		})
	}
}

func TestParseBackupResource(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []config.BackupResource
	}{
		{
			name: "valid input with query and resource",
			input: []string{
				`product="tag:premium AND created_at:>=2025-01-01"`,
				"customer",
			},
			expected: []config.BackupResource{
				{Resource: "product", Query: "\"tag:premium AND created_at:>=2025-01-01\""},
				{Resource: "customer", Query: ""},
			},
		},
		{
			name: "input with multiple resources having queries",
			input: []string{
				`order="status:completed"`,
				`customer="country:US"`,
			},
			expected: []config.BackupResource{
				{Resource: "order", Query: "\"status:completed\""},
				{Resource: "customer", Query: "\"country:US\""},
			},
		},
		{
			name: "input with no query",
			input: []string{
				"product",
				"order",
			},
			expected: []config.BackupResource{
				{Resource: "product", Query: ""},
				{Resource: "order", Query: ""},
			},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []config.BackupResource{},
		},
		{
			name: "input with empty query string",
			input: []string{
				"product=",
				"customer=",
			},
			expected: []config.BackupResource{
				{Resource: "product", Query: ""},
				{Resource: "customer", Query: ""},
			},
		},
		{
			name: "input with malformed resource",
			input: []string{
				"=query",
				"product",
			},
			expected: []config.BackupResource{
				{Resource: "", Query: "query"},
				{Resource: "product", Query: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ParseBackupResource(tt.input))
		})
	}
}

func TestParseRestoreFilters(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		result map[string][]string
		sep    []string
		err    bool
	}{
		{
			name:  "single expression",
			input: "tag:premium",
			result: map[string][]string{
				"tag": {"premium"},
			},
			sep: nil,
			err: false,
		},
		{
			name:  "single expression, multiple vaules",
			input: "tag:premium,on-sale",
			result: map[string][]string{
				"tag": {"premium", "on-sale"},
			},
			sep: nil,
			err: false,
		},
		{
			name:  "simple AND",
			input: "tag:premium AND status:draft",
			result: map[string][]string{
				"tag":    {"premium"},
				"status": {"draft"},
			},
			sep: []string{"AND"},
			err: false,
		},
		{
			name:  "simple OR",
			input: "tag:premium OR status:draft",
			result: map[string][]string{
				"tag":    {"premium"},
				"status": {"draft"},
			},
			sep: []string{"OR"},
			err: false,
		},
		{
			name:  "multiple AND and OR",
			input: "tag:premium,on-sale AND status:draft OR category:books",
			result: map[string][]string{
				"tag":      {"premium", "on-sale"},
				"status":   {"draft"},
				"category": {"books"},
			},
			sep: []string{"AND", "OR"},
			err: false,
		},
		{
			name:  "space separated values",
			input: "title:'first title,second title' AND status:draft",
			result: map[string][]string{
				"title":  {"first title", "second title"},
				"status": {"draft"},
			},
			sep: []string{"AND"},
			err: false,
		},
		{
			name:  "repeated keys",
			input: "tag:premium OR tag:basic AND status:active",
			result: map[string][]string{
				"tag":    {"premium", "basic"},
				"status": {"active"},
			},
			sep: []string{"OR", "AND"},
			err: false,
		},
		{
			name:   "invalid separator",
			input:  "tag:premium XOR status:draft",
			result: nil,
			sep:    nil,
			err:    true,
		},
		{
			name:   "invalid condition format",
			input:  "tag-premium AND status:draft",
			result: nil,
			sep:    nil,
			err:    true,
		},
		{
			name:   "empty input",
			input:  "",
			result: nil,
			sep:    nil,
			err:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, sep, err := ParseRestoreFilters(tc.input)

			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.result, result)
			assert.Equal(t, tc.sep, sep)
		})
	}
}

func TestGetBackupIDFromName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty filename",
			input:    "",
			expected: "",
		},
		{
			name:     "valid name with .tar.gz",
			input:    "daily_2025_02_22_18_18_32_3820045c0c.tar.gz",
			expected: "3820045c0c",
		},
		{
			name:     "valid name without .tar.gz",
			input:    "daily_2025_02_22_18_18_32_3820045c0c",
			expected: "3820045c0c",
		},
		{
			name:     "invalid name missing parts",
			input:    "daily_2025_02_22_18_18_32",
			expected: "",
		},
		{
			name:     "completely invalid name",
			input:    "random_text",
			expected: "",
		},
		{
			name:     "name with additional underscores",
			input:    "daily___2025_02_22_18_18_32_3820045c0c.tar.gz",
			expected: "3820045c0c",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, GetBackupIDFromName(tc.input))
		})
	}
}

func TestSplitKeyVal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		key   string
		value string
		err   bool
	}{
		{
			name:  "Valid key-value",
			input: "Key:Value",
			key:   "Key",
			value: "Value",
			err:   false,
		},
		{
			name:  "Extra whitespace",
			input: "Key:  Value",
			key:   "Key",
			value: "Value",
			err:   false,
		},
		{
			name:  "Multiple colons",
			input: "Key:Value:Extra",
			key:   "Key",
			value: "Value:Extra",
			err:   false,
		},
		{
			name:  "Value with colon wrapped in single quotes",
			input: "Key:'Value:WithColon'",
			key:   "Key",
			value: "'Value:WithColon'",
			err:   false,
		},
		{
			name:  "Missing colon",
			input: "KeyValue",
			key:   "",
			value: "",
			err:   true,
		},
		{
			name:  "Empty key",
			input: ":Value",
			key:   "",
			value: "Value",
			err:   false,
		},
		{
			name:  "Empty value",
			input: "Key:",
			key:   "Key",
			value: "",
			err:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotKey, gotValue, err := SplitKeyVal(tc.input)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.key, gotKey)
				assert.Equal(t, tc.value, gotValue)
			}
		})
	}
}
