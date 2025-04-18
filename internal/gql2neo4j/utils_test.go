package gql2neo4j

import (
	"testing"
)

func Test_isVersionSupported(t *testing.T) {
	type args struct {
		current string
		minimum string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "empty current version",
			args: args{current: "", minimum: "5.9"},
			want: true,
		},
		{
			name: "current version less than minimum",
			args: args{current: "5.8", minimum: "5.9"},
			want: false,
		},
		{
			name: "current version equal to minimum",
			args: args{current: "5.9", minimum: "5.9"},
			want: true,
		},
		{
			name: "current version greater than minimum",
			args: args{current: "6.0", minimum: "5.9"},
			want: true,
		},
		{
			name: "new-style version greater than minimum",
			args: args{current: "2025.1", minimum: "5.9"},
			want: true,
		},
		{
			name:    "invalid current version",
			args:    args{current: "invalid", minimum: "5.9"},
			wantErr: true,
		},
		{
			name:    "invalid minimum version",
			args:    args{current: "5.9", minimum: "invalid"},
			wantErr: true,
		},
		{
			name:    "invalid current and minimum version",
			args:    args{current: "invalid", minimum: "invalid"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isVersionSupported(tt.args.current, tt.args.minimum)
			if (err != nil) != tt.wantErr {
				t.Errorf("isVersionSupported() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isVersionSupported() got = %v, want %v", got, tt.want)
			}
		})
	}
}
