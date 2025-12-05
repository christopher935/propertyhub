# PropertyHub Design System Specification v2.0

## 1. Design Philosophy

### Brand Essence
**Premium. Professional. Trustworthy.**

PropertyHub serves two audiences with one cohesive brand:
- **Consumers**: Prospective tenants seeking their next home
- **Administrators**: Property managers running the business

### Design Principles

1. **Clarity Over Cleverness** — Every element serves a purpose
2. **Premium Restraint** — Luxury is shown through quality, not quantity
3. **Confident Motion** — Quick, purposeful transitions; never playful
4. **Inverse Harmony** — Consumer and admin are mirrors of each other
5. **Color is Earned** — Color for meaning, not decoration

---

## 2. Color Philosophy

### Core Rule
**Backgrounds are ALWAYS white or subtle grays. Color is reserved for interactive elements, headings, data visualization, and status indicators.**

### Use Color For

| Purpose | Examples |
|---------|----------|
| **Interactive elements** | Buttons, links, form focus states |
| **Headings** | H1-H3 in navy for hierarchy |
| **Data visualization** | Charts, stats, progress bars, metrics |
| **Status indicators** | Badges, alerts, notifications |
| **Brand moments** | Logo, key CTAs |

### Never Use Color For

| Avoid | Use Instead |
|-------|-------------|
| Page backgrounds | White (`#FFFFFF`), gray-50 (`#F9FAFB`) |
| Hero sections | White/gray with colored text & buttons |
| Sidebar backgrounds | White/gray with colored accents |
| Large color blocks | Subtle borders, shadows for depth |
| Decorative purposes | Whitespace and typography |

---

## 3. Color Palette

### Brand Colors

| Token | Hex | RGB | Usage |
|-------|-----|-----|-------|
| `--navy-primary` | `#1b3559` | rgb(27, 53, 89) | THE brand navy - headings, admin buttons |
| `--navy-light` | `#2d4a73` | rgb(45, 74, 115) | Hover states |
| `--navy-dark` | `#0f1f35` | rgb(15, 31, 53) | Pressed states, high contrast text |
| `--gold-primary` | `#c4a962` | rgb(196, 169, 98) | THE brand gold - consumer CTAs, accents |
| `--gold-light` | `#d4bc7a` | rgb(212, 188, 122) | Hover states on gold |
| `--gold-dark` | `#b8941f` | rgb(184, 148, 31) | Pressed states on gold |

### Semantic Colors

| Token | Hex | Light Variant | Usage |
|-------|-----|---------------|-------|
| `--success` | `#10B981` | `#D1FAE5` | Success states, positive metrics |
| `--warning` | `#F59E0B` | `#FEF3C7` | Warnings, attention needed |
| `--error` | `#EF4444` | `#FEE2E2` | Errors, destructive actions, negative metrics |
| `--info` | `#3B82F6` | `#DBEAFE` | Informational states |

### Neutral Colors (Backgrounds & Text)

| Token | Hex | Usage |
|-------|-----|-------|
| `--white` | `#FFFFFF` | Primary backgrounds, cards |
| `--gray-50` | `#F9FAFB` | Page backgrounds, alternating sections |
| `--gray-100` | `#F3F4F6` | Secondary backgrounds, footer |
| `--gray-200` | `#E5E7EB` | Borders, dividers |
| `--gray-300` | `#D1D5DB` | Disabled states, subtle borders |
| `--gray-400` | `#9CA3AF` | Placeholder text |
| `--gray-500` | `#6B7280` | Secondary text, captions |
| `--gray-600` | `#4B5563` | Body text |
| `--gray-700` | `#374151` | Primary body text |
| `--gray-800` | `#1F2937` | Dark text emphasis |
| `--gray-900` | `#111827` | Maximum contrast text |

---

## 4. The Inverse Relationship

Consumer and admin experiences use the **same brand colors** but apply them inversely for primary actions.

### Consumer Experience (Public-Facing)
*Gold buttons, navy text*

| Element | Style |
|---------|-------|
| **Page Background** | White / gray-50 |
| **Primary Button** | Gold background, navy text |
| **Secondary Button** | Transparent, navy border, navy text |
| **Headings** | Navy |
| **Body Text** | Gray-600 / gray-700 |
| **Links** | Navy, gold on hover |
| **Active/Focus States** | Gold accent |
| **Hero Section** | White/gray bg, navy heading, gold CTA |

### Admin Experience (Internal Tools)
*Navy buttons, gold accents*

| Element | Style |
|---------|-------|
| **Page Background** | Gray-50 |
| **Primary Button** | Navy background, white text |
| **Secondary Button** | Transparent, navy border, navy text |
| **Headings** | Navy |
| **Body Text** | Gray-600 / gray-700 |
| **Links** | Navy, gold on hover |
| **Sidebar** | White background, gray border |
| **Active Nav Item** | Gold left border, navy text |
| **Stat Numbers** | Navy (large), gold for highlights |
| **Charts** | Navy primary, gold secondary |

### Visual Summary
```
CONSUMER                              ADMIN
┌─────────────────────┐              ┌─────────────────────┐
│ White Background    │              │ Gray-50 Background  │
│                     │              │                     │
│ [███ GOLD ███]      │              │ [███ NAVY ███]      │
│   Navy Text         │              │   White Text        │
│                     │              │                     │
│ [ ─── NAVY ─── ]    │              │ [ ─── NAVY ─── ]    │
│   Navy Text         │              │   Navy Text         │
│                     │              │                     │
│ Accent: Gold        │              │ Accent: Gold        │
└─────────────────────┘              └─────────────────────┘
```

---

## 5. Typography

### Font Families

| Token | Value | Usage |
|-------|-------|-------|
| `--font-display` | `'Playfair Display', serif` | Headings H1-H3, hero text |
| `--font-body` | `'Inter', sans-serif` | Body text, UI elements, H4-H6 |

### Type Scale

| Token | Size | Line Height | Usage |
|-------|------|-------------|-------|
| `--text-xs` | 0.75rem (12px) | 1.5 | Captions, labels, badges |
| `--text-sm` | 0.875rem (14px) | 1.5 | Small text, metadata, helper text |
| `--text-base` | 1rem (16px) | 1.6 | Body text, buttons |
| `--text-lg` | 1.125rem (18px) | 1.6 | Large body, intro paragraphs |
| `--text-xl` | 1.25rem (20px) | 1.5 | H5, card titles |
| `--text-2xl` | 1.5rem (24px) | 1.4 | H4, section titles |
| `--text-3xl` | 1.875rem (30px) | 1.3 | H3 |
| `--text-4xl` | 2.25rem (36px) | 1.2 | H2, stat numbers |
| `--text-5xl` | 3rem (48px) | 1.1 | H1, page titles |
| `--text-6xl` | 3.75rem (60px) | 1.1 | Hero headlines |

### Font Weights

| Token | Value | Usage |
|-------|-------|-------|
| `--font-normal` | 400 | Body text |
| `--font-medium` | 500 | Emphasis, labels |
| `--font-semibold` | 600 | Subheadings, buttons |
| `--font-bold` | 700 | Headings, strong emphasis |

### Heading Hierarchy

| Level | Font | Size | Weight | Color |
|-------|------|------|--------|-------|
| H1 | Playfair Display | text-5xl | Bold | `--navy-primary` |
| H2 | Playfair Display | text-4xl | Bold | `--navy-primary` |
| H3 | Playfair Display | text-3xl | Semibold | `--navy-primary` |
| H4 | Inter | text-2xl | Semibold | `--navy-primary` |
| H5 | Inter | text-xl | Semibold | `--navy-primary` |
| H6 | Inter | text-lg | Semibold | `--navy-primary` |

---

## 6. Spacing System

### Base Unit: 4px (0.25rem)

| Token | Value | Pixels | Common Usage |
|-------|-------|--------|--------------|
| `--space-0` | 0 | 0px | Reset |
| `--space-1` | 0.25rem | 4px | Tight gaps |
| `--space-2` | 0.5rem | 8px | Icon gaps, inline spacing |
| `--space-3` | 0.75rem | 12px | Button padding (vertical) |
| `--space-4` | 1rem | 16px | Standard gap, input padding |
| `--space-5` | 1.25rem | 20px | Card padding (small) |
| `--space-6` | 1.5rem | 24px | Card padding, component spacing |
| `--space-8` | 2rem | 32px | Section spacing |
| `--space-10` | 2.5rem | 40px | Large component gaps |
| `--space-12` | 3rem | 48px | Section margins |
| `--space-16` | 4rem | 64px | Page section padding |
| `--space-20` | 5rem | 80px | Hero padding |
| `--space-24` | 6rem | 96px | Large hero padding |

---

## 7. Border Radius

| Token | Value | Usage |
|-------|-------|-------|
| `--radius-none` | 0 | Sharp corners |
| `--radius-sm` | 2px | Subtle rounding |
| `--radius-base` | 4px | Inputs, small elements |
| `--radius-md` | 6px | Buttons, badges |
| `--radius-lg` | 8px | Cards, dropdowns |
| `--radius-xl` | 12px | Large cards, modals |
| `--radius-2xl` | 16px | Feature cards |
| `--radius-full` | 9999px | Pills, avatars, circular elements |

---

## 8. Shadows

| Token | Value | Usage |
|-------|-------|-------|
| `--shadow-sm` | `0 1px 2px rgba(0,0,0,0.05)` | Subtle depth, resting state |
| `--shadow-base` | `0 1px 3px rgba(0,0,0,0.1), 0 1px 2px rgba(0,0,0,0.06)` | Cards at rest |
| `--shadow-md` | `0 4px 6px rgba(0,0,0,0.1), 0 2px 4px rgba(0,0,0,0.06)` | Hover state |
| `--shadow-lg` | `0 10px 15px rgba(0,0,0,0.1), 0 4px 6px rgba(0,0,0,0.05)` | Elevated elements |
| `--shadow-xl` | `0 20px 25px rgba(0,0,0,0.1), 0 10px 10px rgba(0,0,0,0.04)` | Modals, dropdowns |

---

## 9. Button System

### Button Base (All Buttons)
```css
.btn {
  font-family: var(--font-body);
  font-size: var(--text-base);
  font-weight: var(--font-semibold);
  padding: var(--space-3) var(--space-6);
  border-radius: var(--radius-md);
  border: 2px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn:focus-visible {
  outline: 2px solid var(--gold-primary);
  outline-offset: 2px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
```

### Consumer Buttons

**Primary (Gold)**
```css
.consumer .btn-primary,
.btn-primary-consumer {
  background: var(--gold-primary);
  color: var(--navy-dark);
  border-color: var(--gold-primary);
}

.consumer .btn-primary:hover,
.btn-primary-consumer:hover {
  background: var(--gold-dark);
  border-color: var(--gold-dark);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}
```

**Secondary (Outline Navy)**
```css
.consumer .btn-secondary,
.btn-secondary-consumer {
  background: transparent;
  color: var(--navy-primary);
  border-color: var(--navy-primary);
}

.consumer .btn-secondary:hover,
.btn-secondary-consumer:hover {
  background: var(--navy-primary);
  color: var(--white);
  transform: translateY(-2px);
}
```

### Admin Buttons

**Primary (Navy)**
```css
.admin .btn-primary,
.btn-primary-admin {
  background: var(--navy-primary);
  color: var(--white);
  border-color: var(--navy-primary);
}

.admin .btn-primary:hover,
.btn-primary-admin:hover {
  background: var(--navy-light);
  border-color: var(--navy-light);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}
```

**Secondary (Outline Navy)**
```css
.admin .btn-secondary,
.btn-secondary-admin {
  background: transparent;
  color: var(--navy-primary);
  border-color: var(--navy-primary);
}

.admin .btn-secondary:hover,
.btn-secondary-admin:hover {
  background: var(--navy-primary);
  color: var(--white);
  transform: translateY(-2px);
}
```

### Semantic Buttons (Both Contexts)

```css
.btn-success { background: var(--success); color: white; border-color: var(--success); }
.btn-warning { background: var(--warning); color: white; border-color: var(--warning); }
.btn-danger { background: var(--error); color: white; border-color: var(--error); }
```

### Button Sizes

| Size | Padding | Font Size |
|------|---------|-----------|
| `.btn-sm` | `var(--space-2) var(--space-4)` | `var(--text-sm)` |
| `.btn` (default) | `var(--space-3) var(--space-6)` | `var(--text-base)` |
| `.btn-lg` | `var(--space-4) var(--space-8)` | `var(--text-lg)` |

---

## 10. Animation & Motion

### Philosophy
**Premium motion is confident, quick, and purposeful. No decorative animations.**

### Timing Tokens

| Token | Duration | Usage |
|-------|----------|-------|
| `--duration-fast` | 150ms | Micro-interactions, color changes |
| `--duration-base` | 200ms | Standard hover transitions |
| `--duration-slow` | 300ms | Complex state changes, modals |

### Easing

| Token | Value | Usage |
|-------|-------|-------|
| `--ease-default` | `ease` | All standard transitions |

### Allowed Transforms

| Transform | Value | Usage |
|-----------|-------|-------|
| Hover lift | `translateY(-2px)` | Buttons, cards |
| Active press | `translateY(0)` | Return to rest |
| Slide | `translateX()` | Modals, drawers, mobile menus |
| Fade | `opacity: 0 → 1` | Page transitions, reveals |

### FORBIDDEN (Never Use)

- ❌ `rotate()` — No rotation on any UI element
- ❌ `scale()` on buttons/icons — Feels cheap, not premium
- ❌ `skew()` — Distortion is not professional
- ❌ Bounce/spring easing — Too playful
- ❌ Decorative keyframe animations — No pulsing, glowing, floating

### Standard Interaction Pattern

```css
.interactive {
  transition: transform var(--duration-base) var(--ease-default),
              box-shadow var(--duration-base) var(--ease-default),
              background var(--duration-fast) var(--ease-default);
}

.interactive:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.interactive:active {
  transform: translateY(0);
  box-shadow: var(--shadow-sm);
}
```

### Reduced Motion Support (Required)

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

## 11. Component Patterns

### Cards

```css
.card {
  background: var(--white);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-base);
  padding: var(--space-6);
  transition: transform var(--duration-base) var(--ease-default),
              box-shadow var(--duration-base) var(--ease-default);
}

.card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.card-title {
  font-size: var(--text-xl);
  font-weight: var(--font-semibold);
  color: var(--navy-primary);
  margin-bottom: var(--space-2);
}
```

### Form Inputs

```css
.form-input {
  width: 100%;
  padding: var(--space-3) var(--space-4);
  font-size: var(--text-base);
  font-family: var(--font-body);
  color: var(--gray-700);
  background: var(--white);
  border: 1px solid var(--gray-300);
  border-radius: var(--radius-base);
  transition: border-color var(--duration-fast) var(--ease-default),
              box-shadow var(--duration-fast) var(--ease-default);
}

.form-input::placeholder {
  color: var(--gray-400);
}

.form-input:focus {
  outline: none;
  border-color: var(--navy-primary);
  box-shadow: 0 0 0 3px rgba(27, 53, 89, 0.1);
}

.form-input:invalid:not(:placeholder-shown) {
  border-color: var(--error);
}
```

### Badges

```css
.badge {
  display: inline-flex;
  align-items: center;
  padding: var(--space-1) var(--space-2);
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  border-radius: var(--radius-full);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.badge-success { background: var(--success-light); color: var(--success); }
.badge-warning { background: var(--warning-light); color: var(--warning); }
.badge-danger { background: var(--error-light); color: var(--error); }
.badge-info { background: var(--info-light); color: var(--info); }
```

### Navigation

**Consumer Header Nav:**
```css
.consumer .nav-link {
  color: var(--navy-primary);
  font-weight: var(--font-medium);
  padding: var(--space-2) var(--space-3);
  text-decoration: none;
  transition: color var(--duration-fast) var(--ease-default);
}

.consumer .nav-link:hover {
  color: var(--gold-primary);
}

.consumer .nav-link.active {
  color: var(--gold-primary);
  font-weight: var(--font-semibold);
}
```

**Admin Sidebar Nav:**
```css
.admin-sidebar {
  background: var(--white);
  border-right: 1px solid var(--gray-200);
  width: 260px;
}

.admin .nav-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  color: var(--gray-600);
  padding: var(--space-3) var(--space-4);
  border-left: 3px solid transparent;
  text-decoration: none;
  transition: all var(--duration-fast) var(--ease-default);
}

.admin .nav-item:hover {
  color: var(--navy-primary);
  background: var(--gray-50);
}

.admin .nav-item.active {
  color: var(--navy-primary);
  font-weight: var(--font-semibold);
  background: var(--gray-50);
  border-left-color: var(--gold-primary);
}
```

---

## 12. Layout Structure

### Consumer Layout
```
┌──────────────────────────────────────────────────────┐
│  HEADER                                              │
│  Background: white                                   │
│  Logo: Navy | Nav: Navy links, gold hover            │
├──────────────────────────────────────────────────────┤
│  HERO                                                │
│  Background: white or gray-50                        │
│  Heading: Navy (Playfair Display)                    │
│  Subtext: Gray-600                                   │
│  CTA: Gold button                                    │
├──────────────────────────────────────────────────────┤
│  CONTENT                                             │
│  Background: white (gray-50 for alternating)         │
│  Cards: white + shadow                               │
│  Headings: Navy                                      │
├──────────────────────────────────────────────────────┤
│  FOOTER                                              │
│  Background: gray-100                                │
│  Text: Navy, links gold on hover                     │
└──────────────────────────────────────────────────────┘
```

### Admin Layout
```
┌────────────────┬─────────────────────────────────────┐
│  SIDEBAR       │  HEADER                             │
│  Bg: white     │  Bg: white                          │
│  Border: gray  │  Search, notifications, user menu   │
│                ├─────────────────────────────────────┤
│  Nav: gray-600 │  CONTENT                            │
│  Active: navy  │  Background: gray-50                │
│  + gold border │                                     │
│                │  Cards: white + shadow              │
│                │  Stats: Navy numbers, gold accents  │
│                │  Tables: white, gray borders        │
│                │                                     │
└────────────────┴─────────────────────────────────────┘
```

---

## 13. Responsive Breakpoints

| Token | Value | Target |
|-------|-------|--------|
| `--breakpoint-sm` | 640px | Large phones |
| `--breakpoint-md` | 768px | Tablets |
| `--breakpoint-lg` | 1024px | Small desktops |
| `--breakpoint-xl` | 1280px | Large desktops |

### Mobile-First Approach
```css
/* Base: Mobile */
.element { ... }

/* Tablet+ */
@media (min-width: 768px) { ... }

/* Desktop+ */
@media (min-width: 1024px) { ... }
```

---

## 14. Z-Index Scale

| Token | Value | Usage |
|-------|-------|-------|
| `--z-base` | 0 | Default |
| `--z-dropdown` | 10 | Dropdowns, popovers |
| `--z-sticky` | 20 | Sticky headers |
| `--z-fixed` | 30 | Fixed elements |
| `--z-modal-backdrop` | 40 | Modal overlay |
| `--z-modal` | 50 | Modal content |
| `--z-tooltip` | 60 | Tooltips |
| `--z-toast` | 70 | Toast notifications |

---

## 15. Accessibility Requirements

### Color Contrast (WCAG AA)
- Normal text: 4.5:1 minimum
- Large text (18px+): 3:1 minimum
- UI components: 3:1 minimum

**Verified Combinations:**
- Navy on white: ✅ 10.5:1
- White on navy: ✅ 10.5:1
- Gold on navy: ✅ 4.6:1
- Navy on gold: ✅ 4.6:1
- Gray-600 on white: ✅ 5.7:1

**Caution:**
- Gold on white: ⚠️ 2.8:1 (use only for large text or with underline)

### Focus States
All interactive elements must have visible focus:
```css
:focus-visible {
  outline: 2px solid var(--gold-primary);
  outline-offset: 2px;
}
```

### Touch Targets
- Minimum size: 44px × 44px
- Minimum gap between targets: 8px

---

## 16. File Structure (Target)

```
web/static/css/
├── design-tokens.css       # All variables (single source of truth)
├── reset.css               # CSS reset/normalize
├── utilities.css           # Utility classes
├── components/
│   ├── buttons.css
│   ├── cards.css
│   ├── forms.css
│   ├── badges.css
│   ├── navigation.css
│   └── modals.css
├── layouts/
│   ├── consumer.css        # Consumer header, footer, hero
│   └── admin.css           # Admin sidebar, dashboard grid
└── pages/                  # Minimal page-specific overrides
    ├── homepage.css
    └── dashboard.css
```

---

## Summary: Quick Reference

| Element | Consumer | Admin |
|---------|----------|-------|
| Page background | White / gray-50 | Gray-50 |
| Primary button | **Gold** bg, navy text | **Navy** bg, white text |
| Secondary button | Navy outline | Navy outline |
| Headings | Navy | Navy |
| Body text | Gray-600/700 | Gray-600/700 |
| Links | Navy → gold hover | Navy → gold hover |
| Active accent | Gold | Gold |
| Cards | White + shadow | White + shadow |
| Sidebar | — | White + gold active border |

---

*Last Updated: December 2024*
*Version: 2.0*
