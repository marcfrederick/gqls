package gql2neo4j_test

import (
	"testing"

	"github.com/marcfrederick/gqls/internal/gql2neo4j"
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
			name: "calendar version vs semver",
			args: args{current: "2024.02.0", minimum: "5.9"},
			want: true,
		},
		{
			name: "semver vs calendar version",
			args: args{current: "6.0", minimum: "2024.02.0"},
			want: false,
		},
		{
			name: "calendar version vs calendar version - newer",
			args: args{current: "2025.02.0", minimum: "2024.12.0"},
			want: true,
		},
		{
			name: "calendar version vs calendar version - older",
			args: args{current: "2024.02.0", minimum: "2024.12.0"},
			want: false,
		},
		{
			name: "calendar version vs calendar version - same year different month",
			args: args{current: "2024.12.0", minimum: "2024.02.0"},
			want: true,
		},
		{
			name: "calendar version vs calendar version - same year and month different patch",
			args: args{current: "2024.02.1", minimum: "2024.02.0"},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gql2neo4j.IsVersionSupported(tt.args.current, tt.args.minimum)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsVersionSupported() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsVersionSupported() got = %v, want %v", got, tt.want)
			}
		})
	}
}
