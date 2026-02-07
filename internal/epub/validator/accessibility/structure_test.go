package accessibility

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub/testutil"
)

func TestEpubTypeWithoutRole(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en">
<head><title>Test</title></head>
<body>
  <section epub:type="chapter">
    <h1>Chapter 1</h1>
  </section>
</body>
</html>`)

	v := &StructureValidator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if !testutil.HasCode(diags, "epub-type-has-matching-role") {
		t.Error("expected epub-type-has-matching-role for chapter without role")
	}
}

func TestEpubTypeWithMatchingRole(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en">
<head><title>Test</title></head>
<body>
  <section epub:type="chapter" role="doc-chapter">
    <h1>Chapter 1</h1>
  </section>
</body>
</html>`)

	v := &StructureValidator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if testutil.HasCode(diags, "epub-type-has-matching-role") {
		t.Error("unexpected epub-type-has-matching-role when role matches")
	}
}

func TestPageBreakWithoutLabel(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en">
<head><title>Test</title></head>
<body>
  <span epub:type="pagebreak" id="pg1"/>
</body>
</html>`)

	v := &StructureValidator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if !testutil.HasCode(diags, "pagebreak-label") {
		t.Error("expected pagebreak-label for pagebreak without label")
	}
}

func TestPageBreakWithAriaLabel(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en">
<head><title>Test</title></head>
<body>
  <span epub:type="pagebreak" id="pg1" aria-label="1"/>
</body>
</html>`)

	v := &StructureValidator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if testutil.HasCode(diags, "pagebreak-label") {
		t.Error("unexpected pagebreak-label when aria-label is present")
	}
}

func TestPageBreakWithTitle(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en">
<head><title>Test</title></head>
<body>
  <span epub:type="pagebreak" id="pg1" title="Page 1"/>
</body>
</html>`)

	v := &StructureValidator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if testutil.HasCode(diags, "pagebreak-label") {
		t.Error("unexpected pagebreak-label when title is present")
	}
}

func TestHeadingLevelSkip(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head><title>Test</title></head>
<body>
  <h1>Title</h1>
  <h3>Skipped h2</h3>
</body>
</html>`)

	v := &StructureValidator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if !testutil.HasCode(diags, "heading-order") {
		t.Error("expected heading-order for h1 -> h3 skip")
	}
}

func TestHeadingLevelNoSkip(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head><title>Test</title></head>
<body>
  <h1>Title</h1>
  <h2>Section</h2>
  <h3>Subsection</h3>
  <h2>Another Section</h2>
</body>
</html>`)

	v := &StructureValidator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if testutil.HasCode(diags, "heading-order") {
		t.Error("unexpected heading-order when levels don't skip")
	}
}

func TestTableWithoutCaption(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head><title>Test</title></head>
<body>
  <table>
    <tr><td>Data</td></tr>
  </table>
</body>
</html>`)

	v := &StructureValidator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if !testutil.HasCode(diags, "table-caption") {
		t.Error("expected table-caption for table without caption")
	}
}

func TestTableWithCaption(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head><title>Test</title></head>
<body>
  <table>
    <caption>My Table</caption>
    <tr><td>Data</td></tr>
  </table>
</body>
</html>`)

	v := &StructureValidator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if testutil.HasCode(diags, "table-caption") {
		t.Error("unexpected table-caption when caption is present")
	}
}

func TestTableWithAriaLabel(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head><title>Test</title></head>
<body>
  <table aria-label="Data table">
    <tr><td>Data</td></tr>
  </table>
</body>
</html>`)

	v := &StructureValidator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if testutil.HasCode(diags, "table-caption") {
		t.Error("unexpected table-caption when aria-label is present")
	}
}

func TestInputWithoutLabel(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head><title>Test</title></head>
<body>
  <form>
    <input type="text" id="name"/>
  </form>
</body>
</html>`)

	v := &StructureValidator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if !testutil.HasCode(diags, "input-label") {
		t.Error("expected input-label for input without label")
	}
}

func TestInputWithLabel(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head><title>Test</title></head>
<body>
  <form>
    <label for="name">Name:</label>
    <input type="text" id="name"/>
  </form>
</body>
</html>`)

	v := &StructureValidator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if testutil.HasCode(diags, "input-label") {
		t.Error("unexpected input-label when label for= matches")
	}
}

func TestInputHiddenNoLabel(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head><title>Test</title></head>
<body>
  <form>
    <input type="hidden" name="token" value="abc"/>
    <input type="submit" value="Go"/>
  </form>
</body>
</html>`)

	v := &StructureValidator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if testutil.HasCode(diags, "input-label") {
		t.Error("unexpected input-label for hidden/submit inputs")
	}
}

func TestSelectWithoutLabel(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head><title>Test</title></head>
<body>
  <form>
    <select id="choice">
      <option>A</option>
    </select>
  </form>
</body>
</html>`)

	v := &StructureValidator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if !testutil.HasCode(diags, "input-label") {
		t.Error("expected input-label for select without label")
	}
}
