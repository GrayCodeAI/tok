package selfupdate

type Updater struct {
	CurrentVersion string `json:"current_version"`
	RepoURL        string `json:"repo_url"`
}

func NewUpdater(currentVersion string) *Updater {
	return &Updater{
		CurrentVersion: currentVersion,
		RepoURL:        "https://github.com/GrayCodeAI/tokman",
	}
}

func (u *Updater) CheckForUpdates() (string, bool) {
	return u.CurrentVersion, false
}

func (u *Updater) GetUpdateURL() string {
	return u.RepoURL + "/releases/latest"
}
