package models

// RepoStatus représente l'état brut du dépôt Git fourni par le backend Python.
type RepoStatus struct {
	IsGitRepo     bool   `json:"is_git_repo"`
	RepoPath      string `json:"repo_path"`
	Branch        string `json:"branch"`
	StagedCount   int    `json:"staged_count"`
	UnstagedCount int    `json:"unstaged_count"`
	Error         string `json:"error,omitempty"`
}

// DiffFile représente un fichier modifié analysé et assaini par le backend.
type DiffFile struct {
	Path             string `json:"path"`
	Status           string `json:"status"` // M, A, D, R, etc.
	Added            int    `json:"added"`
	Removed          int    `json:"removed"`
	Binary           bool   `json:"binary"`
	Patch            string `json:"patch"`
	HasSecretsMasked bool   `json:"has_secrets_masked"`
}

// DiffResponse contient la liste des fichiers en attente et les métadonnées.
type DiffResponse struct {
	Success     bool       `json:"success"`
	Error       string     `json:"error,omitempty"`
	Files       []DiffFile `json:"files"`
	StagedCount int        `json:"staged_count"`
}

// CommitProposal représente la suggestion de commit au format Conventional Commits.
type CommitProposal struct {
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	Type         string `json:"type"`
	Scope        string `json:"scope,omitempty"`
	Subject      string `json:"subject"`
	Body         string `json:"body,omitempty"`
	RawFormatted string `json:"raw_formatted"`
}

// ActionResult représente le retour d'une action exécutée par Python (commit, stage, push, save).
type ActionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Sha     string `json:"sha,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ReleaseNotesResponse contient le texte Markdown synthétisé par le LLM.
type ReleaseNotesResponse struct {
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
	FromTag     string `json:"from_tag"`
	ToTag       string `json:"to_tag"`
	Markdown    string `json:"markdown"`
	CommitCount int    `json:"commit_count"`
}

// ConfigResponse contient les paramètres lus ou écrits dans le .env par Python.
type ConfigResponse struct {
	Success       bool    `json:"success"`
	Error         string  `json:"error,omitempty"`
	OllamaBaseURL string  `json:"ollama_base_url"`
	OllamaModel   string  `json:"ollama_model"`
	TimeoutS      float64 `json:"timeout_s"`
	Language      string  `json:"language"`
}

// PingResponse contient le diagnostic de joignabilité du serveur Ollama IUT.
type PingResponse struct {
	Success         bool     `json:"success"`
	Reachable       bool     `json:"reachable"`
	LatencyMs       int64    `json:"latency_ms"`
	InstalledModels []string `json:"installed_models"`
	Error           string   `json:"error,omitempty"`
}
