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

        # Skip 'self' parameter for methods (when we're inside a class)
        start_index = 0
        if self.current_class and num_args > 0 and node.args.args[0].arg == "self":
            start_index = 1

        for i, arg in enumerate(node.args.args[start_index:], start=start_index):
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
            # *args should be typed as a tuple of the annotation or Any
            vararg_type = "tuple"
            if node.args.vararg.annotation:
                vararg_type = self._normalize_type(
                    ast.unparse(node.args.vararg.annotation)
                )
            parameters.append({"name": f"*{node.args.vararg.arg}", "type": vararg_type})
        if node.args.kwarg:
            # **kwargs should be typed as a dict of the annotation or Any
            kwarg_type = "dict"
            if node.args.kwarg.annotation:
                kwarg_type = self._normalize_type(
                    ast.unparse(node.args.kwarg.annotation)
                )
            parameters.append({"name": f"**{node.args.kwarg.arg}", "type": kwarg_type})

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
        # Handle relative imports by preserving the level (number of dots)
        module = node.module or ""
        if node.level > 0:
            # Add the appropriate number of dots for relative imports
            module = "." * node.level + module

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
            value_type = type(node.value).__name__
            # Handle special constant types
            if value_type == "NoneType":
                return "None"
            elif value_type in ("bytes", "bytearray"):
                return value_type
            return value_type
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
                "bytes",
                "bytearray",
                "list",
                "dict",
                "set",
                "frozenset",
                "tuple",
                "complex",
                "range",
                "enumerate",
                "zip",
                "filter",
                "map",
                "slice",
                "object",
                "type",
            ]:
                return func_name
            return "Any"
        elif isinstance(node, ast.Name):
            # Handle special name constants
            if node.id in ("None", "True", "False"):
                return {"None": "None", "True": "bool", "False": "bool"}[node.id]
            return "Any"
        elif isinstance(node, ast.Attribute):
            # Handle attribute access like obj.attr
            return "Any"
        elif isinstance(node, ast.BinOp):
            # Handle binary operations
            return "Any"
        elif isinstance(node, ast.UnaryOp):
            # Handle unary operations
            return "Any"
        elif isinstance(node, ast.Compare):
            # Comparison operations return bool
            return "bool"
        elif isinstance(node, ast.BoolOp):
            # Boolean operations return bool
            return "bool"
        elif isinstance(node, (ast.Lambda, ast.FunctionDef, ast.AsyncFunctionDef)):
            # Functions and lambdas
            return "Callable"
        return "Any"

    def _normalize_type(self, type_str: str) -> str:
        """Normalize Python type annotations, keeping them as Python types."""
        # Handle None and NoneType
        if type_str in ("None", "NoneType"):
            return "None"

        # Handle empty type annotation
        if not type_str or type_str.strip() == "":
            return "Any"

        # Strip whitespace
        type_str = type_str.strip()

        # Handle forward references (quoted types)
        if type_str.startswith('"') and type_str.endswith('"'):
            return self._normalize_type(type_str[1:-1])

        if type_str.startswith("'") and type_str.endswith("'"):
            return self._normalize_type(type_str[1:-1])

        # For Python types, we want to keep them as-is, just clean them up
        # Handle generic types with parameters
        if "[" in type_str and "]" in type_str:
            return self._parse_generic_type_python(type_str)

        # Return the type as-is for Python
        return type_str

    def _parse_generic_type_python(self, type_str: str) -> str:
        """Parse generic type annotations keeping them as Python types."""
        try:
            # Find the base type and parameters
            bracket_start = type_str.find("[")
            bracket_end = type_str.rfind("]")

            if bracket_start == -1 or bracket_end == -1:
                return type_str

            base_type = type_str[:bracket_start].strip()
            params_str = type_str[bracket_start + 1 : bracket_end].strip()

            # Handle empty parameters
            if not params_str:
                return base_type

            # Parse parameters (handle nested brackets)
            params = self._parse_type_parameters(params_str)

            # Recursively normalize parameters while keeping Python syntax
            normalized_params = []
            for param in params:
                normalized_params.append(self._normalize_type(param))

            # Reconstruct the type with normalized parameters
            return f"{base_type}[{', '.join(normalized_params)}]"

        except Exception:
            # If parsing fails, return the original string
            return type_str

    def _parse_type_parameters(self, params_str: str) -> list:
        """Parse type parameters from a string like 'str, int' or 'Dict[str, int], bool'."""
        params = []
        current_param = ""
        bracket_depth = 0

        for char in params_str:
            if char == "[":
                bracket_depth += 1
                current_param += char
            elif char == "]":
                bracket_depth -= 1
                current_param += char
            elif char == "," and bracket_depth == 0:
                # We've found a parameter boundary
                params.append(current_param.strip())
                current_param = ""
            else:
                current_param += char

        # Add the last parameter
        if current_param.strip():
            params.append(current_param.strip())

        return params

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
        """Build comprehensive call graph with caller relationships and metadata."""
        all_functions = self.functions[:]
        for cls in self.classes:
            all_functions.extend(cls["methods"])

        # Create a map of function names to function objects for quick lookup
        func_map = {func["name"]: func for func in all_functions}
        func_names = set(func_map.keys())

        # Initialize called_by field for all functions
        for func in all_functions:
            func["called_by"] = []

        # Build caller relationships with detailed metadata
        for func in all_functions:
            for call in func.get("calls", []):
                call_name = call["name"]

                # Handle both local calls (same file) and external calls
                if "." not in call_name and call_name in func_names:
                    # Local function call within same file
                    target_func = func_map[call_name]
                    caller_info = {
                        "function_name": func["name"],
                        "file": self.file_path,  # Same file for local calls
                        "line": call["line"],
                        "call_type": call["type"],
                    }
                    target_func["called_by"].append(caller_info)
                elif "." in call_name:
                    # Handle method calls - check if it's a call to a local method
                    parts = call_name.split(".")
                    if len(parts) == 2:
                        obj_name, method_name = parts
                        # Check if this is a method call on self or a known class
                        if obj_name == "self" or method_name in func_names:
                            target_name = (
                                method_name if obj_name == "self" else call_name
                            )
                            if target_name in func_map:
                                target_func = func_map[target_name]
                                caller_info = {
                                    "function_name": func["name"],
                                    "file": self.file_path,
                                    "line": call["line"],
                                    "call_type": call["type"],
                                }
                                target_func["called_by"].append(caller_info)

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
