---
path: modules/ui.md
page-type: module
summary: UI module documentation for internal/ui page handlers, page packages, and customization.
tags: [module, ui, internal, html, pages, bootstrap]
created: 2025-12-06
updated: 2026-07-16
version: 2.0.0
---

# UI Module

**Package**: `internal/ui`

This module provides the HTML pages for the authentication flow. Each page is a sub-package.

## Structure

*   `page_login`: The Login form.
*   `page_login_code_verify`: Login code verification form (Passwordless).
*   `page_register`: The Registration form.
*   `page_register_code_verify`: Registration code verification form.
*   `page_logout`: Logout page.
*   `page_password_reset`: Password reset confirmation form.
*   `page_password_restore`: Password restore request form.
*   `shared`: Shared UI components and helpers.

## Customization

Wrap all pages in your own application layout by providing a `FuncLayout` in the configuration:

```go
FuncLayout: func(content string) string {
    return "<html><body>" + content + "</body></html>"
}
```

If no `FuncLayout` is provided, the default `helpers.Layout` is used.

## Styling

The default pages use **Bootstrap 5** classes. If your application uses Bootstrap, they will blend in automatically. HTML is generated using the `dracory/hb` HTML builder library.

## See Also

- [Core Module](core.md)
- [Configuration](../configuration.md)

## Changelog
- **v2.0.0** (2026-07-16): Added page_login_code_verify, page_register_code_verify, page_logout, shared package. Added hb library mention.
- **v1.0.0** (2025-12-04): Initial creation
