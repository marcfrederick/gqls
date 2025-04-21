package gql2neo4j

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/gql/parser"
)

// isVersionSupported checks if the current Neo4j version satisfies the minimum requirement.
func isVersionSupported(current, minimum string) (bool, error) {
	if current == "" || current == "latest" {
		return true, nil
	}

	currentParts := strings.Split(current, ".")
	minimumParts := strings.Split(minimum, ".")

	for i := 0; i < len(currentParts) && i < len(minimumParts); i++ {
		currentStr := currentParts[i]
		currentPart, err := strconv.Atoi(currentStr)
		if err != nil {
			return false, fmt.Errorf("invalid current version part %q: %w", currentStr, err)
		}

		minimumStr := minimumParts[i]
		minimumPart, err := strconv.Atoi(minimumStr)
		if err != nil {
			return false, fmt.Errorf("invalid minimum version part %q: %w", minimumStr, err)
		}

		if currentPart < minimumPart {
			return false, nil
		} else if currentPart > minimumPart {
			return true, nil
		}
	}

	return len(currentParts) >= len(minimumParts), nil
}

// parseGQL parses a GQL input string into a listener structure.
func parseGQL(input string) (*gql.CreateGraphStatementListener, error) {
	lexer := parser.NewGQLLexer(antlr.NewInputStream(input))
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewGQLParser(stream)

	errorListener := &gql.ErrorListener{}
	p.AddErrorListener(errorListener)

	createGraphTypeListener := &gql.CreateGraphStatementListener{}
	antlr.ParseTreeWalkerDefault.Walk(createGraphTypeListener, p.GqlProgram())

	if len(errorListener.SyntaxErrors) > 0 {
		return nil, fmt.Errorf("syntax errors: %v", errorListener.SyntaxErrors)
	}

	return createGraphTypeListener, nil
}
