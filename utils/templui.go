// templui util templui.go - version: v1.11.0 installed by templui v1.11.0
package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"time"

	twmerge "github.com/Oudwins/tailwind-merge-go"
	"github.com/a-h/templ"
)

// TwMerge combines Tailwind classes and resolves conflicts.
func TwMerge(classes ...string) string {
	return twmerge.Merge(classes...)
}

// If returns value if condition is true, otherwise the zero value of T.
func If[T any](condition bool, value T) T {
	var empty T

	if condition {
		return value
	}

	return empty
}

// IfElse returns trueValue if condition is true, otherwise falseValue.
func IfElse[T any](condition bool, trueValue T, falseValue T) T {
	if condition {
		return trueValue
	}

	return falseValue
}

// MergeAttributes combines multiple Attributes into one.
func MergeAttributes(attrs ...templ.Attributes) templ.Attributes {
	merged := templ.Attributes{}

	for _, attr := range attrs {
		for k, v := range attr {
			merged[k] = v
		}
	}

	return merged
}

// RandomID generates a random ID string.
func RandomID() string {
	return fmt.Sprintf("id-%s", rand.Text())
}

// ScriptVersion is a timestamp generated at app start for cache busting.
var ScriptVersion = fmt.Sprintf("%d", time.Now().Unix())

// ScriptURL generates cache-busted script URLs.
var ScriptURL = func(path string) string {
	return path + "?v=" + ScriptVersion
}

// componentScriptBasePath is the public URL path where component JS files are served.
var componentScriptBasePath = "/static/js"

// UseUnminifiedScripts switches component script loading to the unminified files.
var UseUnminifiedScripts = false

// ComponentScript renders a deferred script tag for a component JavaScript file.
func ComponentScript(component string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		nonce := templ.GetNonce(ctx)

		fileName := component + ".min.js"
		if UseUnminifiedScripts {
			fileName = component + ".js"
		}

		src := ScriptURL(componentScriptBasePath + "/" + fileName)

		if _, err := io.WriteString(w, `<script type="module"`); err != nil {
			return err
		}

		if nonce != "" {
			if _, err := io.WriteString(w, ` nonce="`); err != nil {
				return err
			}

			if _, err := io.WriteString(w, templ.EscapeString(nonce)); err != nil {
				return err
			}

			if _, err := io.WriteString(w, `"`); err != nil {
				return err
			}
		}

		if _, err := io.WriteString(w, ` src="`); err != nil {
			return err
		}

		if _, err := io.WriteString(w, templ.EscapeString(src)); err != nil {
			return err
		}

		if _, err := io.WriteString(w, `"></script>`); err != nil {
			return err
		}

		return nil
	})
}
