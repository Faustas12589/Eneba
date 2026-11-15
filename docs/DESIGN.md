# Eneba design reference

Values come from eneba.com's own stylesheets (`main.*.css` + the store page chunk,
fetched 2026-11-15; hex counts are how often each color appears there) and from the
screenshots in `docs/screens/` (`eneba-home.png`, `eneba-cart.png`). They are mirrored
into `frontend/src/styles/tokens.css`; components only use those variables.

Eneba sets `html { font-size: 62.5% }`, so its `1.4rem` = 14px. Below everything is in px.

## Colors
| Role | Hex | Where Eneba uses it |
|---|---|---|
| background | `#4618AC` | `body`, `<meta name="theme-color">` |
| band / nav row | `#3C1395` | second header row, alternating home sections |
| surface (panels) | `#351183` | cart panels, menus, summary box |
| surface-2 (inner) | `#2C0E6C` | cart line inside a panel, pills |
| card / footer | `#1F0A4D` | product card info area, footer |
| border / secondary | `#5825CC` | dividers, side menus, secondary buttons |
| primary CTA | `#FAD318` (hover `#F0A613`) | yellow buttons, active page number, hover borders |
| text on primary | `#351183` | dark purple text on yellow |
| text | `#FFFFFF` | |
| muted text | `#B3AAC9` | secondary text, table borders |
| success | `#23C299` (light `#3DDCB3`) | region label, cashback, success states |
| danger | `#E76A8C` (strong `#D32252`) | errors, warnings |

## Typography
- Font: **Metropolis** (SIL OFL, self-hosted in `frontend/public/fonts/metropolis/`), fallback Arial, Helvetica.
- Body: 16px / 1.5. Most UI text is 12px or 14px (the two most used sizes in their CSS).
- Weights: 400 body, 600 nav/labels, 700 prices/buttons, 800 headings and card titles.
- Headings: 20px section titles ("Rekomenduojama tau"), ~24-28px page titles.

## Spacing & shape
- Spacing: multiples of 5px in their CSS (5/10/15/20); we keep a 4/8/12/16/24/32 scale (close enough, tokens).
- Radii: small. 3-5px on buttons, inputs and panels; 50% for avatars; pills 20px+.
- Cards are almost square-cornered; the image fills the top edge to edge.
- Container: max 1200px.

## Components
- **Header:** two rows on the background color. Row 1: logo, a wide search box (transparent,
  1px light border, magnifier icon, placeholder "Ieškoti žaidimų, papildymų ir daugiau"),
  then icons on the right (wishlist, cart, account). Row 2 (slightly darker band):
  "Kategorijos" + text links (Pigūs žaidimai, Progr. įranga, E. kortelės, ...), 12-14px, semibold.
- **Product card:** narrow (~180-200px). Cover image 3:4 on top, then a dark `#1F0A4D` block:
  title bold 12-14px (2 lines max), region in green caps (`EUROPA`, `VISAS PASAULIS`),
  a muted line "Nuo" with the struck-through list price and `-NN%`, then the price big and bold
  (~20px). Hover: yellow border. Cashback items get a green border and "grįžta" tag (we skip cashback).
- **Price / discount:** discount is plain text `-NN%` next to the old price, not a big badge.
- **Buttons:** primary = yellow `#FAD318`, dark purple text, bold, small radius, full width
  in the cart summary ("Pereiti prie apmokėjimo"). Secondary = `#5825CC` on purple.
- **Cart page:** step indicator at top (1 Krepšelis, 2 Mokėjimas, 3 Atsiimti prekę).
  Left panel "Mano krepšelis" (`#351183`) with each line on a darker inner block: image,
  "Parduoda <seller>", title, `- 2 +` stepper, price. Right panel "Suvestinė": yellow CTA,
  item count + subtotal, "Iš viso" with a large total.
- **Offer table (product page):** sellers listed cheapest first on panels, price + yellow buy button.
- **Footer:** white strip with payment logos, then a dark `#1F0A4D` block with help text and
  a small copyright line.

## Homepage layout (from `docs/screens/desired.png`, Eneba at ~940px wide)
- Header row ~92px: logo, search 56px tall with a 2px white border and 18px semibold text,
  then a cart icon with a red count badge (`#D32252`) and a round avatar.
- Category row ~52px on the darker band: "☰ Kategorijos" + 16px semibold links.
- Hero: full-bleed, no gaps or radius, ~280px tall. First banner wide (~60%), the rest narrow tiles.
- Category icon row under the hero: line icon, name, (ours) product count.
- Section titles ~32px extra-bold; cards ~214px wide with ~24px gaps; no visible scrollbar.
