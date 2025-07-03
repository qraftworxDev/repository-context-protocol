package golang

import (
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"strings"
	"time"
	"unicode"

	"repository-context-protocol/internal/models"
)

const (
	languageGo    = "go"
	kindStruct    = "struct"
	kindInterface = "interface"
	kindAlias     = "alias"
	kindBasic     = "basic"
	kindComposite = "composite"
	kindPointer   = "pointer"
	kindNamed     = "named"

	// Type inference constants
	typeInt        = "int"
	typeFloat64    = "float64"
	typeString     = "string"
	typeRune       = "rune"
	typeBool       = "bool"
	typeComplex128 = "complex128"
)

// Go AST parser implementation
type GoParser struct {
	fset *token.FileSet
}

func NewGoParser() *GoParser {
	return &GoParser{
		fset: token.NewFileSet(),
	}
}

func (p *GoParser) GetSupportedExtensions() []string {
	return []string{".go"}
}

func (p *GoParser) GetLanguageName() string {
	return languageGo
}

func (p *GoParser) ParseFile(path string, content []byte) (*models.FileContext, error) {
	// Parse Go AST and extract functions, types, imports
	file, err := parser.ParseFile(p.fset, path, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Calculate checksum of content
	hash := sha256.Sum256(content)
	checksum := fmt.Sprintf("%x", hash)

	// Get modification time
	var modTime time.Time
	if fileInfo, err := os.Stat(path); err == nil {
		modTime = fileInfo.ModTime()
	} else {
		// If file doesn't exist (e.g., in-memory parsing), use current time
		modTime = time.Now()
	}

	ctx := &models.FileContext{
		Path:      path,
		Language:  languageGo,
		Checksum:  checksum,
		ModTime:   modTime,
		Functions: []models.Function{},
		Types:     []models.TypeDef{},
		Variables: []models.Variable{},
		Imports:   []models.Import{},
		Exports:   []models.Export{},
	}

	// Extract imports
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		}
		ctx.Imports = append(ctx.Imports, models.Import{
			Path:  importPath,
			Alias: alias,
		})
	}

	// Extract functions, types, etc. from AST
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			// Extract all functions (not just exported ones for testing)
			ctx.Functions = append(ctx.Functions, p.extractFunction(node))
		case *ast.TypeSpec:
			ctx.Types = append(ctx.Types, p.extractType(node))
		case *ast.GenDecl:
			if node.Tok == token.VAR {
				for _, spec := range node.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range valueSpec.Names {
							// Extract all variables (not just exported ones)
							ctx.Variables = append(ctx.Variables, p.extractVariable(name, valueSpec))
						}
					}
				}
			} else if node.Tok == token.CONST {
				for _, spec := range node.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range valueSpec.Names {
							// Extract all constants (not just exported ones)
							ctx.Constants = append(ctx.Constants, p.extractConstant(name, valueSpec))
						}
					}
				}
			}
		}
		return true
	})

	// Extract methods for types (second pass)
	p.extractMethods(file, ctx)

	// Build call graph relationships (third pass)
	p.buildCallGraph(ctx)

	// Extract exports (fourth pass)
	p.extractExports(ctx)

	return ctx, nil
}

func (p *GoParser) extractFunction(node *ast.FuncDecl) models.Function {
	fn := models.Function{
		Name:       node.Name.Name,
		Parameters: []models.Parameter{},
		Returns:    []models.Type{},

		// Deprecated fields for backward compatibility
		Calls:    []string{},
		CalledBy: []string{},

		// Enhanced fields with CallReference metadata
		LocalCalls:       []string{},
		CrossFileCalls:   []models.CallReference{},
		LocalCallers:     []string{},
		CrossFileCallers: []models.CallReference{},
	}

	// Extract position information
	if node.Pos().IsValid() {
		pos := p.fset.Position(node.Pos())
		fn.StartLine = pos.Line
	}
	if node.End().IsValid() {
		pos := p.fset.Position(node.End())
		fn.EndLine = pos.Line
	}

	// Extract parameters and returns
	fn.Parameters = p.extractFunctionParameters(node)
	fn.Returns = p.extractFunctionReturns(node)

	// Extract function calls
	p.populateFunctionCalls(node, &fn)

	// Build signature
	fn.Signature = p.buildFunctionSignature(node)

	return fn
}

// extractFunctionParameters extracts parameter information from a function declaration
func (p *GoParser) extractFunctionParameters(node *ast.FuncDecl) []models.Parameter {
	var parameters []models.Parameter

	if node.Type.Params != nil {
		for _, field := range node.Type.Params.List {
			paramType := p.typeToString(field.Type)
			if len(field.Names) > 0 {
				for _, name := range field.Names {
					parameters = append(parameters, models.Parameter{
						Name: name.Name,
						Type: paramType,
					})
				}
			} else {
				// Anonymous parameter
				parameters = append(parameters, models.Parameter{
					Name: "",
					Type: paramType,
				})
			}
		}
	}

	return parameters
}

// extractFunctionReturns extracts return type information from a function declaration
func (p *GoParser) extractFunctionReturns(node *ast.FuncDecl) []models.Type {
	var returns []models.Type

	if node.Type.Results != nil {
		for _, field := range node.Type.Results.List {
			returnType := p.typeToString(field.Type)
			returns = append(returns, models.Type{
				Name: returnType,
				Kind: p.getTypeKind(returnType),
			})
		}
	}

	return returns
}

// populateFunctionCalls extracts function calls from the body and populates call fields
func (p *GoParser) populateFunctionCalls(node *ast.FuncDecl, fn *models.Function) {
	if node.Body != nil {
		// Extract calls for deprecated field (backward compatibility)
		fn.Calls = p.extractFunctionCalls(node.Body)

		// Extract calls with metadata for enhanced fields
		callsWithMetadata := p.extractFunctionCallsWithMetadata(node.Body)

		// Populate LocalCalls (all calls initially - enrichment will categorize)
		for _, call := range callsWithMetadata {
			fn.LocalCalls = append(fn.LocalCalls, call.FunctionName)
		}
	}

	// Ensure fields are never nil for JSON serialization
	if fn.Calls == nil {
		fn.Calls = []string{}
	}
	if fn.LocalCalls == nil {
		fn.LocalCalls = []string{}
	}
}

func (p *GoParser) extractType(node *ast.TypeSpec) models.TypeDef {
	typeDef := models.TypeDef{
		Name:     node.Name.Name,
		Fields:   []models.Field{},
		Methods:  []models.Method{},
		Embedded: []string{},
	}

	// Extract position information
	if node.Pos().IsValid() {
		pos := p.fset.Position(node.Pos())
		typeDef.StartLine = pos.Line
	}
	if node.End().IsValid() {
		pos := p.fset.Position(node.End())
		typeDef.EndLine = pos.Line
	}

	switch t := node.Type.(type) {
	case *ast.StructType:
		typeDef.Kind = kindStruct
		if t.Fields != nil {
			for _, field := range t.Fields.List {
				fieldType := p.typeToString(field.Type)
				tag := ""
				if field.Tag != nil {
					tag = field.Tag.Value
				}

				if len(field.Names) > 0 {
					// Named fields
					for _, name := range field.Names {
						typeDef.Fields = append(typeDef.Fields, models.Field{
							Name: name.Name,
							Type: fieldType,
							Tag:  tag,
						})
					}
				} else {
					// Embedded field
					typeDef.Embedded = append(typeDef.Embedded, fieldType)
				}
			}
		}
	case *ast.InterfaceType:
		typeDef.Kind = kindInterface
		if t.Methods != nil {
			for _, method := range t.Methods.List {
				if len(method.Names) > 0 {
					// Method
					for _, name := range method.Names {
						typeDef.Methods = append(typeDef.Methods, p.extractInterfaceMethod(name, method))
					}
				} else {
					// Embedded interface
					embeddedType := p.typeToString(method.Type)
					typeDef.Embedded = append(typeDef.Embedded, embeddedType)
				}
			}
		}
	default:
		typeDef.Kind = kindAlias
	}

	return typeDef
}

func (p *GoParser) extractVariable(name *ast.Ident, spec *ast.ValueSpec) models.Variable {
	typeName, startLine, endLine := p.extractValueSpecInfo(name, spec)
	return models.Variable{
		Name:      name.Name,
		Type:      typeName,
		StartLine: startLine,
		EndLine:   endLine,
	}
}

func (p *GoParser) extractConstant(name *ast.Ident, spec *ast.ValueSpec) models.Constant {
	typeName, startLine, endLine := p.extractValueSpecInfo(name, spec)
	return models.Constant{
		Name:      name.Name,
		Type:      typeName,
		StartLine: startLine,
		EndLine:   endLine,
	}
}

// extractValueSpecInfo extracts common information from a ValueSpec
func (p *GoParser) extractValueSpecInfo(name *ast.Ident, spec *ast.ValueSpec) (typeName string, startLine, endLine int) {
	if spec.Type != nil {
		typeName = p.typeToString(spec.Type)
	} else if len(spec.Values) > 0 {
		// Try to infer type from the value if no explicit type is given
		typeName = p.inferTypeFromValue(spec.Values[0])
	}

	// Extract position information
	if name.Pos().IsValid() {
		pos := p.fset.Position(name.Pos())
		startLine = pos.Line
		endLine = pos.Line
	}

	return typeName, startLine, endLine
}

func (p *GoParser) inferTypeFromValue(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		switch e.Kind {
		case token.INT:
			return typeInt
		case token.FLOAT:
			return typeFloat64
		case token.STRING:
			return typeString
		case token.CHAR:
			return typeRune
		case token.IMAG:
			return typeComplex128
		case token.ILLEGAL, token.EOF, token.COMMENT, token.IDENT,
			token.ADD, token.SUB, token.MUL, token.QUO, token.REM,
			token.AND, token.OR, token.XOR, token.SHL, token.SHR, token.AND_NOT,
			token.ADD_ASSIGN, token.SUB_ASSIGN, token.MUL_ASSIGN, token.QUO_ASSIGN,
			token.REM_ASSIGN, token.AND_ASSIGN, token.OR_ASSIGN, token.XOR_ASSIGN,
			token.SHL_ASSIGN, token.SHR_ASSIGN, token.AND_NOT_ASSIGN,
			token.LAND, token.LOR, token.ARROW, token.INC, token.DEC,
			token.EQL, token.LSS, token.GTR, token.ASSIGN, token.NOT,
			token.NEQ, token.LEQ, token.GEQ, token.DEFINE, token.ELLIPSIS,
			token.LPAREN, token.LBRACK, token.LBRACE, token.COMMA, token.PERIOD,
			token.RPAREN, token.RBRACK, token.RBRACE, token.SEMICOLON, token.COLON,
			token.BREAK, token.CASE, token.CHAN, token.CONST, token.CONTINUE,
			token.DEFAULT, token.DEFER, token.ELSE, token.FALLTHROUGH, token.FOR,
			token.FUNC, token.GO, token.GOTO, token.IF, token.IMPORT,
			token.INTERFACE, token.MAP, token.PACKAGE, token.RANGE, token.RETURN,
			token.SELECT, token.STRUCT, token.SWITCH, token.TYPE, token.VAR, token.TILDE:
			return ""
		}
	case *ast.Ident:
		// For identifiers like true/false
		if e.Name == "true" || e.Name == "false" {
			return typeBool
		}
	case *ast.BinaryExpr:
		// For complex expressions like 1 + 2i, we can't easily infer the type
		// without more sophisticated analysis, so return empty for now
		return ""
	}
	return ""
}

func (p *GoParser) extractInterfaceMethod(name *ast.Ident, field *ast.Field) models.Method {
	method := models.Method{
		Name:       name.Name,
		Parameters: []models.Parameter{},
		Returns:    []models.Type{},
	}

	// Extract position information
	if name.Pos().IsValid() {
		pos := p.fset.Position(name.Pos())
		method.StartLine = pos.Line
	}
	if name.End().IsValid() {
		pos := p.fset.Position(name.End())
		method.EndLine = pos.Line
	}

	if funcType, ok := field.Type.(*ast.FuncType); ok {
		// Extract parameters
		if funcType.Params != nil {
			for _, param := range funcType.Params.List {
				paramType := p.typeToString(param.Type)
				if len(param.Names) > 0 {
					for _, paramName := range param.Names {
						method.Parameters = append(method.Parameters, models.Parameter{
							Name: paramName.Name,
							Type: paramType,
						})
					}
				} else {
					method.Parameters = append(method.Parameters, models.Parameter{
						Name: "",
						Type: paramType,
					})
				}
			}
		}

		// Extract return types
		if funcType.Results != nil {
			for _, result := range funcType.Results.List {
				returnType := p.typeToString(result.Type)
				method.Returns = append(method.Returns, models.Type{
					Name: returnType,
					Kind: p.getTypeKind(returnType),
				})
			}
		}

		method.Signature = p.buildMethodSignature(name.Name, funcType)
	}

	return method
}

func (p *GoParser) typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return p.typeToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + p.typeToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + p.typeToString(t.Elt)
		}
		return "[" + p.exprToString(t.Len) + "]" + p.typeToString(t.Elt)
	case *ast.MapType:
		return "map[" + p.typeToString(t.Key) + "]" + p.typeToString(t.Value)
	case *ast.ChanType:
		switch t.Dir {
		case ast.SEND:
			return "chan<- " + p.typeToString(t.Value)
		case ast.RECV:
			return "<-chan " + p.typeToString(t.Value)
		default:
			return "chan " + p.typeToString(t.Value)
		}
	case *ast.FuncType:
		return "func" + p.buildFuncTypeSignature(t)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	default:
		return "unknown"
	}
}

func (p *GoParser) exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return e.Value
	case *ast.Ident:
		return e.Name
	default:
		return "..."
	}
}

func (p *GoParser) getTypeKind(typeName string) string {
	switch typeName {
	case typeBool, typeString, typeInt, "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"byte", typeRune, "float32", typeFloat64, "complex64", typeComplex128:
		return kindBasic
	case "error":
		return kindInterface
	default:
		if strings.HasPrefix(typeName, "[]") || strings.HasPrefix(typeName, "map[") ||
			strings.HasPrefix(typeName, "chan") || strings.HasPrefix(typeName, "func") {
			return kindComposite
		}
		if strings.HasPrefix(typeName, "*") {
			return kindPointer
		}
		return kindNamed
	}
}

func (p *GoParser) buildFunctionSignature(node *ast.FuncDecl) string {
	var sig strings.Builder
	sig.WriteString("func ")

	// Add receiver if it's a method
	if node.Recv != nil && len(node.Recv.List) > 0 {
		sig.WriteString("(")
		for i, recv := range node.Recv.List {
			if i > 0 {
				sig.WriteString(", ")
			}
			if len(recv.Names) > 0 {
				sig.WriteString(recv.Names[0].Name + " ")
			}
			sig.WriteString(p.typeToString(recv.Type))
		}
		sig.WriteString(") ")
	}

	sig.WriteString(node.Name.Name)
	sig.WriteString(p.buildFuncTypeSignature(node.Type))

	return sig.String()
}

func (p *GoParser) buildMethodSignature(name string, funcType *ast.FuncType) string {
	return name + p.buildFuncTypeSignature(funcType)
}

func (p *GoParser) buildFuncTypeSignature(funcType *ast.FuncType) string {
	var sig strings.Builder

	// Parameters
	sig.WriteString("(")
	if funcType.Params != nil {
		for i, param := range funcType.Params.List {
			if i > 0 {
				sig.WriteString(", ")
			}
			paramType := p.typeToString(param.Type)
			if len(param.Names) > 0 {
				for j, name := range param.Names {
					if j > 0 {
						sig.WriteString(", ")
					}
					sig.WriteString(name.Name + " " + paramType)
				}
			} else {
				sig.WriteString(paramType)
			}
		}
	}
	sig.WriteString(")")

	// Return types
	if funcType.Results != nil && len(funcType.Results.List) > 0 {
		sig.WriteString(" ")
		if len(funcType.Results.List) > 1 {
			sig.WriteString("(")
		}
		for i, result := range funcType.Results.List {
			if i > 0 {
				sig.WriteString(", ")
			}
			sig.WriteString(p.typeToString(result.Type))
		}
		if len(funcType.Results.List) > 1 {
			sig.WriteString(")")
		}
	}

	return sig.String()
}

func (p *GoParser) extractMethods(file *ast.File, ctx *models.FileContext) {
	// Find methods for each type
	for i := range ctx.Types {
		typeName := ctx.Types[i].Name

		// Look for methods with this type as receiver
		ast.Inspect(file, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
					// This is a method
					recv := funcDecl.Recv.List[0]
					receiverType := p.extractReceiverType(recv.Type)

					if receiverType == typeName {
						method := p.extractMethodFromFunc(funcDecl)
						ctx.Types[i].Methods = append(ctx.Types[i].Methods, method)
					}
				}
			}
			return true
		})
	}
}

func (p *GoParser) extractReceiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return p.extractReceiverType(t.X)
	default:
		return ""
	}
}

func (p *GoParser) extractMethodFromFunc(node *ast.FuncDecl) models.Method {
	method := models.Method{
		Name:       node.Name.Name,
		Parameters: []models.Parameter{},
		Returns:    []models.Type{},
	}

	// Extract position information
	if node.Pos().IsValid() {
		pos := p.fset.Position(node.Pos())
		method.StartLine = pos.Line
	}
	if node.End().IsValid() {
		pos := p.fset.Position(node.End())
		method.EndLine = pos.Line
	}

	// Extract parameters (same as function)
	if node.Type.Params != nil {
		for _, field := range node.Type.Params.List {
			paramType := p.typeToString(field.Type)
			if len(field.Names) > 0 {
				for _, name := range field.Names {
					method.Parameters = append(method.Parameters, models.Parameter{
						Name: name.Name,
						Type: paramType,
					})
				}
			} else {
				method.Parameters = append(method.Parameters, models.Parameter{
					Name: "",
					Type: paramType,
				})
			}
		}
	}

	// Extract return types
	if node.Type.Results != nil {
		for _, field := range node.Type.Results.List {
			returnType := p.typeToString(field.Type)
			method.Returns = append(method.Returns, models.Type{
				Name: returnType,
				Kind: p.getTypeKind(returnType),
			})
		}
	}

	// Build signature
	method.Signature = p.buildFunctionSignature(node)

	return method
}

// extractFunctionCalls analyzes a function body to find all function calls
func (p *GoParser) extractFunctionCalls(body *ast.BlockStmt) []string {
	var calls []string
	callMap := make(map[string]bool) // To avoid duplicates

	ast.Inspect(body, func(n ast.Node) bool {
		if node, ok := n.(*ast.CallExpr); ok {
			callName := p.extractCallName(node.Fun)
			if callName != "" && !callMap[callName] {
				calls = append(calls, callName)
				callMap[callName] = true
			}
		}
		return true
	})

	return calls
}

// extractCallName extracts the function name from a call expression
func (p *GoParser) extractCallName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		// Simple function call: foo()
		return e.Name
	case *ast.SelectorExpr:
		// Method call or package function: obj.Method() or pkg.Func()
		if x, ok := e.X.(*ast.Ident); ok {
			return x.Name + "." + e.Sel.Name
		}
		// For more complex expressions, just return the selector
		return e.Sel.Name
	case *ast.FuncLit:
		// Anonymous function call
		return "<anonymous>"
	default:
		return ""
	}
}

// extractFunctionCallsWithMetadata analyzes a function body to find all function calls with metadata
func (p *GoParser) extractFunctionCallsWithMetadata(body *ast.BlockStmt) []models.CallReference {
	var calls []models.CallReference
	callMap := make(map[string]models.CallReference) // Deduplicate by name but keep metadata

	ast.Inspect(body, func(n ast.Node) bool {
		if callExpr, ok := n.(*ast.CallExpr); ok {
			callName := p.extractCallName(callExpr.Fun)
			if callName != "" {
				pos := p.fset.Position(callExpr.Pos())
				callType := p.classifyCallType(callExpr)

				callRef := models.CallReference{
					FunctionName: callName,
					File:         "", // Will be set during enrichment
					Line:         pos.Line,
					CallType:     callType,
				}

				// Store the call (will overwrite duplicates with potentially better metadata)
				callMap[callName] = callRef
			}
		}
		return true
	})

	// Convert map to slice
	for _, call := range callMap {
		calls = append(calls, call)
	}

	return calls
}

// classifyCallType determines the type of call (function, method, external, etc.)
func (p *GoParser) classifyCallType(callExpr *ast.CallExpr) string {
	switch expr := callExpr.Fun.(type) {
	case *ast.Ident:
		// Simple function call: foo()
		return models.CallTypeFunction
	case *ast.SelectorExpr:
		// Method call or package function: obj.Method() or pkg.Func()
		if x, ok := expr.X.(*ast.Ident); ok {
			// Check if it's a package call (like fmt.Println) or method call
			if p.isPackageCall(x.Name) {
				return models.CallTypeExternal
			}
			return models.CallTypeMethod
		}
		return models.CallTypeMethod
	case *ast.FuncLit:
		// Anonymous function call
		return models.CallTypeComplex
	default:
		return models.CallTypeComplex
	}
}

// isPackageCall checks if the identifier represents a package name
func (p *GoParser) isPackageCall(name string) bool {
	// Common Go standard library packages and patterns
	commonPackages := map[string]bool{
		"fmt":      true,
		"os":       true,
		"io":       true,
		"net":      true,
		"http":     true,
		"json":     true,
		"time":     true,
		"strings":  true,
		"strconv":  true,
		"errors":   true,
		"context":  true,
		"sync":     true,
		"log":      true,
		"math":     true,
		"sort":     true,
		"bytes":    true,
		"path":     true,
		"url":      true,
		"crypto":   true,
		"hash":     true,
		"encoding": true,
		"reflect":  true,
		"runtime":  true,
		"testing":  true,
	}

	return commonPackages[name]
}

// buildCallGraph builds the CalledBy relationships between functions (both deprecated and enhanced fields)
func (p *GoParser) buildCallGraph(ctx *models.FileContext) {
	// Create a map of function names to their indices for quick lookup
	funcMap := make(map[string]int)
	for i := range ctx.Functions {
		funcMap[ctx.Functions[i].Name] = i
	}

	// For each function, update caller relationships for functions it calls
	for i := range ctx.Functions {
		caller := &ctx.Functions[i]

		// Process calls from LocalCalls (enhanced field) to populate LocalCallers
		for _, calledName := range caller.LocalCalls {
			if targetIdx, exists := funcMap[calledName]; exists {
				// Add to deprecated CalledBy field (backward compatibility)
				p.addToCalledBy(&ctx.Functions[targetIdx], caller.Name)

				// Add to enhanced LocalCallers field
				p.addToLocalCallers(&ctx.Functions[targetIdx], caller.Name)
			}
		}

		// Also process deprecated Calls field for backward compatibility
		for _, calledName := range caller.Calls {
			// Handle simple function names and method calls
			targetName := calledName
			if strings.Contains(calledName, ".") {
				// For method calls like "fmt.Println", we might want to track just "Println"
				// or handle package calls differently. For now, keep the full name.
				parts := strings.Split(calledName, ".")
				const expectedParts = 2
				if len(parts) == expectedParts {
					// Check if it's a method call on a local type
					targetName = parts[1]
				}
			}

			if targetIdx, exists := funcMap[targetName]; exists {
				// Add to deprecated CalledBy field (backward compatibility)
				p.addToCalledBy(&ctx.Functions[targetIdx], caller.Name)
			}
		}
	}
}

// addToCalledBy adds a caller to the deprecated CalledBy field if not already present
func (p *GoParser) addToCalledBy(target *models.Function, callerName string) {
	// Check if already exists in CalledBy using slices.Contains
	if !slices.Contains(target.CalledBy, callerName) {
		target.CalledBy = append(target.CalledBy, callerName)
	}
}

// addToLocalCallers adds a caller to the enhanced LocalCallers field if not already present
func (p *GoParser) addToLocalCallers(target *models.Function, callerName string) {
	// Check if already exists in LocalCallers using slices.Contains
	if !slices.Contains(target.LocalCallers, callerName) {
		target.LocalCallers = append(target.LocalCallers, callerName)
	}
}

// extractExports extracts exported symbols and populates the Exports array
func (p *GoParser) extractExports(ctx *models.FileContext) {
	// Extract exported functions
	for i := range ctx.Functions {
		fn := &ctx.Functions[i]
		if isExported(fn.Name) {
			ctx.Exports = append(ctx.Exports, models.Export{
				Name: fn.Name,
				Type: fn.Signature,
				Kind: "function",
			})
		}
	}

	// Extract exported types
	for _, typ := range ctx.Types {
		if isExported(typ.Name) {
			ctx.Exports = append(ctx.Exports, models.Export{
				Name: typ.Name,
				Type: typ.Kind,
				Kind: "type",
			})
		}
	}

	// Extract exported variables
	for _, variable := range ctx.Variables {
		if isExported(variable.Name) {
			ctx.Exports = append(ctx.Exports, models.Export{
				Name: variable.Name,
				Type: variable.Type,
				Kind: "variable",
			})
		}
	}

	// Extract exported constants
	for _, constant := range ctx.Constants {
		if isExported(constant.Name) {
			ctx.Exports = append(ctx.Exports, models.Export{
				Name: constant.Name,
				Type: constant.Type,
				Kind: "constant",
			})
		}
	}
}

// isExported checks if a Go identifier is exported (starts with uppercase letter)
func isExported(name string) bool {
	if name == "" {
		return false
	}
	// In Go, exported identifiers start with an uppercase letter
	// We need to properly decode the first rune from UTF-8
	for _, r := range name {
		return unicode.IsUpper(r)
	}
	return false
}
