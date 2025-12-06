---
trigger: manual
---

# LiveWiki Create

Create a comprehensive LiveWiki documentation for this codebase in the `docs/livewiki` folder.

## Requirements

- Create all documentation as Markdown files in `docs/livewiki/`
- Every page MUST include a metadata header with:
  - Created: YYYY-MM-DD
  - Last Updated: YYYY-MM-DD
  - Version: X.Y.Z (start at 1.0.0)

## Structure to Create

1. **table_of_contents.md** - Master index of all wiki pages with brief descriptions
2. **overview.md** - High-level architecture and project overview
3. **getting_started.md** - Setup, installation, and quick start guide
4. **architecture.md** - System architecture, design patterns, and key decisions
5. **modules/** folder - One page per major module/package with:
   - Purpose and responsibilities
   - Key types/interfaces
   - Dependencies and relationships
   - Usage examples
6. **api_reference.md** - Public APIs, functions, and interfaces
7. **data_flow.md** - How data moves through the system
8. **configuration.md** - Configuration options and environment variables
9. **development.md** - Development workflow, testing, and contributing
10. **troubleshooting.md** - Common issues and solutions

## Metadata Template

Use this at the top of every page:
```markdown
---
Created: YYYY-MM-DD
Last Updated: YYYY-MM-DD
Version: 1.0.0
---
```

## Instructions

1. Analyze the entire codebase structure
2. Identify all major modules, packages, and components
3. Create the folder structure `docs/livewiki/` and `docs/livewiki/modules/`
4. Generate each page with proper metadata headers
5. Ensure table_of_contents.md links to all generated pages
6. Use clear, concise language appropriate for developers
7. Include code examples where relevant
8. Create architecture diagrams using Mermaid syntax where appropriate

Begin by creating the table of contents, then generate all pages systematically.