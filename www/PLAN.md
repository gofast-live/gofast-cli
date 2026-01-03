# GoFast Marketing Page - Design Plan

## Overview

A minimalist, dark-themed marketing page that showcases GoFast CLI through an **interactive fake CLI experience**. Users click through commands, watching generation animations and discovering what each command creates.

**Stack:** SvelteKit (static) + GSAP + Tailwind CSS

---

## Design Direction

### Color Palette

```
Background:     #0a0a0a (near black)
Surface:        #141414 (card/terminal bg)
Border:         #262626 (subtle borders)
Primary:        #10b981 (emerald-500)
Primary Glow:   #059669 (emerald-600)
Text:           #fafafa (white-ish)
Text Muted:     #737373 (gray-500)
Success:        #22c55e (green-500)
```

### Typography

- **CLI/Code:** `JetBrains Mono` or `Fira Code` (monospace)
- **UI Text:** `Inter` or system sans-serif
- **Sizing:** Large, readable, lots of whitespace

### Visual Style

- Ultra-minimal, developer-focused
- Terminal aesthetic without being cheesy
- Generous spacing, centered content
- Subtle glow effects on primary actions
- No gradients, no noise textures - pure clean

---

## Page Structure

**Key concept:** Each step is a **full viewport height (100vh) section**. Users scroll/transition between sections like a presentation. No sidebar - status shown at the end.

---

### Section 1: Hero / Landing (100vh)

The user lands here. Show the stack upfront, make it clear what this is.

```
┌─────────────────────────────────────────────────────────────┐
│                                    [Discord] [V1 →]         │
│                                                             │
│                          [LOGO]                             │
│                                                             │
│                  "Building blocks for Go"                   │
│                                                             │
│   ┌─────────────────────────────────────────────────────┐   │
│   │                                                     │   │
│   │     Go  +  ConnectRPC  +  SvelteKit                 │   │
│   │                                                     │   │
│   │     Production-ready. Type-safe. Deployable.        │   │
│   │                                                     │   │
│   └─────────────────────────────────────────────────────┘   │
│                                                             │
│   ┌─────────────────────────────────────────────────────┐   │
│   │  $ gof init myproject                          [▶]  │   │
│   └─────────────────────────────────────────────────────┘   │
│                                                             │
│                    Click to start building                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Content:**
- Logo + tagline at top
- Stack highlight: **Go + ConnectRPC + SvelteKit**
- Brief value props (production-ready, type-safe, deployable)
- The `gof init` command with play button
- "Click to start building" prompt

---

### Section 2+: Command Flow (100vh each)

After clicking init, the line starts drawing and items appear. Each command gets its own full-height section with the animated line flow.

**Visual: Init command flow**
```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│         ┌───────────────────────────────────┐               │
│         │  $ gof init myproject             │               │
│         └───────────────────────────────────┘               │
│                         │                                   │
│                         │                                   │
│                         ├──── Go HTTP server + ConnectRPC   │
│                         │                                   │
│                         ├──── PostgreSQL + SQLC queries     │
│                         │                                   │
│                         ├──── OAuth2 (GitHub + Google)      │
│                         │                                   │
│                         ├──── Role-based authorization      │
│                         │                                   │
│                         ├──── Docker Compose                │
│                         │                                   │
│                         ├──── GitHub Actions CI/CD          │
│                         │                                   │
│                         ├──── PR preview deployments        │
│                         │                                   │
│                         ▼                                   │
│         ┌───────────────────────────────────┐               │
│         │  What's next?                     │               │
│         │  [model] [client] [stripe]        │               │
│         │  [r2] [postmark] [infra]          │               │
│         │                    [→ finish]     │               │
│         └───────────────────────────────────┘               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Visual: Model command flow**
```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│         ┌───────────────────────────────────┐               │
│         │  $ gof model task                 │               │
│         │    title:string due:date done:bool│               │
│         └───────────────────────────────────┘               │
│                         │                                   │
│                         │                                   │
│              Database   ├──── SQL migration                 │
│                         ├──── SQLC queries (CRUD + List)    │
│                         ├──── Type-safe Go structs          │
│                         │                                   │
│              Service    ├──── Domain service layer          │
│                         ├──── Input validation              │
│                         ├──── ConnectRPC handlers           │
│                         │                                   │
│              Testing    ├──── Service tests                 │
│                         ├──── Transport tests               │
│                         │                                   │
│              + Context  ├──── Svelte pages (if client)      │
│                         │                                   │
│                         ▼                                   │
│         ┌───────────────────────────────────┐               │
│         │  [model] [client] [stripe] ...    │               │
│         └───────────────────────────────────┘               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Key elements:**
- Command box at top (clean, minimal)
- Animated line draws downward
- Items appear along the line as it draws
- Category labels on the left (Database, Service, Testing)
- Context-aware items at the bottom (grayed or highlighted)
- Next command picker at bottom

---

### Final Section: Summary + CTA (100vh)

When user clicks "finish" or exhausts options:

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│                     Your stack is ready                     │
│                                                             │
│   ┌─ What you built ─────────────────────────────────────┐  │
│   │                                                      │  │
│   │  ✓ init     ✓ model: task, post    ✓ client         │  │
│   │  ✓ stripe   ○ r2    ○ postmark     ✓ infra          │  │
│   │                                                      │  │
│   │  Stack: Go + ConnectRPC + Svelte + Stripe + K8s     │  │
│   │                                                      │  │
│   └──────────────────────────────────────────────────────┘  │
│                                                             │
│   ┌─ Pricing ────────────────────────────────────────────┐  │
│   │                                                      │  │
│   │  $40 one-time · Lifetime access                      │  │
│   │                                                      │  │
│   │  ✓ GoFast V2 (this CLI)                              │  │
│   │  ✓ GoFast V1 (Next.js, Vue, AWS, GCP...)            │  │
│   │  ✓ All future updates                                │  │
│   │                                                      │  │
│   │              [ Get Access → ]                        │  │
│   │                                                      │  │
│   └──────────────────────────────────────────────────────┘  │
│                                                             │
│   ──────────────────────────────────────────────────────    │
│                                                             │
│   [Discord] Join 100+ devs · Free, no purchase needed      │
│                                                             │
│   Want more stacks? V1 has Next.js, Vue, AWS S3, GCP...    │
│   → gofast.live                                             │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Content:**
- Summary box showing what was built (checked) vs available (unchecked)
- Dynamic stack summary
- Pricing card with CTA
- Discord link (free)
- V1 link for more options

---

## Scroll/Transition Behavior

- Each section is `100vh` with `scroll-snap-align: start`
- Use CSS `scroll-snap-type: y mandatory` on container
- When command completes → auto-scroll to next section
- Smooth scroll animation (~600ms)
- Optional: GSAP ScrollToPlugin for more control

---

## Key Changes from Previous Plan

1. **Full viewport sections** - not stacking terminals
2. **More descriptive output** - grouped by category (Backend, DevOps, etc.)
3. **No sidebar** - status shown only at the end
4. **Stack shown upfront** - Go + ConnectRPC + SvelteKit on landing
5. **Scroll-snap pagination** - each step fills the screen

---

#### Context-Aware Output

Output adapts based on what's already been added.

**Model Sub-Picker:**
When user clicks "model", a sub-picker appears with 3 options:

```
┌─────────────────────────────────────────┐
│  Pick a model:                          │
│  ┌──────┐ ┌──────┐ ┌───────┐           │
│  │ task │ │ post │ │ event │           │
│  └──────┘ └──────┘ └───────┘           │
└─────────────────────────────────────────┘
```

| Model | Columns | Shows off |
|-------|---------|-----------|
| task | title:string, due:date, done:bool | Mixed types |
| post | title:string, body:string, views:number | Strings + number |
| event | name:string, start:date, end:date | Multiple dates |

User can add multiple models (click "model" again → shows remaining options).

**Example: model → client**
```
$ gof model task title:string due:date done:bool
✓ SQL migration
✓ SQLC queries
✓ Proto definitions
✓ Domain service
✓ Transport handlers
✓ Tests

$ gof client svelte
✓ SvelteKit scaffold
✓ Auth integration
✓ Generated pages for: task    ← SMART: knows model exists
✓ Type-safe API client
```

**Example: multiple models + client**
```
$ gof model task ...
$ gof model post ...
$ gof client svelte
✓ SvelteKit scaffold
✓ Auth integration
✓ Generated pages for: task, post  ← SMART: lists all models
✓ Type-safe API client
```

**Example: stripe → later commands**
```
$ gof add stripe
✓ Payment domain service
✓ Subscriptions migration
✓ Webhook handlers
✓ Access control integration

$ gof model task ...
...
✓ Subscription checks wired    ← SMART: stripe integration
```

### 3. State Tracker

Minimal status showing current project state:

```
┌─────────────────────────────────────────┐
│  myproject                              │
│  ✓ init                                 │
│  ✓ models: task, post                   │
│  ✓ client  ✓ stripe                     │
│  ○ r2  ○ postmark  ○ infra              │
└─────────────────────────────────────────┘
```

- Shows what's been "built"
- Models listed by name
- Remaining options shown as available

### 4. Final CTA + Pricing

Appears after exploring (or clicking "finish"):

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  Your stack:                                                │
│  Go + ConnectRPC + Svelte + Stripe                          │
│  Ready for production.                                      │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  $40 one-time payment                               │    │
│  │  Lifetime access to everything                      │    │
│  │                                                     │    │
│  │  ✓ GoFast V2 (this CLI)                             │    │
│  │  ✓ GoFast V1 (more stacks: Next.js, Vue, AWS, GCP)  │    │
│  │  ✓ All future updates                               │    │
│  │                                                     │    │
│  │            [ Get Access → ]                         │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  ───────────────────────────────────────────────────────    │
│                                                             │
│  Join the community                                         │
│  [ Discord ] ← free, no purchase required                   │
│  Modern web dev, Go, articles, discussions                  │
│                                                             │
│  ───────────────────────────────────────────────────────    │
│                                                             │
│  Want more options? V1 has Next.js, Vue, AWS S3, GCP...     │
│  [ Visit gofast.live → ]                                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Key info to display:**
- **Price:** $40 one-time, lifetime access
- **Includes both:** V2 (this focused CLI) + V1 (kitchen sink)
- **Discord:** Free access, community link
- **V1 link:** gofast.live for those who want more stack options

---

## Animation System (GSAP) - Line Flow Concept

**Core idea:** No fake terminals. Commands in clean boxes, connected by animated lines. Generated items appear ALONG the line as it draws.

### Visual Concept

```
   ┌─────────────────────────────┐
   │  $ gof init myproject   [▶] │
   └─────────────────────────────┘
                │
                │──── ✓ Go + ConnectRPC server
                │
                │──── ✓ OAuth2 (GitHub + Google)
                │
                │──── ✓ PostgreSQL + SQLC
                │
                │──── ✓ Docker Compose
                │
                │──── ✓ GitHub Actions CI/CD
                │
                ▼
   ┌─────────────────────────────┐
   │  What's next?               │
   │  [model] [client] [stripe]  │
   └─────────────────────────────┘
```

### Animation Flow

1. **User clicks command** → line starts drawing downward
2. **As line draws**, items fade in along the line (left or right side, alternating)
3. **Each item** has a small branch from the main line
4. **Line completes** → next command picker fades in
5. **User picks next** → line continues, scroll to next section

### Line Drawing (SVG + GSAP)

```javascript
// Draw the main vertical line
gsap.fromTo(line,
  { strokeDashoffset: lineLength },
  {
    strokeDashoffset: 0,
    duration: 2,
    ease: "none"
  }
);

// Items appear as line passes them
items.forEach((item, i) => {
  gsap.fromTo(item,
    { opacity: 0, x: -20 },
    {
      opacity: 1,
      x: 0,
      duration: 0.4,
      delay: i * 0.3  // staggered with line progress
    }
  );
});
```

### Item Appearance Styles

**Option A: Branches from line**
```
        │
        ├──── ✓ OAuth2 authentication
        │
        ├──── ✓ PostgreSQL + SQLC
        │
```

**Option B: Dots on line with labels**
```
        │
        ●──── OAuth2 authentication
        │
        ●──── PostgreSQL + SQLC
        │
```

**Option C: Alternating sides**
```
                        │
   OAuth2 auth ─────────┤
                        │
                        ├───────── PostgreSQL + SQLC
                        │
   Docker Compose ──────┤
                        │
```

### Colors

- Line: `var(--primary)` emerald with subtle glow
- Dots/branches: Same emerald
- Items text: `var(--text)` white
- Checkmarks: `var(--success)` green

### Glow Effect

```css
.flow-line {
  stroke: var(--primary);
  filter: drop-shadow(0 0 4px var(--primary-glow));
}
```

### Scroll Behavior

- Each section is 100vh
- Line animation triggers on scroll-snap to section
- Or: Line animation triggers on click, then auto-scrolls when complete

---

## Interaction Flow

```
1. Page loads → Hero with `gof init` command
                        │
2. User clicks [▶] ─────┤
                        ▼
3. Typing + generation ─┤
   - Command types      │
   - Checkmarks appear  │
   - "Done" shows       │
                        ▼
4. Choice buttons appear (model, client, stripe, r2, postmark, infra)
                        │
5. User picks one ──────┤
                        ▼
6. Flow line animates ──┤
   - Connects terminals │
   - Auto-scroll down   │
                        ▼
7. New terminal runs ───┤
   - Context-aware output│
   - Updates state tracker│
                        ▼
8. New choices appear ──┤
   - Already-used options grayed │
   - Repeat from step 5          │
                        ▼
9. User clicks "Finish" or exhausts options
                        ▼
10. Final CTA with their custom stack summary
```

---

## Sidebar vs Hints (A/B Testing)

### Option A: Subtle Hints (Default)

- Checkmarks appear inline in terminal
- No persistent UI
- Clean, focused on the terminal
- User scrolls up to review

### Option B: Progress Sidebar

```
┌────────────────────────────────────────────────┐
│  ┌──────┐                                      │
│  │ init │ ✓ OAuth                              │
│  │      │ ✓ Auth                               │
│  │      │ ✓ CICD                               │
│  ├──────┤                                      │
│  │model │ ✓ Migration     [TERMINAL]           │
│  │      │ ✓ Queries                            │
│  │      │ ● Proto...                           │
│  ├──────┤                                      │
│  │client│ (pending)                            │
│  └──────┘                                      │
└────────────────────────────────────────────────┘
```

- Left sidebar showing all chapters
- Expands to show generated items
- Acts as navigation (click to jump)
- Shows overall progress

**Recommendation:** Start with Option A (hints), add sidebar toggle later if needed.

---

## Component Structure

```
src/
├── routes/
│   └── +page.svelte          # Main page
├── lib/
│   ├── components/
│   │   ├── Hero.svelte       # Logo + tagline + first command
│   │   ├── Terminal.svelte   # Reusable terminal component
│   │   ├── CommandPicker.svelte  # Choice buttons grid
│   │   ├── FlowLine.svelte   # GSAP animated connecting lines
│   │   ├── StateTracker.svelte   # Current project state display
│   │   └── CTA.svelte        # Final call-to-action
│   ├── data/
│   │   └── commands.js       # Command definitions + context logic
│   ├── animations/
│   │   └── gsap.js           # GSAP animation utilities
│   └── stores/
│       └── state.svelte.js   # Svelte 5 runes-based state
├── static/
│   └── logo.svg              # GoFast logo (user provides)
└── app.css                   # Global styles
```

**Note:** Using JS with JSDoc comments, not TypeScript. Svelte 5 runes for state.

---

## Command Data Structure

```javascript
/**
 * @typedef {Object} Output
 * @property {string} text - What was generated
 * @property {boolean} [contextual] - Only show if condition met
 * @property {(state: State) => boolean} [showIf] - Condition function
 */

/**
 * @typedef {Object} Command
 * @property {string} id - Unique identifier
 * @property {string} label - Button label
 * @property {(state: State) => string} command - Dynamic command string
 * @property {string} description - "Generating..."
 * @property {Output[]} baseOutputs - Always shown
 * @property {Output[]} contextOutputs - Shown based on state
 */

/** @type {Command[]} */
export const commands = [
  {
    id: 'init',
    label: 'init',
    command: () => 'gof init myproject',
    description: 'Creating project structure...',
    baseOutputs: [
      { text: 'OAuth (GitHub + Google)' },
      { text: 'Bitwise role authorization' },
      { text: 'Docker Compose setup' },
      { text: 'GitHub Actions CI/CD' },
      { text: 'PR preview deployments' },
      { text: 'PostgreSQL + SQLC' },
      { text: 'ConnectRPC transport' },
    ],
    contextOutputs: []
  },
  {
    id: 'model',
    label: 'model',
    hasSubPicker: true,
    variants: [
      {
        id: 'model:task',
        name: 'task',
        command: 'gof model task title:string due:date done:bool',
        tagline: 'mixed types',
      },
      {
        id: 'model:post',
        name: 'post',
        command: 'gof model post title:string body:string views:number',
        tagline: 'strings + number',
      },
      {
        id: 'model:event',
        name: 'event',
        command: 'gof model event name:string start:date end:date',
        tagline: 'multiple dates',
      },
    ],
    description: 'Generating full CRUD stack...',
    baseOutputs: [
      { text: 'SQL migration' },
      { text: 'SQLC queries' },
      { text: 'Proto definitions' },
      { text: 'Domain service layer' },
      { text: 'Transport handlers' },
      { text: 'Validation + tests' },
    ],
    contextOutputs: [
      { text: 'Svelte pages generated', showIf: (s) => s.has('client') },
      { text: 'Subscription checks wired', showIf: (s) => s.has('stripe') },
    ]
  },
  {
    id: 'client',
    label: 'client',
    command: () => 'gof client svelte',
    description: 'Adding Svelte frontend...',
    baseOutputs: [
      { text: 'SvelteKit scaffold' },
      { text: 'Auth integration' },
      { text: 'Type-safe API client' },
    ],
    contextOutputs: [
      // Dynamic: shows actual model names
      { text: (s) => `Generated pages for: ${s.models.join(', ')}`, showIf: (s) => s.models.length > 0 },
      { text: 'Stripe billing UI', showIf: (s) => s.has('stripe') },
      { text: 'File management UI', showIf: (s) => s.has('r2') },
      { text: 'Email dashboard', showIf: (s) => s.has('postmark') },
    ]
  },
  {
    id: 'stripe',
    label: 'stripe',
    command: () => 'gof add stripe',
    description: 'Adding Stripe payments...',
    baseOutputs: [
      { text: 'Payment domain service' },
      { text: 'Subscriptions migration' },
      { text: 'Webhook handlers' },
      { text: 'Access control integration' },
    ],
    contextOutputs: [
      { text: 'Billing UI components', showIf: (s) => s.has('client') },
    ]
  },
  {
    id: 'r2',
    label: 'r2',
    command: () => 'gof add r2',
    description: 'Adding file storage...',
    baseOutputs: [
      { text: 'File domain service' },
      { text: 'Files migration' },
      { text: 'S3-compatible uploads' },
    ],
    contextOutputs: [
      { text: 'File manager UI', showIf: (s) => s.has('client') },
    ]
  },
  {
    id: 'postmark',
    label: 'postmark',
    command: () => 'gof add postmark',
    description: 'Adding email service...',
    baseOutputs: [
      { text: 'Email domain service' },
      { text: 'Emails migration' },
      { text: 'Template support' },
    ],
    contextOutputs: [
      { text: 'Email dashboard UI', showIf: (s) => s.has('client') },
    ]
  },
  {
    id: 'infra',
    label: 'infra',
    command: () => 'gof infra',
    description: 'Adding production infrastructure...',
    baseOutputs: [
      { text: 'Kubernetes manifests' },
      { text: 'Terraform configs' },
      { text: 'OpenTelemetry setup' },
      { text: 'GitHub Actions deploy' },
    ],
    contextOutputs: [
      { text: 'Cloudflare Workers (client)', showIf: (s) => s.has('client') },
      { text: 'Stripe secrets configured', showIf: (s) => s.has('stripe') },
      { text: 'R2 bucket configured', showIf: (s) => s.has('r2') },
      { text: 'Postmark configured', showIf: (s) => s.has('postmark') },
    ]
  },
];
```

### State Store (Svelte 5 Runes)

```javascript
// src/lib/stores/state.svelte.js
import { commands } from '$lib/data/commands.js';

/** @type {Set<string>} */
let completed = $state(new Set());

/** @type {string[]} */
let models = $state([]);

/** @type {boolean} */
let initialized = $state(false);

export const state = {
  get completed() { return completed; },
  get models() { return models; },
  get initialized() { return initialized; },

  has(id) { return completed.has(id); },

  hasModel(name) { return models.includes(name); },

  init() { initialized = true; },

  add(id) {
    completed = new Set([...completed, id]);
  },

  addModel(name) {
    if (!models.includes(name)) {
      models = [...models, name];
    }
  },

  /** Commands still available (excluding completed, handling model specially) */
  get availableCommands() {
    return commands.filter(c => {
      if (c.id === 'init') return false;
      if (c.id === 'model') {
        // Show model if any variants remain
        const usedVariants = c.variants.filter(v => models.includes(v.name));
        return usedVariants.length < c.variants.length;
      }
      return !completed.has(c.id);
    });
  },

  /** For model sub-picker: remaining variants */
  get availableModelVariants() {
    const modelCmd = commands.find(c => c.id === 'model');
    return modelCmd.variants.filter(v => !models.includes(v.name));
  }
};
```

---

## Technical Notes

### Existing Setup

Project already exists in `www/` with:
- SvelteKit + Svelte 5
- Cloudflare adapter (deploy to Cloudflare Pages)
- JS with JSDoc (not TypeScript)

### Dependencies to Add

```bash
cd www
npm install gsap
```

### GSAP Setup

```javascript
// src/lib/animations/gsap.js
import gsap from 'gsap';
import { TextPlugin } from 'gsap/TextPlugin';

// Register plugins (no SSR issues with TextPlugin)
if (typeof window !== 'undefined') {
  gsap.registerPlugin(TextPlugin);
}

export { gsap };
```

### Fonts (via Google Fonts or self-hosted)

```html
<!-- In app.html or via @font-face -->
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500&family=Inter:wght@400;500;600&display=swap" rel="stylesheet">
```

### CSS Variables

```css
/* app.css */
:root {
  --bg: #0a0a0a;
  --surface: #141414;
  --border: #262626;
  --primary: #10b981;
  --primary-glow: #059669;
  --text: #fafafa;
  --text-muted: #737373;
  --success: #22c55e;
}

body {
  background: var(--bg);
  color: var(--text);
  font-family: 'Inter', system-ui, sans-serif;
}

.font-mono {
  font-family: 'JetBrains Mono', monospace;
}
```

---

## Content: What Each Command Generates

### `gof init`
- ✓ OAuth authentication (GitHub + Google)
- ✓ Bitwise permission system (admin/user/custom roles)
- ✓ Docker Compose (PostgreSQL, services)
- ✓ GitHub Actions workflows
- ✓ PR preview deployments
- ✓ PostgreSQL with SQLC type-safe queries
- ✓ ConnectRPC API transport
- ✓ Project structure + configs

### `gof model`
- ✓ SQL migration file
- ✓ SQLC queries (Create, Read, Update, Delete, List)
- ✓ Proto service definitions
- ✓ Domain service layer with validation
- ✓ Transport handlers (ConnectRPC)
- ✓ Comprehensive test suite

### `gof client svelte`
- ✓ SvelteKit frontend scaffold
- ✓ Auto-generated CRUD pages per model
- ✓ Type-safe API client (from proto)
- ✓ TailwindCSS styling
- ✓ Auth integration (login/logout)

### `gof add stripe`
- ✓ Subscription management
- ✓ Checkout sessions
- ✓ Customer portal
- ✓ Webhook handling
- ✓ Access control integration

### `gof add r2`
- ✓ File upload/download
- ✓ S3-compatible storage
- ✓ File management UI

### `gof add postmark`
- ✓ Transactional emails
- ✓ Email templates
- ✓ Attachment support

### `gof infra`
- ✓ Kubernetes manifests
- ✓ Terraform configurations
- ✓ OpenTelemetry observability
- ✓ Cloudflare Workers (client)
- ✓ GCP PostgreSQL setup
- ✓ Production deployment scripts

---

## Implementation Order

1. **Phase 1: Foundation**
   - [ ] Install GSAP dependency
   - [ ] Add fonts (JetBrains Mono, Inter)
   - [ ] Create app.css with CSS variables + dark theme
   - [ ] Add logo to static/

2. **Phase 2: Core Components**
   - [ ] Terminal.svelte - reusable terminal with typing animation
   - [ ] Hero.svelte - logo + tagline + first command
   - [ ] CommandPicker.svelte - choice buttons grid
   - [ ] StateTracker.svelte - current project state display

3. **Phase 3: State + Data**
   - [ ] commands.js - all command definitions with context logic
   - [ ] state.svelte.js - Svelte 5 runes state store
   - [ ] Model sub-picker logic (3 variants)

4. **Phase 4: Animations**
   - [ ] GSAP typing effect for terminal
   - [ ] Checkmark pop-in animation
   - [ ] FlowLine.svelte - connecting lines between terminals
   - [ ] Auto-scroll on command complete

5. **Phase 5: Full Flow**
   - [ ] Wire up: init → choices → command → choices → ...
   - [ ] Context-aware output rendering
   - [ ] "Finish" button → CTA section

6. **Phase 6: Polish**
   - [ ] CTA.svelte with dynamic stack summary
   - [ ] Mobile responsiveness
   - [ ] Hover states on buttons
   - [ ] Final testing of all paths

---

## Decisions

- **Tagline:** "Building blocks for Go"
- **Mobile:** Same interactive flow, responsive design (terminal scales, may scroll horizontally)
- **Loading:** Instant render (static site)
- **Analytics:** TBD (can add Plausible later if needed)

---

## External Links (User to Provide)

| Link | Purpose | Required |
|------|---------|----------|
| Admin app URL | "Get Access" button → OAuth login + Stripe checkout | Yes |
| Discord invite | Community link (free access) | Yes |
| gofast.live | V1 info page | Already known |
| Logo SVG | Place in static/logo.svg | Yes |

---

## Messaging

**V2 positioning:** Focused, opinionated, modern stack (Go + ConnectRPC + Svelte)

**V1 positioning:** Kitchen sink, more options (Next.js, Vue, AWS, GCP, etc.)

**Why both?**
- V2 = "I want the best modern stack, just give it to me"
- V1 = "I need specific tech (Next.js, AWS S3, etc.)"
- $40 gets you both = no wrong choice

---

## Brand Notes (from V1 at gofast.live)

**V1 visual identity:**
- Dark theme + orange (#f90) primary accent
- Clean sans-serif typography
- Honest, anti-hype tone

**V2 differentiation:**
- Same dark theme
- **Emerald/green** primary (fresh, distinct from V1)
- Same honest tone - no marketing fluff
- More interactive/playful (the CLI journey)

**Tone to match V1:**
- Direct, no bullshit
- "You will need to code. But it will work."
- Technical but approachable

**Discord:** 100+ developers in the community (mention this)
