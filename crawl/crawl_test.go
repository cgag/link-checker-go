package crawl

import (
	"testing"
)

func TestTesting(t *testing.T) {

	tests := []struct {
		input          string
		expectedOutput string
	}{
		{"", ""},
		{"a", "a"},
		{"http://curtis.io/with#link", "http://curtis.io/with"},
		{"http://curtis.io/#tag", "http://curtis.io/"},
		{"http://curtis.io/good", "http://curtis.io/good"},
	}

	for _, test := range tests {
		out := stripInPageLink(test.input)
		if out != test.expectedOutput {
			t.Error("failed on: ", test.input, "with output: ", out)
		}
	}

}
