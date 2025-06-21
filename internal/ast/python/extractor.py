#!/usr/bin/env python3
"""
Python AST extractor for repository-context-protocol.

This script parses Python source code and extracts structured information
about functions, classes, variables, imports, and call relationships
for consumption by the Go-based parser.
"""

import ast
import json
import sys
import traceback
from typing import Dict, Any, Optional


class PythonASTExtractor(ast.NodeVisitor):
    def __init__(self, source_code: str, file_path: str = ""):
        self.source_code = source_code
        self.file_path = file_path
        self.lines = source_code.split("\n")
        self.functions = []
        self.classes = []
        self.variables = []
        self.constants = []
        self.imports = []
        self.exports = []
        self.current_class = None
        self.call_stack = []
        self.scope_stack = ["module"]  # Track current scope for variable resolution

    def extract(self) -> Dict[str, Any]:
        """Extract all Python AST information into a structured format."""
        try:
            tree = ast.parse(self.source_code)
            self.visit(tree)

            # Build call graph relationships
            self._build_call_graph()

            # Extract exports (public API)
            self._extract_exports()

            return {
                "path": self.file_path,
                "language": "python",
                "functions": self.functions,
                "types": self.classes,  # Map to 'types' for Go compatibility
                "variables": self.variables,
                "constants": self.constants,
                "imports": self.imports,
                "exports": self.exports,
                "errors": [],
            }
        except Exception as e:
            return {
                "path": self.file_path,
                "language": "python",
                "functions": [],
                "types": [],
                "variables": [],
                "constants": [],
                "imports": [],
                "exports": [],
                "errors": [f"Parse error: {str(e)}"],
            }

    def visit_FunctionDef(self, node: ast.FunctionDef):
        func_info = self._extract_function(node)

        if self.current_class:
            # This is a method
            func_info["is_method"] = True
            func_info["class_name"] = self.current_class["name"]
            self.current_class["methods"].append(func_info)
        else:
            # This is a standalone function
            func_info["is_method"] = False
            self.functions.append(func_info)

        # Visit function body to find calls
        old_stack = self.call_stack[:]
        self.call_stack.append(func_info["name"])
        self.generic_visit(node)
        self.call_stack = old_stack

    def visit_AsyncFunctionDef(self, node: ast.AsyncFunctionDef):
        func_info = self._extract_function(node)
        func_info["is_async"] = True

        if self.current_class:
            func_info["is_method"] = True
            func_info["class_name"] = self.current_class["name"]
            self.current_class["methods"].append(func_info)
        else:
            self.functions.append(func_info)

        old_stack = self.call_stack[:]
        self.call_stack.append(func_info["name"])
        self.generic_visit(node)
        self.call_stack = old_stack

    def _extract_function(self, node) -> Dict[str, Any]:
        """Extract function information with Go model compatibility."""
        # Extract function parameters with defaults
        parameters = []
        defaults = node.args.defaults
        num_defaults = len(defaults)
        num_args = len(node.args.args)

        for i, arg in enumerate(node.args.args):
            param_info = {
                "name": arg.arg,
                "type": self._normalize_type(ast.unparse(arg.annotation))
                if arg.annotation
                else "Any",
            }

            # Check if this parameter has a default value
            default_index = i - (num_args - num_defaults)
            if default_index >= 0:
                try:
                    param_info["default"] = ast.unparse(defaults[default_index])
                except:  # noqa: E722
                    param_info["default"] = "None"

            parameters.append(param_info)

        # Handle *args and **kwargs
        if node.args.vararg:
            parameters.append({"name": f"*{node.args.vararg.arg}", "type": "tuple"})
        if node.args.kwarg:
            parameters.append({"name": f"**{node.args.kwarg.arg}", "type": "dict"})

        # Extract return type
        returns = []
        if node.returns:
            returns.append(
                {
                    "name": self._normalize_type(ast.unparse(node.returns)),
                    "kind": "builtin",
                }
            )
        else:
            returns.append({"name": "None", "kind": "builtin"})

        # Extract decorators
        decorators = []
        for decorator in node.decorator_list:
            decorators.append(ast.unparse(decorator))

        return {
            "name": node.name,
            "parameters": parameters,
            "returns": returns,
            "calls": [],  # Will be populated by visit_Call
            "called_by": [],
            "start_line": node.lineno,
            "end_line": node.end_lineno or node.lineno,
            "decorators": decorators,
            "is_async": isinstance(node, ast.AsyncFunctionDef),
            "docstring": ast.get_docstring(node) or "",
        }

    def visit_ClassDef(self, node: ast.ClassDef):
        """Extract class information with Go model compatibility."""
        class_info = {
            "name": node.name,
            "kind": "class",
            "fields": [],
            "methods": [],
            "embedded": [ast.unparse(base) for base in node.bases],
            "start_line": node.lineno,
            "end_line": node.end_lineno or node.lineno,
            "decorators": [ast.unparse(d) for d in node.decorator_list],
            "docstring": ast.get_docstring(node) or "",
        }

        old_class = self.current_class
        self.current_class = class_info

        # Enter class scope
        self.scope_stack.append(f"class:{node.name}")

        self.generic_visit(node)

        # Exit class scope
        self.scope_stack.pop()
        self.current_class = old_class
        self.classes.append(class_info)

    def visit_Call(self, node: ast.Call):
        if self.call_stack:  # We're inside a function
            current_func = self.call_stack[-1]
            call_name = self._extract_call_name(node)
            if call_name:
                # Find the function we're currently in and add this call
                self._add_call_to_current_function(
                    current_func,
                    {
                        "name": call_name,
                        "line": node.lineno,
                        "type": self._classify_call(node),
                    },
                )

        self.generic_visit(node)

    def _extract_call_name(self, node: ast.Call) -> Optional[str]:
        try:
            if isinstance(node.func, ast.Name):
                return node.func.id
            elif isinstance(node.func, ast.Attribute):
                return ast.unparse(node.func)
            else:
                return ast.unparse(node.func)
        except:  # noqa: E722
            return None

    def _classify_call(self, node: ast.Call) -> str:
        """Classify the type of call (local, method, external, etc.)"""
        if isinstance(node.func, ast.Name):
            return "function"
        elif isinstance(node.func, ast.Attribute):
            if isinstance(node.func.value, ast.Name) and node.func.value.id == "self":
                return "method"
            else:
                return "attribute"
        return "complex"

    def visit_Import(self, node: ast.Import):
        for alias in node.names:
            self.imports.append(
                {
                    "path": alias.name,
                    "alias": alias.asname or "",
                    "items": [],
                    "line": node.lineno,
                    "is_star_import": False,
                }
            )

    def visit_ImportFrom(self, node: ast.ImportFrom):
        module = node.module or ""
        for alias in node.names:
            self.imports.append(
                {
                    "path": module,
                    "alias": alias.asname or "",
                    "items": [alias.name] if alias.name != "*" else [],
                    "line": node.lineno,
                    "is_star_import": alias.name == "*",
                }
            )

    def visit_Assign(self, node: ast.Assign):
        """Extract variable assignments."""
        if len(self.scope_stack) == 1:  # Module level
            for target in node.targets:
                if isinstance(target, ast.Name):
                    var_info = self._extract_variable(target, node)
                    if self._is_constant(target.id):
                        self.constants.append(var_info)
                    else:
                        self.variables.append(var_info)
        self.generic_visit(node)

    def visit_AnnAssign(self, node: ast.AnnAssign):
        """Extract annotated assignments (type hints)."""
        if len(self.scope_stack) == 1 and isinstance(
            node.target, ast.Name
        ):  # Module level
            var_info = self._extract_annotated_variable(node)
            if self._is_constant(node.target.id):
                self.constants.append(var_info)
            else:
                self.variables.append(var_info)
        self.generic_visit(node)

    def _extract_variable(self, target: ast.Name, node: ast.Assign) -> Dict[str, Any]:
        """Extract variable information from assignment."""
        var_type = "Any"
        if node.value:
            var_type = self._infer_type(node.value)

        return {
            "name": target.id,
            "type": var_type,
            "line": node.lineno,
            "is_exported": not target.id.startswith("_"),
        }

    def _extract_annotated_variable(self, node: ast.AnnAssign) -> Dict[str, Any]:
        """Extract variable information from annotated assignment."""
        var_type = ast.unparse(node.annotation) if node.annotation else "Any"
        var_name = node.target.id if isinstance(node.target, ast.Name) else "unknown"

        return {
            "name": var_name,
            "type": self._normalize_type(var_type),
            "line": node.lineno,
            "is_exported": not var_name.startswith("_"),
        }

    def _is_constant(self, name: str) -> bool:
        """Determine if a variable name represents a constant."""
        return name.isupper() or name.startswith("_") and name[1:].isupper()

    def _infer_type(self, node: ast.AST) -> str:
        """Infer Python type from AST node."""
        if isinstance(node, ast.Constant):
            return type(node.value).__name__
        elif isinstance(node, (ast.List, ast.ListComp)):
            return "list"
        elif isinstance(node, (ast.Dict, ast.DictComp)):
            return "dict"
        elif isinstance(node, (ast.Set, ast.SetComp)):
            return "set"
        elif isinstance(node, ast.Tuple):
            return "tuple"
        elif isinstance(node, ast.Call):
            func_name = self._extract_call_name(node)
            if func_name in [
                "int",
                "float",
                "str",
                "bool",
                "list",
                "dict",
                "set",
                "tuple",
            ]:
                return func_name
            return "Any"
        return "Any"

    def _normalize_type(self, type_str: str) -> str:
        """Normalize Python type annotations to Go-compatible types."""
        type_mapping = {
            "int": "int",
            "float": "float64",
            "str": "string",
            "bool": "bool",
            "bytes": "[]byte",
            "bytearray": "[]byte",
            "list": "[]interface{}",
            "dict": "map[string]interface{}",
            "set": "map[interface{}]bool",
            "frozenset": "map[interface{}]bool",
            "tuple": "[]interface{}",
            "None": "nil",
            "Any": "interface{}",
            "List": "[]interface{}",
            "Dict": "map[string]interface{}",
            "Optional": "*interface{}",
            "Union": "interface{}",
        }

        # Handle generic types
        for python_type, go_type in type_mapping.items():
            if type_str.startswith(python_type):
                return go_type

        return type_str  # Return as-is if no mapping found

    def _add_call_to_current_function(self, func_name: str, call_info: Dict[str, Any]):
        """Add a call to the current function's call list."""
        # Find the current function and add the call
        for func in reversed(self.functions):
            if func["name"] == func_name:
                func["calls"].append(call_info)
                break

        # Also check methods in current class
        if self.current_class:
            for method in self.current_class["methods"]:
                if method["name"] == func_name:
                    method["calls"].append(call_info)
                    break

    def _build_call_graph(self):
        """Build call graph relationships between functions."""
        all_functions = self.functions[:]
        for cls in self.classes:
            all_functions.extend(cls["methods"])

        # Create a map of function names for quick lookup
        func_names = {func["name"] for func in all_functions}

        # For each function, analyze its calls
        for func in all_functions:
            func_calls = []
            called_by = []

            for call in func.get("calls", []):
                call_name = call["name"]
                if "." not in call_name and call_name in func_names:
                    func_calls.append(call_name)

            func["calls_functions"] = func_calls
            func["called_by"] = called_by  # Will be populated in second pass

        # Second pass: populate called_by relationships
        for func in all_functions:
            for called_func_name in func.get("calls_functions", []):
                for target_func in all_functions:
                    if target_func["name"] == called_func_name:
                        if "called_by" not in target_func:
                            target_func["called_by"] = []
                        target_func["called_by"].append(func["name"])

    def _extract_exports(self):
        """Extract public API elements (exports)."""
        # Functions that don't start with underscore are exported
        for func in self.functions:
            if not func["name"].startswith("_"):
                self.exports.append(
                    {
                        "name": func["name"],
                        "type": "function",
                        "line": func["start_line"],
                    }
                )

        # Classes that don't start with underscore are exported
        for cls in self.classes:
            if not cls["name"].startswith("_"):
                self.exports.append(
                    {"name": cls["name"], "type": "class", "line": cls["start_line"]}
                )

        # Variables and constants that don't start with underscore are exported
        for var in self.variables:
            if var.get("is_exported", False):
                self.exports.append(
                    {"name": var["name"], "type": "variable", "line": var["line"]}
                )

        for const in self.constants:
            if const.get("is_exported", False):
                self.exports.append(
                    {"name": const["name"], "type": "constant", "line": const["line"]}
                )


def main():
    """Main entry point for the Python AST extractor."""
    try:
        # Get file path from command line argument or use stdin
        file_path = ""
        if len(sys.argv) > 1:
            file_path = sys.argv[1]
            with open(file_path, "r", encoding="utf-8") as f:
                source_code = f.read()
        else:
            source_code = sys.stdin.read()

        extractor = PythonASTExtractor(source_code, file_path)
        result = extractor.extract()
        print(json.dumps(result, indent=2))

    except FileNotFoundError as e:
        error_result = {
            "path": file_path,
            "language": "python",
            "functions": [],
            "types": [],
            "variables": [],
            "constants": [],
            "imports": [],
            "exports": [],
            "errors": [f"File not found: {str(e)}"],
        }
        print(json.dumps(error_result, indent=2))
        sys.exit(1)

    except Exception as e:
        error_result = {
            "path": file_path,
            "language": "python",
            "functions": [],
            "types": [],
            "variables": [],
            "constants": [],
            "imports": [],
            "exports": [],
            "errors": [f"Fatal error: {str(e)}\n{traceback.format_exc()}"],
        }
        print(json.dumps(error_result, indent=2))
        sys.exit(1)


if __name__ == "__main__":
    main()
