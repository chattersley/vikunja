// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Package markdown converts between the HTML stored in task descriptions
// and task comments and Markdown for API clients that prefer plain text
// (LLM agents, scripts, exports).
//
// HTML remains canonical. Markdown is a transport view only — round-trip
// (HTML -> MD -> HTML) is not byte-identical, but is idempotent after the
// second pass once the editor's preferred normalisation is applied.
package markdown

import (
	"bytes"
	"fmt"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var (
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Strikethrough,
			extension.Table,
			extension.TaskList,
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)

	sanitizer = buildSanitizer()
)

func buildSanitizer() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("class").OnElements("code", "pre", "span", "div", "li", "ul", "ol", "p")
	p.AllowAttrs("data-checked", "data-type", "data-id").OnElements("li", "span", "div")
	p.AllowAttrs("checked", "disabled", "type").OnElements("input")
	p.AllowElements("input")
	return p
}

// FromHTML converts stored HTML into a Markdown view. Returns the input
// unchanged when it's empty or pure whitespace.
func FromHTML(in string) (string, error) {
	if strings.TrimSpace(in) == "" {
		return in, nil
	}
	out, err := htmltomarkdown.ConvertString(in)
	if err != nil {
		return "", fmt.Errorf("html->markdown: %w", err)
	}
	return out, nil
}

// ToHTML parses a Markdown body and returns sanitised HTML suitable
// for storage. Sanitisation applies the same allow-list used elsewhere so
// API clients can't bypass the editor's input rules.
func ToHTML(in string) (string, error) {
	if strings.TrimSpace(in) == "" {
		return in, nil
	}
	var buf bytes.Buffer
	if err := md.Convert([]byte(in), &buf); err != nil {
		return "", fmt.Errorf("markdown->html: %w", err)
	}
	return sanitizer.Sanitize(buf.String()), nil
}
