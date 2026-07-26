package docker

import (
	"reflect"
	"testing"
)

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "sqlcmd with single-quoted password placeholder",
			input:    `/opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P '{{.PASSWORD}}'`,
			expected: []string{"/opt/mssql-tools/bin/sqlcmd", "-S", "localhost", "-U", "sa", "-P", "{{.PASSWORD}}"},
		},
		{
			name:     "psql connection string with embedded password",
			input:    `psql 'postgresql://postgres:{{.PASSWORD}}@localhost:5432'`,
			expected: []string{"psql", "postgresql://postgres:{{.PASSWORD}}@localhost:5432"},
		},
		{
			name:     "mysql adjacent quoted flag",
			input:    `mysql -uroot -p'{{.PASSWORD}}'`,
			expected: []string{"mysql", "-uroot", "-p{{.PASSWORD}}"},
		},
		{
			name:     "single word command",
			input:    `redis-cli`,
			expected: []string{"redis-cli"},
		},
		{
			name:     "command with plain argument",
			input:    `./rvn admin-channel`,
			expected: []string{"./rvn", "admin-channel"},
		},
		{
			name:     "no password",
			input:    `cqlsh`,
			expected: []string{"cqlsh"},
		},
		{
			name:     "double quotes",
			input:    `mongo -u admin -p "{{.PASSWORD}}"`,
			expected: []string{"mongo", "-u", "admin", "-p", "{{.PASSWORD}}"},
		},
		{
			name:     "empty string",
			input:    ``,
			expected: nil,
		},
		{
			name:     "extra whitespace",
			input:    `  redis-cli   -a  {{.PASSWORD}}  `,
			expected: []string{"redis-cli", "-a", "{{.PASSWORD}}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitArgs(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("splitArgs(%q) = %#v, want %#v", tt.input, result, tt.expected)
			}
		})
	}
}
