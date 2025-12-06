---
trigger: manual
---

# LiveWiki Update

Update the existing LiveWiki documentation in `docs/livewiki` based on recent codebase changes.

## Requirements

- Preserve existing metadata headers but update "Last Updated" and increment "Version"
- Version increment rules:
  - Patch (X.Y.Z+1): Minor updates, typo fixes, small additions
  - Minor (X.Y+1.0): New sections, significant content additions
  - Major (X+1.0.0): Major restructuring or complete rewrites

## Update Process

### 1. Analyze Changes
- Review git diff or recent commits since last wiki update
- Identify which modules/files have changed
- Determine scope of documentation impact

### 2. Update Existing Pages
- Find all wiki pages related to changed code
- Update content to reflect current state
- Update "Last Updated" to today's date
- Increment "Version" appropriately
- Add changelog entry at bottom of page if significant

### 3. Create New Pages
- If new modules/packages were added, create corresponding wiki pages
- Follow the same metadata template:
```markdown
---
Created: YYYY-MM-DD
Last Updated: YYYY-MM-DD
Version: 1.0.0
---
```
- Add entries to table_of_contents.md
- Link from related existing pages

### 4. Update Cross-References
- Ensure all internal links still work
- Update architecture.md if system structure changed
- Update data_flow.md if data paths changed
- Update api_reference.md if public APIs changed

### 5. Provide Summary Report

After updates, provide:
- List of pages updated with version changes
- List of new pages created
- Brief description of major changes
- Recommendations for further documentation needs

## Metadata Update Example
```markdown
---
Created: 2024-12-04
Last Updated: 2024-12-04  # ← Update this
Version: 1.1.0  # ← Increment this
---

## Changelog
- **v1.1.0** (2024-12-04): Added new authentication flow documentation
- **v1.0.0** (2024-12-04): Initial creation
```

Begin by analyzing what has changed in the codebase, then systematically update affected documentation.