package chat

import (
	"testing"
)

func TestSessionConfigManagerCreateConfig(t *testing.T) {
	scm := NewSessionConfigManager()

	config := map[string]interface{}{
		"model":       "gpt-4",
		"temperature": 0.7,
	}

	cfg, err := scm.CreateConfig("session-1", config)
	if err != nil {
		t.Errorf("CreateConfig failed: %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("Expected version 1")
	}
}

func TestSessionConfigManagerGetConfig(t *testing.T) {
	scm := NewSessionConfigManager()

	config := map[string]interface{}{
		"model": "gpt-4",
	}

	scm.CreateConfig("session-1", config)

	cfg, err := scm.GetConfig("session-1")
	if err != nil {
		t.Errorf("GetConfig failed: %v", err)
	}

	if cfg.Config["model"] != "gpt-4" {
		t.Errorf("Expected model gpt-4")
	}
}

func TestSessionConfigManagerUpdateConfig(t *testing.T) {
	scm := NewSessionConfigManager()

	config := map[string]interface{}{
		"model": "gpt-3.5",
	}

	scm.CreateConfig("session-1", config)

	newConfig := map[string]interface{}{
		"model": "gpt-4",
	}

	cfg, err := scm.UpdateConfig("session-1", newConfig)
	if err != nil {
		t.Errorf("UpdateConfig failed: %v", err)
	}

	if cfg.Version != 2 {
		t.Errorf("Expected version 2")
	}

	if cfg.Config["model"] != "gpt-4" {
		t.Errorf("Expected updated model")
	}
}

func TestSessionConfigManagerMergeConfig(t *testing.T) {
	scm := NewSessionConfigManager()

	config := map[string]interface{}{
		"model":       "gpt-4",
		"temperature": 0.7,
	}

	scm.CreateConfig("session-1", config)

	updates := map[string]interface{}{
		"temperature": 0.9,
	}

	cfg, err := scm.MergeConfig("session-1", updates)
	if err != nil {
		t.Errorf("MergeConfig failed: %v", err)
	}

	if cfg.Config["temperature"] != 0.9 {
		t.Errorf("Expected updated temperature")
	}

	if cfg.Config["model"] != "gpt-4" {
		t.Errorf("Expected preserved model")
	}
}

func TestSessionConfigManagerGetConfigVersion(t *testing.T) {
	scm := NewSessionConfigManager()

	config := map[string]interface{}{
		"model": "gpt-3.5",
	}

	scm.CreateConfig("session-1", config)

	newConfig := map[string]interface{}{
		"model": "gpt-4",
	}

	scm.UpdateConfig("session-1", newConfig)

	cfg, err := scm.GetConfigVersion("session-1", 1)
	if err != nil {
		t.Errorf("GetConfigVersion failed: %v", err)
	}

	if cfg.Config["model"] != "gpt-3.5" {
		t.Errorf("Expected first version model")
	}
}

func TestSessionConfigManagerExportConfig(t *testing.T) {
	scm := NewSessionConfigManager()

	config := map[string]interface{}{
		"model": "gpt-4",
	}

	scm.CreateConfig("session-1", config)

	jsonStr, err := scm.ExportConfig("session-1")
	if err != nil {
		t.Errorf("ExportConfig failed: %v", err)
	}

	if jsonStr == "" {
		t.Errorf("Expected non-empty JSON")
	}
}

func TestSessionConfigManagerImportConfig(t *testing.T) {
	scm := NewSessionConfigManager()

	config := map[string]interface{}{
		"model": "gpt-3.5",
	}

	scm.CreateConfig("session-1", config)

	jsonStr := `{"model": "gpt-4", "temperature": 0.8}`

	cfg, err := scm.ImportConfig("session-1", jsonStr)
	if err != nil {
		t.Errorf("ImportConfig failed: %v", err)
	}

	if cfg.Config["model"] != "gpt-4" {
		t.Errorf("Expected imported model")
	}
}

func TestSessionConfigManagerGetConfigValue(t *testing.T) {
	scm := NewSessionConfigManager()

	config := map[string]interface{}{
		"model":       "gpt-4",
		"temperature": 0.7,
	}

	scm.CreateConfig("session-1", config)

	value, exists := scm.GetConfigValue("session-1", "model")
	if !exists || value != "gpt-4" {
		t.Errorf("Expected model value")
	}
}

func TestSessionConfigManagerSetConfigValue(t *testing.T) {
	scm := NewSessionConfigManager()

	config := map[string]interface{}{
		"model": "gpt-3.5",
	}

	scm.CreateConfig("session-1", config)

	cfg, err := scm.SetConfigValue("session-1", "model", "gpt-4")
	if err != nil {
		t.Errorf("SetConfigValue failed: %v", err)
	}

	if cfg.Config["model"] != "gpt-4" {
		t.Errorf("Expected updated value")
	}
}

func TestInheritanceManagerCreateInheritance(t *testing.T) {
	scm := NewSessionConfigManager()
	im := NewInheritanceManager(scm)

	config := map[string]interface{}{
		"model": "gpt-4",
	}

	scm.CreateConfig("session-1", config)
	scm.CreateConfig("session-2", config)

	err := im.CreateInheritance("session-1", "session-2")
	if err != nil {
		t.Errorf("CreateInheritance failed: %v", err)
	}
}

func TestInheritanceManagerGetEffectiveConfig(t *testing.T) {
	scm := NewSessionConfigManager()
	im := NewInheritanceManager(scm)

	parentConfig := map[string]interface{}{
		"model":       "gpt-4",
		"temperature": 0.7,
	}

	scm.CreateConfig("session-1", parentConfig)
	scm.CreateConfig("session-2", map[string]interface{}{})

	im.CreateInheritance("session-1", "session-2")

	im.SetOverride("session-2", "temperature", 0.9)

	effectiveConfig, err := im.GetEffectiveConfig("session-2")
	if err != nil {
		t.Errorf("GetEffectiveConfig failed: %v", err)
	}

	if effectiveConfig["model"] != "gpt-4" {
		t.Errorf("Expected inherited model")
	}

	if effectiveConfig["temperature"] != 0.9 {
		t.Errorf("Expected overridden temperature")
	}
}

func TestSessionConfigManagerGetConfigVersions(t *testing.T) {
	scm := NewSessionConfigManager()

	config := map[string]interface{}{
		"model": "gpt-3.5",
	}

	scm.CreateConfig("session-1", config)

	for i := 0; i < 3; i++ {
		newConfig := map[string]interface{}{
			"model": "gpt-4",
		}
		scm.UpdateConfig("session-1", newConfig)
	}

	versions := scm.GetConfigVersions("session-1")

	if len(versions) < 1 {
		t.Errorf("Expected versions")
	}
}

func TestSessionConfigManagerDeleteConfig(t *testing.T) {
	scm := NewSessionConfigManager()

	config := map[string]interface{}{
		"model": "gpt-4",
	}

	scm.CreateConfig("session-1", config)

	err := scm.DeleteConfig("session-1")
	if err != nil {
		t.Errorf("DeleteConfig failed: %v", err)
	}

	_, err = scm.GetConfig("session-1")
	if err == nil {
		t.Errorf("Expected error after deletion")
	}
}

func TestSessionConfigManagerGetStatistics(t *testing.T) {
	scm := NewSessionConfigManager()

	config := map[string]interface{}{
		"model": "gpt-4",
	}

	scm.CreateConfig("session-1", config)
	scm.CreateConfig("session-2", config)

	stats := scm.GetStatistics()

	if stats == nil {
		t.Errorf("Expected statistics")
	}

	if totalConfigs, ok := stats["total_configs"].(int); !ok || totalConfigs != 2 {
		t.Errorf("Expected 2 configs")
	}
}

func BenchmarkCreateConfig(b *testing.B) {
	scm := NewSessionConfigManager()

	config := map[string]interface{}{
		"model": "gpt-4",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scm.CreateConfig("session-"+string(rune(i)), config)
	}
}

func BenchmarkGetConfig(b *testing.B) {
	scm := NewSessionConfigManager()

	config := map[string]interface{}{
		"model": "gpt-4",
	}

	scm.CreateConfig("session-1", config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scm.GetConfig("session-1")
	}
}

func BenchmarkMergeConfig(b *testing.B) {
	scm := NewSessionConfigManager()

	config := map[string]interface{}{
		"model":       "gpt-4",
		"temperature": 0.7,
	}

	scm.CreateConfig("session-1", config)

	updates := map[string]interface{}{
		"temperature": 0.9,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scm.MergeConfig("session-1", updates)
	}
}

