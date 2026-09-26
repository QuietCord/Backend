package util

import "testing"

func TestMergeSettingsBundles_pluginConflict(t *testing.T) {
	base := &SettingsBundle{
		Settings: map[string]any{
			"plugins": map[string]any{
				"QuietPerformance": map[string]any{"profile": "balanced"},
			},
		},
		QuickCss: "a",
	}
	local := &SettingsBundle{
		Settings: map[string]any{
			"plugins": map[string]any{
				"QuietPerformance": map[string]any{"profile": "minimal"},
			},
		},
		QuickCss: "a",
	}
	remote := &SettingsBundle{
		Settings: map[string]any{
			"plugins": map[string]any{
				"QuietPerformance": map[string]any{"profile": "performance"},
			},
		},
		QuickCss: "b",
	}

	res, err := MergeSettingsBundles(base, local, remote)
	if err != nil {
		t.Fatal(err)
	}
	if res.Complete {
		t.Fatalf("expected conflicts")
	}
	if len(res.Conflicts) < 2 {
		t.Fatalf("expected plugin + quickCss conflicts, got %d", len(res.Conflicts))
	}
}

func TestMergeSettingsBundles_nonConflict(t *testing.T) {
	base := &SettingsBundle{
		Settings: map[string]any{
			"plugins": map[string]any{
				"A": map[string]any{"x": 1},
				"B": map[string]any{"y": 1},
			},
		},
	}
	local := &SettingsBundle{
		Settings: map[string]any{
			"plugins": map[string]any{
				"A": map[string]any{"x": 2},
				"B": map[string]any{"y": 1},
			},
		},
	}
	remote := &SettingsBundle{
		Settings: map[string]any{
			"plugins": map[string]any{
				"A": map[string]any{"x": 1},
				"B": map[string]any{"y": 2},
			},
		},
	}

	res, err := MergeSettingsBundles(base, local, remote)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Complete {
		t.Fatalf("expected complete merge, conflicts: %+v", res.Conflicts)
	}
	plugins := res.Merged.Settings["plugins"].(map[string]any)
	if plugins["A"].(map[string]any)["x"] != float64(2) {
		t.Fatalf("A should be local change")
	}
	if plugins["B"].(map[string]any)["y"] != float64(2) {
		t.Fatalf("B should be remote change")
	}
}
