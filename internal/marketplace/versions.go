package marketplace

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func ParseVersion(v string) (Version, error) {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version format: %s", v)
	}

	var ver Version
	_, err := fmt.Sscanf(v, "%d.%d.%d", &ver.Major, &ver.Minor, &ver.Patch)
	if err != nil {
		return Version{}, err
	}
	return ver, nil
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		return v.Major - other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor - other.Minor
	}
	return v.Patch - other.Patch
}

func (v Version) IsCompatible(other Version) bool {
	return v.Major == other.Major
}

type Dependency struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Required   bool   `json:"required"`
	MinVersion string `json:"min_version,omitempty"`
	MaxVersion string `json:"max_version,omitempty"`
}

type FilterSpec struct {
	Meta         *FilterMeta
	Dependencies []Dependency
	Content      string
}

type VersionManager struct {
	registry *FilterRegistry
}

func NewVersionManager(registry *FilterRegistry) *VersionManager {
	return &VersionManager{registry: registry}
}

func (vm *VersionManager) CheckForUpdates(ctx context.Context, name string, currentVer string) (string, bool, error) {
	current, err := ParseVersion(currentVer)
	if err != nil {
		return "", false, fmt.Errorf("invalid current version: %w", err)
	}

	meta, ok := vm.registry.builtin[name]
	if !ok {
		return "", false, fmt.Errorf("filter not found: %s", name)
	}

	latest, err := ParseVersion(meta.Version)
	if err != nil {
		return "", false, fmt.Errorf("invalid latest version: %w", err)
	}

	if latest.Compare(current) > 0 {
		return latest.String(), true, nil
	}

	return "", false, nil
}

func (vm *VersionManager) ResolveDependencies(ctx context.Context, deps []Dependency) ([]Dependency, error) {
	resolved := make([]Dependency, 0, len(deps))

	for _, dep := range deps {
		if dep.MinVersion != "" {
			current, err := vm.registry.LoadStore()
			if err == nil {
				if installed, ok := current.Installed[dep.Name]; ok {
					currentVer, _ := ParseVersion(installed.Version)
					minVer, _ := ParseVersion(dep.MinVersion)
					if currentVer.Compare(minVer) < 0 {
						return nil, fmt.Errorf("dependency %s requires version >= %s, got %s", dep.Name, dep.MinVersion, installed.Version)
					}
				}
			}
		}
		resolved = append(resolved, dep)
	}

	return resolved, nil
}

type ConflictDetector struct {
	installed map[string]FilterSpec
}

func NewConflictDetector() *ConflictDetector {
	return &ConflictDetector{installed: make(map[string]FilterSpec)}
}

func (cd *ConflictDetector) AddFilter(name string, spec FilterSpec) error {
	if existing, ok := cd.installed[name]; ok {
		ver1, _ := ParseVersion(existing.Meta.Version)
		ver2, _ := ParseVersion(spec.Meta.Version)

		if !ver1.IsCompatible(ver2) {
			return fmt.Errorf("version conflict: %s has incompatible versions %s and %s", name, ver1.String(), ver2.String())
		}
	}
	cd.installed[name] = spec
	return nil
}

func (cd *ConflictDetector) CheckConflicts(deps []Dependency) []string {
	conflicts := make([]string, 0)

	for _, dep := range deps {
		if existing, ok := cd.installed[dep.Name]; ok {
			currentVer, _ := ParseVersion(existing.Meta.Version)
			minVer, _ := ParseVersion(dep.MinVersion)

			if dep.MinVersion != "" && currentVer.Compare(minVer) < 0 {
				conflicts = append(conflicts, fmt.Sprintf("%s: requires %s but installed %s", dep.Name, dep.MinVersion, existing.Meta.Version))
			}
		}
	}

	return conflicts
}

type FilterBundler struct {
	registry *FilterRegistry
}

func NewFilterBundler(registry *FilterRegistry) *FilterBundler {
	return &FilterBundler{registry: registry}
}

type Bundle struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Filters     []string `json:"filters"`
	Author      string   `json:"author"`
	Version     string   `json:"version"`
}

func (fb *FilterBundler) CreateBundle(name string, filterNames []string) (*Bundle, error) {
	installable := make(map[string]bool)
	for _, f := range filterNames {
		installable[f] = true
	}

	store, err := fb.registry.LoadStore()
	if err != nil {
		return nil, err
	}

	var notInstalled []string
	for _, f := range filterNames {
		if _, ok := store.Installed[f]; !ok {
			notInstalled = append(notInstalled, f)
		}
	}

	return &Bundle{
		Name:        name,
		Description: fmt.Sprintf("Bundle with %d filters (%d installed)", len(filterNames), len(filterNames)-len(notInstalled)),
		Filters:     filterNames,
		Author:      "TokMan",
		Version:     "1.0.0",
	}, nil
}

func (fb *FilterBundler) GetBundleFilters(bundle *Bundle) []string {
	return bundle.Filters
}

type FilterAnalytics struct {
	downloads    map[string]int64
	ratings      map[string][]float64
	installTimes map[string]time.Time
}

func NewFilterAnalytics() *FilterAnalytics {
	return &FilterAnalytics{
		downloads:    make(map[string]int64),
		ratings:      make(map[string][]float64),
		installTimes: make(map[string]time.Time),
	}
}

func (fa *FilterAnalytics) RecordDownload(name string) {
	fa.downloads[name]++
}

func (fa *FilterAnalytics) RecordRating(name string, rating float64) {
	fa.ratings[name] = append(fa.ratings[name], rating)
}

func (fa *FilterAnalytics) RecordInstall(name string) {
	fa.installTimes[name] = time.Now()
}

func (fa *FilterAnalytics) GetStats(name string) (int64, float64, int) {
	downloads := fa.downloads[name]

	var avgRating float64
	ratings := fa.ratings[name]
	if len(ratings) > 0 {
		sum := 0.0
		for _, r := range ratings {
			sum += r
		}
		avgRating = sum / float64(len(ratings))
	}

	return downloads, avgRating, len(ratings)
}

type Collection struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Filters     []string `json:"filters"`
	Author      string   `json:"author"`
	Public      bool     `json:"public"`
}

type CollectionManager struct {
	collections map[string]Collection
}

func NewCollectionManager() *CollectionManager {
	return &CollectionManager{
		collections: make(map[string]Collection),
	}
}

func (cm *CollectionManager) Create(name, desc, author string) *Collection {
	coll := Collection{
		Name:        name,
		Description: desc,
		Author:      author,
		Public:      false,
		Filters:     []string{},
	}
	cm.collections[name] = coll
	return &coll
}

func (cm *CollectionManager) AddFilter(name, collName string) error {
	coll, ok := cm.collections[collName]
	if !ok {
		return fmt.Errorf("collection not found: %s", collName)
	}
	coll.Filters = append(coll.Filters, name)
	cm.collections[collName] = coll
	return nil
}

func (cm *CollectionManager) GetFilters(collName string) ([]string, error) {
	coll, ok := cm.collections[collName]
	if !ok {
		return nil, fmt.Errorf("collection not found: %s", collName)
	}
	return coll.Filters, nil
}

type AuthorProfile struct {
	Name        string   `json:"name"`
	Filters     []string `json:"filters"`
	TotalDls    int64    `json:"total_downloads"`
	AvgRating   float64  `json:"avg_rating"`
	JoinDate    string   `json:"join_date"`
	Description string   `json:"description"`
}

func (r *FilterRegistry) GetAuthorProfiles() map[string]*AuthorProfile {
	profiles := make(map[string]*AuthorProfile)

	for _, f := range r.builtin {
		if _, ok := profiles[f.Author]; !ok {
			profiles[f.Author] = &AuthorProfile{
				Name:        f.Author,
				JoinDate:    time.Now().Format("2006-01-02"),
				Description: fmt.Sprintf("Author of %s filters", f.Name),
			}
		}
		profiles[f.Author].Filters = append(profiles[f.Author].Filters, f.Name)
		profiles[f.Author].TotalDls += f.Downloads
	}

	for _, p := range profiles {
		if len(p.Filters) > 0 {
			var totalRating float64
			for _, fn := range p.Filters {
				if f, ok := r.builtin[fn]; ok {
					totalRating += f.Rating
				}
			}
			p.AvgRating = totalRating / float64(len(p.Filters))
		}
	}

	return profiles
}

type FilterModerator struct {
	reportQueue  []FilterReport
	allowedUsers map[string]bool
}

type FilterReport struct {
	Filter   string    `json:"filter"`
	Reason   string    `json:"reason"`
	Reporter string    `json:"reporter"`
	Date     time.Time `json:"date"`
	Status   string    `json:"status"`
}

func NewFilterModerator() *FilterModerator {
	return &FilterModerator{
		reportQueue:  make([]FilterReport, 0),
		allowedUsers: map[string]bool{"admin": true},
	}
}

func (fm *FilterModerator) ReportFilter(filter, reason, reporter string) error {
	report := FilterReport{
		Filter:   filter,
		Reason:   reason,
		Reporter: reporter,
		Date:     time.Now(),
		Status:   "pending",
	}
	fm.reportQueue = append(fm.reportQueue, report)
	return nil
}

func (fm *FilterModerator) GetPendingReports() []FilterReport {
	return fm.reportQueue
}

func (fm *FilterModerator) ResolveReport(idx int, decision string) error {
	if idx < 0 || idx >= len(fm.reportQueue) {
		return fmt.Errorf("invalid report index")
	}
	fm.reportQueue[idx].Status = decision
	return nil
}

func (r *FilterRegistry) ListByCategory() map[string][]*FilterMeta {
	byCat := make(map[string][]*FilterMeta)
	for _, f := range r.builtin {
		byCat[f.Category] = append(byCat[f.Category], f)
	}

	for cat := range byCat {
		sort.Slice(byCat[cat], func(i, j int) bool {
			return byCat[cat][i].Downloads > byCat[cat][j].Downloads
		})
	}

	return byCat
}

func (r *FilterRegistry) ListByAuthor() map[string][]*FilterMeta {
	byAuthor := make(map[string][]*FilterMeta)
	for _, f := range r.builtin {
		byAuthor[f.Author] = append(byAuthor[f.Author], f)
	}
	return byAuthor
}

func (r *FilterRegistry) GetFilterVersions(name string) []Version {
	versions := make([]Version, 0)
	for i := 1; i <= 10; i++ {
		v := fmt.Sprintf("1.%d.0", i)
		ver, err := ParseVersion(v)
		if err == nil {
			versions = append(versions, ver)
		}
	}
	return versions
}
