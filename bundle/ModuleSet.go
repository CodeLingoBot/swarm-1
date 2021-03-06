package bundle

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"github.com/mrcrowl/swarm/config"
	"github.com/mrcrowl/swarm/monitor"
	"github.com/mrcrowl/swarm/source"
	"sync"
)

// ModuleSet is
type ModuleSet struct {
	modules       []*Module
	mutex         *sync.Mutex
	runtimeConfig *config.RuntimeConfig
}

// CreateModuleSet creates a ModuleSet from a list of NormalisedModuleDescriptions
func CreateModuleSet(ws *source.Workspace, moduleDescriptions []*config.NormalisedModuleDescription, runtimeConfig *config.RuntimeConfig) *ModuleSet {
	modules := make([]*Module, len(moduleDescriptions))
	runtimeConfig.SetPathInterpolationValues(ws.ReadInterpolationValues(runtimeConfig))
	for i, descr := range moduleDescriptions {
		modules[i] = NewModule(ws, descr, runtimeConfig)
	}

	set := &ModuleSet{
		modules: modules,
		mutex:   &sync.Mutex{},
	}

	for _, mod := range set.modules {
		mod.attachExcludedModules(set)
	}

	set.sort()

	for _, mod := range set.modules {
		mod.buildInitialFileSet()
	}

	return set
}

// NotifyChanges absorbs an EventChangeset, triggering artefacts to be recompiled, when necessary
func (set *ModuleSet) NotifyChanges(changes *monitor.EventChangeset) {
	set.mutex.Lock()
	if changes != nil {
		for _, mod := range set.modules {
			mod.absorbChanges(changes)
		}
	}

	// TODO: could this be parallelised?
	for _, mod := range set.modules {
		if mod.dirty() {
			mod.generateBundle()
			if changes != nil {
				changes.FlagDidBundle()
			}
		}
	}
	set.mutex.Unlock()
}

// FindFileByPath finds and returns a file by path name
func (set *ModuleSet) FindFileByPath(path string) *source.File {
	for _, mod := range set.modules {
		if file := mod.GetFileByPath(path); file != nil {
			return file
		}
	}
	return nil
}

func (set *ModuleSet) getModule(name string) *Module {
	for _, mod := range set.modules {
		if mod.description.Name == name {
			return mod
		}
	}
	log.Panicf("getModule: could not find module with name '%s'", name)
	return nil
}

// names gets the module names (sorted topographical, assuming CreateModuleSet has finished!)
func (set *ModuleSet) names() []string {
	names := make([]string, len(set.modules))
	for i, module := range set.modules {
		names[i] = module.Name()
	}
	return names
}

func (set *ModuleSet) linksMap() map[string][]string {
	allLinks := make(map[string][]string)
	for _, module := range set.modules {
		allLinks[module.Name()] = module.links()
	}
	return allLinks
}

func (set *ModuleSet) sort() {
	sortedModules := make([]*Module, len(set.modules))
	graph := source.NewIDGraph(set.linksMap())
	sortedNames := graph.SortTopologically(set.names())
	for i, name := range sortedNames {
		sortedModules[i] = set.getModule(name)
	}
	set.modules = sortedModules
}

// GenerateHTTPHandlers creates http.HandlerFunc's that will return the bundled javascript
func (set *ModuleSet) GenerateHTTPHandlers() map[string]http.HandlerFunc {
	createJSHandler := func(module *Module) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, module.bundledJavascript)
			io.WriteString(w, fmt.Sprintf("//# sourceMappingURL=%s", module.SourceMapName()))
		}
	}

	createMapHandler := func(module *Module) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			sourceMap := module.bundledSourcemap
			sourceMap = strings.Replace(sourceMap, `["BaseController.ts"]`, `["ui/base/BaseController.ts"]`, 1)
			io.WriteString(w, sourceMap)
		}
	}

	handlers := map[string]http.HandlerFunc{}
	for _, module := range set.modules {
		entryPoint := module.PrimaryEntryPoint()
		handlers["/"+entryPoint+".js"] = createJSHandler(module)
		if set.runtimeConfig.SourceMapsEnabled() {
			handlers["/"+entryPoint+".js.map"] = createMapHandler(module)
		}
	}
	return handlers
}
