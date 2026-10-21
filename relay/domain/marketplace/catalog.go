package marketplace

// ConfigField describes one form field in the Coordinator marketplace modal.
type ConfigField struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Type     string `json:"type"` // string | secret | url | duration
	Required bool   `json:"required"`
	Help     string `json:"help,omitempty"`
}

// CatalogEntry is a built-in connector tile.
type CatalogEntry struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Detail       string        `json:"detail"`
	Icon         string        `json:"icon"`
	Capabilities []string      `json:"capabilities"` // webhook | poll (transport hints for UI)
	Actions      []ActionMeta  `json:"actions,omitempty"`
	Connectable  bool          `json:"connectable"`
	ConfigSchema []ConfigField `json:"config_schema"`
	SecretKeys   []string      `json:"-"`
}

// ActionMeta is declared in catalog; handler is registered on connector.ActionRegistry.
type ActionMeta struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// Catalog is legacy built-in tiles; stock connectors register via connectors/*/Options().
func Catalog() []CatalogEntry {
	return nil
}

func EntryByID(id string) (CatalogEntry, bool) {
	return lookupCatalogEntry(id)
}
