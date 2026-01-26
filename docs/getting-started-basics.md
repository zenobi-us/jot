# Getting Started with OpenNotes: A Beginner's Guide

Welcome to OpenNotes! This guide is for you if you want to **manage your markdown notes with a simple CLI tool** without diving into SQL or advanced features. We'll get you up and running in 15 minutes.

**Note**: If you're already comfortable with SQL and want to unlock advanced querying capabilities, check out the [Getting Started for Power Users](getting-started-power-users.md) guide instead.

---

## Part 1: Installation & First Steps (5 minutes)

Let's get OpenNotes installed and verify it's working.

### What You'll Need

**System Requirements** (super minimal!):

- macOS, Linux, or Windows with WSL
- Terminal or command prompt
- ~5 MB of disk space
- No dependencies to install—OpenNotes is a single binary

### Installation

Choose the method that works best for you:

#### Option 1: Homebrew (macOS/Linux) - Easiest

If you have Homebrew installed:

```bash
brew tap zenobi-us/tools
brew install opennotes
```

Verify it worked:

```bash
opennotes --version
```

#### Option 2: Download Binary (All Platforms)

Go to the [OpenNotes Releases](https://github.com/zenobi-us/opennotes/releases) page and download the binary for your system:

- macOS: `opennotes-darwin-arm64` (Apple Silicon) or `opennotes-darwin-amd64` (Intel)
- Linux: `opennotes-linux-amd64`
- Windows: `opennotes-windows-amd64.exe`

Make it executable (macOS/Linux):

```bash
chmod +x opennotes
# optionally move it to your PATH
mv opennotes /usr/local/bin/
```

#### Option 3: Build from Source (For Developers)

```bash
git clone https://github.com/zenobi-us/opennotes.git
cd opennotes
go build -o opennotes
./opennotes --version
```

### Verify Installation

Run this command to confirm everything is working:

```bash
opennotes --help
```

You should see helpful information about OpenNotes commands. Great! You're ready to go.

### First Startup

The first time you run OpenNotes, it creates a configuration file:

```bash
opennotes init
```

This creates `~/.config/opennotes/config.json` where your notebooks are registered. You don't need to edit this file—OpenNotes manages it automatically.

---

## Part 2: Create Your First Notebook (5 minutes)

Now let's create a notebook from your existing markdown files. A **notebook** is just a folder containing markdown files that OpenNotes can search and manage.

### Find Your Notes Folder

First, think about where your markdown notes currently live. Common locations:

- `~/Documents/Notes`
- `~/my-notes`
- `~/Desktop/Notes`
- `~/projects/documentation`

If you don't have any markdown files yet, we can create a sample notebook to learn with.

### Create a Notebook from Existing Files

If you have markdown files (or even just a few `.md` files), run:

```bash
opennotes notebook create ~/Documents/Notes --name "My Notes"
```

Replace `~/Documents/Notes` with the actual path to your markdown files.

**What just happened?**

- OpenNotes scanned your folder for all `.md` files
- It extracted titles from frontmatter (YAML at the top of files) or used filenames
- It registered your notebook so you can use it anytime

### Create an Empty Notebook to Learn

Don't have markdown files yet? No problem! Create an empty notebook:

```bash
# Create a folder for your notes
mkdir -p ~/learning-notes

# Create a notebook from that folder
opennotes notebook create ~/learning-notes --name "Learning"
```

### List Your Notebooks

See all your registered notebooks:

```bash
opennotes notebook list
```

You should see output like:

```
## Notebooks (2)

### My Notes

• **Path:** ~/Documents/Notes/.opennotes.json
• **Root:** ~/Documents/Notes

### Learning

• **Path:** ~/learning-notes/.opennotes.json
• **Root:** ~/learning-notes
```

Perfect! You now have a notebook registered. OpenNotes will automatically use the most recent notebook you created.

---

## Part 3: Add and List Your Notes (5 minutes)

Time to add some notes and see them in action!

### Add Your First Note

Create a new note in your current notebook:

```bash
opennotes notes add "My First Note"
```

This opens your default text editor (vim, nano, VS Code, etc.) so you can write content. Add some text like:

```markdown
# My First Note

This is my first note in OpenNotes!

- Point 1
- Point 2
- Point 3

I can write anything here in **markdown**.
```

Save and close the editor. OpenNotes automatically saves your note with a timestamp.

### Add More Notes

Let's add a few more so we have something to work with:

```bash
opennotes notes add "Shopping List"
# (add some grocery items)

opennotes notes add "Meeting Notes 2024"
# (add some meeting notes)
```

### List All Your Notes

Now see all the notes you've created:

```bash
opennotes notes list
```

Output looks like:

```
### Notes (3)

• [My First Note] my-first-note.md
• [Shopping List] shopping-list.md
• [Meeting Notes 2024] meeting-notes-2024.md
```

Each note shows:

- **Title** (extracted from the note or filename)
- **Filename** (the path where it's stored)

### Add Notes with Content

You can also create notes directly from the command line without opening an editor:

```bash
# Create a note and pipe content to it
echo "Quick reminder to buy milk" | opennotes notes add "Quick Note"
```

Or create a note with a custom path:

```bash
opennotes notes add "Projects/My Project" --path projects/my-project.md
```

---

## Part 4: Simple Searches (5 minutes)

Now let's find notes without needing to know SQL. We'll use simple text searches.

### Search by Text

Find notes containing specific words:

```bash
opennotes notes search "shopping"
```

This searches through all your notes and shows you which files contain "shopping". Perfect for finding that old note!

### Search by Filename

Find notes by their filename or path:

```bash
opennotes notes search "meeting"
```

### View Note Files

Your notes are stored as regular markdown files. You can open them with your favorite editor:

```bash
# With your default editor
cat my-first-note.md

# Or open in VS Code, Sublime, etc.
code my-first-note.md
vim my-first-note.md
```

Or search for content within your notes using the search functionality (see below).

### Search Tips

- **Case-insensitive**: Searching for "SHOPPING" finds "shopping"
- **Partial matches**: Searching for "milk" finds "Buy milk tomorrow"
- **Multiple words**: Searching for "project alpha" finds both words

For more powerful searches (once you're comfortable), see **Part 5: Next Steps**.

---

## Part 5: Next Steps & Learning Paths (5 minutes)

You've got OpenNotes working! Here's what you can do next:

### Continue with the Basics

**Master Notebook Management**:

- Learn how to organize notes across multiple notebooks
- Understand how to use notebooks for different projects or topics
- See [Notebook Management Guide](notebook-discovery.md)

**Better Search Techniques**:

- Use fuzzy matching for approximate searches
- Search by file patterns and locations
- See examples in the [Troubleshooting Guide](getting-started-troubleshooting.md)

### Graduation Path to Advanced Features

Once you're comfortable with basic note management, you have two paths:

#### 🚀 Path 1: SQL Power User (30-60 minutes)

Want to supercharge your searches? OpenNotes can query your entire note collection using SQL:

```bash
# Find notes by word count
opennotes notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) \
   WHERE md_stats(content).word_count > 100 \
   ORDER BY file_path"
```

This lets you:

- Search by metadata (dates, tags, word count)
- Find relationships between notes
- Extract statistics and patterns
- Automate workflows with JSON output

**Get Started**: [Getting Started for Power Users](getting-started-power-users.md) (15 minutes)

#### 📚 Path 2: Structured Note-Taking (20 minutes)

Use frontmatter (metadata at the top of your notes) to organize information:

```markdown
---
title: "Project Alpha"
tags: [work, planning, 2024]
status: active
priority: high
---

# Project Alpha Notes

Content here...
```

Then search by metadata:

```bash
opennotes notes search --data status=active
```

**Get Started**: [Import Workflow Guide](import-workflow-guide.md)

### Integration with Your Workflow

**With Git**:

```bash
# Initialize your notes folder as a git repo
cd ~/Documents/Notes
git init
git add .
git commit -m "Initial notes commit"

# Now your notes are version controlled!
```

**With Shell Scripts**:
OpenNotes outputs JSON that works great with `jq`:

```bash
# Get all notes as JSON (once you learn the power user guide)
opennotes notes list --format json | jq '.notes[] | .title'
```

---

## Troubleshooting

### "Command not found: opennotes"

**Solution**: Add OpenNotes to your PATH. If you downloaded the binary directly:

```bash
# Find where you downloaded it
which opennotes

# If empty, move it to your PATH
mv ./opennotes /usr/local/bin/opennotes
chmod +x /usr/local/bin/opennotes
```

### "No notebooks found"

**Solution**: Create a notebook first:

```bash
opennotes notebook create "My Notes" --path ~/my-notes
```

### "Editor didn't open when I tried to add a note"

**Solution**: Set a default editor:

```bash
export EDITOR=nano  # or vim, code, etc.
opennotes notes add "New Note"
```

### "I can't find the note I created"

**Solutions**:

1. Check you're in the right directory
2. List all notebooks to see which one you created the note in:
   ```bash
   opennotes notebook list
   ```
3. List notes in that specific notebook:
   ```bash
   opennotes notes list
   ```

### Stuck?

Check out the [Troubleshooting Guide](getting-started-troubleshooting.md) or see the [Notebook Discovery Guide](notebook-discovery.md) for more help.

---

## What's Next?

You now know:
✅ How to install OpenNotes
✅ How to create notebooks
✅ How to add and list notes
✅ How to search for notes
✅ Where to learn more

**Next Steps**:

- Start using OpenNotes daily with your real notes
- When you're comfortable, explore the [Power Users Guide](getting-started-power-users.md) for SQL superpowers
- Check out [Automation Recipes](automation-recipes.md) to integrate with other tools

Happy note-taking! 📝
