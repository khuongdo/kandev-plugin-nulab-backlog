package plugin

import "context"

// GetConfig passes the plugin's operator config through from the injected
// Host, so the connection can read the OAuth app (R-02).
func (s hostStores) GetConfig(ctx context.Context) (map[string]any, error) {
	h, err := s.get(ctx)
	if err != nil {
		return nil, err
	}
	return h.GetConfig(ctx)
}
