package notify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"time"
)

// MaskSentinel is what clients send back for a secret they did not change.
const MaskSentinel = "********"

var ErrSecretRequired = errors.New("notify: destination changed, secrets must be re-sent")

type Message struct {
	ReminderID int64
	Title      string
	Notes      string
	PersonName string
	DueAt      time.Time
	// Location is the user's timezone; adapters format times with it, never in UTC.
	Location *time.Location
}

// Channel is implemented once per kind. Send must return a *SendError (or nil);
// adapters must use the injected HTTPClient.
type Channel interface {
	Kind() string
	// Validate checks config shape and URL/topic/token syntax. No network.
	Validate(cfg json.RawMessage) error
	// SecretFields lists JSON keys that hold write-only secrets.
	SecretFields() []string
	// PublicFields is the allow-list of keys the API may return.
	PublicFields() []string
	Send(ctx context.Context, c *HTTPClient, cfg json.RawMessage, m Message) error
}

// PublicView builds API output from the allow-list only. It fails closed: an
// unparsable config yields an empty map.
func PublicView(ch Channel, cfg json.RawMessage) map[string]any {
	out := map[string]any{}

	var all map[string]any
	if err := json.Unmarshal(cfg, &all); err != nil {
		return out
	}

	for _, k := range ch.PublicFields() {
		if v, ok := all[k]; ok && !slices.Contains(ch.SecretFields(), k) {
			out[k] = v
		}
	}

	return out
}

// MergeConfig applies an update to a stored config. Omitted or masked secrets
// keep their stored value, but changing any non-secret field (the destination)
// requires every stored secret to be re-sent, so a stolen session cannot
// redirect an existing secret to a new host.
func MergeConfig(ch Channel, stored, incoming json.RawMessage) (json.RawMessage, error) {
	var in, old map[string]any
	if err := json.Unmarshal(incoming, &in); err != nil || in == nil {
		return nil, fmt.Errorf("notify: config must be a JSON object")
	}

	if err := json.Unmarshal(stored, &old); err != nil {
		return nil, fmt.Errorf("notify: stored config: %w", err)
	}

	secrets := ch.SecretFields()
	destChanged := false

	for _, k := range unionKeys(in, old) {
		if !slices.Contains(secrets, k) && !reflect.DeepEqual(in[k], old[k]) {
			destChanged = true
		}
	}

	for _, k := range secrets {
		if v, ok := in[k]; ok && v != MaskSentinel {
			continue
		}

		if _, had := old[k]; !had {
			delete(in, k)
			continue
		}

		if destChanged {
			return nil, ErrSecretRequired
		}

		in[k] = old[k]
	}

	merged, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	if err := ch.Validate(merged); err != nil {
		return nil, err
	}

	return merged, nil
}

func unionKeys(a, b map[string]any) []string {
	keys := make([]string, 0, len(a)+len(b))
	for k := range a {
		keys = append(keys, k)
	}

	for k := range b {
		if _, ok := a[k]; !ok {
			keys = append(keys, k)
		}
	}

	return keys
}
