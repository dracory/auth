# UI Patterns

This document outlines the User Interface (UI) patterns used in the `dracory/auth` codebase.

## Overview

The UI is built server-side using Go, leveraging the `github.com/dracory/hb` library to construct HTML programmatically. It relies heavily on **Bootstrap 5** for styling and layout, with **jQuery** used for client-side interactivity.

## Technology Stack

-   **HTML Generation**: `github.com/dracory/hb` (Go-based HTML builder).
-   **CSS Framework**: Bootstrap 5.2.1 (via `github.com/dracory/uncdn`).
-   **JavaScript Library**: jQuery 3.6.0 (via `github.com/dracory/uncdn`).
-   **Icons**: Bootstrap Icons (`bi`).
-   **Fonts**: "Nunito" (primary) and "Ubuntu" (secondary/fallback).

## Core Patterns

### 1. Page Structure

Pages are constructed using a "Shell" pattern defined in `internal/ui/shared/webpage.go`.

-   **`buildPage` Function**:
    -   Sets up the full HTML document structure (`<html>`, `<head>`, `<body>`).
    -   Injects external resources (Bootstrap CSS/JS, jQuery, Fonts).
    -   Injects global CSS overrides.
    -   Wraps the specific page content.

### 2. Layout

-   **Global Layout**: Defined in `layout.go`. It provides a basic wrapper, primarily setting the background color and font, and wrapping the content in a centered section with padding.
-   **Container Pattern**: Most pages use a Bootstrap `.container` holding a centered `.card`.
    -   **Card**: Used to frame the content (Login, Register, etc.).
        -   `.card-header`: Contains the page title.
        -   `.card-body`: Contains the form and alerts.
        -   `.card-footer`: Contains secondary actions (e.g., "Forgot password?", "Register").

### 3. Components

-   **Forms**: Standard Bootstrap form groups (`.form-group`) with labels and inputs (`.form-control`).
-   **Buttons**: Bootstrap buttons (`.btn`) often paired with icons.
    -   *Example*: `<button class="btn btn-success"><i class="bi bi-door-open"></i> Log in</button>`
-   **Alerts**: Used for feedback.
    -   `.alert-success`: For success messages.
    -   `.alert-danger`: For error messages.
    -   These are typically hidden by default and toggled via JavaScript.
-   **Spinners**: Bootstrap spinners (`.spinner-border`) used inside buttons to indicate loading states.

### 4. Client-Side Logic

-   **Inline Scripts**: JavaScript logic is defined as Go strings (e.g., `LoginScripts` in `content.go`) and injected into the page.
-   **Validation**: Simple client-side validation (checking for empty fields) before sending AJAX requests.
-   **AJAX**: `$.post` is used to submit forms to API endpoints.
-   **Feedback**: Functions like `loginFormRaiseError` and `loginFormRaiseSuccess` handle showing/hiding alerts.

## Email Templates

-   **Implementation**: `html/template` package.
-   **Structure**: Simple HTML strings defined in `email_*_template.go` files.
-   **Styling**: Minimal inline styling.

---

## Design Improvements

The following improvements are recommended to enhance consistency, maintainability, and user experience.

### 1. Unify Typography
**Current State**: `layout.go` imports "Nunito", while `shared/webpage.go` sets `font-family: Ubuntu, sans-serif` on `html,body` and then overrides `body` with `Nunito`.
**Proposal**:
-   Standardize on a single font family (e.g., Nunito) across all files.
-   Remove the conflicting `Ubuntu` declaration in `shared/webpage.go` unless it serves a specific fallback purpose.

### 2. Extract Reusable Components
**Current State**: The "Card" structure (Header, Body, Footer) is repeated in `LoginContent`, `RegisterContent`, etc.
**Proposal**:
-   Create a `shared.Card(header, body, footer)` function in `internal/ui/shared`.
-   This will reduce code duplication and ensure consistent styling (e.g., max-width, margins) across all auth pages.

### 3. Externalize JavaScript
**Current State**: JavaScript is embedded as string literals in Go functions. This makes it hard to lint, format, and test.
**Proposal**:
-   Move JavaScript to `.js` files and embed them using `//go:embed`.
-   Alternatively, if dynamic values (URLs) are needed, keep the minimal injection logic in Go but move the core logic (validation, UI toggling) to a static `.js` file served separately or embedded.

### 4. Centralize Styles
**Current State**: Styles are scattered across `layout.go` (inline CSS), `shared/webpage.go` (injected CSS string), and `hb` calls (`.Style(...)`).
**Proposal**:
-   Define a central `styles.go` or `css.go` in `internal/ui/shared` to hold all CSS strings.
-   Replace hardcoded hex codes (e.g., `#f8fafc`) with named constants (e.g., `ColorBackground`).

### 5. Enhance Email Templates
**Current State**: Basic HTML structure.
**Proposal**:
-   Use a proper HTML email layout with a table-based structure for better compatibility across email clients.
-   Add a consistent header (logo) and footer to all emails.
-   Extract the common email layout into a helper function, similar to the web page layout.

## Technical Improvements

### 1. Expand Public Interfaces
**Current State**: The `authImplementation` struct acts as a "shallow proxy" for UI pages. Methods like `pageLogin` manually construct a `Dependencies` struct, copying fields from `authImplementation` one by one. This is verbose, brittle, and violates the "Open/Closed" principle.

**Problematic Code Example (`auth_implementation_pages.go`)**:
```go
func (a authImplementation) pageLogin(w http.ResponseWriter, r *http.Request) {
    page_login.PageLogin(w, r, page_login.Dependencies{
        Passwordless:       a.passwordless,
        EnableRegistration: a.enableRegistration,
        Endpoint:           a.endpoint,
        RedirectOnSuccess:  a.LinkRedirectOnSuccess(),
        Layout:             a.funcLayout,
        Logger:             a.GetLogger(),
    })
}
```

**Proposal**:
Expand the existing `AuthSharedInterface` (and its specialized counterparts) to include accessors for common dependencies. This allows the `authImplementation` (which already implements these interfaces) to be passed directly to UI handlers.

**Recommended Approach: Expand Public Interfaces**

1.  **Expand Interface**: Add accessor methods to `types.AuthSharedInterface`.
    ```go
    // types/auth_interfaces.go
    type AuthSharedInterface interface {
        // ... existing methods ...

        // New Accessors
        GetEndpoint() string
        GetLogger() *slog.Logger
        IsRegistrationEnabled() bool
        GetLayout() func(content string) string
    }
    ```

2.  **Update Implementation**: Ensure `authImplementation` satisfies these new methods. (Most are already present or trivial to add).

3.  **Update Handlers**: Update page handlers to accept the interface.
    ```go
    // internal/ui/page_login/page_login.go
    func PageLogin(w http.ResponseWriter, r *http.Request, auth types.AuthSharedInterface) {
        // Use auth.GetEndpoint(), auth.IsRegistrationEnabled(), etc.
    }
    ```

**Benefits**:
-   **Simplicity**: Reuses the existing main interface; no need for a new parallel hierarchy.
-   **Direct Passing**: The `authImplementation` can be passed directly to handlers without wrapping or conversion.
-   **Consistency**: A single interface defines the capabilities of the auth system, both for external consumers and internal UI components.
