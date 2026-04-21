package agent

import "testing"

func TestExtractCPPCode(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "cpp fence",
			raw:  "text\n```cpp\nint main(){return 0;}\n```",
			want: "int main(){return 0;}",
		},
		{
			name: "generic fence",
			raw:  "```\n#include <bits/stdc++.h>\n```",
			want: "#include <bits/stdc++.h>",
		},
		{
			name: "no fence",
			raw:  "int main(){return 0;}",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractCPPCode(tt.raw); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
