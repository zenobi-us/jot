# Views System Examples

Real-world examples and workflows using OpenNotes Views System.

## Table of Contents

- [Daily Workflows](#daily-workflows)
- [Project Management](#project-management)
- [Knowledge Graph Maintenance](#knowledge-graph-maintenance)
- [Team Collaboration](#team-collaboration)
- [Custom View Patterns](#custom-view-patterns)
- [Advanced Use Cases](#advanced-use-cases)

---

## Daily Workflows

### Daily Standup Preparation

**Goal**: Review today's work and prepare standup notes

**Workflow**:
```bash
# 1. See what you worked on today
opennotes notes view today

# 2. Check recent activity
opennotes notes view recent

# 3. Find in-progress tasks
opennotes notes view kanban --param status=in-progress

# 4. Export to JSON for reporting
opennotes notes view today --format json > /tmp/standup-$(date +%Y-%m-%d).json
```

**Custom View**: `standup` (add to `~/.config/opennotes/config.json`)
```json
{
  "views": [
    {
      "name": "standup",
      "description": "Today's work + in-progress tasks",
      "query": {
        "conditions": [
          {
            "field": "updated_at",
            "operator": ">=",
            "value": "{{today}}"
          },
          {
            "field": "data.status",
            "operator": "IN",
            "value": ["in-progress", "done"]
          }
        ],
        "orderBy": "data.status ASC, updated_at DESC"
      }
    }
  ]
}
```

**Usage**:
```bash
opennotes notes view standup
```

---

### End-of-Day Review

**Goal**: Capture daily accomplishments and plan tomorrow

**Workflow**:
```bash
# 1. Review today's completed work
opennotes notes view kanban --param status=done --format json \
  | jq '.[] | select(.updated_at | startswith("'$(date +%Y-%m-%d)'"))'

# 2. Find unfinished tasks
opennotes notes view kanban --param status=todo,in-progress

# 3. Create tomorrow's plan
# (Manual: create new daily note with links to carry-over tasks)
```

**Custom View**: `completed-today`
```json
{
  "views": [
    {
      "name": "completed-today",
      "description": "Tasks completed today",
      "query": {
        "conditions": [
          {
            "field": "updated_at",
            "operator": ">=",
            "value": "{{today}}"
          },
          {
            "field": "data.status",
            "operator": "=",
            "value": "done"
          }
        ],
        "orderBy": "updated_at DESC"
      }
    }
  ]
}
```

---

### Weekly Review

**Goal**: Summarize week's accomplishments

**Custom View**: `this-week` (add to global config)
```json
{
  "views": [
    {
      "name": "this-week",
      "description": "Notes modified this week",
      "query": {
        "conditions": [
          {
            "field": "updated_at",
            "operator": ">=",
            "value": "{{this_week}}"
          }
        ],
        "orderBy": "updated_at DESC"
      }
    }
  ]
}
```

**Workflow**:
```bash
# 1. Review week's activity
opennotes notes view this-week

# 2. Count completed tasks this week
opennotes notes view this-week --format json \
  | jq '[.[] | select(.data.status == "done")] | length'

# 3. Generate weekly report
opennotes notes view this-week --format json \
  | jq 'group_by(.data.status) | map({status: .[0].data.status, count: length})'
```

**Output**:
```json
[
  {"status": "done", "count": 15},
  {"status": "in-progress", "count": 3},
  {"status": "todo", "count": 8}
]
```

---

## Project Management

### Sprint Planning

**Goal**: Organize work for 2-week sprint

**Custom View**: `sprint` (add to notebook `.opennotes.json`)
```json
{
  "views": [
    {
      "name": "sprint",
      "description": "Tasks for current sprint",
      "parameters": [
        {
          "name": "sprint_number",
          "type": "string",
          "required": true,
          "description": "Sprint number (e.g., 'sprint-12')"
        }
      ],
      "query": {
        "conditions": [
          {
            "field": "data.sprint",
            "operator": "=",
            "value": "{{sprint_number}}"
          }
        ],
        "orderBy": "data.priority DESC, data.status ASC"
      }
    }
  ]
}
```

**Usage**:
```bash
# View current sprint
opennotes notes view sprint --param sprint_number=sprint-12

# Export sprint backlog
opennotes notes view sprint --param sprint_number=sprint-12 --format json \
  > sprint-12-backlog.json
```

---

### Kanban Board Tracking

**Goal**: Track project status across multiple states

**Workflow**:
```bash
# 1. See full kanban board
opennotes notes view kanban

# 2. Filter specific columns
opennotes notes view kanban --param status=todo
opennotes notes view kanban --param status=in-progress

# 3. Export to JSON for dashboard
opennotes notes view kanban --format json \
  | jq 'group_by(.data.status) | map({status: .[0].data.status, tasks: [.[] | .title]})'
```

**Custom Kanban States**:

Add to notebook config for project-specific workflow:
```json
{
  "views": [
    {
      "name": "dev-kanban",
      "description": "Development workflow board",
      "parameters": [
        {
          "name": "status",
          "type": "list",
          "required": false,
          "default": ["backlog", "in-dev", "in-review", "testing", "done"]
        }
      ],
      "query": {
        "conditions": [
          {
            "field": "data.status",
            "operator": "IN",
            "value": "{{status}}"
          }
        ],
        "orderBy": "data.priority DESC"
      }
    }
  ]
}
```

---

### Priority Management

**Goal**: Focus on high-priority work

**Custom View**: `urgent`
```json
{
  "views": [
    {
      "name": "urgent",
      "description": "High-priority tasks",
      "query": {
        "conditions": [
          {
            "field": "data.priority",
            "operator": "IN",
            "value": ["urgent", "high"]
          },
          {
            "field": "data.status",
            "operator": "!=",
            "value": "done"
          }
        ],
        "orderBy": "data.priority ASC, data.due_date ASC"
      }
    }
  ]
}
```

**Workflow**:
```bash
# Morning: Check urgent tasks
opennotes notes view urgent

# Filter by category
opennotes notes view urgent --format json \
  | jq '.[] | select(.data.category == "bugs")'
```

---

## Knowledge Graph Maintenance

### Finding Orphaned Notes

**Goal**: Identify disconnected content for integration

**Built-in Views**:
```bash
# Notes with no incoming links (common orphans)
opennotes notes view orphans

# Notes with no links at all
opennotes notes view orphans --param definition=no-links

# Completely isolated notes
opennotes notes view orphans --param definition=isolated
```

**Maintenance Workflow**:
```bash
# 1. Find orphans
opennotes notes view orphans --format json > /tmp/orphans.json

# 2. Review and categorize
cat /tmp/orphans.json | jq '.[] | {path, title}'

# 3. Decide action:
#    - Link to existing notes
#    - Tag for later integration
#    - Archive if obsolete

# 4. Verify reduction
opennotes notes view orphans --format json | jq '. | length'
```

---

### Fixing Broken Links

**Goal**: Maintain link integrity

**Workflow**:
```bash
# 1. Find notes with broken links
opennotes notes view broken-links

# 2. Get JSON for programmatic processing
opennotes notes view broken-links --format json > /tmp/broken.json

# 3. Analyze broken link patterns
cat /tmp/broken.json | jq -r '.[] | .broken_links[]' | sort | uniq -c | sort -rn

# 4. Fix common issues:
#    - Rename files to match references
#    - Update links to correct paths
#    - Create missing notes
```

**Example Output**:
```
### Notes with Broken Links (4)

- [Project Overview] projects/main.md
  Broken: [[old-architecture]] → File doesn't exist
  
- [Meeting Notes] meetings/2026-01-15.md
  Broken: [[action-items]] → File doesn't exist
```

---

### Content Organization Audit

**Goal**: Find and tag unorganized content

**Workflow**:
```bash
# 1. Find untagged notes
opennotes notes view untagged

# 2. Export for batch tagging
opennotes notes view untagged --format json > /tmp/untagged.json

# 3. Analyze paths to infer tags
cat /tmp/untagged.json | jq -r '.[].path' | xargs dirname | sort | uniq -c

# 4. Batch update tags
# (Manual: edit notes or use script to add tags based on path)

# 5. Verify progress
opennotes notes view untagged --format json | jq '. | length'
```

---

## Team Collaboration

### Author-based Views

**Goal**: Track individual contributions

**Custom View**: `by-author`
```json
{
  "views": [
    {
      "name": "by-author",
      "description": "Notes by specific author",
      "parameters": [
        {
          "name": "author",
          "type": "string",
          "required": true,
          "description": "Author name"
        }
      ],
      "query": {
        "conditions": [
          {
            "field": "data.author",
            "operator": "=",
            "value": "{{author}}"
          }
        ],
        "orderBy": "updated_at DESC"
      }
    }
  ]
}
```

**Usage**:
```bash
# View Alice's notes
opennotes notes view by-author --param author="Alice"

# Export for review
opennotes notes view by-author --param author="Bob" --format json \
  | jq '[.[] | {title, path, updated_at}]'
```

---

### Team Standup Aggregation

**Goal**: Aggregate team's daily updates

**Custom View**: `team-today`
```json
{
  "views": [
    {
      "name": "team-today",
      "description": "All team updates today",
      "query": {
        "conditions": [
          {
            "field": "updated_at",
            "operator": ">=",
            "value": "{{today}}"
          },
          {
            "field": "data.type",
            "operator": "=",
            "value": "standup"
          }
        ],
        "orderBy": "data.author ASC, updated_at DESC"
      }
    }
  ]
}
```

**Workflow**:
```bash
# Generate team standup report
opennotes notes view team-today --format json \
  | jq 'group_by(.data.author) | map({
      author: .[0].data.author,
      updates: [.[] | {title, status: .data.status}]
    })'
```

---

### Release Planning

**Goal**: Track release-specific work

**Custom View**: `release`
```json
{
  "views": [
    {
      "name": "release",
      "description": "Notes for specific release",
      "parameters": [
        {
          "name": "version",
          "type": "string",
          "required": true,
          "description": "Release version (e.g., 'v1.2.0')"
        }
      ],
      "query": {
        "conditions": [
          {
            "field": "data.release",
            "operator": "=",
            "value": "{{version}}"
          }
        ],
        "orderBy": "data.priority DESC"
      }
    }
  ]
}
```

**Usage**:
```bash
# View release notes and tasks
opennotes notes view release --param version=v1.2.0

# Count remaining work
opennotes notes view release --param version=v1.2.0 --format json \
  | jq '[.[] | select(.data.status != "done")] | length'
```

---

## Custom View Patterns

### Time-based Filters

**Last 7 Days**:
```json
{
  "name": "last-week",
  "query": {
    "conditions": [
      {
        "field": "updated_at",
        "operator": ">=",
        "value": "{{this_week}}"
      }
    ]
  }
}
```

**This Month**:
```json
{
  "name": "this-month",
  "query": {
    "conditions": [
      {
        "field": "updated_at",
        "operator": ">=",
        "value": "{{this_month}}"
      }
    ]
  }
}
```

**Before Date**:
```json
{
  "name": "before-date",
  "parameters": [
    {
      "name": "date",
      "type": "date",
      "required": true
    }
  ],
  "query": {
    "conditions": [
      {
        "field": "created_at",
        "operator": "<",
        "value": "{{date}}"
      }
    ]
  }
}
```

---

### Multi-field Filtering

**Filtered by Multiple Criteria**:
```json
{
  "name": "active-bugs",
  "description": "Open bugs by priority",
  "parameters": [
    {
      "name": "priority",
      "type": "list",
      "default": ["high", "critical"]
    }
  ],
  "query": {
    "conditions": [
      {
        "field": "data.type",
        "operator": "=",
        "value": "bug"
      },
      {
        "field": "data.status",
        "operator": "!=",
        "value": "resolved"
      },
      {
        "field": "data.priority",
        "operator": "IN",
        "value": "{{priority}}"
      }
    ],
    "orderBy": "data.priority ASC, created_at ASC"
  }
}
```

---

### Negation Filters

**Exclude Specific Tags**:
```json
{
  "name": "no-archive",
  "description": "Active notes (exclude archive tag)",
  "query": {
    "conditions": [
      {
        "field": "data.tags",
        "operator": "NOT LIKE",
        "value": "%archive%"
      }
    ]
  }
}
```

**Not in Specific Status**:
```json
{
  "name": "incomplete",
  "description": "Tasks not yet done",
  "query": {
    "conditions": [
      {
        "field": "data.status",
        "operator": "!=",
        "value": "done"
      }
    ]
  }
}
```

---

## Advanced Use Cases

### Automated Reporting

**Daily Summary Email**:
```bash
#!/bin/bash
# daily-report.sh

DATE=$(date +%Y-%m-%d)
REPORT="/tmp/daily-report-$DATE.md"

cat > "$REPORT" << EOF
# Daily Report - $DATE

## Completed Today
$(opennotes notes view completed-today --format table)

## In Progress
$(opennotes notes view kanban --param status=in-progress --format table)

## Urgent Items
$(opennotes notes view urgent --format table)

EOF

# Send via email (requires mail command)
mail -s "Daily Report - $DATE" team@example.com < "$REPORT"
```

---

### Dashboard Integration

**Export to JSON for Web Dashboard**:
```bash
#!/bin/bash
# export-dashboard.sh

mkdir -p /tmp/dashboard

# Export multiple views
opennotes notes view kanban --format json > /tmp/dashboard/kanban.json
opennotes notes view orphans --format json > /tmp/dashboard/orphans.json
opennotes notes view broken-links --format json > /tmp/dashboard/broken-links.json
opennotes notes view this-week --format json > /tmp/dashboard/activity.json

# Combine into single dashboard JSON
jq -n \
  --slurpfile kanban /tmp/dashboard/kanban.json \
  --slurpfile orphans /tmp/dashboard/orphans.json \
  --slurpfile broken /tmp/dashboard/broken-links.json \
  --slurpfile activity /tmp/dashboard/activity.json \
  '{
    kanban: $kanban[0],
    orphans: $orphans[0],
    broken_links: $broken[0],
    weekly_activity: $activity[0],
    generated_at: now | strftime("%Y-%m-%dT%H:%M:%SZ")
  }' > /tmp/dashboard/dashboard.json

echo "Dashboard exported to /tmp/dashboard/dashboard.json"
```

---

### CI/CD Integration

**Pre-commit Hook (Check for Broken Links)**:
```bash
#!/bin/bash
# .git/hooks/pre-commit

# Check for broken links before commit
BROKEN=$(opennotes notes view broken-links --format json | jq '. | length')

if [ "$BROKEN" -gt 0 ]; then
  echo "❌ Commit blocked: $BROKEN notes have broken links"
  echo "Run: opennotes notes view broken-links"
  exit 1
fi

echo "✅ No broken links detected"
```

---

### Weekly Metrics Tracking

**Track Productivity Trends**:
```bash
#!/bin/bash
# weekly-metrics.sh

WEEK=$(date +%Y-W%U)
METRICS_FILE="metrics/$WEEK.json"

mkdir -p metrics

jq -n \
  --argjson total $(opennotes notes list --format json | jq '. | length') \
  --argjson completed $(opennotes notes view this-week --format json | jq '[.[] | select(.data.status == "done")] | length') \
  --argjson in_progress $(opennotes notes view kanban --param status=in-progress --format json | jq '. | length') \
  --argjson orphans $(opennotes notes view orphans --format json | jq '. | length') \
  --argjson untagged $(opennotes notes view untagged --format json | jq '. | length') \
  '{
    week: $ENV.WEEK,
    total_notes: $total,
    completed_this_week: $completed,
    in_progress: $in_progress,
    orphans: $orphans,
    untagged: $untagged,
    completion_rate: (($completed / ($completed + $in_progress)) * 100 | round)
  }' > "$METRICS_FILE"

echo "Metrics saved to $METRICS_FILE"
cat "$METRICS_FILE" | jq .
```

---

## Next Steps

- **User Guide**: See [views-guide.md](views-guide.md) for complete documentation
- **API Reference**: See [views-api.md](views-api.md) for schema details
- **SQL Guide**: See [sql-guide.md](sql-guide.md) for custom queries
- **Automation**: See [automation-recipes.md](automation-recipes.md) for more scripts

---

**Last Updated**: 2026-01-24  
**Version**: 1.0.0
