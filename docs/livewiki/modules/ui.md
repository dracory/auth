---
Created: 2025-12-06
Last Updated: 2025-12-06
Version: 1.0.0
---

# UI Module

**Package**: `internal/ui`

This module provides the HTML pages for the authentication flow.

## Structure

Similar to `internal/api`, each page is a sub-package:

*   `page_login`: The Login form.
*   `page_register`: The Registration form.
*   `page_password_reset`: Password reset request and confirmation forms.

## Customization

The HTML is generated via Go code strings (unless customized).

You can wrap all these pages in your own application layout by providing a `FuncLayout` in the configuration.

```go
FuncLayout: func(content string) string {
    return "<html><body>" + content + "</body></html>"
}
```

## Styling

The default pages use **Bootstrap 5** classes. If your application uses Bootstrap, they will blend in automatically. If not, the structure is simple enough to style via CSS overrides.
