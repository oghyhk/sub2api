# Onboarding Page Overrides

> **PROJECT:** Five Model API
> **Generated:** 2026-07-23 09:30:04
> **Page Type:** Product Detail

> ⚠️ **IMPORTANT:** Rules in this file **override** the Master file (`design-system/MASTER.md`).
> Only deviations from the Master are documented here. For all other rules, refer to the Master.

---

## Page-Specific Rules

### Product Direction Override (Authoritative)

- Build an actual guided start experience, not a generic marketing landing page.
- First viewport: literal product/site name, the offer "Five models. One API key.", the five model choices, and one primary action. Leave the next setup section visible.
- Present GPT-5.6 Sol, Terra, Luna and Gemini 3.1 Pro, 3.6 Flash as selectable product rows or tiles with provider marks and plain-language roles.
- Show the three-step path directly: create account, create key, connect an OpenAI- or Google-compatible client.
- Authenticated users continue to the dashboard; unauthenticated users continue to login/registration. Preserve custom home-content behavior.
- Use white/black neutral full-width bands, Apple blue actions, emerald GPT and cobalt Gemini accents. Do not use gradients, orbs, fake terminal decoration, or card-in-card layouts.
- Mobile controls are at least 44px high, content has no horizontal overflow, and long model names wrap cleanly.

### Layout Overrides

- **Max Width:** 1200px (standard)
- **Layout:** Full-width sections, centered content
- **Sections:** 1. Hero, 2. Step 1 (problem), 3. Step 2 (solution), 4. Step 3 (action), 5. CTA progression

### Spacing Overrides

- No overrides — use Master spacing

### Typography Overrides

- No overrides — use Master typography

### Color Overrides

- **Strategy:** Step colors: 1 (Red/Problem), 2 (Orange/Process), 3 (Green/Solution). CTA: Brand color

### Component Overrides

- Avoid: Force linear unskippable tour

---

## Page-Specific Components

- No unique components for this page

---

## Recommendations

- Effects: Inner+outer shadows (subtle, no hard lines), soft press (200ms ease-out), fluffy elements, smooth transitions
- Onboarding: Provide Skip and Back buttons
- CTA Placement: Each step: mini-CTA. Final: main CTA
