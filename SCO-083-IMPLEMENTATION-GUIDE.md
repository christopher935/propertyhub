# SCO-083: Phase 2 - Apply Architecture to All 27 Admin Templates

## Implementation Status

### ✅ Completed

1. **UI_ARCHITECTURE.md** - Comprehensive architecture documentation created
2. **Shared Partials Created** (4 files):
   - `_head.html` - Standardized head section with fonts, CSS, and Alpine.js
   - `_sidebar.html` - Complete sidebar navigation with dynamic badges
   - `_topbar.html` - Top bar with search, user profile, visitor counter
   - `_footer_scripts.html` - Script loading (Chart.js, page-specific JS)

3. **High Priority Templates Refactored** (2/4):
   - ✅ `admin-dashboard.html` - Fully refactored, JS extracted to `/static/js/dashboard.js`
   - ✅ `settings.html` - Refactored, CSS extracted to `/static/css/settings-page.css`
   - ⏳ `lead-management.html` - Needs refactoring
   - ⏳ `property-list.html` - Needs refactoring

4. **External Assets Created**:
   - `/static/js/dashboard.js` - All dashboard Alpine.js functions
   - `/static/css/settings-page.css` - Settings page form styles

### 🔄 Remaining Work

**High Priority (2 files):**
- `lead-management.html`
- `property-list.html`

**Medium Priority (16 files):**
- `application-workflow.html`
- `booking-detail.html`
- `business-intelligence.html`
- `command-center.html`
- `communication-center.html`
- `customer-feedback.html`
- `feedback-detail.html`
- `lead-add.html`
- `lead-assignment.html`
- `property-create.html`
- `property-detail.html`
- `property-edit.html`
- `system-settings.html`
- `team-dashboard.html`
- `team-member-add.html`
- `team-member-edit.html`

**Success Pages (6 files):**
- `commission-updated-success.html`
- `lead-added-success.html`
- `property-added-success.html`
- `property-updated-success.html`
- `team-member-added-success.html`
- `team-member-updated-success.html`

**Special Case (1 file):**
- `agent-mobile.html`

---

## Refactoring Pattern

### Step-by-Step Process

#### 1. Read the Current Template
```bash
wc -l <template-file>
grep -n "<script>" <template-file>
grep -c "style=" <template-file>
```

#### 2. Extract Inline Scripts (if any)
- Identify all Alpine.js functions and data structures
- Create external JS file: `/static/js/<page-name>.js`
- Move all JavaScript functions to external file
- Keep only small Alpine.js component definitions inline if necessary

#### 3. Extract Inline Styles (if any)
- Identify all `<style>` blocks and inline `style=` attributes
- Create external CSS file: `/static/css/<page-name>.css`
- Move all CSS to external file
- Remove `<style>` tags and inline styles

#### 4. Replace Head Section
**Before:**
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <!-- ... more head content ... -->
    <title>Page Title - PropertyHub</title>
    <link rel="stylesheet" href="/static/css/...">
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
```

**After:**
```html
<!DOCTYPE html>
<html lang="en">
{{template "_head.html" .}}
<body>
```

**Note:** The Go handler must pass `PageTitle` in the template data.

#### 5. Replace Sidebar
**Before:**
```html
<aside class="admin-sidebar">
    <div class="sidebar-header">
        <a href="/" class="sidebar-logo">
            <!-- ... entire sidebar navigation ... -->
        </a>
    </div>
    <nav class="admin-sidebar-nav">
        <!-- ... many nav sections ... -->
    </nav>
</aside>
```

**After:**
```html
{{template "_sidebar.html" .}}
```

**Note:** The Go handler must pass:
- `ActivePage` - Current page identifier (e.g., "dashboard", "lead-management")
- `ActiveSection` - Current section (e.g., "leads", "properties", "intelligence")
- `SidebarCounts` - Object with badge counts

#### 6. Replace Top Bar
**Before:**
```html
<header class="admin-header">
    <div class="admin-header-content">
        <h1 class="admin-header-title">Page Title</h1>
        <div class="admin-header-actions">
            <!-- search, user menu, logout -->
        </div>
    </div>
</header>
```

**After:**
```html
{{template "_topbar.html" .}}
```

**Note:** The Go handler must pass:
- `PageTitle` - Page heading
- `ShowVisitorCounter` - Boolean (true for dashboard only)
- `UserProfile` - User data object

#### 7. Structure Main Content
**Before:**
```html
<div class="admin-layout">
    <aside class="admin-sidebar">...</aside>
    <main class="admin-main">
        <header class="admin-header">...</header>
        <div class="admin-content" x-data="pageData()" x-init="init()">
            <!-- content -->
        </div>
    </main>
</div>
```

**After:**
```html
<div class="admin-layout" x-data="pageData()" x-init="init()">
    {{template "_sidebar.html" .}}
    
    <main class="admin-main">
        {{template "_topbar.html" .}}
        
        {{if .ShowActionBar}}
        <div class="action-bar">
            <!-- Page-specific action bar content -->
        </div>
        {{end}}
        
        <div class="admin-content">
            <!-- Page-specific content -->
        </div>
    </main>
</div>
```

**Note:** Alpine.js `x-data` goes on `.admin-layout` wrapper, not on `.admin-content`.

#### 8. Add Footer Scripts
**Before:**
```html
        </main>
    </div>
    
    <script>
        function pageData() { ... }
    </script>
</body>
</html>
```

**After:**
```html
        </main>
    </div>
    
    {{template "_footer_scripts.html" .}}
</body>
</html>
```

**Note:** The Go handler must pass:
- `NeedsCharts` - Boolean (true if page uses Chart.js)
- `PageScript` - Filename of page-specific JS (e.g., "dashboard.js")

---

## Template Patterns by Page Type

### Dashboard Pages
**Example:** `admin-dashboard.html`

**Structure:**
```html
<!DOCTYPE html>
<html lang="en">
{{template "_head.html" .}}
<body>
    <div class="admin-layout" x-data="dashboardData()" x-init="init()">
        {{template "_sidebar.html" .}}
        
        <main class="admin-main">
            {{template "_topbar.html" .}}
            
            <div class="admin-content">
                <section class="dashboard-section">
                    <div class="section-label admin-section-label">
                        <span class="section-label-line"></span>
                        <span class="section-label-text">Section Title</span>
                        <span class="section-label-line"></span>
                    </div>
                    <div class="stats-grid stats-grid-4">
                        <!-- stat cards -->
                    </div>
                </section>
            </div>
        </main>
    </div>
    
    {{template "_footer_scripts.html" .}}
</body>
</html>
```

**Go Handler Requirements:**
```go
data := map[string]interface{}{
    "PageTitle": "Dashboard",
    "ActivePage": "dashboard",
    "ActiveSection": "",
    "ShowVisitorCounter": true,
    "UserProfile": userProfile,
    "SidebarCounts": sidebarCounts,
    "NeedsCharts": true,
    "PageScript": "dashboard.js",
}
```

### List Pages
**Examples:** `lead-management.html`, `property-list.html`

**Structure:**
```html
<!DOCTYPE html>
<html lang="en">
{{template "_head.html" .}}
<body>
    <div class="admin-layout" x-data="listPageData()" x-init="init()">
        {{template "_sidebar.html" .}}
        
        <main class="admin-main">
            {{template "_topbar.html" .}}
            
            <div class="action-bar">
                <div class="action-bar-left">
                    <div class="filter-group">
                        <select x-model="filters.status">
                            <option value="">All Status</option>
                            <option value="active">Active</option>
                        </select>
                    </div>
                    <div class="search-box">
                        <input type="search" placeholder="Search..." x-model="searchQuery">
                    </div>
                </div>
                <div class="action-bar-right">
                    <button class="btn btn-primary" @click="addNew()">
                        <svg>...</svg>
                        Add New
                    </button>
                </div>
            </div>
            
            <div class="admin-content">
                <div class="data-table">
                    <table>
                        <thead>...</thead>
                        <tbody>
                            <template x-for="item in items" :key="item.id">
                                <tr>...</tr>
                            </template>
                        </tbody>
                    </table>
                </div>
                
                <div class="pagination">
                    <!-- pagination controls -->
                </div>
            </div>
        </main>
    </div>
    
    {{template "_footer_scripts.html" .}}
</body>
</html>
```

**Go Handler Requirements:**
```go
data := map[string]interface{}{
    "PageTitle": "Lead Management",
    "ActivePage": "lead-management",
    "ActiveSection": "leads",
    "ShowActionBar": true,
    "ShowVisitorCounter": false,
    "UserProfile": userProfile,
    "SidebarCounts": sidebarCounts,
    "PageScript": "lead-management.js",
}
```

### Form Pages
**Examples:** `lead-add.html`, `property-create.html`, `team-member-add.html`

**Structure:**
```html
<!DOCTYPE html>
<html lang="en">
{{template "_head.html" .}}
<body>
    <div class="admin-layout" x-data="formPageData()" x-init="init()">
        {{template "_sidebar.html" .}}
        
        <main class="admin-main">
            {{template "_topbar.html" .}}
            
            <div class="admin-content">
                <form class="form-card" @submit.prevent="handleSubmit">
                    <div class="form-section">
                        <h3>Section Title</h3>
                        <div class="form-row">
                            <div class="form-group">
                                <label>Field Label</label>
                                <input type="text" x-model="formData.field">
                            </div>
                        </div>
                    </div>
                    
                    <div class="form-actions">
                        <button type="button" class="btn btn-secondary" @click="cancel()">Cancel</button>
                        <button type="submit" class="btn btn-primary">Save</button>
                    </div>
                </form>
            </div>
        </main>
    </div>
    
    {{template "_footer_scripts.html" .}}
</body>
</html>
```

### Detail Pages
**Examples:** `property-detail.html`, `feedback-detail.html`, `booking-detail.html`

**Structure:**
```html
<!DOCTYPE html>
<html lang="en">
{{template "_head.html" .}}
<body>
    <div class="admin-layout" x-data="detailPageData()" x-init="init()">
        {{template "_sidebar.html" .}}
        
        <main class="admin-main">
            {{template "_topbar.html" .}}
            
            <div class="admin-content">
                <div class="detail-header">
                    <h2>Item Title</h2>
                    <div class="detail-actions">
                        <button class="btn btn-secondary">Edit</button>
                        <button class="btn btn-danger">Delete</button>
                    </div>
                </div>
                
                <div class="detail-tabs">
                    <button @click="activeTab = 'overview'" :class="{ 'active': activeTab === 'overview' }">Overview</button>
                    <button @click="activeTab = 'history'" :class="{ 'active': activeTab === 'history' }">History</button>
                </div>
                
                <div class="detail-content">
                    <div x-show="activeTab === 'overview'">
                        <!-- overview content -->
                    </div>
                    <div x-show="activeTab === 'history'">
                        <!-- history content -->
                    </div>
                </div>
            </div>
        </main>
    </div>
    
    {{template "_footer_scripts.html" .}}
</body>
</html>
```

### Settings Pages
**Example:** `settings.html`, `system-settings.html`

**Structure:**
```html
<!DOCTYPE html>
<html lang="en">
{{template "_head.html" .}}
<body>
    <link rel="stylesheet" href="/static/css/settings-page.css">
    <div class="admin-layout">
        {{template "_sidebar.html" .}}
        
        <main class="admin-main">
            {{template "_topbar.html" .}}
            
            <div class="admin-content">
                <div class="settings-tabs">
                    <button class="settings-tab active" data-tab="general">General</button>
                    <button class="settings-tab" data-tab="advanced">Advanced</button>
                </div>
                
                <div id="general-tab" class="tab-content active">
                    <div class="settings-card">
                        <div class="settings-card-header">
                            <div class="section-label admin-section-label">
                                <span class="section-label-line"></span>
                                <span class="section-label-text">Section Title</span>
                                <span class="section-label-line"></span>
                            </div>
                        </div>
                        <div class="settings-card-body">
                            <form>
                                <!-- form fields -->
                            </form>
                        </div>
                    </div>
                </div>
            </div>
        </main>
    </div>
    
    {{template "_footer_scripts.html" .}}
</body>
</html>
```

### Success Pages
**Examples:** All `*-success.html` files

**Structure:**
```html
<!DOCTYPE html>
<html lang="en">
{{template "_head.html" .}}
<body>
    <div class="admin-layout">
        {{template "_sidebar.html" .}}
        
        <main class="admin-main">
            {{template "_topbar.html" .}}
            
            <div class="admin-content">
                <div class="success-card">
                    <div class="success-icon">✓</div>
                    <h2>Action Completed Successfully</h2>
                    <p>Details about what was completed.</p>
                    <div class="success-actions">
                        <a href="/admin/list" class="btn btn-primary">View All Items</a>
                        <a href="/admin/add" class="btn btn-secondary">Add Another</a>
                    </div>
                </div>
            </div>
        </main>
    </div>
    
    {{template "_footer_scripts.html" .}}
</body>
</html>
```

---

## Backend Requirements

### Go Template Data Structure

Every admin page handler must provide this data structure:

```go
type AdminPageData struct {
    // Required for all pages
    PageTitle          string
    ActivePage         string
    ActiveSection      string
    UserProfile        UserProfile
    SidebarCounts      SidebarCounts
    CSRFToken          string
    
    // Optional (page-specific)
    ShowVisitorCounter bool
    ShowActionBar      bool
    NeedsCharts        bool
    PageScript         string
    AdditionalCSS      string
    
    // Page-specific data
    // ... custom fields per page
}

type UserProfile struct {
    Username  string
    Role      string
    Initials  string
    AvatarURL string
}

type SidebarCounts struct {
    TotalLeads          int
    HotLeads            int
    WarmLeads           int
    ActiveProperties    int
    PendingApplications int
    ConfirmedBookings   int
    ClosingPipeline     int
    PendingImages       int
}
```

### Example Handler

```go
func AdminDashboardHandler(c *gin.Context) {
    user := GetCurrentUser(c)
    
    data := AdminPageData{
        PageTitle:          "Dashboard",
        ActivePage:         "dashboard",
        ActiveSection:      "",
        ShowVisitorCounter: true,
        NeedsCharts:        true,
        PageScript:         "dashboard.js",
        AdditionalCSS:      "dashboard-interactions.css",
        CSRFToken:          csrf.Token(c.Request),
        UserProfile: UserProfile{
            Username: user.Username,
            Role:     user.Role,
            Initials: getInitials(user.Username),
        },
        SidebarCounts: fetchSidebarCounts(),
    }
    
    c.HTML(http.StatusOK, "admin/pages/admin-dashboard.html", data)
}
```

---

## Checklist for Each Template

Use this checklist when refactoring each template:

### File Preparation
- [ ] Read current template and note line count
- [ ] Identify inline scripts (locate `<script>` tags)
- [ ] Identify inline styles (locate `<style>` tags and `style=` attributes)
- [ ] Identify Alpine.js data functions
- [ ] Note any unique page features

### Extract Assets
- [ ] Create `/static/js/<page-name>.js` for scripts (if needed)
- [ ] Create `/static/css/<page-name>.css` for styles (if needed)
- [ ] Move all JavaScript functions to external file
- [ ] Move all CSS rules to external file

### Template Refactoring
- [ ] Replace `<head>` section with `{{template "_head.html" .}}`
- [ ] Remove `class` attribute from `<body>` tag
- [ ] Add `.admin-layout` wrapper with `x-data` and `x-init`
- [ ] Replace sidebar with `{{template "_sidebar.html" .}}`
- [ ] Replace top bar with `{{template "_topbar.html" .}}`
- [ ] Wrap content in `<main class="admin-main">` (if not present)
- [ ] Add action bar for list pages (if applicable)
- [ ] Remove all inline `<style>` tags
- [ ] Remove all inline `style=` attributes
- [ ] Remove large inline `<script>` blocks
- [ ] Add `{{template "_footer_scripts.html" .}}` before `</body>`
- [ ] Link external CSS file (if created)

### Backend Updates
- [ ] Update Go handler to provide `PageTitle`
- [ ] Update Go handler to provide `ActivePage`
- [ ] Update Go handler to provide `ActiveSection` (if applicable)
- [ ] Update Go handler to provide `UserProfile`
- [ ] Update Go handler to provide `SidebarCounts`
- [ ] Update Go handler to provide `ShowVisitorCounter` (if applicable)
- [ ] Update Go handler to provide `ShowActionBar` (if applicable)
- [ ] Update Go handler to provide `NeedsCharts` (if applicable)
- [ ] Update Go handler to provide `PageScript` (if applicable)
- [ ] Update Go handler to provide `AdditionalCSS` (if applicable)

### Testing
- [ ] Template compiles without errors
- [ ] Sidebar displays correctly
- [ ] Top bar displays correctly with user info
- [ ] Page title is correct
- [ ] Navigation active states work
- [ ] Sidebar badges show counts
- [ ] Alpine.js components initialize
- [ ] Charts render (if applicable)
- [ ] Forms submit correctly (if applicable)
- [ ] Responsive layout works on mobile

---

## Common Issues and Solutions

### Issue: Template Parse Error
**Error:** `template: admin/pages/xxx.html: unexpected "<" in operand`

**Solution:** Ensure all template directives use `{{` and `}}`, not HTML tags.

### Issue: Sidebar Not Displaying
**Cause:** Missing template data

**Solution:** Ensure Go handler passes `ActivePage`, `ActiveSection`, and `SidebarCounts`.

### Issue: Alpine.js Not Working
**Cause:** Scripts loaded in wrong order or `x-data` in wrong location

**Solution:** 
1. Ensure Alpine.js loads in `_head.html` with `defer`
2. Place `x-data` on `.admin-layout` wrapper
3. External JS files load after Alpine.js

### Issue: Styles Not Applied
**Cause:** CSS file not linked or loaded after Alpine renders

**Solution:** Link CSS files immediately after opening `<body>` tag if page-specific.

### Issue: Charts Not Rendering
**Cause:** Chart.js not loaded

**Solution:** Set `NeedsCharts: true` in Go handler data.

---

## Quick Reference: File Locations

### Templates
```
web/templates/admin/
├── partials/
│   ├── _head.html
│   ├── _sidebar.html
│   ├── _topbar.html
│   └── _footer_scripts.html
└── pages/
    ├── admin-dashboard.html ✅
    ├── settings.html ✅
    ├── lead-management.html ⏳
    ├── property-list.html ⏳
    └── ... (23 more files)
```

### Static Assets
```
web/static/
├── css/
│   ├── admin-design-tokens.css
│   ├── admin-scout-utilities.css
│   ├── admin-scout-layouts.css
│   ├── scout-components-admin.css
│   ├── dashboard-interactions.css
│   ├── settings-page.css ✅
│   └── ... (add more page-specific CSS)
└── js/
    ├── admin.js
    ├── ux-enhancements.js
    ├── dashboard.js ✅
    └── ... (add more page-specific JS)
```

---

## Next Steps

1. **Complete High Priority Templates**:
   - Refactor `lead-management.html`
   - Refactor `property-list.html`

2. **Refactor Medium Priority Templates** (16 files):
   - Follow the pattern established
   - Extract scripts and styles
   - Use appropriate page type pattern

3. **Refactor Success Pages** (6 files):
   - These are simpler, follow success page pattern
   - No complex logic, mostly static content

4. **Handle Special Case**:
   - `agent-mobile.html` - May need different approach or rebuilding

5. **Update Go Handlers**:
   - Ensure all handlers provide required template data
   - Add sidebar count fetching logic
   - Add user profile fetching logic

6. **Testing**:
   - Test each page after refactoring
   - Verify navigation works
   - Check responsive layouts
   - Test Alpine.js interactions

7. **Documentation**:
   - Update handler documentation
   - Document any new Alpine.js components
   - Update deployment notes

---

## Estimated Effort

- **High Priority Remaining**: 2 files × 1 hour = 2 hours
- **Medium Priority**: 16 files × 0.5 hours = 8 hours
- **Success Pages**: 6 files × 0.25 hours = 1.5 hours
- **Special Case**: 1 file × 2 hours = 2 hours
- **Backend Updates**: 3 hours
- **Testing**: 3 hours

**Total: ~19.5 hours**

---

## Success Criteria

- ✅ All 27 admin templates use shared partials
- ✅ No inline `<style>` tags in any template
- ✅ No inline `style=` attributes in any template
- ✅ No large inline `<script>` blocks (small Alpine.js components OK)
- ✅ Sidebar identical across all pages
- ✅ Top bar identical across all pages (except page title)
- ✅ All pages follow zone architecture from UI_ARCHITECTURE.md
- ✅ All pages responsive and working on mobile
- ✅ All Alpine.js interactions functional
- ✅ All navigation links work correctly
- ✅ No browser console errors

---

**Document Version:** 1.0  
**Last Updated:** December 5, 2024  
**Status:** Phase 2 In Progress - 2/27 templates completed
