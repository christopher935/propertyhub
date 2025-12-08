# Base Template System - ENFORCEMENT LAYER

**Purpose:** Make it IMPOSSIBLE to create inconsistent layouts. Every page MUST extend a base template.

---

## Architecture

```
admin-base.html     → All admin pages inherit from this
consumer-base.html  → All consumer pages inherit from this
```

**Result:** Change sidebar once, it updates everywhere. No more duplicate HTML.

---

## Admin Template Usage

### Minimal Example

```html
{{define "property-list"}}
{{template "admin-base" .}}

{{define "content"}}
    <h2>My Page Content</h2>
    <p>This is the only HTML you write.</p>
{{end}}

{{end}}
```

### Full Example with All Options

```html
{{define "my-admin-page"}}
{{template "admin-base" .}}

<!-- Optional: Add page-specific CSS -->
{{define "additional-css"}}
    <link rel="stylesheet" href="/static/css/my-page.css">
{{end}}

<!-- Optional: Add extra head content (meta tags, etc) -->
{{define "head"}}
    <meta name="description" content="My admin page">
{{end}}

<!-- Optional: Alpine.js data initialization -->
{{define "layout-data"}}
x-data="myPageData()" x-init="init()"
{{end}}

<!-- Required: Your page content -->
{{define "content"}}
    <section class="dashboard-section">
        <h2>Dashboard Stats</h2>
        <!-- Your page content here -->
    </section>
{{end}}

<!-- Optional: Page-specific JavaScript -->
{{define "scripts"}}
    <script src="/static/js/my-page.js"></script>
{{end}}

{{end}}
```

### What You Get Automatically

✅ Sidebar with all navigation (from `shared/components/admin-sidebar.html`)  
✅ Top bar with search, notifications, user menu  
✅ All CSS in correct order (tokens → utilities → layouts → components)  
✅ Alpine.js loaded  
✅ Mobile-responsive sidebar toggle  
✅ Admin notifications system  
✅ Consistent user avatar and menu  

---

## Consumer Template Usage

### Minimal Example

```html
{{define "about-page"}}
{{template "consumer-base" .}}

{{define "content"}}
    <section class="hero">
        <h1>About Us</h1>
        <p>Welcome to PropertyHub</p>
    </section>
{{end}}

{{end}}
```

### Full Example

```html
{{define "my-consumer-page"}}
{{template "consumer-base" .}}

<!-- Optional: Page-specific CSS -->
{{define "additional-css"}}
    <link rel="stylesheet" href="/static/css/homepage.css">
{{end}}

<!-- Required: Page content -->
{{define "content"}}
    <section class="hero">
        <h1>{{.HeroTitle}}</h1>
    </section>
    
    <section class="features">
        <!-- Your content -->
    </section>
{{end}}

<!-- Optional: Page-specific JavaScript -->
{{define "scripts"}}
    <script src="/static/js/homepage-properties.js"></script>
{{end}}

{{end}}
```

### What You Get Automatically

✅ Consumer header with navigation (from `shared/components/consumer-header.html`)  
✅ Footer with links, legal, contact info  
✅ Cookie consent banner  
✅ All CSS in correct order  
✅ Alpine.js loaded  
✅ Equal Housing logo  
✅ Mobile-responsive navigation  

---

## Template Data Requirements

Pass this data from your Go handler:

```go
type AdminPageData struct {
    PageTitle string              // "Property List" (shows in top bar)
    User      *models.User        // Current user (for avatar, name, role)
    // ... your page-specific data
}

// In handler
data := AdminPageData{
    PageTitle: "Property List",
    User:      currentUser,
}
tmpl.ExecuteTemplate(w, "property-list", data)
```

---

## Block Definitions

### Admin Base Blocks

| Block Name | Required | Purpose |
|------------|----------|---------|
| `content` | **YES** | Your page content (everything between top bar and footer) |
| `additional-css` | No | Page-specific CSS files |
| `head` | No | Extra `<head>` content (meta tags, etc) |
| `layout-data` | No | Alpine.js data initialization for layout div |
| `scripts` | No | Page-specific JavaScript at end of body |

### Consumer Base Blocks

| Block Name | Required | Purpose |
|------------|----------|---------|
| `content` | **YES** | Your page content (everything between header and footer) |
| `additional-css` | No | Page-specific CSS files |
| `head` | No | Extra `<head>` content |
| `scripts` | No | Page-specific JavaScript |

---

## Migration Guide

### Before (Old Way - DON'T DO THIS)

```html
<!DOCTYPE html>
<html>
<head>
    <title>Property List</title>
    <link rel="stylesheet" href="/static/css/design-tokens.css">
    <!-- ... 20 more lines of head content -->
</head>
<body>
    <aside class="admin-sidebar">
        <!-- 150 lines of sidebar HTML copied from other pages -->
    </aside>
    
    <main>
        <div class="admin-top-bar">
            <!-- 50 lines of top bar HTML -->
        </div>
        
        <div class="admin-content">
            <!-- Your actual page content (10 lines) -->
        </div>
    </main>
    
    <script src="/static/js/admin.js"></script>
    <!-- ... more scripts -->
</body>
</html>
```

**Problems:**
- 220 lines of boilerplate per page
- Change sidebar? Edit 50 files
- Forget a CSS file? Broken styles
- Wrong CSS order? Components don't work

### After (New Way - DO THIS)

```html
{{define "property-list"}}
{{template "admin-base" .}}

{{define "content"}}
    <!-- Your actual page content (10 lines) -->
{{end}}

{{end}}
```

**Benefits:**
- 10 lines instead of 220
- Change sidebar? Edit 1 file (`shared/components/admin-sidebar.html`)
- CSS order enforced by base template
- Can't forget to include CSS - it's automatic

---

## Rules (MUST FOLLOW)

### ✅ DO

1. **Always extend a base template** - Never create standalone HTML
2. **Use shared components** - Sidebar, header, footer come from base
3. **Define only the `content` block** - Unless you need optional blocks
4. **Pass required data** - `.PageTitle`, `.User`, etc.
5. **Keep content focused** - Only YOUR page's content in the block

### ❌ DON'T

1. **Don't create new base templates** - Use admin-base or consumer-base
2. **Don't copy base template code** - Extend it, don't duplicate
3. **Don't define sidebar/header in pages** - They come from base
4. **Don't change CSS loading order** - It's set in base template
5. **Don't use inline styles** - Use CSS classes from design system

---

## Examples of Good Templates

### Admin Dashboard

```html
{{define "admin-dashboard"}}
{{template "admin-base" .}}

{{define "layout-data"}}
x-data="dashboardData()" x-init="init()"
{{end}}

{{define "content"}}
    <!-- MFA Prompt -->
    <div x-data="mfaPromptData()" x-init="init()">
        <div class="mfa-prompt" x-show="!dismissed && !hasMFA">
            <!-- MFA prompt content -->
        </div>
    </div>
    
    <!-- Critical Actions Bar -->
    <section class="dashboard-section">
        <div class="critical-actions-bar">
            <!-- Critical items -->
        </div>
    </section>
    
    <!-- Stats Grid -->
    <section class="dashboard-section">
        <div class="stats-grid-4">
            <!-- Stat cards -->
        </div>
    </section>
{{end}}

{{define "scripts"}}
    <script src="/static/js/dashboard.js"></script>
    <script src="/static/js/mfa-prompt.js"></script>
{{end}}

{{end}}
```

### Consumer Homepage

```html
{{define "homepage"}}
{{template "consumer-base" .}}

{{define "content"}}
    <!-- Hero Section -->
    <section class="hero-section">
        <div class="container">
            <h1 class="hero-title">Find Your Dream Rental</h1>
            <p class="hero-subtitle">Premium properties in Houston's best neighborhoods</p>
            <a href="/properties" class="btn btn-primary-consumer">Browse Properties</a>
        </div>
    </section>
    
    <!-- Featured Properties -->
    <section class="featured-section">
        <div class="container">
            <h2 class="section-label">Featured Properties</h2>
            <div class="property-grid">
                {{range .FeaturedProperties}}
                    {{template "property-card" .}}
                {{end}}
            </div>
        </div>
    </section>
{{end}}

{{define "scripts"}}
    <script src="/static/js/homepage-properties.js"></script>
{{end}}

{{end}}
```

---

## Enforcement

### Build-Time Checks (Future)

```bash
# Check that all pages extend a base template
./scripts/validate-templates.sh

# Check for inline styles
grep -r 'style="' web/templates/admin/pages/ && exit 1
grep -r 'style="' web/templates/consumer/pages/ && exit 1

# Check for duplicate sidebar/header definitions
# (should only exist in shared/components/)
```

---

## Support

**Questions?**
- Check existing templates in `admin/pages/` and `consumer/pages/`
- Review the base templates in `shared/layouts/`
- Ask the team

**Found a bug in base template?**
- Fix it in `shared/layouts/admin-base.html` or `consumer-base.html`
- All pages get the fix automatically

---

**Last Updated:** December 2025  
**Version:** 1.0 - Initial enforcement architecture
