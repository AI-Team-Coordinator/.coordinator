package marketplace

// RegisterCatalog adds instance-only marketplace tiles (Alina connectors, etc.).
func RegisterCatalog(entries ...CatalogEntry) {
	extraCatalog = append(extraCatalog, entries...)
}

var extraCatalog []CatalogEntry

func allCatalogEntries() []CatalogEntry {
	out := Catalog()
	out = append(out, extraCatalog...)
	return out
}

func lookupCatalogEntry(id string) (CatalogEntry, bool) {
	for _, e := range allCatalogEntries() {
		if e.ID == id {
			return e, true
		}
	}
	return CatalogEntry{}, false
}
