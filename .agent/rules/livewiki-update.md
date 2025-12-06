---
trigger: manual
---

# LiveWiki Update

Update the existing LiveWiki documentation in `docs/livewiki` based on recent codebase changes.

LiveWiki serves two purposes:
- **A**: Live documentation for humans (browsable via Docsify)
- **B**: Live documentation for LLMs (structured metadata for AI consumption)

## Requirements

- Render the documentation using **Docsify**.
- Preserve existing metadata headers but update required fields
- Maintain LLM-friendly metadata (tags, summary, path)
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
- Update the following metadata fields:
  - `updated:` to today's date
  - `version:` increment appropriately
  - `tags:` add/remove tags if scope changed
  - `summary:` update if page purpose evolved
- Add changelog entry at bottom of page if significant
- **Important**: If you update `overview.md`, you MUST copy its content to `README.md` to keep the default entry point in sync.

### 3. Create New Pages
- If new modules/packages were added, create corresponding wiki pages
- Use the enhanced metadata template:

```markdown
---
path: <filename>.md
page-type: overview | reference | tutorial | module | changelog
summary: One-line description of this page's content and purpose.
tags: [tag1, tag2, tag3]
created: YYYY-MM-DD
updated: YYYY-MM-DD
version: 1.0.0
---
```
- Add entries to `table_of_contents.md`
- **Add entries to `_sidebar.md`** so they appear in the navigation menu.

### 4. Update Cross-References
- Ensure all internal links still work
- Update `architecture.md` if system structure changed
- Update `data_flow.md` if data paths changed
- Update `api_reference.md` if public APIs changed
- Update "See Also" sections if relationships changed

### 5. Update LLM Context

**Critical for LLM consumption:**
- Regenerate `llm-context.md` if:
  - Major architectural changes occurred
  - New modules/packages were added
  - Core concepts or patterns changed
- Update the "Important Files" section if key files changed

### 6. Validate Frontmatter

Before completing updates, verify:
- [ ] All pages have valid `path:` field matching their filename
- [ ] All pages have appropriate `page-type:` value
- [ ] All pages have a meaningful `summary:` (not placeholder)
- [ ] All pages have relevant `tags:` array
- [ ] All dates use YYYY-MM-DD format

### 7. Provide Summary Report

After updates, provide:
- List of pages updated with version changes
- List of new pages created
- Tags added or removed
- Brief description of major changes
- Whether `llm-context.md` was regenerated
- Recommendations for further documentation needs

## Metadata Update Example

```markdown
---
path: api_reference.md
page-type: reference
summary: Complete API endpoint documentation for authentication service.
tags: [api, endpoints, rest, authentication]
created: 2024-12-04
updated: 2024-12-06
version: 1.1.0
---

## Changelog
- **v1.1.0** (2024-12-06): Added new OAuth2 endpoints documentation
- **v1.0.0** (2024-12-04): Initial creation
```

Begin by analyzing what has changed in the codebase, then systematically update affected documentation.
