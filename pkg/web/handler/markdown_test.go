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

package handler

import (
	"strings"
	"testing"
)

type fakeModel struct {
	Description string
	Other       string
}

func (f *fakeModel) MarkdownFields() []*string { return []*string{&f.Description} }

func TestApplyMarkdownIn(t *testing.T) {
	m := &fakeModel{Description: "**bold**", Other: "untouched"}
	applyMarkdownIn(m)
	if !strings.Contains(m.Description, "<strong>bold</strong>") {
		t.Errorf("expected HTML strong, got %q", m.Description)
	}
	if m.Other != "untouched" {
		t.Errorf("non-markdown field mutated: %q", m.Other)
	}
}

func TestApplyMarkdownOut(t *testing.T) {
	m := &fakeModel{Description: "<p><strong>bold</strong></p>"}
	applyMarkdownOut(m)
	if strings.TrimSpace(m.Description) != "**bold**" {
		t.Errorf("expected markdown bold, got %q", m.Description)
	}
}

func TestApplyMarkdownOutSlice(t *testing.T) {
	items := []*fakeModel{
		{Description: "<p>a</p>"},
		{Description: "<p>b</p>"},
	}
	applyMarkdownOut(items)
	for i, it := range items {
		if strings.TrimSpace(it.Description) == "" || strings.Contains(it.Description, "<p>") {
			t.Errorf("item %d not converted: %q", i, it.Description)
		}
	}
}

func TestApplyMarkdownIgnoresNonFormattable(t *testing.T) {
	v := struct{ X string }{X: "**bold**"}
	applyMarkdownIn(&v)
	if v.X != "**bold**" {
		t.Errorf("non-formattable struct mutated: %q", v.X)
	}
	applyMarkdownOut(&v)
	if v.X != "**bold**" {
		t.Errorf("non-formattable struct mutated: %q", v.X)
	}
}

func TestApplyMarkdownOutNestedStruct(t *testing.T) {
	type bucket struct {
		Name  string
		Tasks []*fakeModel
	}
	buckets := []*bucket{
		{Name: "todo", Tasks: []*fakeModel{
			{Description: "<p>one</p>"},
			{Description: "<p>two</p>"},
		}},
		{Name: "done", Tasks: []*fakeModel{{Description: "<p>three</p>"}}},
	}
	applyMarkdownOut(buckets)
	for bi, b := range buckets {
		for ti, task := range b.Tasks {
			if strings.Contains(task.Description, "<p>") {
				t.Errorf("bucket %d task %d not converted: %q", bi, ti, task.Description)
			}
		}
	}
}

func TestApplyMarkdownEmptyStringNoOp(t *testing.T) {
	m := &fakeModel{}
	applyMarkdownIn(m)
	applyMarkdownOut(m)
	if m.Description != "" {
		t.Errorf("empty field mutated: %q", m.Description)
	}
}
