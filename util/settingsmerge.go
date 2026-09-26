package util

import (
	"encoding/json"
	"fmt"
)

type SettingsBundle struct {
	Settings map[string]any `json:"settings"`
	QuickCss string         `json:"quickCss"`
}

type MergeConflict struct {
	Namespace string `json:"namespace"`
	Yours     any    `json:"yours"`
	Theirs    any    `json:"theirs"`
}

type MergeResult struct {
	Merged    *SettingsBundle  `json:"merged,omitempty"`
	Conflicts []MergeConflict  `json:"conflicts"`
	Complete  bool             `json:"complete"`
}

func ParseSettingsBundle(raw []byte) (*SettingsBundle, error) {
	var b SettingsBundle
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, err
	}
	if b.Settings == nil {
		b.Settings = map[string]any{}
	}
	return &b, nil
}

func jsonEqual(a, b any) bool {
	ab, errA := json.Marshal(a)
	bb, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return false
	}
	return string(ab) == string(bb)
}

func pickMerged(base, local, remote any) (merged any, conflict bool, yours, theirs any) {
	if jsonEqual(local, remote) {
		return local, false, nil, nil
	}
	if jsonEqual(base, local) {
		return remote, false, nil, nil
	}
	if jsonEqual(base, remote) {
		return local, false, nil, nil
	}
	return nil, true, local, remote
}

func pluginNamespaces(settings map[string]any) map[string]any {
	plugins, _ := settings["plugins"].(map[string]any)
	if plugins == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(plugins))
	for name, val := range plugins {
		out["plugins."+name] = val
	}
	return out
}

func applyPluginMerge(mergedSettings map[string]any, namespace string, value any) {
	if len(namespace) < 9 || namespace[:8] != "plugins." {
		return
	}
	name := namespace[8:]
	plugins, ok := mergedSettings["plugins"].(map[string]any)
	if !ok || plugins == nil {
		plugins = map[string]any{}
		mergedSettings["plugins"] = plugins
	}
	plugins[name] = value
}

// MergeSettingsBundles performs a 3-way merge using base when provided; missing base treats empty values as base.
func MergeSettingsBundles(base, local, remote *SettingsBundle) (*MergeResult, error) {
	if local == nil || remote == nil {
		return nil, fmt.Errorf("local and remote bundles required")
	}

	result := &MergeResult{
		Conflicts: []MergeConflict{},
		Complete:  true,
		Merged: &SettingsBundle{
			Settings: map[string]any{},
			QuickCss: "",
		},
	}

	baseSettings := map[string]any{}
	baseQuickCss := ""
	if base != nil {
		baseSettings = base.Settings
		baseQuickCss = base.QuickCss
	}

	localPlugins := pluginNamespaces(local.Settings)
	remotePlugins := pluginNamespaces(remote.Settings)
	basePlugins := pluginNamespaces(baseSettings)

	allPluginNs := map[string]struct{}{}
	for k := range localPlugins {
		allPluginNs[k] = struct{}{}
	}
	for k := range remotePlugins {
		allPluginNs[k] = struct{}{}
	}

	for ns := range allPluginNs {
		l := localPlugins[ns]
		r := remotePlugins[ns]
		b := basePlugins[ns]
		m, conflict, yours, theirs := pickMerged(b, l, r)
		if conflict {
			result.Complete = false
			result.Conflicts = append(result.Conflicts, MergeConflict{
				Namespace: ns,
				Yours:     yours,
				Theirs:    theirs,
			})
			continue
		}
		if m != nil {
			applyPluginMerge(result.Merged.Settings, ns, m)
		}
	}

	// Top-level settings keys (except plugins) — one namespace per key.
	skip := map[string]struct{}{"plugins": {}}
	for key := range local.Settings {
		if _, ok := skip[key]; !ok {
			skip[key] = struct{}{}
		}
	}
	for key := range remote.Settings {
		if _, ok := skip[key]; !ok {
			skip[key] = struct{}{}
		}
	}
	for key := range skip {
		if key == "plugins" {
			continue
		}
		l := local.Settings[key]
		r := remote.Settings[key]
		b := baseSettings[key]
		m, conflict, yours, theirs := pickMerged(b, l, r)
		if conflict {
			result.Complete = false
			result.Conflicts = append(result.Conflicts, MergeConflict{
				Namespace: "settings." + key,
				Yours:     yours,
				Theirs:    theirs,
			})
			continue
		}
		if m != nil {
			result.Merged.Settings[key] = m
		}
	}

	m, conflict, yours, theirs := pickMerged(baseQuickCss, local.QuickCss, remote.QuickCss)
	if conflict {
		result.Complete = false
		result.Conflicts = append(result.Conflicts, MergeConflict{
			Namespace: "quickCss",
			Yours:     yours,
			Theirs:    theirs,
		})
	} else if s, ok := m.(string); ok {
		result.Merged.QuickCss = s
	} else if m != nil {
		result.Merged.QuickCss = fmt.Sprint(m)
	}

	return result, nil
}
