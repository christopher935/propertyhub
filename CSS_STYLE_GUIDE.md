# PropertyHub CSS Style Guide

**Version**: 2.0.0  
**Last Updated**: December 4, 2025

---

## 📐 Architecture Overview

PropertyHub uses a **4-layer CSS architecture** based on design tokens:

```
1. design-tokens.css    → Variables (colors, spacing, typography)
2. scout-utilities.css  → Utility classes (.text-center, .mt-4)
3. scout-layouts.css    → Layout patterns (.container, .grid, .flex)
4. scout-components.css → Component styles (.btn, .card, .property-card)
```

### File Loading Order (in HTML)
```html
<link rel="stylesheet" href="/static/css/design-tokens.css">
<link rel="stylesheet" href="/static/css/scout-utilities.css">
<link rel="stylesheet" href="/static/css/scout-layouts.css">
<link rel="stylesheet" href="/static/css/scout-components.css">
```

**Never change this order!** Each layer depends on the previous one.

---

## 🎨 Design Tokens

### How to Use Tokens

**✅ DO THIS:**
```css
.my-component {
  color: var(--navy-primary);
  padding: var(--space-4);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-md);
}
```

**❌ DON'T DO THIS:**
```css
.my-component {
  color: #1b3559;  /* Hardcoded! */
  padding: 16px;    /* Hardcoded! */
  border-radius: 8px;  /* Hardcoded! */
}
```

---

### Color Palette

#### Brand Colors
```css
--navy-primary: #1b3559
--navy-light: #2d4a73
--navy-dark: #0f1f35

--gold-primary: #c4a962  /* Main accent color */
--gold-light: #e6c966
--gold-dark: #b8941f
```

#### Semantic Colors
```css
--success: #10B981  (green)
--success-light: #D1FAE5
--success-dark: #059669

--error: #EF4444  (red)
--error-light: #FEE2E2
--error-dark: #DC2626

--warning: #F59E0B  (orange)
--warning-light: #FEF3C7
--warning-dark: #D97706

--info: #3B82F6  (blue)
--info-light: #DBEAFE
--info-dark: #1E40AF
```

#### Grayscale
```css
--gray-50: #F9FAFB   (lightest)
--gray-100: #F3F4F6
--gray-200: #E5E7EB
--gray-300: #D1D5DB
--gray-400: #9CA3AF
--gray-500: #6B7280  (middle)
--gray-600: #4B5563
--gray-700: #374151
--gray-800: #1F2937
--gray-900: #111827  (darkest)
```

---

### Spacing Scale

**Based on 4px increments:**
```css
--space-1: 0.25rem  (4px)
--space-2: 0.5rem   (8px)
--space-3: 0.75rem  (12px)
--space-4: 1rem     (16px)
--space-5: 1.25rem  (20px)
--space-6: 1.5rem   (24px)
--space-8: 2rem     (32px)
--space-10: 2.5rem  (40px)
--space-12: 3rem    (48px)
--space-16: 4rem    (64px)
--space-20: 5rem    (80px)
--space-24: 6rem    (96px)
```

**Usage:**
```css
.card {
  padding: var(--space-6);  /* 24px */
  margin-bottom: var(--space-4);  /* 16px */
  gap: var(--space-2);  /* 8px */
}
```

---

### Typography Scale

#### Font Families
```css
--font-family-primary: 'Playfair Display', serif  (headings)
--font-family-secondary: 'Inter', sans-serif  (body)
```

#### Font Sizes
```css
--font-size-xs: 0.75rem    (12px)
--font-size-sm: 0.875rem   (14px)
--font-size-base: 1rem     (16px)  ← Default
--font-size-lg: 1.125rem   (18px)
--font-size-xl: 1.25rem    (20px)
--font-size-2xl: 1.5rem    (24px)
--font-size-3xl: 1.875rem  (30px)
--font-size-4xl: 2.25rem   (36px)
--font-size-5xl: 3rem      (48px)
--font-size-6xl: 3.75rem   (60px)
```

#### Font Weights
```css
--font-weight-normal: 400
--font-weight-medium: 500
--font-weight-semibold: 600
--font-weight-bold: 700
```

---

### Shadows

```css
--shadow-sm: subtle shadow
--shadow-base: default shadow
--shadow-md: medium shadow
--shadow-lg: large shadow  ← Use for cards
--shadow-xl: extra large shadow
--shadow-2xl: huge shadow
```

**Usage:**
```css
.card {
  box-shadow: var(--shadow-base);
}

.card:hover {
  box-shadow: var(--shadow-lg);
}
```

---

### Border Radius

```css
--radius-sm: 2px
--radius-base: 4px
--radius-md: 6px
--radius-lg: 8px     ← Use for buttons
--radius-xl: 12px    ← Use for cards
--radius-2xl: 16px
--radius-3xl: 24px
--radius-full: 9999px  ← Use for circles/pills
```

---

## 🧱 Component Patterns

### Property Card

**The most important component.** Used on homepage, search, saved properties, AI recommendations.

**HTML Structure:**
```html
<div class="property-card">
  <div class="property-image">
    <img src="..." alt="...">
    <button class="save-heart">❤️</button>
  </div>
  <div class="property-details">
    <div class="property-price">$1,500/month</div>
    <h3 class="property-address">123 Main St</h3>
    <p class="property-location">Houston, TX 77001</p>
    <p class="property-specs">3 bed • 2 bath • 1,200 sq ft</p>
    <a href="..." class="btn btn-primary">Schedule A Tour</a>
  </div>
</div>
```

**Features:**
- ✅ Hover effects (lift + shadow)
- ✅ Image zoom on hover
- ✅ Save heart button with animation
- ✅ Responsive (mobile-optimized)
- ✅ Loading skeleton states
- ✅ AI badge support

**Page-specific variants:**
```css
.page-saved-properties .property-card {
  /* Custom styles for saved properties page */
}
```

---

### Buttons

**Classes:**
```css
.btn              Base button style
.btn-primary      Navy button (main CTA)
.btn-secondary    Gray outline button
.btn-outline      Outline variant
.btn-sm           Small size
.btn-lg           Large size
```

**Usage:**
```html
<button class="btn btn-primary">Schedule A Tour</button>
<button class="btn btn-secondary">Learn More</button>
<button class="btn btn-outline">View Details</button>
```

---

### Forms

**Input fields:**
```html
<div class="form-group">
  <label for="email">Email</label>
  <input type="email" id="email" class="form-input" placeholder="you@example.com">
</div>
```

**Select dropdowns:**
```html
<select class="form-select">
  <option>Choose...</option>
  <option>Option 1</option>
</select>
```

**Validation states:**
```html
<input class="form-input is-valid">  <!-- Green border -->
<input class="form-input is-invalid">  <!-- Red border -->
```

---

## 📝 Naming Conventions

### Component Classes

**Pattern**: `component-element-modifier`

```css
.property-card              Component base
.property-card-image        Element
.property-card--featured    Modifier (-- for variants)
.property-card.is-loading   State (. for states)
```

### Page Scoping

**Pattern**: `.page-{pagename} .component`

```css
.page-homepage .property-card { }
.page-saved-properties .property-card { }
.page-search-results .property-card { }
```

---

## 🚫 Anti-Patterns (Don't Do This!)

### ❌ Hardcoded Values
```css
/* BAD */
.my-component {
  color: #1b3559;
  padding: 16px;
  margin: 20px;
}

/* GOOD */
.my-component {
  color: var(--navy-primary);
  padding: var(--space-4);
  margin: var(--space-5);
}
```

### ❌ Duplicate Components
```css
/* BAD - Creating similar components */
.property-card { }
.saved-property-card { }  /* Don't duplicate! */
.ai-property-card { }     /* Don't duplicate! */

/* GOOD - Use modifiers */
.property-card { }
.property-card--saved { }
.property-card--ai { }
```

### ❌ Inline Styles
```html
<!-- BAD -->
<div style="color: blue; padding: 20px;">

<!-- GOOD -->
<div class="text-info p-5">
```

### ❌ !important
```css
/* BAD */
.my-class {
  color: red !important;  /* Avoid unless absolutely necessary */
}

/* GOOD */
.my-class {
  color: var(--error);  /* Use specificity correctly */
}
```

---

## 📱 Responsive Design

### Breakpoints

```css
Mobile:   < 768px   (--breakpoint-sm)
Tablet:   768-1024px  (--breakpoint-md)
Desktop:  > 1024px  (--breakpoint-lg)
```

### Mobile-First Approach

```css
/* Base styles (mobile) */
.component {
  font-size: var(--font-size-sm);
  padding: var(--space-4);
}

/* Tablet and up */
@media (min-width: 768px) {
  .component {
    font-size: var(--font-size-base);
    padding: var(--space-6);
  }
}

/* Desktop and up */
@media (min-width: 1024px) {
  .component {
    font-size: var(--font-size-lg);
    padding: var(--space-8);
  }
}
```

---

## ✨ Animation Guidelines

### Transitions

**Use consistent timing:**
```css
transition: all 0.2s ease;   /* Fast (hover effects) */
transition: all 0.3s ease;   /* Medium (most animations) */
transition: all 0.5s ease;   /* Slow (large movements) */
```

**Common easing:**
```css
ease-in-out              Smooth both directions
cubic-bezier(0.4, 0, 0.2, 1)   Material Design easing
```

### Keyframe Animations

```css
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.component {
  animation: fadeIn 0.3s ease;
}
```

---

## 🎯 Adding New Components

### Step-by-Step Process

1. **Check if it exists** - Search scout-components.css first
2. **Use tokens** - Never hardcode colors/spacing
3. **Follow patterns** - Match existing component structure
4. **Add comments** - Document complex components
5. **Test responsive** - Check mobile, tablet, desktop
6. **Update this guide** - Document new patterns

### Template

```css
/* ========================================================================
   MY NEW COMPONENT
   ======================================================================== */

/* Base component */
.my-component {
  /* Layout */
  display: flex;
  padding: var(--space-4);
  
  /* Visual */
  background: var(--white);
  border: 1px solid var(--gray-200);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-base);
  
  /* Animation */
  transition: all 0.3s ease;
}

/* Hover state */
.my-component:hover {
  box-shadow: var(--shadow-lg);
  transform: translateY(-2px);
}

/* Child elements */
.my-component-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--navy-primary);
}

/* Modifiers */
.my-component--large {
  padding: var(--space-8);
}

/* States */
.my-component.is-active {
  border-color: var(--navy-primary);
}

/* Responsive */
@media (max-width: 768px) {
  .my-component {
    padding: var(--space-3);
  }
}
```

---

## 🔧 Maintenance

### Regular Audits

**Check for:**
- [ ] Hardcoded colors (search for `#`)
- [ ] Hardcoded spacing (search for `px` outside tokens)
- [ ] Duplicate selectors
- [ ] Unused styles
- [ ] Missing responsive styles

### Tools

```bash
# Find hardcoded colors
grep -E "#[0-9a-fA-F]{6}" scout-components.css

# Find hardcoded px values
grep -E "[0-9]+px" scout-components.css | grep -v "var("

# Count CSS lines
wc -l *.css
```

---

## 📚 Resources

- **Design Tokens**: `/static/css/design-tokens.css`
- **Components**: `/static/css/scout-components.css`
- **Architecture Audit**: `/CSS_ARCHITECTURE_AUDIT.md`
- **Figma** (if exists): Link to design files
- **Color Contrast Checker**: https://webaim.org/resources/contrastchecker/

---

## 🆘 Common Issues

### Issue: My styles aren't applying
**Solution**: Check CSS specificity and file loading order

### Issue: Colors look different than design
**Solution**: Use design tokens, don't hardcode colors

### Issue: Component looks broken on mobile
**Solution**: Add responsive media queries

### Issue: Animation is janky
**Solution**: Use `transform` and `opacity` (GPU-accelerated)

---

**Questions?** Check the codebase or ask the team.

**Last refactored**: December 4, 2025 - Property card consolidation & token migration
