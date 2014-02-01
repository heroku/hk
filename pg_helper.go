package main

import (
	"os"
	"strings"

	"github.com/bgentry/heroku-go"
)

// the names of heroku postgres addons vary in dev environments
func hpgAddonName() string {
	if e := os.Getenv("SHOGUN"); e != "" {
		return "shogun-" + e
	}
	if e := os.Getenv("HEROKU_POSTGRESQL_ADDON_NAME"); e != "" {
		return e
	}
	return "heroku-postgresql"
}

func pgEnvToDBName(key string) string {
	return strings.ToLower(strings.Replace(strings.TrimSuffix(key, "_URL"), "_", "-", -1))
}

func dbNameToPgEnv(name string) string {
	return ensurePrefix(
		strings.ToUpper(strings.Replace(name, "-", "_", -1)),
		strings.ToUpper(strings.Replace(hpgAddonName()+"_", "-", "_", -1)),
	) + "_URL"
}

type pgAddonMap struct {
	addonToEnv map[string][]string
	appConf    map[string]string
}

func (p *pgAddonMap) FindAddonFromValue(value string) (key string, ok bool) {
	for addonName, envs := range p.addonToEnv {
		for _, e := range envs {
			if p.appConf[e] == value {
				return addonName, true
			}
		}
	}
	return "", false
}

func (p *pgAddonMap) FindEnvsFromValue(value string) []string {
	addonName, ok := p.FindAddonFromValue(value)
	if !ok {
		return []string{}
	}
	return p.addonToEnv[addonName]
}

func newPgAddonMap(addons []heroku.Addon, appConf map[string]string) pgAddonMap {
	m := make(map[string][]string)
	for _, addon := range addons {
		if strings.HasPrefix(addon.Name, hpgAddonName()+"-") {
			if len(addon.ConfigVars) > 0 {
				m[addon.Name] = addon.ConfigVars
				includesDbURL := false
				for _, k := range addon.ConfigVars {
					if k == "DATABASE_URL" {
						includesDbURL = true
					}
				}
				// add DATABASE_URL if it's not already included and the values match
				if !includesDbURL && appConf["DATABASE_URL"] == appConf[addon.ConfigVars[0]] {
					m[addon.Name] = append([]string{"DATABASE_URL"}, m[addon.Name]...)
				}
			}
		}
	}
	return pgAddonMap{m, appConf}
}
