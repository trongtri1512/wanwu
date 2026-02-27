package i18n

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func loadJsonlTextConfigs(i18nJsonlPath string) ([]*textConfig, error) {
	// open jsonl
	f, err := os.Open(i18nJsonlPath)
	if err != nil {
		return nil, fmt.Errorf("i18n open %v err: %v", i18nJsonlPath, err)
	}
	defer func() { _ = f.Close() }()
	// load jsonl
	var ret []*textConfig
	var line int
	decoder := json.NewDecoder(f)
	for {
		line++
		var textCfg textConfig
		if err := decoder.Decode(&textCfg); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("i18n load %v line %v err: %v", i18nJsonlPath, line, err)
		}
		ret = append(ret, &textCfg)
	}
	return ret, nil
}
