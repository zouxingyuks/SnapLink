package GenerateShortLink

import "testing"

func TestGenerateHash(t *testing.T) {
	type args struct {
		uri1 string
		uri2 string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				uri1: "/",
				uri2: "/",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g1 := GenerateHash(tt.args.uri1)
			g2 := GenerateHash(tt.args.uri2)
			if (g1 == g2) != tt.want {
				t.Errorf("uri1: %s, uri2: %s, want %v", g1, g2, tt.want)
			}
		})
	}
}
