---
trigger: manual
---

# LiveWiki Create

Create a comprehensive LiveWiki documentation for this codebase in the `docs/livewiki` folder.

LiveWiki serves two purposes:
- **A**: Live documentation for humans (browsable via Docsify)
- **B**: Live documentation for LLMs (structured metadata for AI consumption)

## Requirements

- Create all documentation as Markdown files in `docs/livewiki/`
- Render the documentation using **Docsify** (no-build HTML interface).
- Every page MUST include the enhanced metadata header (see below)

## Enhanced Metadata Template

Every page MUST include this frontmatter for both human and LLM consumption:

```markdown
---
path: <filename>.md
page-type: overview | reference | tutorial | module | changelog
summary: One-line description of this page's content and purpose.
tags: [tag1, tag2, tag3]
created: YYYY-MM-DD
updated: YYYY-MM-DD
version: X.Y.Z
---
```

### Page-Type Taxonomy

| Type | Description |
|------|-------------|
| `overview` | High-level introductions and architectural summaries |
| `reference` | API docs, configuration options, technical specifications |
| `tutorial` | Step-by-step guides and getting started content |
| `module` | Documentation for a specific module/package |
| `changelog` | Version history and change logs |

## Structure to Create

### Core Files

1.  **index.html** - Docsify entry point (Dark Theme enabled)
2.  **_sidebar.md** - Docsify navigation sidebar
3.  **README.md** - Default entry point (copy content from overview.md)
4.  **table_of_contents.md** - Master index of all wiki pages

### Documentation Pages

5.  **overview.md** - High-level architecture and project overview
6.  **getting_started.md** - Setup, installation, and quick start guide
7.  **architecture.md** - System architecture, design patterns, and key decisions
8.  **api_reference.md** - Public APIs, functions, and interfaces
9.  **data_flow.md** - How data moves through the system
10. **configuration.md** - Configuration options and environment variables
11. **development.md** - Development workflow, testing, and contributing
12. **troubleshooting.md** - Common issues and solutions

### LLM & Human Enhancement Files

13. **llm-context.md** - Single-file summary of entire codebase for LLM ingestion
14. **cheatsheet.md** - Quick reference for common operations
15. **conventions.md** - Coding/documentation conventions for contributors

### Module Documentation

16. **modules/** folder - One page per major module/package with:
    - Purpose and responsibilities
    - Key types/interfaces
    - Dependencies and relationships
    - Usage examples

## Docsify Configuration

The `index.html` must:
- Use **Dark Theme** (`themes/dark.css`)
- Configure `window.$docsify` (NOT `window.docsify`)
- Enable `loadSidebar: true`
- Enable `auto2top: true`
- Enable `search` plugin
- Enable `mermaid` support

## Instructions

1.  Analyze the entire codebase structure.
2.  Identify all major modules, packages, and components.
3.  Create the folder structure `docs/livewiki/` and `docs/livewiki/modules/`.
4.  Generate Markdown content for all pages **using the enhanced metadata template**.
5.  **Create `index.html`** with the required Docsify configuration.
6.  **Create `_sidebar.md`** linking to all top-level pages and modules.
7.  **Create `README.md`** by duplicating `overview.md` (prevents 404s on load).
8.  **Create `llm-context.md`** summarizing the entire codebase in one file.
9.  **Create `cheatsheet.md`** with quick-reference commands and patterns.
10. **Create `conventions.md`** documenting project conventions.
11. Ensure `table_of_contents.md` links to all generated pages.
12. Use clear, concise language appropriate for developers.
13. Include code examples where relevant.
14. Create architecture diagrams using Mermaid syntax where appropriate.
15. Add a "See Also" section at the bottom of each page linking to related pages.

## llm-context.md Template

This file provides a single-file context for LLMs:

```markdown
---
path: llm-context.md
page-type: overview
summary: Complete codebase summary optimized for LLM consumption.
tags: [llm, context, summary]
created: YYYY-MM-DD
updated: YYYY-MM-DD
version: 1.0.0
---

# LLM Context: [Project Name]

## Project Summary
[1-2 paragraph description of what this project does]

## Key Technologies
- [Tech 1]
- [Tech 2]

## Directory Structure
[Brief outline of main directories]

## Core Concepts
[Numbered list of main concepts an LLM needs to understand]

## Common Patterns
[Patterns used throughout the codebase]

## Important Files
[List of key files with brief descriptions]
```
