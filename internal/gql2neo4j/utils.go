package gql2neo4j

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/antlr4-go/antlr/v4"

	"github.com/marcfrederick/gqls/internal/gql"
	"github.com/marcfrederick/gqls/internal/gql/parser"
)

var (
	calendarVersionRegex = regexp.MustCompile(`^(\d{4})\.(\d{2})(?:\.(\d+))?$`)
	semverVersionRegex   = regexp.MustCompile(`^(\d+)\.(\d+)(?:\.(\d+))?$`)
)

// IsVersionSupported checks if the current Neo4j version satisfies the minimum requirement.
func IsVersionSupported(current, minimum string) (bool, error) {
	if current == "" {
		return true, nil
	}

	// Determine version types
	currentIsDate := calendarVersionRegex.MatchString(current)
	minimumIsDate := calendarVersionRegex.MatchString(minimum)

	// Calendar versions are always newer than semver versions
	if currentIsDate && !minimumIsDate {
		return true, nil
	}
	if !currentIsDate && minimumIsDate {
		return false, nil
	}

	// Direct string comparison for calendar versions
	if currentIsDate && minimumIsDate {
		currParts := strings.Split(current, ".")
		minParts := strings.Split(minimum, ".")

		// Compare year
		currYear, _ := strconv.Atoi(currParts[0])
		minYear, _ := strconv.Atoi(minParts[0])
		if currYear != minYear {
			return currYear > minYear, nil
		}

		// Compare month
		currMonth, _ := strconv.Atoi(currParts[1])
		minMonth, _ := strconv.Atoi(minParts[1])
		if currMonth != minMonth {
			return currMonth > minMonth, nil
		}

		// Compare patch if present
		currPatch := 0
		minPatch := 0
		if len(currParts) > 2 {
			currPatch, _ = strconv.Atoi(currParts[2])
		}
		if len(minParts) > 2 {
			minPatch, _ = strconv.Atoi(minParts[2])
		}
		return currPatch >= minPatch, nil
	}

	// For semver versions, use semver comparison
	if !currentIsDate && !minimumIsDate {
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

	return false, fmt.Errorf("invalid version comparison")
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
