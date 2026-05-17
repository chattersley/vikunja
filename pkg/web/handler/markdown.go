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
	"reflect"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/markdown"
	"code.vikunja.io/api/pkg/web"

	"github.com/labstack/echo/v5"
)

const markdownFormatValue = "markdown"

// wantsMarkdown returns true when the caller requested ?format=markdown.
func wantsMarkdown(ctx *echo.Context) bool {
	return ctx.QueryParam("format") == markdownFormatValue
}

// applyMarkdownIn converts the Markdown fields of obj from MD -> HTML before
// the model layer sees them. Errors on a single field are logged and the
// field is left untouched — never fails the whole request.
func applyMarkdownIn(obj any) {
	mf, ok := obj.(web.MarkdownFormattable)
	if !ok {
		return
	}
	for _, fp := range mf.MarkdownFields() {
		if fp == nil || *fp == "" {
			continue
		}
		out, err := markdown.ToHTML(*fp)
		if err != nil {
			log.Errorf("format=markdown: input conversion failed: %s", err)
			continue
		}
		*fp = out
	}
}

// applyMarkdownOut converts MarkdownFormattable fields of obj from stored
// HTML to Markdown before the response is serialised. Walks slices, maps,
// struct fields, and pointers up to maxMarkdownDepth so nested collections
// (e.g. TaskCollection returning buckets containing tasks) are handled.
func applyMarkdownOut(obj any) {
	if obj == nil {
		return
	}
	walkMarkdownOut(reflect.ValueOf(obj), 0)
}

const maxMarkdownDepth = 6

func walkMarkdownOut(v reflect.Value, depth int) {
	if depth > maxMarkdownDepth || !v.IsValid() {
		return
	}

	//nolint:exhaustive // we only care about composite kinds; scalars are ignored.
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return
		}
		if v.CanInterface() {
			if mf, ok := v.Interface().(web.MarkdownFormattable); ok {
				convertMarkdownFields(mf)
				return
			}
		}
		walkMarkdownOut(v.Elem(), depth+1)
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			walkMarkdownOut(v.Index(i), depth+1)
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			walkMarkdownOut(iter.Value(), depth+1)
		}
	case reflect.Struct:
		if v.CanAddr() && v.Addr().CanInterface() {
			if mf, ok := v.Addr().Interface().(web.MarkdownFormattable); ok {
				convertMarkdownFields(mf)
				return
			}
		}
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if !f.CanInterface() {
				continue
			}
			walkMarkdownOut(f, depth+1)
		}
	}
}

func convertMarkdownFields(mf web.MarkdownFormattable) {
	for _, fp := range mf.MarkdownFields() {
		if fp == nil || *fp == "" {
			continue
		}
		out, err := markdown.FromHTML(*fp)
		if err != nil {
			log.Errorf("format=markdown: output conversion failed: %s", err)
			continue
		}
		*fp = out
	}
}
