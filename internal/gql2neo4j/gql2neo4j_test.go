package gql2neo4j_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/gql/datatype"
	"github.com/marcfrederick/gqls/internal/gql2neo4j"
	"github.com/marcfrederick/gqls/internal/neo4j/constraints"
)

type testCase struct {
	Input string `yaml:"input"`
	Want  string `yaml:"want"`
}

func TestConvert(t *testing.T) {
	testCasePaths, err := filepath.Glob("testdata/*.yml")
	if err != nil {
		t.Fatalf("failed to find test cases: %v", err)
	}

	for _, path := range testCasePaths {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			tc := loadTestCase(t, path)

			got, err := gql2neo4j.Convert(tc.Input, "")
			if err != nil {
				t.Errorf("Convert() returned error: %v", err)
			}

			got = strings.TrimSpace(got)
			want := strings.TrimSpace(tc.Want)
			if got != want {
				t.Errorf("Convert() got = %v, want %v", got, want)
			}
		})
	}
}

func TestGenerateConstraints(t *testing.T) {
	type args struct {
		graphType    *gql.GraphType
		neo4jVersion string
	}
	tests := []struct {
		name    string
		args    args
		want    []constraints.Constraint
		wantErr bool
	}{
		{
			name: "empty graph type",
			args: args{
				graphType: &gql.GraphType{},
			},
			want: nil,
		},
		{
			name: "node type constraints",
			args: args{
				graphType: &gql.GraphType{
					NodeTypeSet: []*gql.NodeType{
						{
							NodeTypeLabelSet: []string{"Person"},
							PropertyTypeSet: []*gql.PropertyType{
								{
									PropertyName: "name",
									PropertyValueType: &datatype.CharacterStringDataTypeDescriptor{
										Nullable: true,
									},
								},
								{
									PropertyName: "age",
									PropertyValueType: &datatype.NumericDataTypeDescriptor{
										Kind:     datatype.NumericKindExact,
										Radix:    datatype.RadixDecimal,
										Nullable: false,
									},
								},
							},
						},
					},
				},
			},
			want: []constraints.Constraint{
				&constraints.PropertyTypeConstraint{
					BaseConstraint: constraints.BaseConstraint{
						Name:    "constraint_person_name_type",
						Pattern: "(n:Person)",
					},
					PropertyName: "n.name",
					PropertyType: "STRING",
				},
				&constraints.PropertyTypeConstraint{
					BaseConstraint: constraints.BaseConstraint{
						Name:    "constraint_person_age_type",
						Pattern: "(n:Person)",
					},
					PropertyName: "n.age",
					PropertyType: "INTEGER",
				},
				&constraints.PropertyExistenceConstraint{
					BaseConstraint: constraints.BaseConstraint{
						Name:    "constraint_person_age_existence",
						Pattern: "(n:Person)",
					},
					PropertyName: "n.age",
				},
			},
		},
		{
			name: "node type constraints unsupported",
			args: args{
				graphType: &gql.GraphType{
					NodeTypeSet: []*gql.NodeType{
						{
							NodeTypeLabelSet: []string{"Person"},
							PropertyTypeSet: []*gql.PropertyType{
								{
									PropertyName: "foo",
									PropertyValueType: &datatype.MiscellaneousDataTypeDescriptor{
										Kind:     "BAR",
										Nullable: true,
									},
								},
							},
						},
					},
				},
				neo4jVersion: "5.8",
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gql2neo4j.GenerateConstraints(tt.args.graphType, tt.args.neo4jVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateConstraints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, got, tt.want, "GenerateConstraints() got = %v, want %v", got, tt.want)
		})
	}
}

// loadTestCase loads a test case from the specified path.
func loadTestCase(t *testing.T, path string) testCase {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read test case file %q: %v", path, err)
	}

	var tc testCase
	if err := yaml.Unmarshal(data, &tc); err != nil {
		t.Fatalf("failed to unmarshal test case %q: %v", path, err)
	}

	if strings.TrimSpace(tc.Input) == "" || strings.TrimSpace(tc.Want) == "" {
		t.Skipf("test case %q missing input or want", path)
	}

	return tc
}
