# Regex Limitations in Query Engine

## Overview

The Repository Context Protocol query engine supports pattern matching using various methods including regular expressions. However, due to limitations in Go's `regexp` package, some advanced regex features are not supported and are automatically converted to approximate alternatives.

## Unsupported Features and Conversions

### Lookbehind Assertions

**Unsupported:** Negative lookbehind `(?<!pattern)`
- **Example:** `(?<!Process).*Data` (match "Data" not preceded by "Process")
- **Conversion:** The lookbehind assertion is removed, resulting in `.*Data`
- **Impact:** May match more results than intended

### Lookahead Assertions

**Partially Supported:** Positive lookahead `(?=pattern)`
- **Example:** `Handle(?=User)` (match "Handle" followed by "User")
- **Conversion:** Converted to `Handle.*User`
- **Impact:** May match different results due to greedy matching behavior

**Unsupported:** Negative lookahead `(?!pattern)`
- **Example:** `Handle(?!Error)` (match "Handle" not followed by "Error")
- **Conversion:** The negative lookahead assertion is removed, resulting in `Handle`
- **Impact:** May match more results than intended

## Warnings and Error Handling

When unsupported regex features are detected:

1. **Warnings are logged** to help identify potential matching discrepancies
2. **Patterns are automatically converted** to supported alternatives
3. **Search continues** with the converted pattern

To see these warnings, enable logging in your application.

## Best Practices

1. **Test your patterns** against expected data to verify matching behavior
2. **Use simpler alternatives** when possible:
   - Instead of lookbehind: Use multiple patterns or post-process results
   - Instead of negative lookahead: Use exclusion patterns or filtering
3. **Monitor logs** for regex conversion warnings
4. **Consider using glob patterns** for simpler matching scenarios

## Alternative Approaches

For complex pattern matching requirements:

1. **Multiple queries:** Break complex patterns into simpler ones
2. **Post-processing:** Apply additional filtering after initial search
3. **Glob patterns:** Use glob-style patterns for file/path matching
4. **Exact matching:** Use exact string matching when appropriate

## Examples

### Recommended Patterns

```
# Good: Simple patterns
"func.*User"           # Functions containing "User"
"*Handler"             # Glob pattern for handlers
"User*"                # Names starting with "User"

# Good: Character classes
"[A-Z][a-z]*Service"   # Service classes with proper naming

# Good: Anchoring
"^Handle"              # Names starting with "Handle"
"Error$"               # Names ending with "Error"
```

### Patterns That Will Be Converted

```
# Will be converted with warning
"(?<!Test).*Handler"   # Lookbehind removed -> ".*Handler"
"Handle(?=User)"       # Lookahead converted -> "Handle.*User"
"Handle(?!Error)"      # Negative lookahead removed -> "Handle"
```

## Configuration

Currently, regex feature conversion cannot be disabled. This ensures consistent behavior across different environments and prevents search failures due to unsupported features.

Future versions may include options to:
- Strict mode (return errors instead of converting)
- Detailed conversion reports
- Custom conversion strategies
