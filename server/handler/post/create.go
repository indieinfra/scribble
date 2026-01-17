package post

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/indieinfra/scribble/server/auth"
	"github.com/indieinfra/scribble/server/resp"
	"github.com/indieinfra/scribble/server/state"
	"github.com/indieinfra/scribble/server/util"
	"github.com/indieinfra/scribble/storage/content"
)

func Create(st *state.ScribbleState, w http.ResponseWriter, r *http.Request, data map[string]any) {
	if !auth.RequestHasScope(r, auth.ScopeCreate) {
		resp.WriteInsufficientScope(w, "no create scope")
		return
	}

	ct, _ := util.ExtractMediaType(w, r)

	var document util.Mf2Document
	switch ct {
	case "application/json":
		document = normalizeJson(data)
	case "application/x-www-form-urlencoded":
		document = normalizeFormBody(data)
		delete(document.Properties, "access_token")
	}

	err := util.ValidateMf2(document)
	if err != nil {
		resp.WriteInvalidRequest(w, err.Error())
		return
	}

	// Process server commands (mp-* properties)
	suggestedSlug := processMpProperties(&document)

	// If no suggested slug, generate one from content
	if suggestedSlug == "" {
		suggestedSlug = util.GenerateSlug(document)
	}

	// If still no slug, return error - document must have content or name
	if suggestedSlug == "" {
		resp.WriteInvalidRequest(w, "unable to generate slug: document must contain a name or content property")
		return
	}

	// Check if slug already exists
	exists, err := st.ContentStore.ExistsBySlug(r.Context(), suggestedSlug)
	if exists || err != nil {
		uuid, err := uuid.NewRandom()
		if err != nil {
			resp.WriteInternalServerError(w, "slug clash - failed to create UUID while attempting to resolve")
			return
		}

		suggestedSlug += uuid.String()
	}

	// Store the final slug as "slug" property (not mp-slug)
	document.Properties["slug"] = []any{suggestedSlug}

	url, now, err := st.ContentStore.Create(r.Context(), document)
	if err != nil {
		resp.WriteInternalServerError(w, err.Error())
		return
	}

	if now {
		resp.WriteCreated(w, url)
	} else {
		resp.WriteAccepted(w, url)
	}
}

func normalizeJson(input map[string]any) util.Mf2Document {
	doc := util.Mf2Document{
		Type:       []string{"h-entry"},
		Properties: content.MicropubProperties{},
	}

	if rawType, ok := input["type"]; ok {
		switch v := rawType.(type) {
		case string:
			doc.Type = []string{v}
		case []any:
			var types []string
			for _, t := range v {
				if s, ok := t.(string); ok {
					types = append(types, s)
				}
			}

			if len(types) > 0 {
				doc.Type = types
			}
		}
	}

	rawProps, ok := input["properties"]
	if !ok {
		return doc
	}

	props, ok := rawProps.(map[string]any)
	if !ok {
		return doc
	}

	for key, val := range props {
		switch v := val.(type) {
		case string:
			doc.Properties[key] = []any{v}
		case []any:
			doc.Properties[key] = normalizeJsonArray(v)
		case map[string]any:
			// Preserve maps as-is for embedded objects like {html: ["..."], value: ["..."]}
			doc.Properties[key] = []any{v}
		case nil:
			// Skip nil values
		default:
			// Preserve other types (numbers, booleans, etc.)
			doc.Properties[key] = []any{v}
		}
	}

	return doc
}

func normalizeJsonArray(arr []any) []any {
	out := make([]any, 0, len(arr))

	for _, v := range arr {
		switch x := v.(type) {
		case string:
			out = append(out, x)
		case map[string]any:
			// Preserve maps as-is (e.g., {html: ["..."], value: ["..."]})
			// Don't recursively normalize them to avoid losing structure
			out = append(out, x)
		case nil:
			// Skip nil values
		default:
			// Preserve other types (numbers, booleans, etc.)
			out = append(out, x)
		}
	}

	return out
}

func normalizeFormBody(props map[string]any) util.Mf2Document {
	doc := util.Mf2Document{
		Type:       []string{"h-entry"},
		Properties: content.MicropubProperties{},
	}

	for key, val := range props {
		if key == "h" {
			if s, ok := firstString(val); ok {
				doc.Type = []string{"h-" + s}
			}
			continue
		}

		if strings.HasSuffix(key, "[]") {
			key, _ = strings.CutSuffix(key, "[]")
		}

		values := coerceSlice(val)
		if len(values) == 0 {
			continue
		}

		doc.Properties[key] = values
	}

	return doc
}

func firstString(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case []any:
		if len(x) > 0 {
			if s, ok := x[0].(string); ok {
				return s, true
			}
		}
	}
	return "", false
}

// extractStringFromProperty extracts the first string value from an MF2 property ([]any)
func extractStringFromProperty(values []any) string {
	for _, val := range values {
		if s, ok := val.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

// processMpProperties handles server command properties (mp-*) and removes them from the document.
// Returns the suggested slug from mp-slug if present, otherwise returns empty string.
func processMpProperties(doc *util.Mf2Document) string {
	var suggestedSlug string

	// Extract mp-slug if present
	if mpSlugProp, ok := doc.Properties["mp-slug"]; ok {
		suggestedSlug = extractStringFromProperty(mpSlugProp)
	}

	// Collect mp-* keys first to avoid modifying map during iteration
	var mpKeys []string
	for key := range doc.Properties {
		if strings.HasPrefix(key, "mp-") {
			mpKeys = append(mpKeys, key)
		}
	}

	// Remove all mp-* (server command) properties per spec
	for _, key := range mpKeys {
		delete(doc.Properties, key)
	}

	return suggestedSlug
}

func coerceSlice(v any) []any {
	var out []any

	switch x := v.(type) {
	case []any:
		for _, e := range x {
			// Preserve all non-nil types
			if e != nil {
				out = append(out, e)
			}
		}
	default:
		// Preserve single non-nil values
		if x != nil {
			out = append(out, x)
		}
	}

	return out
}
