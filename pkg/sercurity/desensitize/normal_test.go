package desensitize

import "testing"

func TestIDCard(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				id: "123456789012345678",
			},
			want: "1234**********5678",
		},
		{
			name: "test2",
			args: args{
				id: "12345678901234567X",
			},
			want: "1234**********567X",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IDCard(tt.args.id); got != tt.want {
				t.Errorf("IDCard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPhoneNumber(t *testing.T) {
	type args struct {
		phone string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				phone: "12345678901",
			},
			want: "123****901",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PhoneNumber(tt.args.phone); got != tt.want {
				t.Errorf("PhoneNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}
