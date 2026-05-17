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

package markdown

import (
	"strings"
	"testing"
)

func TestFromHTML(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{"empty", "", ""},
		{"whitespace passthrough", "   \n\t  ", "   \n\t  "},
		{"paragraph", "<p>hello world</p>", "hello world"},
		{"bold and italic", "<p><strong>bold</strong> and <em>italic</em></p>", "**bold** and *italic*"},
		{"link", `<p><a href="https://vikunja.io">site</a></p>`, "[site](https://vikunja.io)"},
		{"unordered list", "<ul><li>one</li><li>two</li></ul>", "- one\n- two"},
		{"ordered list", "<ol><li>one</li><li>two</li></ol>", "1. one\n2. two"},
		{"code block", "<pre><code>go fmt ./...</code></pre>", "```\ngo fmt ./...\n```"},
		{"inline code", "<p>use <code>git pull</code></p>", "use `git pull`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromHTML(tt.html)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if strings.TrimSpace(got) != strings.TrimSpace(tt.want) {
				t.Errorf("FromHTML(%q)\n got:  %q\n want: %q", tt.html, got, tt.want)
			}
		})
	}
}

func TestToHTML(t *testing.T) {
	tests := []struct {
		name        string
		md          string
		wantSubstrs []string
	}{
		{"empty", "", nil},
		{"paragraph", "hello world", []string{"<p>hello world</p>"}},
		{"bold", "**bold**", []string{"<strong>bold</strong>"}},
		{"link", "[site](https://vikunja.io)", []string{`href="https://vikunja.io"`, `>site</a>`}},
		{"list", "- one\n- two", []string{"<ul>", "<li>one</li>", "<li>two</li>"}},
		{"task list", "- [x] done\n- [ ] todo", []string{`<input`, `checked=""`, `disabled=""`}},
		{"table", "| a | b |\n|---|---|\n| 1 | 2 |", []string{"<table>", "<th>a</th>", "<td>1</td>"}},
		{"strip script", "ok\n\n<script>alert('x')</script>", []string{"<p>ok</p>"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToHTML(tt.md)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			for _, want := range tt.wantSubstrs {
				if !strings.Contains(got, want) {
					t.Errorf("ToHTML(%q) =\n%s\nmissing substring %q", tt.md, got, want)
				}
			}
			if strings.Contains(got, "<script>") {
				t.Errorf("sanitizer failed to strip <script>: %s", got)
			}
		})
	}
}

func TestRoundTripIdempotent(t *testing.T) {
	cases := []string{
		"<p>hello <strong>bold</strong> world</p>",
		`<p><a href="https://vikunja.io">vikunja</a></p>`,
		"<ul><li>one</li><li>two</li></ul>",
		"<p>line one</p><p>line two</p>",
	}

	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			md1, err := FromHTML(in)
			if err != nil {
				t.Fatal(err)
			}
			html1, err := ToHTML(md1)
			if err != nil {
				t.Fatal(err)
			}
			md2, err := FromHTML(html1)
			if err != nil {
				t.Fatal(err)
			}
			if md1 != md2 {
				t.Errorf("round-trip not idempotent\n first md:  %q\n second md: %q", md1, md2)
			}
		})
	}
}
