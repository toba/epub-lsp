---
name: epub
description: "EPUB 3.3 photo book generation from Lightroom Classic collections. Use when: (1) Working on internal/epub/ package code (2) Creating or modifying EPUB output (OPF, XHTML pages, navigation, CSS) (3) Debugging EPUB validation errors (epubcheck) (4) Adding fixed-layout (FXL) support for pre-paginated photo pages (5) Modifying the ZIP container structure (mimetype, META-INF, OEBPS) (6) Working with Dublin Core metadata or rendition properties (7) User mentions epub, photo book, ebook, opf, xhtml content documents, or epub packaging"
---

# EPUB 3.3 Photo Book Generator

Generate EPUB 3.3 photo books from Lightroom Classic collections via the Go package at `internal/epub/`.

## EPUB 3.3 Spec Essentials

### Container Structure

```
mimetype                  ← first file, uncompressed, no extra fields
META-INF/
  container.xml           ← points to OEBPS/content.opf
OEBPS/
  content.opf             ← package document
  toc.xhtml               ← EPUB 3 navigation document
  toc.ncx                 ← EPUB 2 backward compat
  styles/
    book.css
  text/
    cover.xhtml
    title.xhtml
    photo_001.xhtml ...
    map.xhtml             ← optional
  images/
    cover.jpg
    001.jpg ... NNN.jpg
    map.png               ← optional
```

### mimetype File

Must be:
- First entry in the ZIP archive
- Uncompressed (stored, no deflation)
- No extra field data in the local file header
- Content: `application/epub+zip` (no trailing newline)

### OPF Package Document

```xml
<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="pub-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="pub-id">urn:uuid:...</dc:identifier>
    <dc:title>...</dc:title>
    <dc:creator>...</dc:creator>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="toc.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
    <item id="css" href="styles/book.css" media-type="text/css"/>
    <item id="cover" href="text/cover.xhtml" media-type="application/xhtml+xml"/>
    <!-- ... -->
  </manifest>
  <spine toc="ncx">
    <itemref idref="cover"/>
    <itemref idref="title"/>
    <!-- photo pages -->
  </spine>
</package>
```

Required metadata: `dc:identifier`, `dc:title`, `dc:language`, `dcterms:modified`.

### XHTML Content Documents

All content documents must be valid XHTML5 with the EPUB namespace:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops"
      xml:lang="en" lang="en">
<head>
  <meta charset="UTF-8"/>
  <title>Page Title</title>
  <link rel="stylesheet" type="text/css" href="../styles/book.css"/>
</head>
<body>
  <!-- content -->
</body>
</html>
```

Key requirements:
- Must include `xml:lang` and `lang` attributes on `<html>`
- Use `xmlns:epub` for semantic inflection (`epub:type`)
- Self-closing tags for void elements (`<meta/>`, `<link/>`, `<img/>`)
- All attribute values must be quoted

### Navigation Document (toc.xhtml)

Must contain a `<nav epub:type="toc">` element. Should also include landmarks:

```xml
<nav epub:type="landmarks">
  <ol>
    <li><a epub:type="cover" href="text/cover.xhtml">Cover</a></li>
    <li><a epub:type="toc" href="toc.xhtml">Table of Contents</a></li>
    <li><a epub:type="bodymatter" href="text/photo_001.xhtml">Photos</a></li>
  </ol>
</nav>
```

### Semantic Inflection

Use `epub:type` for structural semantics:
- `cover` — cover page
- `titlepage` — title page
- `toc` — table of contents
- `bodymatter` — main content
- `footnote`, `endnote`, `bibliography` — back matter (not used in photo books)

## CSS in EPUB 3.3

### Spec-Guaranteed Support

EPUB 3.3 references **CSS 2.1 as the required baseline**, plus specific CSS3 modules:

| Module | Key Features |
|--------|-------------|
| CSS 2.1 | Selectors, box model, positioning, floats, fonts, colors |
| CSS Writing Modes 3 | `writing-mode`, `direction`, vertical text |
| CSS Speech | Aural rendering (`speak`, `voice-family`) |
| CSS Fonts 3 | `@font-face`, `font-feature-settings` |
| CSS Text 3 | `word-spacing`, `letter-spacing`, `text-transform` |
| CSS Text Decoration 3 | `text-decoration-color`, `text-decoration-style` |

### Widely Supported (Not Spec-Guaranteed)

Major reading systems (Apple Books, Kobo, Readium, Thorium) support more CSS3 than the spec requires. These are **practically safe** but not guaranteed:

- Flexbox (`display: flex`)
- `calc()`
- Viewport units (`vw`, `vh`, `vmin`, `vmax`)
- `object-fit`, `object-position`
- CSS custom properties (`--var`)
- `border-radius`, `box-shadow`, `opacity`
- CSS Grid (Apple Books, Readium — less universal)

### Not Safe / Avoid

- CSS animations and transitions (inconsistent support)
- `position: fixed` (meaningless in paginated context)
- `position: sticky`
- Media queries beyond `min-width`/`max-width` (limited in many readers)
- `backdrop-filter`
- Container queries

### Current Project CSS Approach

The stylesheet is defined in `internal/epub/style.go` as the `bookCSS` constant:
- Georgia serif font stack
- Responsive images: `max-width: 100%`, `object-fit: contain`
- Cover: flexbox centering with `object-fit: cover`
- Minimal spacing, white background, dark text (#333)
- Title page: 2.2em heading, muted gray author/date

### Fixed-Layout (FXL) Considerations

For pre-paginated photo pages, add to OPF metadata:
```xml
<meta property="rendition:layout">pre-paginated</meta>
<meta property="rendition:spread">auto</meta>
```

Each FXL page needs a viewport meta tag:
```xml
<meta name="viewport" content="width=1200, height=1600"/>
```

FXL pages use absolute positioning — CSS2.1 features are sufficient.

## Project Code Structure

### Files

| File | Purpose |
|------|---------|
| `internal/epub/epub.go` | Main `Generate()` entry point, ZIP assembly, OPF/XHTML templates |
| `internal/epub/style.go` | CSS stylesheet constant (`bookCSS`) |
| `internal/epub/image.go` | `ResizeJPEG()` — downscale with CatmullRom resampling |
| `internal/config/config.go` | `Config` struct with author, image size, quality settings |

### Config Defaults

| Field | Default | Purpose |
|-------|---------|---------|
| `Author` | — | DC creator metadata |
| `MaxPhotoEdge` | 1200 | Max long-edge pixels |
| `JPEGQuality` | 85 | Re-encode quality |
| `MapWidth` | 600 | Static map pixel width |
| `MapHeight` | 400 | Static map pixel height |

### Page Types Generated

1. **Cover** (`cover.xhtml`) — full-bleed image, `epub:type="cover"`
2. **Title** (`title.xhtml`) — title, summary, author, date; `epub:type="titlepage"`
3. **Photo** (`photo_NNN.xhtml`) — image + optional h2 title + optional caption
4. **Map** (`map.xhtml`) — optional static map from GPS data

## Validation

Use [epubcheck](https://www.w3.org/publishing/epubcheck/) to validate output:

```bash
# Install
brew install epubcheck

# Validate
epubcheck output.epub
```

Common validation errors:
- Missing `dc:language` or `dcterms:modified` in OPF
- `mimetype` file compressed or not first in ZIP
- Missing `xml:lang` / `lang` on `<html>` element
- Nav document missing `epub:type="toc"` nav element
- Image referenced in OPF but not in ZIP (or vice versa)
- Non-self-closing void elements in XHTML

## Build & Test

```bash
# Build the CLI
scripts/build-mac.sh

# Run Go tests
scripts/test.sh

# End-to-end test
scripts/test-e2e.sh

# Lint Go code
golangci-lint run ./...
```
