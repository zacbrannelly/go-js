# Toy JavaScript Runtime

A toy JavaScript runtime written in Go. This project is a learning exercise to both learn Go and better understand how JavaScript works under the hood.

## Current Status

The project is in very early stages. Currently implementing:

- [x] Lexical Analysis
- [x] Syntax Analysis
- [x] Abstract Syntax Tree Generation
- [ ] Semantic Analysis
- [ ] Naive Evaluation
- [ ] Intermediate Code Generation
- [ ] Code Optimization
- [ ] Code Generation

## Prerequisites

Requires Go 1.20 or later.

## Usage

```bash
go run .
```

You'll be prompted to select a mode:

1) Lexer - Tokenize JavaScript code and display the tokens
2) Parser - Parse JavaScript code and display the AST
3) Runtime - Execute JavaScript code and display the result

The runtime mode is currently very basic and only supports simple expressions.
