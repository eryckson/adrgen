# adrgen

🧱 Simple, fast, and customizable ADR (Architecture Decision Records) generator, based on the model by [Joel Parker Henderson](https://github.com/joelparkerhenderson/architecture-decision-record).

## ✨ Installation

```bash
go install github.com/eryckson/adrgen@latest
```

## 🛠️ Build from Source

```bash
go build -o bin/adrgen main.go
```

## 🚀 Usage

The `adrgen` tool helps you create and manage Architecture Decision Records (ADRs) with a simple command-line interface.

### Basic Usage

```bash
adrgen --number 001 --status Accepted --title "Choose Database Technology"
```

This will:
1. Create a new ADR file in `docs/adr/adr-001-choose-database-technology.md`
2. Use the default template or your custom template if available
3. Automatically update the ADR index file (`docs/adr/README.md`)

### Directory Structure

After running adrgen, your project will have this structure:

```
docs/
└── adr/
    ├── README.md                    # Auto-generated index of all ADRs
    ├── template.md                  # Optional custom template
    ├── adr-001-first-decision.md
    └── adr-002-second-decision.md
```

### Customizing Templates

You can customize the ADR template by creating a `template.md` file in the `docs/adr` directory. The template supports the following variables:

- `{{number}}` - The ADR number
- `{{title}}` - The ADR title
- `{{status}}` - The ADR status
- `{{date}}` - Automatically filled with the current date

### Example Template

```markdown
# ADR {{number}}: {{title}}

**Status**: {{status}}  
**Date**: {{date}}

## Context
[Describe the context and problem statement]

## Decision
[Describe the decision that was made]

## Consequences
[Describe the resulting context]
```

### Command Options

- `--number` - Sequential ADR number (e.g., "001", "002")
- `--status` - Decision status (e.g., "Accepted", "Proposed", "Rejected")
- `--title` - Descriptive title for the ADR (use quotes for multi-word titles)