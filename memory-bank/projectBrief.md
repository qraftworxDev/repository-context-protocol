# Repository Context Protocol - Project Brief

## Project Overview
The Repository Context Protocol is a Go-based system that provides structured analysis and indexing of code repositories across multiple programming languages. It serves as a foundation for AI-powered code understanding and repository navigation tools.

## Core Requirements

### Language Support
- **Go**: Primary language with full AST parsing and analysis
- **Python**: Advanced AST parsing with type hint support
- **TypeScript**: Basic parsing support (**Not implemented yet**)
- **Extensible**: Architecture supports adding new language parsers

### Key Features
1. **AST Analysis**: Deep parsing of code structure, functions, classes, variables
2. **Type System**: Extraction and analysis of type information
3. **Call Graph Generation**: Building relationships between functions and methods
4. **Import/Export Tracking**: Understanding module dependencies
5. **Semantic Chunking**: Intelligent code segmentation for AI processing


### Technical Architecture
- **Modular Parser System**: Language-specific parsers implementing common interfaces
- **SQLite Storage**: Efficient indexing and querying of code metadata
- **Concurrent Processing**: Multi-threaded analysis for large repositories

## Current Focus
The project is actively developing the Python AST parser with emphasis on:
- Complex type annotation support (Union, Optional, Generic types)
- Method signature generation with multiple return types
- Comprehensive test coverage for edge cases
- Integration with the broader indexing system

## Quality Standards
- Test-driven development with comprehensive test suites
- Concurrent-safe implementations
- Clear error handling and logging
- Performance optimization for large codebases
- Clean, maintainable code architecture
