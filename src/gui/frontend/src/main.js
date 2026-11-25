// ==========================================================================
// GITELLIGENCE GUI — LOGIQUE FRONTEND INTERACTIVE (WAILS v2)
// ==========================================================================

let appApi = null;
let currentDiffFiles = [];
let selectedDiffIndex = 0;
let eventListenersInitialized = false;
let rnAllCommits = [];
let rnSelectedStartSha = "";
let rnSelectedEndSha = "";
let dockerPollTimer = null;

// Initialisation au chargement du DOM ou immédiate si déjà chargé
function boot() {
    initNavigation();
    initThemeAndAccents();
    initEventListeners();
    initWailsConnection();
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', boot);
} else {
    boot();
}

// --------------------------------------------------------------------------
// 1. GESTION DE LA NAVIGATION PAR ONGLETS
// --------------------------------------------------------------------------
function resetScrollPositions(targetView = null) {
    const containers = ['.content-stage', '.main-wrapper', '.app-layout'];
    containers.forEach(selector => {
        const el = document.querySelector(selector);
        if (el) {
            el.scrollTop = 0;
            el.scrollLeft = 0;
        }
    });
    if (targetView) {
        targetView.scrollTop = 0;
        targetView.scrollLeft = 0;
    }
    if (document.documentElement) {
        document.documentElement.scrollTop = 0;
        document.documentElement.scrollLeft = 0;
    }
    if (document.body) {
        document.body.scrollTop = 0;
        document.body.scrollLeft = 0;
    }
    window.scrollTo(0, 0);
}

function initNavigation() {
    const navButtons = document.querySelectorAll('.nav-item');
    const tabViews = document.querySelectorAll('.tab-view');
    const crumbActive = document.getElementById('crumb-active');

    navButtons.forEach(btn => {
        btn.addEventListener('click', () => {
            const tabName = btn.dataset.tab;
            if (!tabName) return;

            const targetView = document.getElementById(`view-${tabName}`);

            // Réinitialise les positions de scroll avant et après bascule pour
            // éviter tout blocage de ScrollableArea sous WebKitGTK
            resetScrollPositions(targetView);

            navButtons.forEach(b => b.classList.remove('active'));
            tabViews.forEach(v => v.classList.remove('active'));

            btn.classList.add('active');
            if (targetView) {
                targetView.classList.add('active');
            }

            // Mise à jour du fil d'Ariane
            const label = btn.querySelector('.nav-label')?.textContent || tabName;
            if (crumbActive) {
                crumbActive.textContent = label;
            }

            resetScrollPositions(targetView);
            requestAnimationFrame(() => resetScrollPositions(targetView));

            // Déclenchement des rafraîchissements ciblés à l'affichage
            onTabActivated(tabName);
        });
    });

    // Raccourcis clavier (Ctrl+1 à Ctrl+5)
    window.addEventListener('keydown', (e) => {
        if (e.ctrlKey && !e.shiftKey && !e.altKey) {
            const num = parseInt(e.key, 10);
            if (num >= 1 && num <= navButtons.length) {
                e.preventDefault();
                navButtons[num - 1].click();
            }
        }
    });

    // Sous-onglets Release Notes (Rendu Formaté vs Brut)
    const rnSubtabBtns = document.querySelectorAll('.subtab-btn');
    rnSubtabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            rnSubtabBtns.forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            const subtab = btn.dataset.subtab;
            document.getElementById('rn-subtab-preview').classList.toggle('active', subtab === 'rn-preview');
            document.getElementById('rn-subtab-raw').classList.toggle('active', subtab === 'rn-raw');
        });
    });
}

function onTabActivated(tabName) {
    if (tabName === 'dashboard') {
        refreshDashboard();
    } else if (tabName === 'diff') {
        refreshDiffs();
    } else if (tabName === 'release-notes') {
        loadReleaseNotesCommits();
    } else if (tabName === 'config') {
        loadConfig();
        const autoRefreshChecked = document.getElementById('cfg-docker-log-autorefresh')?.checked ?? false;
        updateDockerStatusAndLogs(autoRefreshChecked);
        if (autoRefreshChecked) {
            startDockerPolling();
        }
    }

    if (tabName !== 'config') {
        stopDockerPolling();
    }
}

// --------------------------------------------------------------------------
// 2. THÈME ET COULEURS D'ACCENTUATION
// --------------------------------------------------------------------------
function initThemeAndAccents() {
    const body = document.body;
    const themeBtn = document.getElementById('theme-toggle');
    const accentBtns = document.querySelectorAll('.accent-btn');

    // Restauration du thème sauvegardé
    const savedTheme = localStorage.getItem('gitelligence_theme') || 'dark';
    if (savedTheme === 'light') {
        body.classList.remove('theme-dark');
        body.classList.add('theme-light');
        if (themeBtn) themeBtn.textContent = '☀️';
    } else {
        body.classList.add('theme-dark');
        body.classList.remove('theme-light');
        if (themeBtn) themeBtn.textContent = '🌙';
    }

    if (themeBtn) {
        themeBtn.addEventListener('click', () => {
            const isLight = body.classList.contains('theme-light');
            if (isLight) {
                body.classList.remove('theme-light');
                body.classList.add('theme-dark');
                themeBtn.textContent = '🌙';
                localStorage.setItem('gitelligence_theme', 'dark');
            } else {
                body.classList.remove('theme-dark');
                body.classList.add('theme-light');
                themeBtn.textContent = '☀️';
                localStorage.setItem('gitelligence_theme', 'light');
            }
        });
    }

    // Couleurs d'accent
    const savedAccent = localStorage.getItem('gitelligence_accent') || 'cyan';
    setAccentColor(savedAccent);

    accentBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            const color = btn.dataset.color;
            if (!color) return;
            accentBtns.forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            setAccentColor(color);
            localStorage.setItem('gitelligence_accent', color);
        });
    });
}

function setAccentColor(colorName) {
    const colors = {
        cyan: { hex: '#38bdf8', glow: 'rgba(56, 189, 248, 0.35)' },
        purple: { hex: '#a855f7', glow: 'rgba(168, 85, 247, 0.35)' },
        green: { hex: '#34d399', glow: 'rgba(52, 211, 153, 0.35)' },
        orange: { hex: '#fb923c', glow: 'rgba(251, 146, 60, 0.35)' },
        rose: { hex: '#fb7185', glow: 'rgba(251, 113, 133, 0.35)' },
    };

    const sel = colors[colorName] || colors.cyan;
    document.documentElement.style.setProperty('--accent-current', sel.hex);
    document.documentElement.style.setProperty('--accent-current-glow', sel.glow);
}

// --------------------------------------------------------------------------
// 3. INITIALISATION ET CONNEXION AUX BINDINGS WAILS
// --------------------------------------------------------------------------
function initWailsConnection() {
    let checkCount = 0;
    const interval = setInterval(() => {
        if (window.go && window.go.main && window.go.main.App) {
            clearInterval(interval);
            appApi = window.go.main.App;
            onWailsReady();
        } else if (checkCount++ > 4) {
            clearInterval(interval);
            console.warn("Wails runtime non détecté après délai. Mode standalone/preview.");
            // Si on est en dev sans backend Wails injecté
            setupFallbackApi();
            onWailsReady();
        }
    }, 100);
}

function onWailsReady() {
    if (!eventListenersInitialized) {
        initEventListeners();
    }
    checkPythonStatus();
    refreshAll();
}

// --------------------------------------------------------------------------
// 4. ATTACHEMENT DES ÉVÉNEMENTS INTERACTIFS
// --------------------------------------------------------------------------
function initEventListeners() {
    if (eventListenersInitialized) return;
    eventListenersInitialized = true;
    // Bouton de rafraîchissement global
    document.getElementById('btn-global-refresh')?.addEventListener('click', () => {
        refreshAll();
        showToast("Statut actualisé !", "info");
    });

    // Sélecteur de répertoire cible
    document.getElementById('btn-browse-repo')?.addEventListener('click', async () => {
        try {
            const dir = await appApi.ChooseDirectory();
            if (dir && dir.trim() !== "") {
                document.getElementById('target-repo-input').value = dir;
                await changeRepository(dir);
            }
        } catch (err) {
            showToast("Erreur sélection dossier : " + err, "error");
        }
    });

    document.getElementById('btn-apply-repo')?.addEventListener('click', async () => {
        const path = document.getElementById('target-repo-input')?.value?.trim();
        if (path) {
            await changeRepository(path);
        }
    });

    // Actions rapides Dashboard
    document.getElementById('dash-btn-gen-commit')?.addEventListener('click', () => {
        document.querySelector('.nav-item[data-tab="commit"]')?.click();
        triggerCommitGeneration();
    });

    document.getElementById('dash-btn-view-diff')?.addEventListener('click', () => {
        document.querySelector('.nav-item[data-tab="diff"]')?.click();
    });

    document.getElementById('dash-btn-stage-all')?.addEventListener('click', handleStageAll);
    document.getElementById('diff-stage-all-btn')?.addEventListener('click', handleStageAll);

    document.getElementById('dash-btn-push')?.addEventListener('click', handlePush);
    document.getElementById('dash-refresh-commits')?.addEventListener('click', loadRecentCommits);

    // Diffs
    document.getElementById('diff-copy-patch-btn')?.addEventListener('click', () => {
        const patch = document.getElementById('diff-patch-content')?.innerText;
        if (patch) {
            navigator.clipboard.writeText(patch);
            showToast("Patch copié dans le presse-papier !", "success");
        }
    });

    // Générateur de Commits
    document.getElementById('commit-generate-btn')?.addEventListener('click', () => triggerCommitGeneration(""));
    document.getElementById('commit-refine-btn')?.addEventListener('click', () => {
        const feedback = document.getElementById('commit-feedback-input')?.value?.trim() || "";
        triggerCommitGeneration(feedback);
    });

    document.getElementById('commit-apply-btn')?.addEventListener('click', handleApplyCommit);
    document.getElementById('commit-copy-btn')?.addEventListener('click', () => {
        const msg = document.getElementById('commit-full-message')?.value;
        if (msg) {
            navigator.clipboard.writeText(msg);
            showToast("Message copié dans le presse-papier !", "success");
        }
    });
    document.getElementById('commit-push-btn')?.addEventListener('click', handlePush);

    // Release Notes (Sélecteur de commits & Génération)
    document.getElementById('rn-generate-btn')?.addEventListener('click', handleGenerateReleaseNotes);
    document.getElementById('rn-refresh-commits-btn')?.addEventListener('click', () => {
        loadReleaseNotesCommits();
        showToast("Historique des commits actualisé !", "info");
    });
    document.getElementById('rn-save-file-btn')?.addEventListener('click', handleSaveReleaseNotes);
    document.getElementById('rn-copy-btn')?.addEventListener('click', () => {
        const md = document.getElementById('rn-raw-markdown')?.value;
        if (md) {
            navigator.clipboard.writeText(md);
            showToast("Release notes copiées dans le presse-papier !", "success");
        }
    });

    // Filtres de recherche instantanés pour les listes de commits
    document.getElementById('rn-search-start')?.addEventListener('input', renderReleaseNotesCommitsList);
    document.getElementById('rn-search-end')?.addEventListener('input', renderReleaseNotesCommitsList);

    // Presets rapides Release Notes
    document.getElementById('rn-preset-all')?.addEventListener('click', () => {
        if (rnAllCommits.length > 0) {
            rnSelectedStartSha = rnAllCommits[rnAllCommits.length - 1].sha;
            rnSelectedEndSha = rnAllCommits[0].sha;
            renderReleaseNotesCommitsList();
            updateReleaseNotesSummary();
            showToast("Plage définie : Historique complet ➔ HEAD", "info");
        }
    });

    document.getElementById('rn-preset-last-tag')?.addEventListener('click', () => {
        if (rnAllCommits.length > 0) {
            // Cherche le premier commit avec tag (en dehors de HEAD)
            const tagged = rnAllCommits.slice(1).find(c => c.tag && c.tag.trim() !== "");
            if (tagged) {
                rnSelectedStartSha = tagged.sha;
            } else if (rnAllCommits[0].tag) {
                rnSelectedStartSha = rnAllCommits[0].sha;
            } else {
                showToast("Aucun tag antérieur détecté dans l'historique.", "warning");
                return;
            }
            rnSelectedEndSha = rnAllCommits[0].sha;
            renderReleaseNotesCommitsList();
            updateReleaseNotesSummary();
            showToast("Plage définie : Dernier tag ➔ HEAD", "info");
        }
    });

    document.getElementById('rn-preset-last-5')?.addEventListener('click', () => {
        if (rnAllCommits.length > 0) {
            const idx = Math.min(4, rnAllCommits.length - 1);
            rnSelectedStartSha = rnAllCommits[idx].sha;
            rnSelectedEndSha = rnAllCommits[0].sha;
            renderReleaseNotesCommitsList();
            updateReleaseNotesSummary();
            showToast("Plage définie : 5 derniers commits", "info");
        }
    });

    document.getElementById('rn-preset-last-10')?.addEventListener('click', () => {
        if (rnAllCommits.length > 0) {
            const idx = Math.min(9, rnAllCommits.length - 1);
            rnSelectedStartSha = rnAllCommits[idx].sha;
            rnSelectedEndSha = rnAllCommits[0].sha;
            renderReleaseNotesCommitsList();
            updateReleaseNotesSummary();
            showToast("Plage définie : 10 derniers commits", "info");
        }
    });

    // Bouton de redirection vers la config Python
    document.getElementById('btn-fix-python')?.addEventListener('click', () => {
        document.querySelector('.nav-item[data-tab="config"]')?.click();
    });

    // Pousser après validation du commit
    document.getElementById('btn-post-commit-push')?.addEventListener('click', async () => {
        document.getElementById('post-commit-push-prompt').style.display = 'none';
        await handlePush();
    });
    document.getElementById('btn-post-commit-dismiss')?.addEventListener('click', () => {
        document.getElementById('post-commit-push-prompt').style.display = 'none';
    });

    // Imposer un type Conventional Commit (Option [T] du CLI)
    document.querySelectorAll('.type-chip').forEach(chip => {
        chip.addEventListener('click', () => {
            const chosenType = chip.dataset.type;
            if (chosenType) {
                const comment = document.getElementById('commit-feedback-input')?.value?.trim() || "";
                triggerCommitGenerationWithType(chosenType, comment);
            }
        });
    });

    // Configuration, Docker & Python
    document.getElementById('cfg-ping-btn')?.addEventListener('click', handlePingOllama);
    document.getElementById('config-save-btn')?.addEventListener('click', handleSaveConfig);
    document.getElementById('cfg-docker-start')?.addEventListener('click', handleStartDocker);
    document.getElementById('cfg-docker-restart')?.addEventListener('click', handleRestartDocker);
    document.getElementById('cfg-docker-stop')?.addEventListener('click', handleStopDocker);
    document.getElementById('cfg-docker-refresh-status-btn')?.addEventListener('click', () => {
        updateDockerStatusAndLogs(true);
        showToast("Statut Docker actualisé !", "info");
    });
    document.getElementById('cfg-docker-ping-btn')?.addEventListener('click', handleDockerPing);
    document.getElementById('cfg-docker-log-refresh')?.addEventListener('click', () => {
        fetchDockerLogs();
        showToast("Logs Docker actualisés !", "info");
    });
    document.getElementById('cfg-docker-log-autorefresh')?.addEventListener('change', (e) => {
        if (e.target.checked) {
            startDockerPolling();
            showToast("Auto-refresh des logs activé (5s)", "info");
        } else {
            stopDockerPolling();
            showToast("Auto-refresh des logs désactivé", "info");
        }
    });
    document.getElementById('cfg-docker-log-lines')?.addEventListener('change', () => fetchDockerLogs());
    document.getElementById('cfg-docker-log-copy')?.addEventListener('click', () => {
        const text = document.getElementById('cfg-docker-logs-terminal')?.textContent || "";
        if (text) {
            navigator.clipboard.writeText(text);
            showToast("Logs Docker copiés dans le presse-papier !", "success");
        }
    });
    document.getElementById('cfg-docker-log-clear')?.addEventListener('click', () => {
        const term = document.getElementById('cfg-docker-logs-terminal');
        if (term) term.textContent = "Console effacée. Cliquez sur Actualiser pour recharger.";
        showToast("Affichage des logs vidé.", "info");
    });

    // Actions rapides Python
    document.getElementById('cfg-py-quick-monenv')?.addEventListener('click', () => {
        const input = document.getElementById('cfg-python-bin');
        if (input) input.value = '/home/maxence/monenv/bin/python';
    });
    document.getElementById('cfg-py-quick-auto')?.addEventListener('click', () => {
        const input = document.getElementById('cfg-python-bin');
        if (input) input.value = '';
    });
    document.getElementById('cfg-py-test-candidate')?.addEventListener('click', handleTestPythonCandidate);
    document.getElementById('cfg-py-save-btn')?.addEventListener('click', handleSavePythonBin);
    document.getElementById('cfg-py-diagnose-btn')?.addEventListener('click', handleDiagnosePython);
}

// --------------------------------------------------------------------------
// 5. CHARGEMENT ET RAFRAÎCHISSEMENT GLOBAL DES DONNÉES
// --------------------------------------------------------------------------
async function refreshAll() {
    await refreshDashboard();
    await refreshDiffs();
}

async function changeRepository(newPath) {
    if (!appApi) return;
    try {
        const status = await appApi.SetTargetRepo(newPath);
        showToast("Dépôt cible mis à jour : " + newPath, "success");
        updateRepoStatusDisplay(status);
        await refreshAll();
    } catch (err) {
        showToast("Échec changement de dépôt : " + err, "error");
    }
}

async function refreshDashboard() {
    if (!appApi) return;
    
    // 1. Récupération résiliente de l'état du dépôt
    try {
        const status = await appApi.GetRepoStatus();
        updateRepoStatusDisplay(status);
    } catch (err) {
        console.warn("Erreur GetRepoStatus :", err);
    }

    // 2. Préremplissage du chemin de dépôt cible
    try {
        const currentRepo = await appApi.GetTargetRepo();
        const input = document.getElementById('target-repo-input');
        if (input && currentRepo) {
            input.value = currentRepo;
        }
    } catch (err) {
        console.warn("Erreur GetTargetRepo :", err);
    }

    // 3. Chargement parallèle et indépendant des métriques, commits et serveur Ollama
    await Promise.allSettled([
        loadRecentCommits(),
        loadSystemSpecs(),
        checkOllamaStatus()
    ]);
}

function updateRepoStatusDisplay(status) {
    const branchEl = document.getElementById('kpi-branch');
    const repoStatusEl = document.getElementById('kpi-repo-status');
    const modCountEl = document.getElementById('kpi-modified-count');
    const stagedUnstagedEl = document.getElementById('kpi-staged-unstaged');
    const sideRepoName = document.getElementById('sidebar-repo-name');
    const sideBranch = document.getElementById('sidebar-branch-name');

    if (!status || !status.is_git_repo) {
        if (branchEl) branchEl.textContent = "Non-Git";
        if (repoStatusEl) repoStatusEl.textContent = "Répertoire cible non initialisé";
        if (sideRepoName) sideRepoName.textContent = "Aucun dépôt Git";
        if (sideBranch) sideBranch.textContent = "Invalide";
        return;
    }

    const branch = status.branch || "HEAD";
    const totalMod = (status.staged_count || 0) + (status.unstaged_count || 0);

    if (branchEl) branchEl.textContent = branch;
    if (sideBranch) sideBranch.textContent = branch;
    if (repoStatusEl) repoStatusEl.textContent = "Dépôt Git opérationnel";
    if (sideRepoName) {
        const parts = status.repo_path.split('/');
        sideRepoName.textContent = parts[parts.length - 1] || status.repo_path;
        sideRepoName.title = status.repo_path;
    }

    if (modCountEl) modCountEl.textContent = totalMod;
    if (stagedUnstagedEl) {
        stagedUnstagedEl.textContent = `${status.staged_count || 0} indexé(s) · ${status.unstaged_count || 0} non-indexé(s)`;
    }

    // Badge de diff dans la sidebar
    const badge = document.getElementById('sidebar-diff-badge');
    if (badge) {
        if (totalMod > 0) {
            badge.textContent = totalMod;
            badge.style.display = 'inline-block';
        } else {
            badge.style.display = 'none';
        }
    }
}

async function loadRecentCommits() {
    if (!appApi) return;
    const container = document.getElementById('dash-commits-list');
    if (!container) return;

    try {
        const resp = await appApi.GetRecentCommits(8);
        if (!resp || !resp.commits || resp.commits.length === 0) {
            container.innerHTML = '<div class="empty-state">Aucun commit trouvé dans l\'historique récent.</div>';
            return;
        }

        container.innerHTML = resp.commits.map(c => `
            <div class="commit-row">
                <span class="commit-sha">${escapeHtml(c.sha.substring(0, 7))}</span>
                <span class="commit-msg" title="${escapeHtml(c.message)}">${escapeHtml(c.message)}</span>
                <span class="commit-author">${escapeHtml(c.author)}</span>
                ${c.tag ? `<span class="badge badge-info">${escapeHtml(c.tag)}</span>` : ''}
            </div>
        `).join('');
    } catch (err) {
        container.innerHTML = `<div class="empty-state">Erreur de chargement : ${escapeHtml(err.toString())}</div>`;
    }
}

async function loadSystemSpecs() {
    if (!appApi) return;
    try {
        const specs = await appApi.GetSystemSpecs();
        if (!specs) return;

        document.getElementById('spec-os').textContent = specs.os || "Linux";
        document.getElementById('spec-arch').textContent = specs.arch || "amd64";
        document.getElementById('spec-go').textContent = specs.go_version || "1.27";
        document.getElementById('spec-mem').textContent = specs.alloc_mb || "0 MB";
        document.getElementById('spec-cpu').textContent = specs.cpus || "1";
        document.getElementById('spec-goroutines').textContent = specs.goroutines || "1";
        document.getElementById('spec-webview').textContent = specs.webview || "WebKitGTK 4.1";
    } catch (err) {
        console.error("Erreur specs :", err);
    }
}

async function checkOllamaStatus() {
    if (!appApi) return;
    const kpiModel = document.getElementById('kpi-ollama-model');
    const kpiStatus = document.getElementById('kpi-ollama-status');
    const kpiDocker = document.getElementById('kpi-docker-status');
    const kpiDockerSub = document.getElementById('kpi-docker-mode');

    // 1. Modèle configuré dans le .env
    try {
        const cfg = await appApi.GetConfig();
        if (cfg) {
            if (kpiModel) kpiModel.textContent = cfg.ollama_model || "Non défini";
            if (kpiDockerSub) {
                kpiDockerSub.textContent = cfg.docker_llm ? "Mode Conteneur Local" : "Mode Distant (.env)";
            }
        }
    } catch (err) {
        console.warn("Erreur GetConfig dashboard :", err);
    }

    // 2. État Docker Compose (isolé pour ne pas bloquer le ping)
    try {
        const dockerStatus = await appApi.DockerOllamaStatus();
        if (dockerStatus && kpiDocker) {
            kpiDocker.textContent = dockerStatus.running ? "Actif (En cours)" : "Inactif";
        }
    } catch (err) {
        if (kpiDocker) kpiDocker.textContent = "Inactif";
    }

    // 3. Test de ping direct sur l'API Ollama
    try {
        const ping = await appApi.PingOllama();
        if (ping && kpiStatus) {
            kpiStatus.textContent = ping.available 
                ? `En ligne (${ping.latency_ms} ms)` 
                : "Indisponible / Hors-ligne";
            kpiStatus.style.color = ping.available ? "#34d399" : "#fb7185";
        }
    } catch (err) {
        if (kpiStatus) {
            kpiStatus.textContent = "Erreur de connexion";
            kpiStatus.style.color = "#fb7185";
        }
    }
}

// --------------------------------------------------------------------------
// 6. GESTION DES DIFFS
// --------------------------------------------------------------------------
async function refreshDiffs() {
    if (!appApi) return;
    const filesContainer = document.getElementById('diff-files-container');
    const fileCountEl = document.getElementById('diff-file-count');

    try {
        const diffResp = await appApi.GetDiff();
        if (!diffResp || !diffResp.files || diffResp.files.length === 0) {
            let unstagedCount = 0;
            try {
                const status = await appApi.GetRepoStatus();
                if (status) unstagedCount = status.unstaged_count || 0;
            } catch (_) {}

            if (filesContainer) {
                if (unstagedCount > 0) {
                    filesContainer.innerHTML = `
                        <div class="empty-state" style="padding: 24px 16px; text-align: center;">
                            <div style="font-size: 2rem; margin-bottom: 8px;">📂</div>
                            <div style="font-weight: 600; margin-bottom: 6px; color: var(--text-main);">Aucun fichier indexé (staged)</div>
                            <div class="text-muted" style="font-size: 0.85rem; margin-bottom: 16px;">${unstagedCount} fichier(s) modifié(s) en attente d'indexation.</div>
                            <button id="diff-empty-stage-all" class="btn btn-primary btn-sm" style="width: 100%; justify-content: center;">➕ Tout indexer (git add -A)</button>
                        </div>
                    `;
                    document.getElementById('diff-empty-stage-all')?.addEventListener('click', handleStageAll);
                } else {
                    filesContainer.innerHTML = '<div class="empty-state">Aucun fichier modifié détecté.</div>';
                }
            }
            if (fileCountEl) fileCountEl.textContent = '0';
            currentDiffFiles = [];
            renderPatch(null);
            return;
        }

        currentDiffFiles = diffResp.files;
        if (fileCountEl) fileCountEl.textContent = currentDiffFiles.length;

        if (filesContainer) {
            filesContainer.innerHTML = currentDiffFiles.map((file, idx) => `
                <div class="diff-file-item ${idx === selectedDiffIndex ? 'active' : ''}" data-idx="${idx}">
                    <span class="diff-file-status ${getStatusClass(file.status)}">${file.status || 'M'}</span>
                    <span class="diff-file-name" title="${escapeHtml(file.path)}">${escapeHtml(file.path)}</span>
                    <span class="diff-file-stat">
                        <span class="text-success">+${file.added}</span>
                        <span class="text-danger">-${file.removed}</span>
                    </span>
                </div>
            `).join('');

            // Clic sur un fichier
            filesContainer.querySelectorAll('.diff-file-item').forEach(item => {
                item.addEventListener('click', () => {
                    const idx = parseInt(item.dataset.idx, 10);
                    selectedDiffIndex = idx;
                    filesContainer.querySelectorAll('.diff-file-item').forEach(i => i.classList.remove('active'));
                    item.classList.add('active');
                    renderPatch(currentDiffFiles[idx]);
                });
            });
        }

        // Afficher le premier fichier par défaut
        if (currentDiffFiles.length > 0) {
            if (selectedDiffIndex >= currentDiffFiles.length) selectedDiffIndex = 0;
            renderPatch(currentDiffFiles[selectedDiffIndex]);
        }
    } catch (err) {
        if (filesContainer) filesContainer.innerHTML = `<div class="empty-state">Erreur : ${escapeHtml(err.toString())}</div>`;
    }
}

function renderPatch(file) {
    const currentFileEl = document.getElementById('diff-current-file');
    const addedBadge = document.getElementById('diff-added-badge');
    const deletedBadge = document.getElementById('diff-deleted-badge');
    const secretBadge = document.getElementById('diff-secret-badge');
    const binaryBadge = document.getElementById('diff-binary-badge');
    const viewer = document.getElementById('diff-patch-content');

    if (!file) {
        if (currentFileEl) currentFileEl.textContent = "Aucun fichier sélectionné";
        if (addedBadge) addedBadge.textContent = "+0";
        if (deletedBadge) deletedBadge.textContent = "-0";
        if (viewer) viewer.innerHTML = '<code class="diff-code">Aucune modification à afficher.</code>';
        return;
    }

    if (currentFileEl) currentFileEl.textContent = file.path;
    if (addedBadge) addedBadge.textContent = `+${file.added}`;
    if (deletedBadge) deletedBadge.textContent = `-${file.removed}`;

    if (secretBadge) secretBadge.style.display = file.has_secrets_masked ? 'inline-flex' : 'none';
    if (binaryBadge) binaryBadge.style.display = file.binary ? 'inline-flex' : 'none';

    if (viewer) {
        if (file.binary) {
            viewer.innerHTML = '<code class="diff-code text-muted">Contenu binaire non affiché.</code>';
            return;
        }

        // Coloration syntaxique ligne par ligne
        const lines = (file.patch || "").split('\n');
        const formatted = lines.map(line => {
            const escaped = escapeHtml(line);
            if (line.startsWith('+++') || line.startsWith('---')) {
                return `<span class="diff-line-hunk">${escaped}</span>`;
            } else if (line.startsWith('+')) {
                return `<span class="diff-line-add">${escaped}</span>`;
            } else if (line.startsWith('-')) {
                return `<span class="diff-line-del">${escaped}</span>`;
            } else if (line.startsWith('@@')) {
                return `<span class="diff-line-hunk">${escaped}</span>`;
            }
            return `<span>${escaped}</span>`;
        }).join('\n');

        viewer.innerHTML = `<code class="diff-code">${formatted || 'Aucun patch différentiel.'}</code>`;
    }
}

function getStatusClass(status) {
    switch (status) {
        case 'A': return 'text-success';
        case 'D': return 'text-danger';
        case 'M': return 'text-warning';
        default: return 'text-primary';
    }
}

// --------------------------------------------------------------------------
// 7. ACTIONS GIT : STAGE ALL & PUSH
// --------------------------------------------------------------------------
async function handleStageAll() {
    if (!appApi) return;
    try {
        const res = await appApi.StageAll();
        if (res && res.success) {
            showToast("Tous les fichiers ont été indexés !", "success");
            await refreshAll();
        } else {
            showToast("Erreur stage all : " + (res?.error || "Inconnue"), "error");
        }
    } catch (err) {
        showToast("Erreur : " + err, "error");
    }
}

async function handlePush() {
    if (!appApi) return;
    try {
        showToast("Envoi vers le dépôt distant en cours...", "info");
        const res = await appApi.Push();
        if (res && res.success) {
            showToast("Branches poussées avec succès vers le remote !", "success");
        } else {
            showToast("Échec git push : " + (res?.error || "Erreur distante"), "error");
        }
    } catch (err) {
        showToast("Erreur push : " + err, "error");
    }
}

// --------------------------------------------------------------------------
// 8. GÉNÉRATEUR CONVENTIONAL COMMITS
// --------------------------------------------------------------------------
async function triggerCommitGeneration(feedback = "") {
    if (!appApi) return;
    const loadingEl = document.getElementById('commit-loading');
    const proposalArea = document.getElementById('commit-proposal-area');
    const feedbackArea = document.getElementById('commit-feedback-area');
    const forcedTypeArea = document.getElementById('commit-forced-type-area');
    const genBtn = document.getElementById('commit-generate-btn');

    if (loadingEl) loadingEl.style.display = 'flex';
    if (proposalArea) proposalArea.style.display = 'none';
    if (feedbackArea) feedbackArea.style.display = 'none';
    if (forcedTypeArea) forcedTypeArea.style.display = 'none';
    if (genBtn) genBtn.disabled = true;

    try {
        const proposal = await appApi.GenerateCommit(feedback);
        applyCommitProposalToView(proposal);
    } catch (err) {
        showToast("Erreur génération commit : " + err, "error");
    } finally {
        if (loadingEl) loadingEl.style.display = 'none';
        if (genBtn) genBtn.disabled = false;
    }
}

async function triggerCommitGenerationWithType(chosenType, comment = "") {
    if (!appApi) return;
    const loadingEl = document.getElementById('commit-loading');
    const proposalArea = document.getElementById('commit-proposal-area');
    const feedbackArea = document.getElementById('commit-feedback-area');
    const forcedTypeArea = document.getElementById('commit-forced-type-area');
    const genBtn = document.getElementById('commit-generate-btn');

    if (loadingEl) loadingEl.style.display = 'flex';
    if (proposalArea) proposalArea.style.display = 'none';
    if (feedbackArea) feedbackArea.style.display = 'none';
    if (forcedTypeArea) forcedTypeArea.style.display = 'none';
    if (genBtn) genBtn.disabled = true;

    try {
        const proposal = await appApi.ImposeCommitType(chosenType, comment);
        applyCommitProposalToView(proposal);
        showToast(`Commit régénéré avec le type imposé : ${chosenType}`, "info");
    } catch (err) {
        showToast("Erreur type imposé : " + err, "error");
    } finally {
        if (loadingEl) loadingEl.style.display = 'none';
        if (genBtn) genBtn.disabled = false;
    }
}

function applyCommitProposalToView(proposal) {
    const proposalArea = document.getElementById('commit-proposal-area');
    const feedbackArea = document.getElementById('commit-feedback-area');
    const forcedTypeArea = document.getElementById('commit-forced-type-area');

    if (!proposal || !proposal.success) {
        showToast("Échec génération commit : " + (proposal?.error || "IA non joignable"), "error");
        return;
    }

    // Remplissage des métadonnées
    document.getElementById('commit-type-badge').textContent = proposal.type || 'feat';
    document.getElementById('commit-scope-badge').textContent = proposal.scope || 'core';
    document.getElementById('commit-subject-display').textContent = proposal.subject || '';

    // Remplissage de la zone de texte éditable
    const fullMsgArea = document.getElementById('commit-full-message');
    if (fullMsgArea) {
        if (proposal.raw_formatted) {
            fullMsgArea.value = proposal.raw_formatted;
        } else {
            let full = `${proposal.type}${proposal.scope ? '(' + proposal.scope + ')' : ''}: ${proposal.subject}`;
            if (proposal.body) full += `\n\n${proposal.body}`;
            fullMsgArea.value = full;
        }
    }

    if (proposalArea) proposalArea.style.display = 'block';
    if (feedbackArea) feedbackArea.style.display = 'block';
    if (forcedTypeArea) forcedTypeArea.style.display = 'block';
    showToast("Message de commit généré avec succès !", "success");
}

async function handleApplyCommit() {
    if (!appApi) return;
    const msg = document.getElementById('commit-full-message')?.value?.trim();
    if (!msg) {
        showToast("Le message de commit ne peut pas être vide !", "warning");
        return;
    }

    try {
        const res = await appApi.ApplyCommit(msg);
        if (res && res.success) {
            showToast("Commit créé et enregistré avec succès ! (" + (res.sha ? res.sha.substring(0, 7) : "OK") + ")", "success");
            await refreshAll();
            // Affiche la proposition de push
            const postPrompt = document.getElementById('post-commit-push-prompt');
            if (postPrompt) postPrompt.style.display = 'flex';
        } else {
            showToast("Échec de l'application du commit : " + (res?.error || "Inconnu"), "error");
        }
    } catch (err) {
        showToast("Erreur : " + err, "error");
    }
}

// --------------------------------------------------------------------------
// 9. RELEASE NOTES & CHANGELOGS (AVEC SÉLECTEUR INTERACTIF DE COMMITS)
// --------------------------------------------------------------------------

async function loadReleaseNotesCommits() {
    if (!appApi) return;
    const startList = document.getElementById('rn-commits-list-start');
    const endList = document.getElementById('rn-commits-list-end');

    try {
        // Chargement de l'historique complet (0 = tous les commits comme dans la CLI)
        const resp = await appApi.GetRecentCommits(0);
        if (!resp || !resp.commits || resp.commits.length === 0) {
            const emptyHtml = '<div class="loading-commits-text">Aucun commit trouvé dans ce dépôt Git.</div>';
            if (startList) startList.innerHTML = emptyHtml;
            if (endList) endList.innerHTML = emptyHtml;
            return;
        }

        rnAllCommits = resp.commits;

        // Pré-sélection intelligente (comme dans la CLI) :
        // 1. ARRIVÉE : toujours HEAD (le commit le plus récent, index 0)
        if (!rnSelectedEndSha || !rnAllCommits.some(c => c.sha === rnSelectedEndSha)) {
            rnSelectedEndSha = rnAllCommits[0].sha;
        }

        // 2. DÉPART : recherche du premier commit taggé antérieur à HEAD, sinon 5 derniers commits
        if (!rnSelectedStartSha || !rnAllCommits.some(c => c.sha === rnSelectedStartSha)) {
            const tagged = rnAllCommits.slice(1).find(c => c.tag && c.tag.trim() !== "");
            if (tagged) {
                rnSelectedStartSha = tagged.sha;
            } else if (rnAllCommits.length > 5) {
                rnSelectedStartSha = rnAllCommits[4].sha;
            } else {
                rnSelectedStartSha = rnAllCommits[rnAllCommits.length - 1].sha;
            }
        }

        renderReleaseNotesCommitsList();
        updateReleaseNotesSummary();
    } catch (err) {
        console.error("Erreur chargement commits release notes:", err);
        const errHtml = `<div class="loading-commits-text text-danger">Erreur : ${escapeHtml(err.toString())}</div>`;
        if (startList) startList.innerHTML = errHtml;
        if (endList) endList.innerHTML = errHtml;
    }
}

function renderReleaseNotesCommitsList() {
    const startList = document.getElementById('rn-commits-list-start');
    const endList = document.getElementById('rn-commits-list-end');
    const qStart = (document.getElementById('rn-search-start')?.value || "").toLowerCase().trim();
    const qEnd = (document.getElementById('rn-search-end')?.value || "").toLowerCase().trim();

    if (startList) {
        const filteredStart = rnAllCommits.filter(c => matchesCommitFilter(c, qStart));
        if (filteredStart.length === 0) {
            startList.innerHTML = '<div class="loading-commits-text">Aucun commit correspondant.</div>';
        } else {
            startList.innerHTML = filteredStart.map(c => `
                <div class="picker-commit-item ${c.sha === rnSelectedStartSha ? 'selected-start' : ''}" data-sha="${escapeHtml(c.sha)}">
                    <div class="picker-item-top">
                        <div class="picker-item-sha-group">
                            <span class="picker-item-sha">${escapeHtml(c.sha.substring(0, 7))}</span>
                            ${c.tag ? `<span class="tag-badge">🏷️ ${escapeHtml(c.tag)}</span>` : ''}
                        </div>
                        <span class="picker-item-date">${escapeHtml(c.date || '')}</span>
                    </div>
                    <div class="picker-item-msg" title="${escapeHtml(c.message)}">${escapeHtml(c.message)}</div>
                    <div class="picker-item-author">${escapeHtml(c.author || '')}</div>
                </div>
            `).join('');

            startList.querySelectorAll('.picker-commit-item').forEach(item => {
                item.addEventListener('click', () => {
                    rnSelectedStartSha = item.dataset.sha;
                    renderReleaseNotesCommitsList();
                    updateReleaseNotesSummary();
                });
            });
        }
    }

    if (endList) {
        const filteredEnd = rnAllCommits.filter(c => matchesCommitFilter(c, qEnd));
        if (filteredEnd.length === 0) {
            endList.innerHTML = '<div class="loading-commits-text">Aucun commit correspondant.</div>';
        } else {
            endList.innerHTML = filteredEnd.map(c => `
                <div class="picker-commit-item ${c.sha === rnSelectedEndSha ? 'selected-end' : ''}" data-sha="${escapeHtml(c.sha)}">
                    <div class="picker-item-top">
                        <div class="picker-item-sha-group">
                            <span class="picker-item-sha">${escapeHtml(c.sha.substring(0, 7))}</span>
                            ${c.tag ? `<span class="tag-badge">🏷️ ${escapeHtml(c.tag)}</span>` : ''}
                        </div>
                        <span class="picker-item-date">${escapeHtml(c.date || '')}</span>
                    </div>
                    <div class="picker-item-msg" title="${escapeHtml(c.message)}">${escapeHtml(c.message)}</div>
                    <div class="picker-item-author">${escapeHtml(c.author || '')}</div>
                </div>
            `).join('');

            endList.querySelectorAll('.picker-commit-item').forEach(item => {
                item.addEventListener('click', () => {
                    rnSelectedEndSha = item.dataset.sha;
                    renderReleaseNotesCommitsList();
                    updateReleaseNotesSummary();
                });
            });
        }
    }
}

function matchesCommitFilter(commit, query) {
    if (!query) return true;
    return (
        commit.sha.toLowerCase().includes(query) ||
        (commit.message && commit.message.toLowerCase().includes(query)) ||
        (commit.author && commit.author.toLowerCase().includes(query)) ||
        (commit.tag && commit.tag.toLowerCase().includes(query))
    );
}

function updateReleaseNotesSummary() {
    const startShaEl = document.getElementById('rn-selected-start-sha');
    const startTagEl = document.getElementById('rn-selected-start-tag');
    const startMsgEl = document.getElementById('rn-selected-start-msg');

    const endShaEl = document.getElementById('rn-selected-end-sha');
    const endTagEl = document.getElementById('rn-selected-end-tag');
    const endMsgEl = document.getElementById('rn-selected-end-msg');

    const intervalCountEl = document.getElementById('rn-interval-count-badge');
    const genBtn = document.getElementById('rn-generate-btn');

    const startCommit = rnAllCommits.find(c => c.sha === rnSelectedStartSha);
    const endCommit = rnAllCommits.find(c => c.sha === rnSelectedEndSha);

    if (startCommit) {
        if (startShaEl) startShaEl.textContent = startCommit.sha.substring(0, 7);
        if (startTagEl) {
            if (startCommit.tag) {
                startTagEl.textContent = `🏷️ ${startCommit.tag}`;
                startTagEl.style.display = 'inline-block';
            } else {
                startTagEl.style.display = 'none';
            }
        }
        if (startMsgEl) startMsgEl.textContent = startCommit.message;
    } else {
        if (startShaEl) startShaEl.textContent = "Sélectionnez...";
        if (startTagEl) startTagEl.style.display = 'none';
        if (startMsgEl) startMsgEl.textContent = "Cliquez sur un commit dans la colonne de départ";
    }

    if (endCommit) {
        const isHead = rnAllCommits.length > 0 && rnAllCommits[0].sha === endCommit.sha;
        if (endShaEl) endShaEl.textContent = isHead ? `HEAD (${endCommit.sha.substring(0, 7)})` : endCommit.sha.substring(0, 7);
        if (endTagEl) {
            if (endCommit.tag) {
                endTagEl.textContent = `🏷️ ${endCommit.tag}`;
                endTagEl.style.display = 'inline-block';
            } else {
                endTagEl.style.display = 'none';
            }
        }
        if (endMsgEl) endMsgEl.textContent = endCommit.message;
    } else {
        if (endShaEl) endShaEl.textContent = "Sélectionnez...";
        if (endTagEl) endTagEl.style.display = 'none';
        if (endMsgEl) endMsgEl.textContent = "Cliquez sur un commit dans la colonne d'arrivée";
    }

    // Calcul du nombre de commits dans l'intervalle sélectionné
    if (startCommit && endCommit) {
        const idxStart = rnAllCommits.findIndex(c => c.sha === rnSelectedStartSha);
        const idxEnd = rnAllCommits.findIndex(c => c.sha === rnSelectedEndSha);
        if (idxStart !== -1 && idxEnd !== -1) {
            const count = Math.abs(idxStart - idxEnd) + 1;
            if (intervalCountEl) intervalCountEl.textContent = `${count} commit(s) sélectionné(s)`;
        }
        if (genBtn) {
            genBtn.textContent = `⚡ Rédiger les Release Notes (${startCommit.sha.substring(0, 7)} ➔ ${endCommit.sha.substring(0, 7)})`;
            genBtn.disabled = false;
        }
    } else {
        if (intervalCountEl) intervalCountEl.textContent = "0 commit(s)";
        if (genBtn) {
            genBtn.textContent = "⚡ Rédiger les Release Notes";
            genBtn.disabled = true;
        }
    }
}

async function handleGenerateReleaseNotes() {
    if (!appApi) return;
    if (!rnSelectedStartSha || !rnSelectedEndSha) {
        showToast("Veuillez sélectionner un commit de départ et un commit d'arrivée !", "warning");
        return;
    }

    const loadingEl = document.getElementById('rn-loading');
    const resultArea = document.getElementById('rn-result-area');
    const genBtn = document.getElementById('rn-generate-btn');

    if (loadingEl) {
        loadingEl.style.display = 'flex';
        const loadingText = loadingEl.querySelector('.loading-text');
        if (loadingText) {
            loadingText.textContent = `Extraction des commits (${rnSelectedStartSha.substring(0, 7)} ➔ ${rnSelectedEndSha.substring(0, 7)}) et synthèse par le LLM...`;
        }
    }
    if (resultArea) resultArea.style.display = 'none';
    if (genBtn) genBtn.disabled = true;

    try {
        const notes = await appApi.GetReleaseNotes(rnSelectedStartSha, rnSelectedEndSha);
        if (!notes || !notes.success) {
            showToast("Échec génération release notes : " + (notes?.error || "Erreur LLM"), "error");
            return;
        }

        const countBadge = document.getElementById('rn-commit-count-badge');
        if (countBadge) countBadge.textContent = `${notes.commit_count || 0} commit(s) analysé(s)`;

        const rawArea = document.getElementById('rn-raw-markdown');
        if (rawArea) rawArea.value = notes.markdown || "";

        const renderedEl = document.getElementById('rn-rendered-html');
        if (renderedEl) {
            renderedEl.innerHTML = renderMarkdownBasic(notes.markdown || "");
        }

        if (resultArea) {
            resultArea.style.display = 'block';
            const scrollContainer = document.getElementById('view-release-notes');
            if (scrollContainer) {
                const containerTop = scrollContainer.getBoundingClientRect().top;
                const targetTop = resultArea.getBoundingClientRect().top;
                const offset = targetTop - containerTop + scrollContainer.scrollTop - 16;
                scrollContainer.scrollTo({ top: Math.max(0, offset), behavior: 'smooth' });
            }
        }
        showToast("Release notes rédigées avec succès !", "success");
    } catch (err) {
        showToast("Erreur : " + err, "error");
    } finally {
        if (loadingEl) loadingEl.style.display = 'none';
        if (genBtn) genBtn.disabled = false;
    }
}

async function handleSaveReleaseNotes() {
    if (!appApi) return;
    const filename = document.getElementById('rn-filename-input')?.value?.trim() || "RELEASE_NOTES.md";
    const content = document.getElementById('rn-raw-markdown')?.value;

    if (!content) {
        showToast("Aucun contenu à enregistrer !", "warning");
        return;
    }

    try {
        await appApi.SaveReleaseNotesMarkdown(filename, content);
        showToast(`Fichier ${filename} enregistré avec succès à la racine !`, "success");
    } catch (err) {
        showToast("Erreur enregistrement : " + err, "error");
    }
}

// --------------------------------------------------------------------------
// 10. CONFIGURATION & DIAGNOSTICS
// --------------------------------------------------------------------------
// --------------------------------------------------------------------------
// 10. CONFIGURATION & DIAGNOSTICS PYTHON ET OLLAMA (COPIE CONFORME CLI)
// --------------------------------------------------------------------------
async function checkPythonStatus() {
    if (!appApi) return;
    const backendLabel = document.getElementById('sidebar-backend-label');
    const backendSub = document.getElementById('sidebar-backend-sub');
    const backendDot = document.getElementById('sidebar-backend-dot');
    const errBanner = document.getElementById('python-error-banner');
    const errDetail = document.getElementById('python-error-detail');

    try {
        const report = await appApi.VerifyPythonEnvironment();
        if (report && report.success) {
            if (backendLabel) backendLabel.textContent = `Python ${report.version || "OK"}`;
            if (backendSub) backendSub.textContent = report.mode ? `Mode ${report.mode}` : "Backend opérationnel";
            if (backendDot) {
                backendDot.classList.remove('dot-error');
                backendDot.style.background = '#34d399';
            }
            if (errBanner) errBanner.style.display = 'none';
        } else {
            const reason = (report && report.missing_deps && report.missing_deps.length > 0)
                ? `Dépendance(s) manquante(s) : ${report.missing_deps.join(', ')}`
                : (report?.app_load_error || "Interpréteur non configuré ou invalide");

            if (backendLabel) backendLabel.textContent = "Erreur Python";
            if (backendSub) backendSub.textContent = "Backend non opérationnel";
            if (backendDot) {
                backendDot.classList.add('dot-error');
                backendDot.style.background = '#ef4444';
            }
            if (errBanner) {
                errBanner.style.display = 'flex';
                if (errDetail) errDetail.textContent = reason;
            }
        }
    } catch (err) {
        if (backendLabel) backendLabel.textContent = "Erreur Python";
        if (backendSub) backendSub.textContent = "Échec du diagnostic";
        if (backendDot) {
            backendDot.classList.add('dot-error');
            backendDot.style.background = '#ef4444';
        }
        if (errBanner) {
            errBanner.style.display = 'flex';
            if (errDetail) errDetail.textContent = err.toString();
        }
    }
}

async function loadConfig() {
    if (!appApi) return;
    try {
        const cfg = await appApi.GetConfig();
        if (!cfg) return;

        document.getElementById('cfg-base-url').value = cfg.ollama_base_url || "";
        document.getElementById('cfg-model').value = cfg.ollama_model || "";
        document.getElementById('cfg-timeout').value = cfg.timeout_s || 30;
        document.getElementById('cfg-docker-llm').checked = !!cfg.docker_llm;
        document.getElementById('cfg-mock-interface').checked = !!cfg.mock_interface;
        document.getElementById('cfg-mock-ai').checked = !!cfg.mock_ai;
        document.getElementById('cfg-notifications').checked = !!cfg.enable_notifications;

        // Configuration Python
        const pyInput = document.getElementById('cfg-python-bin');
        if (pyInput) pyInput.value = cfg.python_bin || "";

        const resolvedBinEl = document.getElementById('cfg-py-resolved-bin');
        if (resolvedBinEl) resolvedBinEl.textContent = cfg.resolved_python_bin || "Non résolu";

        const pyModeEl = document.getElementById('cfg-py-mode');
        if (pyModeEl) {
            pyModeEl.textContent = cfg.python_bin ? "Personnalisé (PYTHON_BIN)" : "Automatique (.venv)";
        }

        await updateDockerStatusAndLogs(true);
        await updatePythonBadgeInConfig();
    } catch (err) {
        console.error("Erreur chargement config :", err);
    }
}

async function updatePythonBadgeInConfig() {
    if (!appApi) return;
    const badge = document.getElementById('cfg-py-badge');
    try {
        const report = await appApi.VerifyPythonEnvironment();
        if (badge) {
            if (report && report.success) {
                badge.className = 'badge badge-success';
                badge.textContent = `Opérationnel (Python ${report.version})`;
            } else {
                badge.className = 'badge badge-danger';
                badge.textContent = 'Invalide';
            }
        }
    } catch (_) {
        if (badge) {
            badge.className = 'badge badge-danger';
            badge.textContent = 'Inaccessible';
        }
    }
}

async function handleTestPythonCandidate() {
    if (!appApi) return;
    const input = document.getElementById('cfg-python-bin')?.value?.trim() || "";
    const isAuto = input === "";
    showToast("Vérification de l'interpréteur candidat en cours...", "info");

    try {
        const report = await appApi.VerifyPythonCandidate(input, isAuto);
        if (report && report.success) {
            showToast(`Interpréteur validé ! (Python ${report.version})`, "success");
        } else {
            const err = (report?.missing_deps?.length > 0)
                ? `Dépendances manquantes : ${report.missing_deps.join(', ')}`
                : (report?.app_load_error || "Erreur inconnue");
            showToast(`Candidat invalide : ${err}`, "error");
        }
    } catch (err) {
        showToast(`Échec du test : ${err}`, "error");
    }
}

async function handleSavePythonBin() {
    if (!appApi) return;
    const input = document.getElementById('cfg-python-bin')?.value?.trim() || "";
    const isAuto = input === "";

    try {
        const res = await appApi.SavePythonBin(input, isAuto);
        if (res && res.success) {
            showToast("Interpréteur Python enregistré et validé !", "success");
            await loadConfig();
            await checkPythonStatus();
            await refreshDashboard();
        } else {
            showToast("Modification refusée : " + (res?.error || "Candidat invalide"), "error");
        }
    } catch (err) {
        showToast("Erreur modification Python : " + err, "error");
    }
}

async function handleDiagnosePython() {
    if (!appApi) return;
    const reportBox = document.getElementById('cfg-py-report-card');
    if (!reportBox) return;

    reportBox.style.display = 'block';
    reportBox.innerHTML = '<span class="text-muted">Analyse approfondie de l\'environnement Python et des dépendances en cours...</span>';

    try {
        const report = await appApi.VerifyPythonEnvironment();
        if (report && report.success) {
            reportBox.innerHTML = `
                <div style="color: #34d399; font-weight: 600; margin-bottom: 6px;">✔ Environnement Python validé et complet</div>
                <div><strong>Exécutable :</strong> <code class="font-mono text-cyan">${escapeHtml(report.executable)}</code></div>
                <div><strong>Version :</strong> Python ${escapeHtml(report.version)}</div>
                <div><strong>Mode d'attribution :</strong> ${escapeHtml(report.mode || "Automatique")}</div>
                <div><strong>Dépendances requises :</strong> <span class="badge badge-success">Toutes installées (GitPython, PyYAML)</span></div>
                <div><strong>Module Backend :</strong> <span class="badge badge-success">Succès (git_commit_release_notes_generator)</span></div>
            `;
        } else {
            const missing = report?.missing_deps || [];
            const loadErr = report?.app_load_error || "Inconnu";
            reportBox.innerHTML = `
                <div style="color: #f87171; font-weight: 600; margin-bottom: 6px;">❌ Environnement Python non conforme</div>
                <div><strong>Exécutable :</strong> <code>${escapeHtml(report?.executable || "Introuvable")}</code></div>
                ${missing.length > 0 ? `<div><strong class="text-danger">Dépendances manquantes :</strong> ${escapeHtml(missing.join(', '))}</div>` : ''}
                ${loadErr ? `<div><strong class="text-danger">Erreur backend :</strong> ${escapeHtml(loadErr)}</div>` : ''}
                <div class="text-muted" style="margin-top: 8px; font-size: 11px;">
                    Astuce : utilisez le bouton <code>~/monenv</code> ci-dessus ou lancez <code>python scripts/setup_python_env.py</code>.
                </div>
            `;
        }
    } catch (err) {
        reportBox.innerHTML = `<div style="color: #f87171;">Échec du diagnostic : ${escapeHtml(err.toString())}</div>`;
    }
}

// --------------------------------------------------------------------------
// GESTION AVANCÉE DU CONTENEUR DOCKER & JOURNAUX D'EXÉCUTION
// --------------------------------------------------------------------------

function startDockerPolling() {
    stopDockerPolling();
    dockerPollTimer = setInterval(async () => {
        const autoRefreshChecked = document.getElementById('cfg-docker-log-autorefresh')?.checked ?? false;
        if (!autoRefreshChecked) {
            stopDockerPolling();
            return;
        }
        await updateDockerStatusAndLogs(true);
    }, 5000);
}

function startActiveDockerWatch() {
    stopDockerPolling();
    let checks = 0;
    dockerPollTimer = setInterval(async () => {
        checks++;
        const status = await updateDockerStatusAndLogs(true);
        if ((status && status.is_ready) || checks >= 15) {
            const autoChecked = document.getElementById('cfg-docker-log-autorefresh')?.checked ?? false;
            if (!autoChecked) {
                stopDockerPolling();
            }
        }
    }, 3000);
}

function stopDockerPolling() {
    if (dockerPollTimer) {
        clearInterval(dockerPollTimer);
        dockerPollTimer = null;
    }
}

async function updateDockerStatusAndLogs(shouldFetchLogs = false) {
    if (!appApi) return;
    const pulseEl = document.getElementById('cfg-docker-pulse');
    const badgeEl = document.getElementById('cfg-docker-badge');
    const stateIconEl = document.getElementById('cfg-docker-state-icon');
    const statusTextEl = document.getElementById('cfg-docker-status-text');
    const subTextEl = document.getElementById('cfg-docker-sub-text');
    const pullBoxEl = document.getElementById('cfg-docker-pull-box');
    const pullInfoEl = document.getElementById('cfg-docker-pull-info');
    const metaModelEl = document.getElementById('cfg-docker-meta-model');
    const metaCidEl = document.getElementById('cfg-docker-meta-cid');
    const metaHealthEl = document.getElementById('cfg-docker-meta-health');
    const metaStateEl = document.getElementById('cfg-docker-meta-state');
    const kpiDocker = document.getElementById('kpi-docker-status');

    try {
        const status = await appApi.DockerOllamaStatus();
        if (!status) return;

        // Modèle cible et conteneur
        if (metaModelEl) metaModelEl.textContent = status.target_model || "gemma4:26b";
        if (metaCidEl) metaCidEl.textContent = status.container_id ? status.container_id.substring(0, 12) : "Aucun";

        // État de santé
        if (metaHealthEl) {
            if (status.health === 'healthy') {
                metaHealthEl.innerHTML = '<span class="text-green font-semibold">🟢 Sain (Healthy)</span>';
            } else if (status.health === 'starting') {
                metaHealthEl.innerHTML = '<span class="text-yellow">⏳ Démarrage (Starting)</span>';
            } else if (status.health === 'unhealthy') {
                metaHealthEl.innerHTML = '<span class="text-danger">🔴 Anomalie (Unhealthy)</span>';
            } else {
                metaHealthEl.textContent = status.health || 'Non configuré';
            }
        }

        // État Docker brut
        if (metaStateEl) {
            metaStateEl.textContent = status.state || (status.running ? 'running' : 'exited');
        }

        // Gestion fine de l'état (prêt, téléchargement, démarrage, arrêté)
        if (status.is_pulling) {
            if (pulseEl) pulseEl.className = 'pulse-indicator pulse-pulling';
            if (badgeEl) {
                badgeEl.className = 'badge badge-info';
                badgeEl.textContent = 'Téléchargement du modèle...';
            }
            if (stateIconEl) stateIconEl.textContent = '📥';
            if (statusTextEl) statusTextEl.textContent = status.status_text || 'Téléchargement du modèle en cours...';
            if (subTextEl) subTextEl.textContent = "Le conteneur télécharge les couches du modèle Ollama";
            if (pullBoxEl) pullBoxEl.style.display = 'flex';
            if (pullInfoEl) pullInfoEl.textContent = status.pull_info || "Téléchargement en cours...";
            if (kpiDocker) {
                kpiDocker.textContent = "Téléchargement...";
                kpiDocker.style.color = "#38bdf8";
            }
        } else if (status.is_ready) {
            if (pulseEl) pulseEl.className = 'pulse-indicator pulse-ready';
            if (badgeEl) {
                badgeEl.className = 'badge badge-success';
                badgeEl.textContent = 'Opérationnel & Prêt';
            }
            if (stateIconEl) stateIconEl.textContent = '🟢';
            if (statusTextEl) statusTextEl.textContent = status.status_text || 'Serveur Ollama prêt à répondre';
            if (subTextEl) subTextEl.textContent = "Le démon et le modèle sont chargés et joignables sur http://localhost:11434";
            if (pullBoxEl) pullBoxEl.style.display = 'none';
            if (kpiDocker) {
                kpiDocker.textContent = "Actif (Prêt)";
                kpiDocker.style.color = "#34d399";
            }
        } else if (status.running) {
            if (pulseEl) pulseEl.className = 'pulse-indicator pulse-starting';
            if (badgeEl) {
                badgeEl.className = 'badge badge-warning';
                badgeEl.textContent = 'Initialisation...';
            }
            if (stateIconEl) stateIconEl.textContent = '⏳';
            if (statusTextEl) statusTextEl.textContent = status.status_text || 'Conteneur actif, initialisation du démon...';
            if (subTextEl) subTextEl.textContent = "Attente de la disponibilité de l'API Ollama";
            if (pullBoxEl) pullBoxEl.style.display = 'none';
            if (kpiDocker) {
                kpiDocker.textContent = "Démarrage...";
                kpiDocker.style.color = "#fbbf24";
            }
        } else {
            if (pulseEl) pulseEl.className = 'pulse-indicator pulse-stopped';
            if (badgeEl) {
                badgeEl.className = 'badge badge-danger';
                badgeEl.textContent = 'Arrêté';
            }
            if (stateIconEl) stateIconEl.textContent = '⏹';
            if (statusTextEl) statusTextEl.textContent = status.status_text || 'Conteneur Ollama arrêté';
            if (subTextEl) subTextEl.textContent = "Cliquez sur Démarrer ci-dessous pour lancer l'orchestration";
            if (pullBoxEl) pullBoxEl.style.display = 'none';
            if (kpiDocker) {
                kpiDocker.textContent = "Inactif";
                kpiDocker.style.color = "var(--text-muted)";
            }
        }

        // Si demandé, rafraîchir les logs
        if (shouldFetchLogs) {
            await fetchDockerLogs();
        }
        return status;
    } catch (err) {
        if (badgeEl) {
            badgeEl.className = 'badge badge-danger';
            badgeEl.textContent = 'Erreur Docker';
        }
        if (pulseEl) pulseEl.className = 'pulse-indicator pulse-error';
        if (statusTextEl) statusTextEl.textContent = 'Démon Docker inaccessible ou arrêté';
        if (subTextEl) subTextEl.textContent = err.toString();
        return null;
    }
}

async function fetchDockerLogs() {
    if (!appApi || !appApi.GetDockerLogs) return;
    const terminal = document.getElementById('cfg-docker-logs-terminal');
    const timestampEl = document.getElementById('cfg-docker-logs-timestamp');
    const linesSelect = document.getElementById('cfg-docker-log-lines');
    const autoScroll = document.getElementById('cfg-docker-log-autoscroll')?.checked ?? true;
    const lines = parseInt(linesSelect?.value || "100", 10);

    try {
        const res = await appApi.GetDockerLogs(lines);
        if (res && res.success) {
            if (timestampEl) {
                const now = new Date();
                timestampEl.textContent = `Actualisé à ${now.toLocaleTimeString()}`;
            }

            if (terminal) {
                const logsRaw = res.logs || "";
                if (!logsRaw.trim()) {
                    terminal.textContent = "Aucun log disponible pour le service ollama (conteneur arrêté ou sans sortie).";
                } else {
                    terminal.innerHTML = formatDockerLogsHtml(logsRaw);
                    if (autoScroll) {
                        const wrap = terminal.parentElement;
                        if (wrap) wrap.scrollTop = wrap.scrollHeight;
                    }
                }
            }
        } else if (terminal) {
            terminal.textContent = `Erreur récupération logs : ${res?.error || "Inconnue"}`;
        }
    } catch (err) {
        if (terminal) {
            terminal.textContent = `Erreur communication backend : ${err}`;
        }
    }
}

function formatDockerLogsHtml(raw) {
    const lines = raw.split('\n');
    return lines.map(line => {
        if (!line.trim()) return '';
        const escaped = escapeHtml(line);
        const lower = line.toLowerCase();

        let extraClass = '';
        if (lower.includes('pulling') || lower.includes('downloading') || lower.includes('%') || lower.includes('verifying')) {
            extraClass = ' log-pull';
        } else if (lower.includes('ready') || lower.includes('already present') || lower.includes('success') || lower.includes('health check passed')) {
            extraClass = ' log-ready';
        } else if (lower.includes('error') || lower.includes('failed') || lower.includes('fatal') || lower.includes('erreur')) {
            extraClass = ' log-error';
        }

        const tsMatch = escaped.match(/^(\d{4}-\d{2}-\d{2}T[^\s]+)\s(.*)$/);
        if (tsMatch) {
            return `<span class="log-line${extraClass}"><span class="log-timestamp">${tsMatch[1]}</span>${tsMatch[2]}</span>`;
        }

        return `<span class="log-line${extraClass}">${escaped}</span>`;
    }).join('');
}

async function handleStartDocker() {
    if (!appApi) return;
    const feedback = document.getElementById('cfg-docker-action-feedback');
    if (feedback) feedback.textContent = "Lancement de docker compose up...";
    showToast("Démarrage du conteneur Ollama...", "info");

    try {
        const res = await appApi.StartDockerOllama();
        if (res && res.success) {
            showToast("Conteneur Ollama démarré !", "success");
            if (feedback) feedback.textContent = "Conteneur démarré ! Initialisation en cours...";
            await updateDockerStatusAndLogs(true);
            startActiveDockerWatch();
        } else {
            showToast("Échec démarrage Docker : " + (res?.error || "Vérifier docker compose"), "error");
            if (feedback) feedback.textContent = "Échec du démarrage";
        }
    } catch (err) {
        showToast("Erreur : " + err, "error");
        if (feedback) feedback.textContent = "Erreur : " + err;
    }
}

async function handleRestartDocker() {
    if (!appApi) return;
    const feedback = document.getElementById('cfg-docker-action-feedback');
    if (feedback) feedback.textContent = "Redémarrage du service ollama...";
    showToast("Redémarrage du conteneur Ollama...", "info");

    try {
        const res = await appApi.RestartDockerOllama();
        if (res && res.success) {
            showToast("Conteneur Ollama redémarré !", "success");
            if (feedback) feedback.textContent = "Redémarrage effectué";
            await updateDockerStatusAndLogs(true);
            startActiveDockerWatch();
        } else {
            showToast("Échec redémarrage Docker : " + (res?.error || "Inconnu"), "error");
            if (feedback) feedback.textContent = "Échec redémarrage";
        }
    } catch (err) {
        showToast("Erreur : " + err, "error");
        if (feedback) feedback.textContent = "Erreur : " + err;
    }
}

async function handleStopDocker() {
    if (!appApi) return;
    const feedback = document.getElementById('cfg-docker-action-feedback');
    if (feedback) feedback.textContent = "Arrêt de docker compose down...";
    showToast("Arrêt du conteneur Ollama...", "info");
    stopDockerPolling();

    try {
        const res = await appApi.StopDockerOllama();
        if (res && res.success) {
            showToast("Conteneur Ollama arrêté !", "success");
            if (feedback) feedback.textContent = "Conteneur arrêté";
            await updateDockerStatusAndLogs(false);
        } else {
            showToast("Échec arrêt Docker : " + (res?.error || "Inconnu"), "error");
            if (feedback) feedback.textContent = "Échec arrêt";
        }
    } catch (err) {
        showToast("Erreur : " + err, "error");
        if (feedback) feedback.textContent = "Erreur : " + err;
    }
}

async function handleDockerPing() {
    if (!appApi) return;
    const feedback = document.getElementById('cfg-docker-action-feedback');
    if (feedback) feedback.textContent = "Test HTTP local http://localhost:11434...";

    try {
        const ping = await appApi.PingOllama();
        if (ping && (ping.available || ping.reachable)) {
            showToast(`Ollama local répond (latence : ${ping.latency_ms || 0} ms) !`, "success");
            if (feedback) feedback.textContent = `🟢 Joignable (${ping.latency_ms || 0} ms)`;
        } else {
            showToast("Ollama local ne répond pas encore sur le port 11434", "warning");
            if (feedback) feedback.textContent = "🔴 Non joignable sur 11434";
        }
        await updateDockerStatusAndLogs(true);
    } catch (err) {
        showToast("Erreur test HTTP Docker : " + err, "error");
        if (feedback) feedback.textContent = "Erreur ping : " + err;
    }
}

async function handleSaveConfig() {
    if (!appApi) return;
    const baseUrl = document.getElementById('cfg-base-url')?.value?.trim() || "";
    const model = document.getElementById('cfg-model')?.value?.trim() || "";
    const timeout = parseFloat(document.getElementById('cfg-timeout')?.value || "30");
    const dockerLLM = document.getElementById('cfg-docker-llm')?.checked || false;
    const mockInterface = document.getElementById('cfg-mock-interface')?.checked || false;
    const mockAI = document.getElementById('cfg-mock-ai')?.checked || false;
    const notifications = document.getElementById('cfg-notifications')?.checked ?? true;
    const pyInput = document.getElementById('cfg-python-bin')?.value?.trim() || "";

    try {
        const res = await appApi.SaveConfig(baseUrl, model, timeout, mockInterface, dockerLLM, mockAI);
        if (res && res.success) {
            // Sauvegarde de l'état des notifications si supporté
            if (appApi.SaveNotifications) {
                try {
                    await appApi.SaveNotifications(notifications);
                } catch (e) {
                    console.warn("Échec sauvegarde notifications :", e);
                }
            }

            // Si le champ Python a été modifié par rapport à l'existant, on le valide et l'enregistre
            const currentCfg = await appApi.GetConfig();
            const configuredPy = (currentCfg?.python_bin || "").trim();
            if (pyInput !== configuredPy && appApi.SavePythonBin) {
                const isAuto = pyInput === "";
                const pyRes = await appApi.SavePythonBin(pyInput, isAuto);
                if (!pyRes || !pyRes.success) {
                    showToast("Avertissement Python : " + (pyRes?.error || "Interpréteur non validé"), "warning");
                }
            }

            showToast("Configuration enregistrée dans .env !", "success");
            await loadConfig();
            await refreshDashboard();
        } else {
            showToast("Erreur enregistrement config : " + (res?.error || "Inconnue"), "error");
        }
    } catch (err) {
        showToast("Erreur : " + err, "error");
    }
}

async function handlePingOllama() {
    if (!appApi) return;
    const resultEl = document.getElementById('cfg-ping-result');
    const modelsBox = document.getElementById('cfg-installed-models-box');
    const modelsList = document.getElementById('cfg-installed-models-list');
    if (resultEl) resultEl.textContent = "Test en cours...";

    try {
        const ping = await appApi.PingOllama();
        if (ping && (ping.available || ping.reachable)) {
            const latency = ping.latency_ms || 0;
            if (resultEl) {
                resultEl.style.color = '#34d399';
                resultEl.textContent = `🟢 Joignable (${latency} ms, v${ping.version || 'Ollama'})`;
            }

            // Affichage des modèles installés comme l'option [T] du CLI
            if (ping.installed_models && ping.installed_models.length > 0) {
                if (modelsBox) modelsBox.style.display = 'block';
                if (modelsList) {
                    modelsList.innerHTML = ping.installed_models.map(m => `<span class="model-chip">${escapeHtml(m)}</span>`).join('');
                }
            } else if (modelsBox) {
                modelsBox.style.display = 'none';
            }

            showToast("Serveur Ollama joignable !", "success");
        } else {
            if (resultEl) {
                resultEl.style.color = '#f87171';
                resultEl.textContent = `🔴 Injoignable : ${ping?.error || 'Serveur hors-ligne'}`;
            }
            if (modelsBox) modelsBox.style.display = 'none';
            showToast("Serveur Ollama injoignable", "warning");
        }
    } catch (err) {
        if (resultEl) {
            resultEl.style.color = '#f87171';
            resultEl.textContent = `🔴 Erreur : ${err}`;
        }
        if (modelsBox) modelsBox.style.display = 'none';
    }
}

// --------------------------------------------------------------------------
// 11. UTILITAIRES & NOTIFICATIONS TOAST
// --------------------------------------------------------------------------
function showToast(message, type = 'info', duration = 3500) {
    const container = document.getElementById('toast-container');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;

    let icon = 'ℹ';
    if (type === 'success') icon = '✅';
    if (type === 'error') icon = '❌';
    if (type === 'warning') icon = '⚠️';

    toast.innerHTML = `<span>${icon}</span> <span>${escapeHtml(message)}</span>`;
    container.appendChild(toast);

    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateX(30px)';
        toast.style.transition = 'all 0.3s ease';
        setTimeout(() => toast.remove(), 300);
    }, duration);
}

function escapeHtml(text) {
    if (!text) return '';
    return text
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
}

// Convertisseur Markdown basique et robuste pour les Release Notes
function renderMarkdownBasic(md) {
    if (!md) return '<p>Aucun contenu.</p>';
    const lines = md.split('\n');
    let html = '';
    let inList = false;

    for (let line of lines) {
        line = line.trim();
        if (line.startsWith('# ')) {
            if (inList) { html += '</ul>'; inList = false; }
            html += `<h2>${escapeHtml(line.substring(2))}</h2>`;
        } else if (line.startsWith('## ')) {
            if (inList) { html += '</ul>'; inList = false; }
            html += `<h3>${escapeHtml(line.substring(3))}</h3>`;
        } else if (line.startsWith('### ')) {
            if (inList) { html += '</ul>'; inList = false; }
            html += `<h4>${escapeHtml(line.substring(4))}</h4>`;
        } else if (line.startsWith('- ') || line.startsWith('* ')) {
            if (!inList) { html += '<ul>'; inList = true; }
            const content = escapeHtml(line.substring(2))
                .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
                .replace(/`(.*?)`/g, '<code>$1</code>');
            html += `<li>${content}</li>`;
        } else if (line === '') {
            if (inList) { html += '</ul>'; inList = false; }
        } else {
            if (inList) { html += '</ul>'; inList = false; }
            const content = escapeHtml(line)
                .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
                .replace(/`(.*?)`/g, '<code>$1</code>');
            html += `<p>${content}</p>`;
        }
    }
    if (inList) html += '</ul>';
    return html;
}

// API de secours si prévisualisé directement dans un navigateur sans Wails
function setupFallbackApi() {
    let mockStagedCount = 1;
    let mockUnstagedCount = 2;
    let mockCurrentRepo = "/home/maxence/Documents/DATA/Ecole/but3_info/sae_applicationIntelligente/SAE_gui";
    let mockPythonBin = "/home/maxence/monenv/bin/python";

    let mockCommits = [
        { sha: "f7a3b1c8d9e01234", message: "feat(gui): interface graphique Wails v2 avec mode Mock IA", author: "Maxence Laurence", tag: "v1.0.0", date: "2026-11-24" },
        { sha: "e6b2c0a7d8e90123", message: "refactor(bridge): synchronisation des variables d'environnement", author: "Maxence Laurence", tag: "", date: "2026-11-23" },
        { sha: "d5a1b9f6c7d89012", message: "docs(architecture): mise à jour de la documentation miroir", author: "Enzo Ciuffa", tag: "", date: "2026-11-22" },
        { sha: "c4f0a8e5b6c78901", message: "test(mock): ajout des suites de tests unitaires mock IA", author: "Dorian Hembert", tag: "v0.9.0", date: "2026-11-21" },
        { sha: "b3e9f7d4a5b67890", message: "feat(commit): support du forçage de type Conventional Commits", author: "Maxence Laurence", tag: "", date: "2026-11-20" },
        { sha: "a2d8e6c3f4a56789", message: "fix(service): élimination des sorties parasites sur stdout", author: "Enzo Ciuffa", tag: "", date: "2026-11-19" },
        { sha: "91c7d5b2e3f45678", message: "chore(docker): configuration ollama docker compose", author: "Dorian Hembert", tag: "", date: "2026-11-18" },
        { sha: "80b6c4a1d2e34567", message: "feat(core): initialisation du générateur intelligent", author: "Maxence Laurence", tag: "v0.1.0", date: "2026-11-17" },
    ];

    let mockDiffFiles = [
        { path: "src/gui/app.go", status: "M", added: 42, removed: 8, patch: "@@ -80,12 +80,16 @@ func NewApp()\n- mockActive := forceMockInterface\n+ mockActive := forceMockInterface || ReadMockSettingFromEnv(appRoot)\n+ mockIAActive := forceMockIA || ReadMockAISettingFromEnv(appRoot)\n+ if mockIAActive {\n+     _ = os.Setenv(\"MOCK_AI\", \"true\")\n+ }" },
        { path: "src/gui/frontend/src/main.js", status: "M", added: 85, removed: 12, patch: "@@ -176,7 +176,7 @@ function initWailsConnection()\n- } else if (checkCount++ > 25) {\n+ } else if (checkCount++ > 4) {\n      clearInterval(interval);\n+     setupFallbackApi();" },
        { path: "src/gui/frontend/src/style.css", status: "A", added: 30, removed: 0, patch: "@@ -0,0 +1,30 @@\n+.badge-mock-ai {\n+    background: rgba(56, 189, 248, 0.2);\n+    color: #38bdf8;\n+}" },
    ];

    let mockConfig = {
        ollama_base_url: "http://localhost:11434",
        ollama_model: "gemma4:26b",
        timeout_s: 30,
        docker_llm: false,
        mock_interface: false,
        mock_ai: true,
        enable_notifications: true,
        python_bin: mockPythonBin,
        resolved_python_bin: mockPythonBin
    };

    appApi = {
        GetRepoStatus: async () => ({
            is_git_repo: true,
            repo_path: mockCurrentRepo,
            branch: "feature/gui-wails",
            staged_count: mockStagedCount,
            unstaged_count: mockUnstagedCount
        }),
        GetTargetRepo: async () => mockCurrentRepo,
        SetTargetRepo: async (p) => {
            mockCurrentRepo = p || mockCurrentRepo;
            return { is_git_repo: true, repo_path: mockCurrentRepo, branch: "feature/gui-wails", staged_count: mockStagedCount, unstaged_count: mockUnstagedCount };
        },
        ChooseDirectory: async () => mockCurrentRepo,
        GetDiff: async () => ({
            success: true,
            files: mockDiffFiles
        }),
        StageAll: async () => {
            mockStagedCount += mockUnstagedCount;
            mockUnstagedCount = 0;
            return { success: true, message: "Tous les fichiers ont été indexés (git add -A)." };
        },
        GenerateCommit: async (fb) => {
            const subject = fb ? `adaptation selon retour : ${fb}` : "intégration complète de l'interface graphique Wails avec Mock IA";
            const body = fb
                ? `Prise en compte de la directive utilisateur :\n- ${fb}\n\nValidation locale effectuée en mode simulation Mock IA.`
                : "Synchronisation de tous les modules graphiques avec le pontage Python et support du mode Mock IA.\n\n- src/gui/app.go : enrichissement des méthodes exposées à Wails\n- src/gui/frontend/src/main.js : implémentation de tous les écouteurs interactifs";
            return {
                success: true,
                type: "feat",
                scope: "gui",
                subject: subject,
                body: body,
                raw_formatted: `feat(gui): ${subject}\n\n${body}`
            };
        },
        ImposeCommitType: async (type, comment) => {
            const subject = `mise à jour de l'interface Wails (${type})`;
            const body = comment
                ? `Commentaire appliqué : ${comment}`
                : `Type imposé [${type}] pour le commit selon les règles Conventional Commits.`;
            return {
                success: true,
                type: type,
                scope: "gui",
                subject: subject,
                body: body,
                raw_formatted: `${type}(gui): ${subject}\n\n${body}`
            };
        },
        ApplyCommit: async (msg) => {
            const newSha = Math.random().toString(16).substring(2, 10);
            const firstLine = msg.split('\n')[0] || "commit généré";
            mockCommits.unshift({
                sha: newSha + "a1b2c3d4",
                message: firstLine,
                author: "Maxence Laurence",
                tag: "",
                date: new Date().toISOString().split('T')[0]
            });
            mockStagedCount = 0;
            mockDiffFiles = [];
            return { success: true, sha: newSha };
        },
        Push: async () => ({ success: true, message: "Branches poussées vers le serveur distant avec succès !" }),
        GetReleaseNotes: async (fromTag, toTag) => {
            let count = mockCommits.length;
            if (fromTag && toTag) {
                const idxStart = mockCommits.findIndex(c => c.sha === fromTag);
                const idxEnd = mockCommits.findIndex(c => c.sha === toTag);
                if (idxStart !== -1 && idxEnd !== -1) {
                    count = Math.abs(idxStart - idxEnd) + 1;
                }
            }
            return {
                success: true,
                commit_count: count,
                markdown: `# Release Notes (${(fromTag || "HEAD").substring(0, 7)} ➔ ${(toTag || "HEAD").substring(0, 7)})\n\n## 🚀 Nouvelles Fonctionnalités\n- **feat(gui)**: Interface graphique native Wails v2 avec support complet du Mock IA.\n- **feat(commit)**: Assistant Conventional Commits interactif avec imposition de type.\n\n## 🛠️ Améliorations & Réfactorisation\n- **refactor(bridge)**: Communication fluide en JSON pur via sous-processus Python.\n- **docs**: Documentation miroir MkDocs synchronisée.\n\n## 📋 Résumé des commits (${count} analysés)\n- ${mockCommits.slice(0, count).map(c => `${c.sha.substring(0, 7)} — ${c.message} (${c.author})`).join('\n- ')}`
            };
        },
        GetRecentCommits: async () => ({
            commits: mockCommits
        }),
        GetConfig: async () => ({ ...mockConfig }),
        SaveConfig: async (baseUrl, model, timeout, mockInterface, dockerLLM, mockAI) => {
            mockConfig.ollama_base_url = baseUrl;
            mockConfig.ollama_model = model;
            mockConfig.timeout_s = timeout;
            mockConfig.mock_interface = mockInterface;
            mockConfig.docker_llm = dockerLLM;
            mockConfig.mock_ai = mockAI;
            return { success: true, message: "Configuration enregistrée avec succès." };
        },
        SaveNotifications: async (enabled) => {
            mockConfig.enable_notifications = enabled;
            return { success: true };
        },
        PingOllama: async () => ({
            available: true,
            latency_ms: 14,
            version: "0.6.2",
            installed_models: ["gemma4:26b", "gemma4:12b", "smollm:135m"]
        }),
        SaveReleaseNotesMarkdown: async (filename, content) => {
            console.log(`[Mock Save] Release notes enregistrées dans ${filename} (${content.length} octets).`);
            return null;
        },
        DockerOllamaStatus: async () => ({
            success: true,
            running: true,
            is_ready: true,
            is_pulling: false,
            state: "running",
            health: "healthy",
            status_text: "Opérationnel & Prêt (Modèle gemma4:26b chargé)",
            target_model: "gemma4:26b",
            container_id: "a89d71c43021"
        }),
        StartDockerOllama: async () => ({ success: true, message: "Docker Ollama démarré avec succès." }),
        RestartDockerOllama: async () => ({ success: true, message: "Docker Ollama redémarré avec succès." }),
        StopDockerOllama: async () => ({ success: true, message: "Docker Ollama arrêté avec succès." }),
        GetDockerLogs: async (lines) => ({
            success: true,
            logs: "2026-11-24T18:00:00Z [entrypoint] Démarre le serveur Ollama en arrière-plan...\n2026-11-24T18:00:01Z [entrypoint] Waiting for Ollama server... (1/30)\n2026-11-24T18:00:02Z [ollama] Listening on 0.0.0.0:11434 (schema: http)\n2026-11-24T18:00:03Z [entrypoint] Ollama server is ready.\n2026-11-24T18:00:04Z [entrypoint] Model gemma4:26b already present.\n2026-11-24T18:00:05Z [ollama] Health check passed (HTTP 200 OK)",
            tail_lines: lines,
            running: true
        }),
        GetSystemSpecs: async () => ({
            os: "linux",
            arch: "amd64",
            go_version: "go1.27.0",
            alloc_mb: "14.20 MB",
            sys_mb: "32.50 MB",
            cpus: 8,
            goroutines: 6,
            webview: "WebKitGTK 4.1 (Wails v2)",
            server_time: new Date().toLocaleTimeString(),
            app_version: "1.0.0"
        }),
        VerifyPythonEnvironment: async () => ({
            success: true,
            executable: mockPythonBin,
            version: "3.14.9",
            mode: "Personnalisé (PYTHON_BIN)",
            missing_deps: [],
            app_load_error: ""
        }),
        VerifyPythonCandidate: async (bin, isAuto) => ({
            success: true,
            executable: bin || mockPythonBin,
            version: "3.14.9",
            mode: isAuto ? "Automatique (.venv)" : "Personnalisé",
            missing_deps: [],
            app_load_error: ""
        }),
        SavePythonBin: async (bin, isAuto) => {
            mockPythonBin = isAuto ? "/home/maxence/monenv/bin/python" : bin;
            mockConfig.python_bin = mockPythonBin;
            mockConfig.resolved_python_bin = mockPythonBin;
            return { success: true, message: `Interpréteur Python validé et configuré avec succès (${isAuto ? 'Automatique' : 'Personnalisé'}, Python 3.14.9) !` };
        }
    };
}
