# Dynamic Flag Parsing Research - Thinking Process

**Research Question**: How can we implement flexible `--data.*` flags in a Cobra-based Go CLI?

**Started**: 2026-01-20 20:45:00
**Status**: In Progress

## Research Methodology

1. **Primary Sources to Investigate**:
   - Official Cobra documentation on flag parsing
   - pflag library documentation and examples
   - Viper integration patterns for structured config
   - GitHub repositories with similar implementations

2. **Key Questions to Answer**:
   - Can Cobra/pflag support dynamic flag registration?
   - What are the patterns for collecting multi-value flags?
   - How do other CLI tools handle this pattern?
   - What validation strategies are most robust?

3. **Search Strategy**:
   - Search GitHub for "cobra dynamic flags" patterns
   - Look for "cobra map flags" implementations
   - Find examples of "cobra nested flags" or "cobra prefix flags"
   - Review kubectl, gh, aws-cli for inspiration

## Initial Observations

The desired pattern is:
```bash
opennotes note add "title" path \
  --data.tag "one" --data.tag "two" \
  --data.status "todo" \
  --data.link "some/path.md"
```

This requires:
- Parsing flags with a common prefix (`--data.`)
- Supporting duplicate flags for array values (`--data.tag` multiple times)
- Converting to structured data (map or struct)
- Validating field names

## Investigation Log

