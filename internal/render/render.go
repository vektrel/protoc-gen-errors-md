package render

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

type Enum struct {
	Name        string
	DefaultCode int32
	Values      []Value
}

type Value struct {
	Name        string
	HTTPCode    int32
	Description string
}

type Document struct {
	Package string
	Enums   []Enum
}

func Run(plugin *protogen.Plugin) error {
	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}
		var enums []Enum
		for _, enum := range file.Enums {
			e, ok := extractEnum(enum)
			if !ok {
				continue
			}
			enums = append(enums, e)
		}
		if len(enums) == 0 {
			continue
		}
		doc := Document{
			Package: string(file.Desc.Package()),
			Enums:   enums,
		}
		out := plugin.NewGeneratedFile(file.GeneratedFilenamePrefix+".md", "")
		if _, err := out.Write([]byte(Render(doc))); err != nil {
			return err
		}
	}
	return nil
}

func extractEnum(enum *protogen.Enum) (Enum, bool) {
	opts := enum.Desc.Options()
	if !proto.HasExtension(opts, errors.E_DefaultCode) {
		return Enum{}, false
	}
	defaultCode := proto.GetExtension(opts, errors.E_DefaultCode).(int32)
	out := Enum{
		Name:        string(enum.Desc.Name()),
		DefaultCode: defaultCode,
	}
	for _, v := range enum.Values {
		if v.Desc.Number() == 0 && strings.HasSuffix(string(v.Desc.Name()), "_UNSPECIFIED") {
			continue
		}
		code := defaultCode
		vopts := v.Desc.Options()
		if proto.HasExtension(vopts, errors.E_Code) {
			code = proto.GetExtension(vopts, errors.E_Code).(int32)
		}
		out.Values = append(out.Values, Value{
			Name:        string(v.Desc.Name()),
			HTTPCode:    code,
			Description: normalizeComment(string(v.Comments.Leading)),
		})
	}
	return out, true
}

var whitespace = regexp.MustCompile(`\s+`)

func normalizeComment(raw string) string {
	if raw == "" {
		return ""
	}
	lines := strings.Split(raw, "\n")
	parts := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(l), "//"))
		if l != "" {
			parts = append(parts, l)
		}
	}
	joined := strings.Join(parts, " ")
	return whitespace.ReplaceAllString(joined, " ")
}

func escapeCell(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "|", `\|`)
	return s
}

func Render(doc Document) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s error codes\n\n", doc.Package)
	multi := len(doc.Enums) > 1
	for i, e := range doc.Enums {
		if multi {
			if i > 0 {
				b.WriteString("\n")
			}
			fmt.Fprintf(&b, "## %s\n\n", e.Name)
		}
		fmt.Fprintf(&b, "> Default HTTP code: `%d` (any reason not listed below uses this code).\n\n", e.DefaultCode)
		b.WriteString("| Reason | HTTP | Description |\n")
		b.WriteString("|---|---|---|\n")
		for _, v := range e.Values {
			fmt.Fprintf(&b, "| `%s` | %d | %s |\n", v.Name, v.HTTPCode, escapeCell(v.Description))
		}
	}
	return b.String()
}
