# PropertyHub Go Template System

## Directory Structure

```
templates/
├── admin/
│   ├── base.html                      # Base template for all admin pages
│   ├── components/
│   │   ├── sidebar.html               # Admin sidebar navigation
│   │   ├── header.html                # Admin header with user menu
│   │   └── footer.html                # Admin footer
│   └── property-crud-manager.html     # Property CRUD page (example)
└── consumer/
    └── (to be created)
```

## How It Works

### 1. Base Template (`admin/base.html`)

The base template defines the overall page structure and includes all components:

```go
{{define "admin-base"}}
<!DOCTYPE html>
<html>
<head>
    <title>{{block "title" .}}Admin - PropertyHub{{end}}</title>
    <!-- CSS files -->
</head>
<body class="{{block "body-class" .}}admin-page{{end}}">
    <div class="admin-layout">
        {{template "admin-sidebar" .}}
        <main class="admin-main">
            {{template "admin-header" .}}
            <div class="admin-content">
                {{block "content" .}}{{end}}
            </div>
            {{template "admin-footer" .}}
        </main>
    </div>
    {{block "scripts" .}}{{end}}
</body>
</html>
{{end}}
```

### 2. Page Templates

Each page only defines its unique content:

```go
{{define "title"}}Property Management{{end}}
{{define "body-class"}}property-crud-page{{end}}

{{define "content"}}
    <!-- Page-specific content here -->
{{end}}

{{define "scripts"}}
    <!-- Page-specific scripts here -->
{{end}}

{{template "admin-base" .}}
```

### 3. Components

Components are reusable pieces included in the base template:

- **sidebar.html**: Navigation sidebar (single source of truth)
- **header.html**: Page header with user menu
- **footer.html**: Page footer

## Template Data Structure

Each page receives a data struct with these fields:

```go
type AdminPageData struct {
    // Common fields (all pages)
    PageTitle      string
    ActiveSection  string  // "dashboard", "properties", "leads", etc.
    ActivePage     string  // "property-crud", "analytics", etc.
    UserName       string
    UserInitials   string
    UserRole       string
    PropertyCount  int
    LeadCount      int
    BookingCount   int
    
    // Page-specific fields
    // (varies by page)
}
```

## Benefits

1. **Single Source of Truth**: Sidebar, header, footer defined once
2. **Automatic Consistency**: All pages use same components
3. **No Duplication**: Pages only define unique content
4. **Type-Safe**: Go compiler catches template errors
5. **Easy Maintenance**: Change sidebar once, updates everywhere

## Creating a New Admin Page

1. Create new template file in `templates/admin/`
2. Define title, body-class, and content blocks
3. Call `{{template "admin-base" .}}`  at the end
4. Load template in Go backend
5. Render with `tmpl.ExecuteTemplate(w, "admin-base", data)`

Example:

```go
// templates/admin/my-new-page.html
{{define "title"}}My New Page{{end}}
{{define "body-class"}}my-new-page{{end}}

{{define "content"}}
    <h2>My Page Content</h2>
{{end}}

{{template "admin-base" .}}
```

## Next Steps

1. Update Go backend to load templates
2. Test Property CRUD page
3. Migrate remaining admin pages
4. Create consumer template system

