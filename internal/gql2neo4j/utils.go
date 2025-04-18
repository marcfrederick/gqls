package gql2neo4j

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/antlr4-go/antlr/v4"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/gql/parser"
)

// isVersionSupported checks if the current Neo4j version satisfies the minimum requirement.
func isVersionSupported(current, minimum string) (bool, error) {
	if current == "" {
		return true, nil
	}

	currVer, err := semver.NewVersion(current)
	if err != nil {
		return false, fmt.Errorf("error parsing current version: %s", current)
	}
	minVer, err := semver.NewVersion(minimum)
	if err != nil {
		return false, fmt.Errorf("error parsing minimum version: %s", minimum)
	}

	return currVer.GreaterThanEqual(minVer), nil
}

// parseGQL parses a GQL input string into a listener structure.
func parseGQL(input string) (*gql.CreateGraphTypeListener, error) {
	lexer := parser.NewGQLLexer(antlr.NewInputStream(input))
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewGQLParser(stream)

	listener := &gql.CreateGraphTypeListener{}
	antlr.ParseTreeWalkerDefault.Walk(listener, p.GqlProgram())

	return listener, nil
}
