package render

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name string
		doc  Document
		want string
	}{
		{
			name: "single enum single value",
			doc: Document{
				Package: "api.helloworld.v1",
				Enums: []Enum{{
					Name:        "ErrorReason",
					DefaultCode: 500,
					Values: []Value{
						{Name: "USER_NOT_FOUND", HTTPCode: 404, Description: "User does not exist."},
					},
				}},
			},
			want: "# api.helloworld.v1 error codes\n\n" +
				"> Default HTTP code: `500` (any reason not listed below uses this code).\n\n" +
				"| Reason | HTTP | Description |\n" +
				"|---|---|---|\n" +
				"| `USER_NOT_FOUND` | 404 | User does not exist. |\n",
		},
		{
			name: "multi values mixed code and fallback",
			doc: Document{
				Package: "api.blog.v1",
				Enums: []Enum{{
					Name:        "BlogErrorReason",
					DefaultCode: 500,
					Values: []Value{
						{Name: "UNAUTHORIZED", HTTPCode: 401, Description: ""},
						{Name: "ARTICLE_NOT_FOUND", HTTPCode: 404, Description: "Article not found or already deleted."},
						{Name: "INTERNAL", HTTPCode: 500, Description: ""},
					},
				}},
			},
			want: "# api.blog.v1 error codes\n\n" +
				"> Default HTTP code: `500` (any reason not listed below uses this code).\n\n" +
				"| Reason | HTTP | Description |\n" +
				"|---|---|---|\n" +
				"| `UNAUTHORIZED` | 401 |  |\n" +
				"| `ARTICLE_NOT_FOUND` | 404 | Article not found or already deleted. |\n" +
				"| `INTERNAL` | 500 |  |\n",
		},
		{
			name: "pipe and backslash escaped",
			doc: Document{
				Package: "api.demo.v1",
				Enums: []Enum{{
					Name:        "Reason",
					DefaultCode: 400,
					Values: []Value{
						{Name: "TITLE_INVALID", HTTPCode: 400, Description: `contains | and \ characters`},
					},
				}},
			},
			want: "# api.demo.v1 error codes\n\n" +
				"> Default HTTP code: `400` (any reason not listed below uses this code).\n\n" +
				"| Reason | HTTP | Description |\n" +
				"|---|---|---|\n" +
				"| `TITLE_INVALID` | 400 | contains \\| and \\\\ characters |\n",
		},
		{
			name: "multi enums use h2 sections",
			doc: Document{
				Package: "api.multi.v1",
				Enums: []Enum{
					{Name: "A", DefaultCode: 500, Values: []Value{{Name: "X", HTTPCode: 400, Description: "x"}}},
					{Name: "B", DefaultCode: 503, Values: []Value{{Name: "Y", HTTPCode: 404, Description: "y"}}},
				},
			},
			want: "# api.multi.v1 error codes\n\n" +
				"## A\n\n" +
				"> Default HTTP code: `500` (any reason not listed below uses this code).\n\n" +
				"| Reason | HTTP | Description |\n" +
				"|---|---|---|\n" +
				"| `X` | 400 | x |\n" +
				"\n" +
				"## B\n\n" +
				"> Default HTTP code: `503` (any reason not listed below uses this code).\n\n" +
				"| Reason | HTTP | Description |\n" +
				"|---|---|---|\n" +
				"| `Y` | 404 | y |\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Render(tc.doc)
			if got != tc.want {
				t.Fatalf("mismatch\n--- want ---\n%s\n--- got ---\n%s", tc.want, got)
			}
		})
	}
}

func TestRenderIdempotent(t *testing.T) {
	doc := Document{
		Package: "api.helloworld.v1",
		Enums: []Enum{{
			Name:        "ErrorReason",
			DefaultCode: 500,
			Values: []Value{
				{Name: "USER_NOT_FOUND", HTTPCode: 404, Description: "User does not exist."},
				{Name: "CONTENT_MISSING", HTTPCode: 400, Description: "Request payload is missing required fields."},
			},
		}},
	}
	first := Render(doc)
	second := Render(doc)
	if first != second {
		t.Fatalf("not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestNormalizeComment(t *testing.T) {
	tests := []struct {
		raw, want string
	}{
		{"", ""},
		{"// simple comment\n", "simple comment"},
		{"// first line\n// second line\n", "first line second line"},
		{"//   multiple   spaces\n", "multiple spaces"},
		{"// trailing whitespace   \n", "trailing whitespace"},
		{"//\ttab\tcompressed\n", "tab compressed"},
	}
	for _, tc := range tests {
		t.Run(tc.raw, func(t *testing.T) {
			got := normalizeComment(tc.raw)
			if got != tc.want {
				t.Fatalf("normalizeComment(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestEscapeCell(t *testing.T) {
	if got := escapeCell(`a|b`); got != `a\|b` {
		t.Errorf("pipe escape: got %q", got)
	}
	if got := escapeCell(`a\b`); got != `a\\b` {
		t.Errorf("backslash escape: got %q", got)
	}
	if got := escapeCell(`a\|b`); !strings.Contains(got, `\\\|`) {
		t.Errorf("backslash must be escaped before pipe; got %q", got)
	}
}
