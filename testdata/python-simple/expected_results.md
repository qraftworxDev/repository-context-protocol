# Python-Simple Testdata Expected Results

This document provides expected results for the Python AST parser when run against the python-simple testdata files.

## File Structure Overview

The python-simple testdata contains:
- `main.py` - Simple functions with type hints and basic logic
- `models.py` - Classes, methods, constants, and variables for testing Python AST parsing

## Expected Parsing Results

### main.py Parsing Results

#### Functions Expected:
1. **format_name**
   - Name: `format_name`
   - Parameters: `[{Name: "name", Type: "str"}]`
   - Returns: `[{Name: "str", Kind: "builtin"}]`
   - Location: Line ~9

2. **validate_email**
   - Name: `validate_email`
   - Parameters: `[{Name: "email", Type: "str"}]`
   - Returns: `[{Name: "bool", Kind: "builtin"}]`
   - Location: Line ~14

3. **process_user_data**
   - Name: `process_user_data`
   - Parameters: `[{Name: "name", Type: "str"}, {Name: "email", Type: "str"}, {Name: "age", Type: "int", Default: "18"}]`
   - Returns: `[{Name: "Dict[str, Union[str, int, bool]]", Kind: "composite"}]`
   - Location: Line ~19
   - Calls: `validate_email`, `format_name`

4. **calculate_statistics**
   - Name: `calculate_statistics`
   - Parameters: `[{Name: "numbers", Type: "List[float]"}]`
   - Returns: `[{Name: "Dict[str, float]", Kind: "composite"}]`
   - Location: Line ~38

5. **search_users**
   - Name: `search_users`
   - Parameters: `[{Name: "users", Type: "List[Dict[str, str]]"}, {Name: "query", Type: "str"}]`
   - Returns: `[{Name: "List[Dict[str, str]]", Kind: "composite"}]`
   - Location: Line ~54

6. **generate_report**
   - Name: `generate_report`
   - Parameters: `[{Name: "data", Type: "List[Dict[str, Union[str, int]]]"}]`
   - Returns: `[{Name: "str", Kind: "builtin"}]`
   - Location: Line ~69

7. **main**
   - Name: `main`
   - Parameters: `[]`
   - Returns: `[{Name: "None", Kind: "builtin"}]`
   - Location: Line ~84
   - Calls: `process_user_data`, `calculate_statistics`, `search_users`, `generate_report`

#### Imports Expected:
- `{Path: "os", Alias: ""}`
- `{Path: "sys", Alias: ""}`
- `{Path: "typing", Alias: "", Items: ["List", "Dict", "Optional", "Union"]}`

#### Call Graph Expected:
- `main` → `process_user_data`
- `process_user_data` → `validate_email`
- `process_user_data` → `format_name`
- `main` → `calculate_statistics`
- `main` → `search_users`
- `main` → `generate_report`

### models.py Parsing Results

#### Constants Expected:
- `MAX_USERS` = 100
- `DEFAULT_TIMEOUT` = 30.0
- `SERVICE_VERSION` = "1.0.0"
- `DEBUG_MODE` = True

#### Variables Expected:
- `global_config`: Type `Optional[Config]`, Value: `None`
- `user_count`: Type `int`, Value: `0`
- `service_registry`: Type `Dict[str, Any]`, Value: `{}`
- `is_initialized`: Type `bool`, Value: `False`

#### Classes Expected:

1. **Address** (dataclass)
   - Type: `class`
   - Fields: `street: str`, `city: str`, `state: str`, `zip_code: str`, `country: str`
   - Methods: `__str__`, `validate`
   - Decorators: `@dataclass`

2. **Config**
   - Type: `class`
   - Methods: `__init__`, `set_feature`, `is_feature_enabled`, `create_default` (classmethod)

3. **User**
   - Type: `class`
   - Methods: `__init__`, `__str__`, `activate`, `deactivate`, `set_profile`

4. **Profile**
   - Type: `class`
   - Methods: `__init__`, `add_skill`, `remove_skill`, `get_skill_count`, `set_address`

5. **Repository** (Abstract Base Class)
   - Type: `class`
   - Kind: `abstract`
   - Methods: `save`, `find_by_id`, `find_all`, `delete` (all abstract)
   - Base: `ABC`

6. **UserService** (Protocol)
   - Type: `protocol`
   - Methods: `get_user`, `create_user`, `update_user`, `delete_user`

7. **InMemoryUserRepository**
   - Type: `class`
   - Base: `Repository`
   - Methods: `__init__`, `save`, `find_by_id`, `find_all`, `delete`, `get_user_count`

#### Functions Expected:
1. **create_default_config**
   - Name: `create_default_config`
   - Parameters: `[]`
   - Returns: `[{Name: "Config", Kind: "class"}]`
   - Calls: `Config.create_default`

#### Imports Expected:
- `{Path: "typing", Alias: "", Items: ["List", "Dict", "Optional", "Any", "Protocol"]}`
- `{Path: "datetime", Alias: "", Items: ["datetime"]}`
- `{Path: "dataclasses", Alias: "", Items: ["dataclass"]}`
- `{Path: "abc", Alias: "", Items: ["ABC", "abstractmethod"]}`

#### Inheritance Relationships Expected:
- `Address` → `dataclass` (decorator)
- `Repository` → `ABC`
- `UserService` → `Protocol`
- `InMemoryUserRepository` → `Repository`

#### Method Call Relationships Expected:
- `Address.__str__` → string formatting
- `Address.validate` → string methods
- `Config.set_feature` → dict operations
- `Config.is_feature_enabled` → dict operations
- `Config.create_default` → `Config.__init__`, `Config.set_feature`
- `User.set_profile` → attribute access
- `Profile.add_skill` → list operations, `datetime.now`
- `Profile.remove_skill` → list operations, `datetime.now`
- `Profile.set_address` → `datetime.now`
- `InMemoryUserRepository.save` → dict operations
- `InMemoryUserRepository.find_by_id` → dict operations
- `InMemoryUserRepository.find_all` → list operations
- `InMemoryUserRepository.delete` → dict operations
- `create_default_config` → `Config.create_default`

## Test Validation Criteria

### Function Parsing:
- All functions should be correctly identified with proper names
- Type hints should be parsed and mapped to appropriate Go model types
- Default parameter values should be captured
- Function calls within bodies should be identified

### Class Parsing:
- All classes should be identified with correct names
- Inheritance relationships should be captured
- Method definitions should be properly parsed
- Class variables and instance variables should be distinguished
- Abstract methods and protocols should be identified

### Import Parsing:
- All import statements should be captured
- `from X import Y` should be properly parsed
- Type imports should be identified

### Type System Mapping:
- `str` → "string"
- `int` → "int"
- `float` → "float64"
- `bool` → "bool"
- `List[T]` → "[]T"
- `Dict[K,V]` → "map[K]V"
- `Optional[T]` → "*T"
- `Union[A,B]` → "A|B"

### Call Graph Analysis:
- Function calls should be properly identified
- Method calls should include receiver information where possible
- Cross-function call relationships should be tracked

## Performance Expectations

- Parse both files in under 1 second
- Memory usage should remain reasonable (< 50MB)
- Error handling should gracefully handle syntax errors
