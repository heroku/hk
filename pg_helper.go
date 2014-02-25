package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/bgentry/heroku-go"
	"github.com/heroku/hk/postgresql"
	"github.com/mgutz/ansi"
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

// addon options that Heroku Postgres needs resolved from database names to
// full postgres URLs
var hpgOptNames = []string{"fork", "follow", "rollback"}

// resolve addon options whose names are in hpgOptNames into their full URLs
func hpgAddonOptResolve(opts *map[string]string, appEnv map[string]string) error {
	if opts != nil {
		for _, k := range hpgOptNames {
			val, ok := (*opts)[k]
			if ok && !strings.HasPrefix(val, "postgres://") {
				envName := dbNameToPgEnv(val)
				url, exists := appEnv[envName]
				if !exists {
					return fmt.Errorf("could not resolve %s option %q to a %s addon", k, val, hpgAddonName())
				}
				(*opts)[k] = url
			}
		}
	}
	return nil
}

func pgEnvToDBName(key string) string {
	return strings.ToLower(strings.Replace(strings.TrimSuffix(key, "_URL"), "_", "-", -1))
}

func dbNameToPgEnv(name string) string {
	return ensureSuffix(ensurePrefix(
		strings.ToUpper(strings.Replace(name, "-", "_", -1)),
		strings.ToUpper(strings.Replace(hpgAddonName()+"_", "-", "_", -1)),
	), "_URL")
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

func mustGetDBInfoAndAddonMap(addonName, appname string) (postgresql.DB, postgresql.DBInfo, pgAddonMap) {
	// list all addons
	addons, err := client.AddonList(appname, nil)
	must(err)

	// locate specific addon
	var addon *heroku.Addon
	for i := range addons {
		if addons[i].Name == addonName {
			addon = &addons[i]
			break
		}
	}
	if addon == nil {
		printFatal("addon %s not found", addonName)
	}

	// fetch app's config concurrently in case we need to resolve DB names
	var appConf map[string]string
	confch := make(chan map[string]string, 1)
	errch := make(chan error, 1)
	go func(appname string) {
		if config, err := client.ConfigVarInfo(appname); err != nil {
			errch <- err
		} else {
			confch <- config
		}
	}(appname)

	db := pgclient.NewDB(addon.ProviderId, addon.Plan.Name)
	dbi, err := db.Info()
	must(err)

	select {
	case err := <-errch:
		printFatal(err.Error())
	case appConf = <-confch:
	}

	addonMap := newPgAddonMap(addons, appConf)
	return db, dbi, addonMap
}

type fullDBInfo struct {
	Name     string
	DBInfo   postgresql.DBInfo
	Parent   *fullDBInfo
	Children []*fullDBInfo
}

func (f *fullDBInfo) MaintenanceString() string {
	valstr, _ := f.DBInfo.Info.GetString("Maintenance")
	if valstr != "" && valstr != "not required" {
		return " " + ansi.Color("!!", "red+b") + ansi.ColorCode("reset")
	}
	return ""
}

// fullDBInfosByName implements sort.Interface for []*fullDBInfo based on the
// Name field.
type fullDBInfosByName []*fullDBInfo

func (f fullDBInfosByName) Len() int           { return len(f) }
func (f fullDBInfosByName) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f fullDBInfosByName) Less(i, j int) bool { return f[i].Name < f[j].Name }

func sortedDBInfoTree(dbinfos []*fullDBInfo, addonMap pgAddonMap) (result []*fullDBInfo) {
	sort.Sort(fullDBInfosByName(dbinfos))
	// get all children organized under their parents
	for _, info := range dbinfos {
		parentName := getResolvedInfoValue(info.DBInfo, "Forked From", &addonMap)
		if parentName == "" {
			parentName = getResolvedInfoValue(info.DBInfo, "Following", &addonMap)
		}
		if parentName != "" {
			for _, parent := range dbinfos {
				if parent.Name == parentName {
					parent.Children = append(parent.Children, info)
					info.Parent = parent
					break
				}
			}
		}
	}
	// keep items that have no parent
	for i := 0; i < len(dbinfos); i++ {
		if dbinfos[i].Parent == nil {
			result = append(result, dbinfos[i])
		}
	}
	return
}

func printDBTree(w io.Writer, dbinfos []*fullDBInfo, addonMap pgAddonMap) {
	for _, info := range dbinfos {
		name := info.Name
		if info.Parent != nil {
			name = printTreeElements(info) + name
		}
		dburlMarker := "  "
		if stringsIndex(addonMap.addonToEnv[info.Name], "DATABASE_URL") != -1 {
			dburlMarker = "* "
		}
		status, _ := info.DBInfo.Info.GetString("Status")
		listRec(w,
			dburlMarker+name,
			info.DBInfo.Plan,
			status+info.MaintenanceString(),
			info.DBInfo.NumConnections,
		)
		if len(info.Children) > 0 {
			printDBTree(w, info.Children, addonMap)
		}
	}
}

const (
	treeMiddleBranch = "├─"
	treeLastBranch   = "└─"
	treeForkSymbol   = " ─┤"
	treeFollowSymbol = "──>"
	treeIndentMiddle = "│     "
	treeIndentLast   = "      "
)

func printTreeElements(info *fullDBInfo) string {
	if info == nil || info.Parent == nil {
		return ""
	}
	prefix := ""
	for curInfo, p := info, info.Parent; p != nil; curInfo, p = p, p.Parent {
		prefix = treePrefix(p, curInfo, prefix == "") + prefix
	}
	symbol := treeForkSymbol
	if info.DBInfo.IsFollower() {
		symbol = treeFollowSymbol
	}
	return prefix + symbol + " "
}

func treePrefix(parent, info *fullDBInfo, firstLevel bool) string {
	if parent.Children[len(parent.Children)-1] == info {
		// this is the parent's last child
		if firstLevel {
			return treeLastBranch
		}
		return treeIndentLast
	}
	if firstLevel {
		return treeMiddleBranch
	}
	return treeIndentMiddle
}

func getResolvedInfoValue(dbi postgresql.DBInfo, key string, addonMap *pgAddonMap) string {
	val, resolve := dbi.Info.GetString(key)
	if val != "" && resolve {
		if addonName, ok := addonMap.FindAddonFromValue(val); ok {
			return addonName
		}
	}
	return val
}
