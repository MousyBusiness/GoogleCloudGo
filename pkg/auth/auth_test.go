package auth

import "testing"

func TestCheckInternal(t *testing.T) {
	var tests = []struct {
		input    string
		expected bool
	}{
		{"10.0.0.2", true},
		{"127.0.0.1", true},
		{"172.16.0.1", true},
		{"172.0.0.1", false},
		{"172.32.0.1", false},
		{"192.168.0.1", true},
		{"192.167.0.1", false},
		{"192.169.0.1", false},
		{"192.168.234.1", true},
		{"2.32.0.222", false},
		{"32.18.1.23", false},
		{"0.0.0.0", false},
	}

	for _, test := range tests {
		if output := checkInternal(test.input); output != test.expected {
			t.Errorf("test failed: input: %v, wanted: %v, got: %v", test.input, test.expected, output)
		}
	}
}
