# UI Architecture Documentation

## Overview
This document defines the zone architecture and shared component patterns for PropertyHub admin interface. All admin templates must follow this architecture to ensure consistency and maintainability.

## Zone Architecture

### Layout Structure
```
<body>
  <div class="admin-layout" x-data="pageData()" x-init="init()">
    Zone A: Sidebar
    <main class="admin-main">
      Zone B: Top Bar
      Zone C: Content Area
        - Action Bar (optional)
        - Page Content
    </main>
  </div>
  Zone D: Footer Scripts
</body>
```

## Zones Description

### Zone A: Sidebar (`_sidebar.html`)
**Purpose:** Global navigation across all admin pages
**Location:** Left side of viewport, fixed position
**Components:**
- Logo/branding
- Navigation sections with expandable menus
- Dynamic badges showing counts (leads, properties, etc.)
- Active state management

**Key Features:**
- Collapsible sections with Alpine.js
- Real-time count updates via sidebar data
- Icon-based navigation
- Active page highlighting

**Usage:** Include in every admin page
```html
{{template "_sidebar.html" .}}
```

### Zone B: Top Bar (`_topbar.html`)
**Purpose:** Page-level controls and user context
**Location:** Top of main content area, below sidebar header
**Components:**
- Page title (dynamic)
- Search functionality
- Visitor counter (dashboard only)
- User profile menu
- Logout button

**Key Features:**
- Responsive layout
- Real-time visitor tracking
- User profile display
- Quick search access

**Usage:** Include in main content area
```html
{{template "_topbar.html" .}}
```

### Zone C: Content Area
**Purpose:** Page-specific content and functionality
**Location:** Main scrollable area
**Sub-zones:**
1. **Action Bar** (optional, for list pages)
   - Filters, sorting, bulk actions
   - Add new item button
   - View toggle (grid/list)

2. **Content Body**
   - Cards, tables, forms
   - Following Scout design system
   - Consistent spacing and typography

**Patterns by Page Type:**

#### Dashboard Pattern
```html
<div class="admin-content">
  <section class="dashboard-section">
    <div class="section-label admin-section-label">...</div>
    <div class="stats-grid">...</div>
  </section>
</div>
```

#### List Page Pattern
```html
<div class="action-bar">
  <!-- Filters, search, add button -->
</div>
<div class="admin-content">
  <div class="data-table">...</div>
  <div class="pagination">...</div>
</div>
```

#### Detail Page Pattern
```html
<div class="admin-content">
  <div class="detail-header">...</div>
  <div class="detail-tabs">...</div>
  <div class="detail-content">...</div>
</div>
```

#### Form Page Pattern
```html
<div class="admin-content">
  <div class="form-card">
    <div class="form-section">...</div>
  </div>
  <div class="form-actions">
    <button type="submit">Save</button>
  </div>
</div>
```

#### Settings Pattern
```html
<div class="admin-content">
  <div class="settings-tabs">...</div>
  <div class="settings-card">
    <div class="settings-card-header">...</div>
    <div class="settings-card-body">...</div>
  </div>
</div>
```

#### Success Page Pattern
```html
<div class="admin-content">
  <div class="success-card">
    <div class="success-icon">✓</div>
    <h2>Action Completed</h2>
    <p>Details...</p>
    <div class="success-actions">
      <a href="..." class="btn btn-primary">Next Action</a>
    </div>
  </div>
</div>
```

### Zone D: Footer Scripts (`_footer_scripts.html`)
**Purpose:** Load Alpine.js and page-specific JavaScript
**Location:** End of body tag
**Components:**
- Alpine.js CDN
- Chart.js (if needed)
- Page-specific external JS files

**Usage:** Include before closing body tag
```html
{{template "_footer_scripts.html" .}}
```

## Shared Partials

### 1. _head.html
**Contains:**
- Meta tags (charset, viewport)
- Google Fonts preconnect and loading
- Title tag with dynamic page name
- All CSS files (design tokens, utilities, layouts, components)
- Alpine.js defer script tag

**Template Variables:**
- `.PageTitle` - Dynamic page title

### 2. _sidebar.html
**Contains:**
- Complete sidebar navigation structure
- Logo and branding
- All navigation sections
- Dynamic badge counters

**Template Variables:**
- `.SidebarCounts` - Object with count data for badges

**Alpine.js Data:**
- Expects parent component to provide `sidebarCounts` object

### 3. _topbar.html
**Contains:**
- Page title header
- Search bar
- Visitor counter (conditional)
- User profile section
- Logout button

**Template Variables:**
- `.PageTitle` - Main page heading
- `.ShowVisitorCounter` - Boolean to show/hide counter
- `.UserProfile` - User data object

### 4. _footer_scripts.html
**Contains:**
- Alpine.js CDN (if not in head)
- Chart.js CDN (conditional)
- Common utility scripts
- Page-specific script includes

## File Organization

```
web/templates/admin/
├── partials/
│   ├── _head.html
│   ├── _sidebar.html
│   ├── _topbar.html
│   └── _footer_scripts.html
└── pages/
    ├── admin-dashboard.html
    ├── lead-management.html
    ├── property-list.html
    └── ... (27 total pages)
```

## Implementation Rules

### Required in ALL Admin Pages

1. **Head Section**
   ```html
   {{template "_head.html" .}}
   ```

2. **Body Structure**
   ```html
   <body>
     <div class="admin-layout" x-data="pageData()" x-init="init()">
       {{template "_sidebar.html" .}}
       <main class="admin-main">
         {{template "_topbar.html" .}}
         <div class="admin-content">
           <!-- Page content -->
         </div>
       </main>
     </div>
     {{template "_footer_scripts.html" .}}
   </body>
   ```

3. **Alpine.js Data Function**
   - Every page must define a unique data function
   - Example: `dashboardData()`, `leadManagementData()`, `propertyListData()`
   - Function must include `init()` method

### Prohibited Patterns

❌ **NO** inline `<style>` tags
❌ **NO** inline styles on elements
❌ **NO** inline `<script>` blocks (except small Alpine.js components)
❌ **NO** class attribute on `<body>` tag
❌ **NO** duplicate sidebar/header markup
❌ **NO** hardcoded user data

### Best Practices

✅ Use external CSS files only
✅ Use external JS files for page logic
✅ Use Alpine.js for interactive components
✅ Use Scout design system classes
✅ Use template variables for dynamic content
✅ Keep Alpine.js functions in external files
✅ Follow consistent naming conventions

## Alpine.js Conventions

### Data Function Structure
```javascript
function pageNameData() {
  return {
    // State variables
    loading: true,
    items: [],
    
    // Initialization
    async init() {
      await this.loadData();
    },
    
    // Methods
    async loadData() {
      // Fetch data
    }
  };
}
```

### Component Naming
- Page data: `pageNameData()`
- Reusable components: `componentName()`
- Utilities: `utilityName()`

## CSS Class Conventions

### Layout Classes
- `.admin-layout` - Root layout wrapper
- `.admin-sidebar` - Sidebar zone
- `.admin-main` - Main content zone
- `.admin-header` - Top bar (topbar)
- `.admin-content` - Page content area

### Component Classes
- `.nav-section` - Sidebar navigation section
- `.stat-card` - Dashboard stat card
- `.data-table` - List page table
- `.form-card` - Form container
- `.settings-card` - Settings section

### Utility Classes
- Use Scout utility classes from `admin-scout-utilities.css`
- Follow design token patterns from `admin-design-tokens.css`

## Template Data Contract

### Every Admin Page Receives:
```go
type AdminPageData struct {
    PageTitle           string
    UserProfile         UserProfile
    SidebarCounts       SidebarCounts
    CSRFToken           string
    ShowVisitorCounter  bool
    ShowActionBar       bool
    // Page-specific data
}
```

### UserProfile
```go
type UserProfile struct {
    Username  string
    Role      string
    Initials  string
    AvatarURL string
}
```

### SidebarCounts
```go
type SidebarCounts struct {
    TotalLeads         int
    HotLeads           int
    WarmLeads          int
    ActiveProperties   int
    PendingApplications int
    ConfirmedBookings  int
    ClosingPipeline    int
    PendingImages      int
}
```

## Migration Checklist

When refactoring an existing page:

- [ ] Remove old `<head>` content, replace with `{{template "_head.html" .}}`
- [ ] Remove `class` attribute from `<body>` tag
- [ ] Add `.admin-layout` wrapper with Alpine.js
- [ ] Remove old sidebar markup, replace with `{{template "_sidebar.html" .}}`
- [ ] Remove old header markup, replace with `{{template "_topbar.html" .}}`
- [ ] Wrap content in `<main class="admin-main">` if not present
- [ ] Move all inline CSS to external files
- [ ] Move all inline JavaScript to external files
- [ ] Update Alpine.js data function names to be unique
- [ ] Add `{{template "_footer_scripts.html" .}}` before `</body>`
- [ ] Test navigation and interactivity
- [ ] Verify responsive layout
- [ ] Check browser console for errors

## Browser Support

- Modern evergreen browsers (Chrome, Firefox, Safari, Edge)
- Alpine.js 3.x compatible
- CSS Grid and Flexbox support required
- ES6+ JavaScript features

## Performance Considerations

1. **CSS Loading**
   - All CSS loaded in head for consistent rendering
   - Minimize CSS file count
   - Use design tokens for consistency

2. **JavaScript Loading**
   - Alpine.js deferred in head
   - Page scripts loaded at end of body
   - Lazy load charts and heavy libraries

3. **Alpine.js Best Practices**
   - Use `x-show` instead of `x-if` for frequently toggled elements
   - Debounce search inputs
   - Use `$nextTick` for DOM updates
   - Avoid watchers when possible

## Testing Requirements

### Visual Regression
- Sidebar consistent across all pages
- Top bar consistent across all pages
- Spacing and typography consistent
- Responsive layouts work on mobile

### Functional Testing
- Navigation works on all pages
- Alpine.js components initialize
- Search functionality works
- User profile displays correctly
- Logout redirects properly

### Accessibility
- Semantic HTML structure
- ARIA labels where needed
- Keyboard navigation support
- Focus states visible

## Version History

- v1.0 - Initial architecture definition (SCO-081)
- v1.1 - Added shared partials specification (SCO-082)
- v1.2 - Phase 2 implementation (SCO-083)
