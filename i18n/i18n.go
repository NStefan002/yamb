package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	mu           sync.RWMutex
	Translations map[string]map[string]string
)

func LoadLocales(dir string) error {
	mu.Lock()
	defer mu.Unlock()

	Translations = map[string]map[string]string{}

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if filepath.Ext(f.Name()) != ".json" {
			continue
		}
		lang := f.Name()[0 : len(f.Name())-len(filepath.Ext(f.Name()))] // en.json -> en

		b, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			return fmt.Errorf("read %s: %w", f.Name(), err)
		}
		var data map[string]string
		if err := json.Unmarshal(b, &data); err != nil {
			return fmt.Errorf("unmarshal %s: %w", f.Name(), err)
		}
		Translations[lang] = data
	}

	return nil
}

// T returns translation for lang,key; falls back to key if missing.
func T(lang, key string) string {
	mu.RLock()
	defer mu.RUnlock()
	if Translations == nil {
		return key
	}
	if lm, ok := Translations[lang]; ok {
		if v, ok := lm[key]; ok {
			return v
		}
	}
	// fallback to en if available
	if lm, ok := Translations["en"]; ok {
		if v, ok := lm[key]; ok {
			return v
		}
	}
	return key
}

func Available() []string {
	mu.RLock()
	defer mu.RUnlock()
	keys := make([]string, 0, len(Translations))
	for k := range Translations {
		keys = append(keys, k)
	}
	return keys
}
