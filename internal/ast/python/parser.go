package python

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"repository-context-protocol/internal/models"
)

const (
	languagePython    = "python"
	versionPython3    = "python3"
	extensionPython   = ".py"
	osWindows         = "windows"
	windowsExecutable = "py"
)

// PythonExtractorOutput represents the JSON output from the Python extractor
type PythonExtractorOutput struct {
	Path      string               `json:"path"`
	Language  string               `json:"language"`
	Functions []PythonFunctionInfo `json:"functions"`
	Types     []PythonClassInfo    `json:"types"`
	Variables []PythonVariableInfo `json:"variables"`
	Constants []PythonVariableInfo `json:"constants"`
	Imports   []PythonImportInfo   `json:"imports"`
	Exports   []PythonExportInfo   `json:"exports"`
	Errors    []string             `json:"errors"`
}

type PythonFunctionInfo struct {
	Name       string                `json:"name"`
	Parameters []PythonParameterInfo `json:"parameters"`
	Returns    []PythonTypeInfo      `json:"returns"`
	Calls      []PythonCallInfo      `json:"calls"`
	CalledBy   []string              `json:"called_by"`
	StartLine  int                   `json:"start_line"`
	EndLine    int                   `json:"end_line"`
	Decorators []string              `json:"decorators"`
	IsAsync    bool                  `json:"is_async"`
	Docstring  string                `json:"docstring"`
}

type PythonParameterInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Default string `json:"default,omitempty"`
}

type PythonTypeInfo struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type PythonCallInfo struct {
	Name string `json:"name"`
	Line int    `json:"line"`
	Type string `json:"type"`
}

type PythonClassInfo struct {
	Name       string               `json:"name"`
	Kind       string               `json:"kind"`
	Fields     []PythonFieldInfo    `json:"fields"`
	Methods    []PythonFunctionInfo `json:"methods"`
	Embedded   []string             `json:"embedded"`
	StartLine  int                  `json:"start_line"`
	EndLine    int                  `json:"end_line"`
	Decorators []string             `json:"decorators"`
	Docstring  string               `json:"docstring"`
}

type PythonFieldInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Line int    `json:"line"`
}

type PythonVariableInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Line       int    `json:"line"`
	IsExported bool   `json:"is_exported"`
}

type PythonImportInfo struct {
	Path         string   `json:"path"`
	Alias        string   `json:"alias"`
	Items        []string `json:"items"`
	Line         int      `json:"line"`
	IsStarImport bool     `json:"is_star_import"`
}

type PythonExportInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Line int    `json:"line"`
}

// PythonParser implements the LanguageParser interface for Python files
type PythonParser struct {
	mu            sync.RWMutex
	pythonPath    string
	extractorPath string
}

// NewPythonParser creates a new Python parser instance
func NewPythonParser() *PythonParser {
	parser := &PythonParser{
		pythonPath:    versionPython3, // Default Python executable
		extractorPath: "",             // Will be determined dynamically
	}

	// Try to detect Python installation
	if err := parser.validatePythonSetup(); err != nil {
		// For now, just log and continue with default - we'll improve error handling later
		// This allows basic interface tests to pass
		log.Printf("Error validating Python setup: %v", err)
	}

	return parser
}

// GetSupportedExtensions returns the file extensions supported by this parser
func (p *PythonParser) GetSupportedExtensions() []string {
	return []string{extensionPython}
}

// GetLanguageName returns the name of the language this parser handles
func (p *PythonParser) GetLanguageName() string {
	return languagePython
}

// ParseFile parses a Python file and returns a FileContext
func (p *PythonParser) ParseFile(path string, content []byte) (*models.FileContext, error) {
	// Ensure Python is available and paths are set
	if err := p.ensureInitialized(); err != nil {
		return nil, fmt.Errorf("parser initialization failed: %w", err)
	}

	// Execute the Python extractor
	extractedData, err := p.executeExtractor(path, content)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Python extractor: %w", err)
	}

	// Parse the JSON output into Go models
	fileContext, err := p.parseJSON(extractedData, path, content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extractor output: %w", err)
	}

	return fileContext, nil
}

// ensureInitialized ensures both Python setup and extractor path are configured with thread safety
func (p *PythonParser) ensureInitialized() error {
	// First, check if we're already initialized (read lock)
	p.mu.RLock()
	initialized := p.pythonPath != "" && p.extractorPath != ""
	p.mu.RUnlock()

	if initialized {
		return nil
	}

	// We need to initialize, acquire write lock
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if p.pythonPath != "" && p.extractorPath != "" {
		return nil
	}

	// Validate Python setup if not done yet
	if p.pythonPath == "" {
		if err := p.validatePythonSetupLocked(); err != nil {
			return fmt.Errorf("python setup validation failed: %w", err)
		}
	}

	// Set extractor path if not done yet
	if p.extractorPath == "" {
		if err := p.setExtractorPathLocked(); err != nil {
			return fmt.Errorf("failed to locate Python extractor: %w", err)
		}
	}

	return nil
}

// setExtractorPathLocked determines the path to the Python extractor script (assumes write lock held)
func (p *PythonParser) setExtractorPathLocked() error {
	// Get the directory of the current Go file
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to determine current file location")
	}

	// The extractor should be in the same directory as this Go file
	extractorPath := filepath.Join(filepath.Dir(currentFile), "extractor.py")

	// Check if the extractor exists
	if _, err := os.Stat(extractorPath); err != nil {
		return fmt.Errorf("python extractor not found at %s: %w", extractorPath, err)
	}

	p.extractorPath = extractorPath
	return nil
}

// executeExtractor runs the Python extractor script and returns the JSON output
func (p *PythonParser) executeExtractor(_ string, content []byte) ([]byte, error) {
	// Get paths with read lock
	p.mu.RLock()
	pythonPath := p.pythonPath
	extractorPath := p.extractorPath
	p.mu.RUnlock()

	// Create command to run Python extractor
	// Don't pass the file path as argument, use stdin instead
	// #nosec G204 - pythonPath and extractorPath are controlled internally and validated
	cmd := exec.Command(pythonPath, extractorPath)

	// Always use content via stdin for consistency
	cmd.Stdin = bytes.NewReader(content)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("python extractor failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// parseJSON converts the Python extractor JSON output to Go models
func (p *PythonParser) parseJSON(data []byte, path string, content []byte) (*models.FileContext, error) {
	var pythonOutput PythonExtractorOutput
	if err := json.Unmarshal(data, &pythonOutput); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Python extractor output: %w", err)
	}

	// Check for extraction errors
	if len(pythonOutput.Errors) > 0 {
		return nil, fmt.Errorf("python extractor errors: %v", pythonOutput.Errors)
	}

	// Calculate checksum and modification time
	hash := sha256.Sum256(content)
	checksum := fmt.Sprintf("%x", hash)

	var modTime time.Time
	if fileInfo, err := os.Stat(path); err == nil {
		modTime = fileInfo.ModTime()
	} else {
		modTime = time.Now()
	}

	// Create FileContext with the correct path (since extractor gets stdin, not file path)
	fileContext := &models.FileContext{
		Path:      path, // Use the actual path passed to ParseFile
		Language:  languagePython,
		Checksum:  checksum,
		ModTime:   modTime,
		Functions: p.convertFunctions(pythonOutput.Functions),
		Types:     p.convertTypes(pythonOutput.Types),
		Variables: p.convertVariables(pythonOutput.Variables),
		Constants: p.convertConstants(pythonOutput.Constants),
		Imports:   p.convertImports(pythonOutput.Imports),
		Exports:   p.convertExports(pythonOutput.Exports),
	}

	return fileContext, nil
}

// convertFunctions converts Python function info to Go models
func (p *PythonParser) convertFunctions(pythonFunctions []PythonFunctionInfo) []models.Function {
	functions := make([]models.Function, len(pythonFunctions))

	for i := range pythonFunctions {
		pFunc := &pythonFunctions[i]
		function := models.Function{
			Name:      pFunc.Name,
			StartLine: pFunc.StartLine,
			EndLine:   pFunc.EndLine,
			Calls:     p.extractCallNames(pFunc.Calls),
			CalledBy:  pFunc.CalledBy,
		}

		// Convert parameters
		for _, param := range pFunc.Parameters {
			parameter := models.Parameter{
				Name: param.Name,
				Type: param.Type,
			}
			function.Parameters = append(function.Parameters, parameter)
		}

		// Convert return types
		for _, ret := range pFunc.Returns {
			returnType := models.Type{
				Name: ret.Name,
				Kind: ret.Kind,
			}
			function.Returns = append(function.Returns, returnType)
		}

		functions[i] = function
	}

	return functions
}

// convertTypes converts Python class info to Go models
func (p *PythonParser) convertTypes(pythonTypes []PythonClassInfo) []models.TypeDef {
	types := make([]models.TypeDef, len(pythonTypes))

	for i := range pythonTypes {
		pType := &pythonTypes[i]
		typeDef := models.TypeDef{
			Name:      pType.Name,
			Kind:      pType.Kind,
			StartLine: pType.StartLine,
			EndLine:   pType.EndLine,
			Embedded:  pType.Embedded,
		}

		// Convert fields
		for _, field := range pType.Fields {
			modelField := models.Field{
				Name: field.Name,
				Type: field.Type,
			}
			typeDef.Fields = append(typeDef.Fields, modelField)
		}

		// Convert methods
		for j := range pType.Methods {
			method := &pType.Methods[j]
			modelMethod := models.Method{
				Name:      method.Name,
				Signature: p.buildMethodSignature(method),
				StartLine: method.StartLine,
				EndLine:   method.EndLine,
			}

			// Convert method parameters
			for _, param := range method.Parameters {
				parameter := models.Parameter{
					Name: param.Name,
					Type: param.Type,
				}
				modelMethod.Parameters = append(modelMethod.Parameters, parameter)
			}

			// Convert method return types
			for _, ret := range method.Returns {
				returnType := models.Type{
					Name: ret.Name,
					Kind: ret.Kind,
				}
				modelMethod.Returns = append(modelMethod.Returns, returnType)
			}

			typeDef.Methods = append(typeDef.Methods, modelMethod)
		}

		types[i] = typeDef
	}

	return types
}

// convertVariables converts Python variable info to Go models
func (p *PythonParser) convertVariables(pythonVars []PythonVariableInfo) []models.Variable {
	variables := make([]models.Variable, len(pythonVars))

	for i, pVar := range pythonVars {
		variables[i] = models.Variable{
			Name: pVar.Name,
			Type: pVar.Type,
		}
	}

	return variables
}

// convertConstants converts Python constant info to Go models
func (p *PythonParser) convertConstants(pythonConsts []PythonVariableInfo) []models.Constant {
	constants := make([]models.Constant, len(pythonConsts))

	for i, pConst := range pythonConsts {
		constants[i] = models.Constant{
			Name: pConst.Name,
			Type: pConst.Type,
		}
	}

	return constants
}

// convertImports converts Python import info to Go models
func (p *PythonParser) convertImports(pythonImports []PythonImportInfo) []models.Import {
	imports := make([]models.Import, len(pythonImports))

	for i, pImport := range pythonImports {
		imports[i] = models.Import{
			Path:  pImport.Path,
			Alias: pImport.Alias,
		}
	}

	return imports
}

// convertExports converts Python export info to Go models
func (p *PythonParser) convertExports(pythonExports []PythonExportInfo) []models.Export {
	exports := make([]models.Export, len(pythonExports))

	for i, pExport := range pythonExports {
		exports[i] = models.Export{
			Name: pExport.Name,
			Type: pExport.Type,
		}
	}

	return exports
}

// extractCallNames extracts just the function names from call info
func (p *PythonParser) extractCallNames(calls []PythonCallInfo) []string {
	names := make([]string, len(calls))
	for i, call := range calls {
		names[i] = call.Name
	}
	return names
}

// buildMethodSignature creates a method signature string
func (p *PythonParser) buildMethodSignature(method *PythonFunctionInfo) string {
	var parts []string

	// Add parameters
	for _, param := range method.Parameters {
		parts = append(parts, fmt.Sprintf("%s: %s", param.Name, param.Type))
	}

	paramStr := strings.Join(parts, ", ")

	// Add return type - handle multiple return types
	returnStr := "None"
	if len(method.Returns) > 0 {
		if len(method.Returns) == 1 {
			returnStr = method.Returns[0].Name
		} else {
			// Multiple return types - format as Union or Tuple depending on context
			var returnTypes []string
			for _, ret := range method.Returns {
				returnTypes = append(returnTypes, ret.Name)
			}

			// If all return types are the same, just use one
			if p.allReturnTypesSame(method.Returns) {
				returnStr = method.Returns[0].Name
			} else {
				// Format as Union for multiple different types
				returnStr = fmt.Sprintf("Union[%s]", strings.Join(returnTypes, ", "))
			}
		}
	}

	return fmt.Sprintf("(%s) -> %s", paramStr, returnStr)
}

// allReturnTypesSame checks if all return types in the slice are identical
func (p *PythonParser) allReturnTypesSame(returns []PythonTypeInfo) bool {
	if len(returns) <= 1 {
		return true
	}

	first := returns[0].Name
	for _, ret := range returns[1:] {
		if ret.Name != first {
			return false
		}
	}
	return true
}

// validatePythonSetup checks if Python is available and accessible
func (p *PythonParser) validatePythonSetup() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.validatePythonSetupLocked()
}

// validatePythonSetupLocked checks if Python is available and accessible (assumes write lock held)
func (p *PythonParser) validatePythonSetupLocked() error {
	// Try python3 first
	if err := p.checkPythonExecutable(versionPython3); err == nil {
		p.pythonPath = versionPython3
		return nil
	}

	// Try python as fallback
	if err := p.checkPythonExecutable(languagePython); err == nil {
		p.pythonPath = languagePython
		return nil
	}

	// Try platform-specific paths
	if runtime.GOOS == osWindows {
		if err := p.checkPythonExecutable(windowsExecutable); err == nil {
			p.pythonPath = windowsExecutable
			return nil
		}
	}

	return fmt.Errorf("python executable not found - please install Python 3.8+ and ensure it's in PATH")
}

// checkPythonExecutable verifies if the given Python executable exists and works
func (p *PythonParser) checkPythonExecutable(executable string) error {
	cmd := exec.Command(executable, "--version")
	return cmd.Run()
}
