# PropertyHub Admin Components

This directory contains the **canonical reusable components** for all admin templates.

## Components

### 1. admin-sidebar.html
**The canonical admin sidebar** extracted from production dashboard.

**Structure:**
- Sidebar Header (PH logo + "PropertyHub")
- 7 Navigation Sections:
  1. **Overview (OV)** - Link to /admin/dashboard
  2. **Properties & Showings (PS)** - 8 sub-items
  3. **Leads & Conversion (LC)** - 5 sub-items
  4. **Communications (CM)** - 2 sub-items
  5. **Workflow (WF)** - 3 sub-items
  6. **Analytics & Reports (AR)** - 5 sub-items
  7. **System (SY)** - 7 sub-items

**Usage:**
```html
<!-- Copy-paste this entire file into every admin template -->
<aside class="admin-sidebar">
    ...
</aside>
```

**Customization:**
- Mark active page: Add `active` class to the current page's nav-item
- Expand active section: Add `expanded` class to nav-section-items

---

### 2. admin-header.html
**The canonical admin top bar** extracted from production dashboard.

**Structure:**
- Left: Page Title (h1)
- Right: Search | Notifications (🔔 + badge) | User Menu with Dropdown

**User Dropdown Menu:**
- Profile
- Settings
- Help & Support
- Logout

**Usage:**
```html
<!-- Copy-paste and replace {{PAGE_TITLE}} with actual page title -->
<div class="admin-top-bar">
    <div class="admin-page-header">
        <h1 class="admin-page-title">{{PAGE_TITLE}}</h1>
    </div>
    ...
</div>
```

**Customization:**
- Replace `{{PAGE_TITLE}}` with the actual page title (e.g., "Property Management", "Lead Management")

---

## Standard Template Structure

Every admin template should follow this exact structure:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{PAGE_TITLE}} - PropertyHub Admin</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Playfair+Display:wght@400;700;800&family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    
    <link rel="stylesheet" href="/static/css/design-tokens.css">
    <link rel="stylesheet" href="/static/css/scout-utilities.css">
    <link rel="stylesheet" href="/static/css/scout-layouts.css">
    <link rel="stylesheet" href="/static/css/scout-components.css">
</head>
<body class="{{page-slug}}-page">
    <div class="admin-layout">

        <!-- COPY FROM: components/admin-sidebar.html -->
        <aside class="admin-sidebar">
            ...
        </aside>

        <main class="admin-main">

            <!-- COPY FROM: components/admin-header.html -->
            <div class="admin-top-bar">
                ...
            </div>

            <!-- YOUR PAGE-SPECIFIC CONTENT -->
            <div class="admin-content">
                <div class="admin-content-inner">
                    
                    <!-- Page content here -->
                    
                </div>
            </div>

        </main>
    </div>

    <script src="/static/js/admin-sidebar.js"></script>
    
    <!-- Page-specific JavaScript -->
    <script>
        // Your JS here
    </script>

</body>
</html>
```

---

## Rules

1. **DO NOT MODIFY** these components without updating ALL admin templates
2. **COPY-PASTE EXACTLY** - Do not create variations
3. **ONLY CUSTOMIZE** the page title and active nav item
4. **USE SCOUT CSS** for all page-specific content (no new CSS)
5. **FOLLOW ADMIN_TEMPLATE_STANDARD.md** for complete guidelines

---

## Files Using These Components

✅ **Lead Management** - `/opt/propertyhub/web/templates/pages/lead-management.html`  
✅ **Property CRUD Manager** - `/opt/propertyhub/web/templates/admin/property-crud-manager.html`  

**Remaining:** 48 templates need standardization

---

**Last Updated:** October 29, 2025  
**Source:** Production Dashboard (https://llotschedule.online/admin/dashboard)
