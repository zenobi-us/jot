# Advanced Note Operations Research - Executive Summary

**Date**: 2026-01-20  
**Epic**: 3e01c563 - Advanced Note Creation and Search Capabilities  
**Status**: Research Complete ✅

---

## Quick Recommendations

### 1. Dynamic Flag Parsing (`--data.*` syntax)

**✅ RECOMMENDED**: Use `StringArray` flag with custom `field=value` parsing

```bash
opennotes note add "title" path \
  --data tag=workflow \
  --data tag=learning \
  --data status=draft
```

**Implementation**: pflag `StringArrayVar` + custom parser to handle multi-value fields as arrays

**Confidence**: ⭐⭐⭐⭐⭐ HIGH

---

### 2. FZF Integration

**✅ RECOMMENDED**: Use `github.com/ktr0731/go-fuzzyfinder` (pure Go)

```bash
opennotes note search --fzf
```

**Why**: No external dependencies, works on all platforms, customizable UI

**Alternative**: Shell out to `fzf` binary if installed

**Confidence**: ⭐⭐⭐⭐ MEDIUM-HIGH

---

### 3. Boolean Query Construction

**✅ RECOMMENDED**: Parameterized queries + whitelist validation

```bash
opennotes note search \
  --and data.tag "workflow" \
  --and data.tag "learnings" \
  --not data.status "archived"
```

**Security**: ALWAYS use `?` placeholders, WHITELIST field names, VALIDATE operators

**Implementation**: Build WHERE clauses programmatically with strict validation

**Confidence**: ⭐⭐⭐⭐⭐ HIGH

---

### 4. View/Alias System

**✅ RECOMMENDED**: YAML configuration format

```yaml
views:
  today:
    description: "Notes created or updated today"
    query:
      conditions:
        - field: "data.created"
          operator: ">="
          value: "{{today}}"
```

**Storage Hierarchy**:
1. Built-in views (code)
2. Global config (`~/.config/opennotes/config.yaml`)
3. Notebook-specific (`.opennotes.json`)

**Confidence**: ⭐⭐⭐⭐⭐ HIGH

---

## Implementation Priority

### Phase 1 - MVP
1. ✅ Dynamic Flag Parsing - CRITICAL
2. ✅ Boolean Query Construction - CRITICAL
3. ✅ Built-in Views - HIGH VALUE
4. ✅ FZF Integration - UX ENHANCEMENT

### Phase 2 - Enhancements
- Custom user views
- View parameterization
- View composition/inheritance
- Advanced FZF features (multi-select, hotkeys)

---

## Key Security Rules

1. ✅ **NEVER concatenate user input into SQL**
2. ✅ **ALWAYS use parameterized queries (`?` placeholders)**
3. ✅ **WHITELIST field names** (don't allow arbitrary columns)
4. ✅ **VALIDATE operators** (only allow known safe operators)
5. ✅ **SANITIZE ORDER BY** (cannot be parameterized, must whitelist)

---

## Expected Performance

| Operation | Dataset | Time |
|-----------|---------|------|
| Simple search | 10k notes | < 10ms |
| Boolean AND (2 conditions) | 10k notes | < 20ms |
| Complex (5+ conditions) | 10k notes | < 100ms |
| FZF interactive | 1k results | < 50ms |

---

## Code Examples Location

All detailed code examples, validation functions, and integration patterns are in:

📄 `.memory/research-3e01c563-advanced-operations.md`

**Sections**:
- Topic 1: Dynamic Flag Parsing (Findings 1.1-1.4)
- Topic 2: FZF Integration (Findings 2.1-2.4)
- Topic 3: Boolean Queries (Findings 3.1-3.4)
- Topic 4: View System (Findings 4.1-4.5)
- Implementation Recommendations
- Security Considerations
- Testing Strategy

---

## Next Steps

1. ✅ Review recommendations with team
2. ✅ Create implementation tasks
3. ✅ Begin Phase 1 MVP development
4. ✅ Iterate based on user feedback

---

**Full Research Document**: `.memory/research-3e01c563-advanced-operations.md`  
**Confidence**: ⭐⭐⭐⭐⭐ HIGH - All patterns proven in production CLI tools
