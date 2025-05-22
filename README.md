# GQLS - GQL to Neo4j Schema Converter

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/apache-2-0)
[![Test](https://github.com/marcfrederick/gqls/actions/workflows/test.yml/badge.svg)](https://github.com/marcfrederick/gqls/actions/workflows/test.yml)

<div align="center">
    <img src="internal/web/static/assets/logo.png" alt="GQLS Logo" width="256px"/>
</div>

GQLS is a CLI tool that converts GQL schema definitions (ISO/IEC 39075:2024) into Neo4j-compatible Cypher statements. It
helps bridge the gap between the Graph Query Language (GQL) standard and Neo4j's implementation, making it easier to
work with both technologies.

## Features

- Parse GQL schema files and generate Neo4j constraints and indexes
- Support for multiple Neo4j versions
- Command-line interface for ease of use
- Web interface for interactive conversion
- Docker image available for containerized usage
- APOC plugin support for advanced type constraints

### Types Supported

#### Boolean Types

- `BOOLEAN` -> `BOOLEAN`
- `BOOL` -> `BOOLEAN`

#### Numeric Types

##### Exact Numeric Types

- `INT` -> `INTEGER`
- `SMALLINT` -> `INTEGER`
- `BIGINT` -> `INTEGER`
- `INT8` -> `INTEGER` + Trigger
- `INT16` -> `INTEGER` + Trigger
- `INT32` -> `INTEGER` + Trigger
- `INT64` -> `INTEGER`
- `INT(<precision>)` -> `INTEGER` + Trigger
- `FLOAT` -> `FLOAT`
- `FLOAT32` -> `FLOAT` + Trigger
- `FLOAT64` -> `FLOAT`
- `FLOAT(<precision>)` -> `FLOAT` + Trigger

##### Approximate Numeric Types

- `FLOAT` -> `FLOAT`
- `REAL` -> `FLOAT`
- `DOUBLE` -> `FLOAT`
- `FLOAT32` -> `FLOAT` + Trigger
- `FLOAT64` -> `FLOAT`
- `FLOAT(<precision>)` -> `FLOAT` + Trigger

#### String Types

- `STRING` -> `STRING`
- `STRING(<max>)` -> `STRING` + Trigger
- `STRING(<min>, <max>)` -> `STRING` + Trigger
- `VARCHAR` -> `STRING`
- `VARCHAR(<max>)` -> `STRING` + Trigger
- `VARCHAR(<min>, <max>)` -> `STRING` + Trigger
- `CHAR(<fixed>)` -> `STRING` + Trigger

## Installation

### Prerequisites

- [Go](https://golang.org/doc/install) (>= 1.23)
- [Make](https://www.gnu.org/software/make/) (for building)
- [Java Runtime Environment (JRE)](https://adoptium.net/de/temurin/releases/) (e.g. Temurin; for parser generation)

### Clone and Build

```shell
git clone https://github.com/marcfrederick/gqls.git
cd gqls
make build
```

### Generate the Parser

Before building, generate the ANTLR4 parser files:

```shell
make generate
```

This will populate the `internal/gql/parser/` directory with Go bindings for the GQL grammar.
You must have a valid Java Runtime (JRE) installed and available in your `$PATH` for this step.
Make will automatically download the required ANTLR4 jar file if it is not already present.

### Docker

You can also use the Docker image:

```shell
docker run --rm -v $(pwd):/data marcfrederick/gqls:latest --input /data/schema.gql --output /data/output.cypher
```

## Usage

### Command Line Interface

Run the `gqls` command with the required options:

```bash
./gqls --input schema.gql --neo4j-version 5.9
```

#### Options

| Flag              | Description                                                             |
|-------------------|-------------------------------------------------------------------------|
| `--input`, `-i`   | Path to the GQL schema file. Use - or omit to read from stdin.          |
| `--output`, `-o`  | Path to the output file. Defaults to stdout.                            |
| `--neo4j-version` | Target Neo4j version (e.g., 5.9). Defaults to latest supported version. |
| `--apoc`          | Generate APOC triggers. Defaults to false.                              |
| `--serve`         | Start web server on the specified address (e.g., :8080)                 |

### Web Interface

To start the web interface:

```bash
./gqls --serve :8080
```

Then open your browser to `http://localhost:8080` to use the interactive converter.

### Example

Given a GQL schema like:

```sql
CREATE
PROPERTY GRAPH socialNetworkType {
    (:Person { name :: STRING NOT NULL, age :: INT8 }),
    (:Post { content :: STRING, created_at :: DATE }),
    (:Person)~[:is_friend { since :: DATE NOT NULL }]~(:Person),
    (:Person)-[:likes { since :: DATE NOT NULL }]->(:Post)
}
```

GQLS will generate Neo4j constraints and indexes for all properties and relationships.

## Development

### Linting

This project uses `golangci-lint` for linting. Run the following command to check for linting issues:

```bash
make lint
```

### Testing

Run unit tests using:

```bash
make test
```

To run all tests, including integration tests, use:

```bash
make test-integration
```

You need to have either `docker` or `podman` installed to run the integration tests.

#### Docker setup

When using Docker, no additional configuration is needed. Integration tests will work out of the box.

#### Podman setup

When using Podman, make sure to use a rootfull container and set the following environment variables:

```bash
export TESTCONTAINERS_DRIVER="podman"
export TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED="true"
```

### Contributing

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request with a detailed description of your changes.

## License

This project is dual-licensed under the [MIT License](LICENSE-MIT) and the [Apache License 2.0](LICENSE-APACHE).
You can choose either license for your use.

## Acknowledgments

- [opengql/grammar](https://github.com/opengql/grammar) for the ANTLR4 translation of the GQL grammar.
