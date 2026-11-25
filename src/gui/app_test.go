package main

import (
	"testing"
)

func TestNewAppAndGetters(t *testing.T) {
	// Création d'une application en mode mock
	app := NewApp("/tmp/fake-repo", "/tmp/fake-app", true, true, true, false)

	if app.GetTargetRepo() != "/tmp/fake-repo" {
		t.Errorf("attendu repoRoot '/tmp/fake-repo', obtenu '%s'", app.GetTargetRepo())
	}
	if app.GetAppRoot() != "/tmp/fake-app" {
		t.Errorf("attendu appRoot '/tmp/fake-app', obtenu '%s'", app.GetAppRoot())
	}
	if !app.IsDevMode() {
		t.Errorf("attendu devMode == true")
	}

	specs := app.GetSystemSpecs()
	if specs["os"] == "" || specs["arch"] == "" {
		t.Errorf("specs système incomplètes : %+v", specs)
	}
}

func TestAppGetRepoStatusMock(t *testing.T) {
	app := NewApp(".", "", true, true, true, false)

	status, err := app.GetRepoStatus()
	if err != nil {
		t.Fatalf("GetRepoStatus a renvoyé une erreur inattendue : %v", err)
	}
	if status == nil {
		t.Fatal("status renvoyé est nil")
	}
	if !status.IsGitRepo {
		t.Errorf("attendu is_git_repo == true en mode mock")
	}
}

func TestAppGetDiffMock(t *testing.T) {
	app := NewApp(".", "", true, true, true, false)

	diff, err := app.GetDiff()
	if err != nil {
		t.Fatalf("GetDiff a renvoyé une erreur : %v", err)
	}
	if diff == nil || !diff.Success {
		t.Fatal("diff invalide ou échec")
	}
}

func TestAppStageAllMock(t *testing.T) {
	app := NewApp(".", "", true, true, true, false)

	res, err := app.StageAll()
	if err != nil {
		t.Fatalf("StageAll a échoué : %v", err)
	}
	if res == nil || !res.Success {
		t.Fatal("résultat stage all invalide")
	}
}

func TestAppGenerateCommitMock(t *testing.T) {
	app := NewApp(".", "", true, true, true, false)

	proposal, err := app.GenerateCommit("")
	if err != nil {
		t.Fatalf("GenerateCommit a échoué : %v", err)
	}
	if proposal == nil || !proposal.Success {
		t.Fatal("proposition de commit invalide")
	}
	if proposal.Type == "" || proposal.Subject == "" {
		t.Errorf("type ou sujet vide dans la proposition : %+v", proposal)
	}

	// Avec feedback
	proposalWithFeedback, err := app.GenerateCommit("ajoute plus de détails")
	if err != nil {
		t.Fatalf("GenerateCommit avec feedback a échoué : %v", err)
	}
	if proposalWithFeedback == nil || !proposalWithFeedback.Success {
		t.Fatal("proposition avec feedback invalide")
	}
}

func TestAppApplyCommitAndPushMock(t *testing.T) {
	app := NewApp(".", "", true, true, true, false)

	resCommit, err := app.ApplyCommit("feat(core): commit de test unitaire")
	if err != nil {
		t.Fatalf("ApplyCommit a échoué : %v", err)
	}
	if resCommit == nil || !resCommit.Success {
		t.Fatal("échec de ApplyCommit")
	}

	resPush, err := app.Push()
	if err != nil {
		t.Fatalf("Push a échoué : %v", err)
	}
	if resPush == nil || !resPush.Success {
		t.Fatal("échec de Push")
	}
}

func TestAppReleaseNotesAndRecentCommitsMock(t *testing.T) {
	app := NewApp(".", "", true, true, true, false)

	rn, err := app.GetReleaseNotes("v1.0.0", "HEAD")
	if err != nil {
		t.Fatalf("GetReleaseNotes a échoué : %v", err)
	}
	if rn == nil || !rn.Success {
		t.Fatal("échec GetReleaseNotes")
	}
	if rn.Markdown == "" {
		t.Errorf("markdown des release notes vide")
	}

	commits, err := app.GetRecentCommits(5)
	if err != nil {
		t.Fatalf("GetRecentCommits a échoué : %v", err)
	}
	if commits == nil || !commits.Success {
		t.Fatal("échec GetRecentCommits")
	}
}

func TestAppConfigAndPingMock(t *testing.T) {
	tempDir := t.TempDir()
	app := NewApp(tempDir, tempDir, true, true, true, false)

	cfg, err := app.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig a échoué : %v", err)
	}
	if cfg == nil {
		t.Fatal("cfg est nil")
	}

	ping, err := app.PingOllama()
	if err != nil {
		t.Fatalf("PingOllama a échoué : %v", err)
	}
	if ping == nil {
		t.Fatal("ping est nil")
	}

	// Test sauvegarde de la configuration (6 arguments stricts)
	resSave, errSaveCfg := app.SaveConfig("http://localhost:11434", "gemma4:12b", 45.0, true, false, true)
	if errSaveCfg != nil {
		t.Fatalf("SaveConfig a échoué : %v", errSaveCfg)
	}
	if resSave == nil || !resSave.Success {
		t.Fatalf("SaveConfig retour invalide : %v", resSave)
	}

	// Test sauvegarde des notifications
	resNotif, errNotif := app.SaveNotifications(true)
	if errNotif != nil {
		t.Fatalf("SaveNotifications a échoué : %v", errNotif)
	}
	if resNotif == nil || !resNotif.Success {
		t.Fatalf("SaveNotifications retour invalide : %v", resNotif)
	}

	// Test sauvegarde fichier markdown
	errSave := app.SaveReleaseNotesMarkdown("TEST_RN.md", "# Test Notes")
	if errSave != nil {
		t.Fatalf("SaveReleaseNotesMarkdown a échoué : %v", errSave)
	}
}

func TestAppSetTargetRepo(t *testing.T) {
	app := NewApp(".", "", true, true, true, false)

	// Échec sur chemin vide
	_, err := app.SetTargetRepo("")
	if err == nil {
		t.Errorf("attendu erreur sur SetTargetRepo avec chaîne vide")
	}

	// Succès sur nouveau chemin
	status, err := app.SetTargetRepo(t.TempDir())
	if err != nil {
		t.Fatalf("SetTargetRepo a échoué : %v", err)
	}
	if status == nil {
		t.Fatal("status renvoyé par SetTargetRepo est nil")
	}
}

func TestAppDockerOperationsMock(t *testing.T) {
	app := NewApp(".", "", true, true, true, false)

	status, err := app.DockerOllamaStatus()
	if err != nil {
		t.Fatalf("DockerOllamaStatus a échoué : %v", err)
	}
	if status == nil {
		t.Fatal("status docker est nil")
	}

	// En mode mock, Start et Stop renvoient un succès sans manipuler de conteneur réel
	startRes, err := app.StartDockerOllama()
	if err != nil || startRes == nil || !startRes.Success {
		t.Errorf("StartDockerOllama invalide en mode mock : %v", err)
	}

	// Test redémarrage simulé
	restartRes, err := app.RestartDockerOllama()
	if err != nil || restartRes == nil || !restartRes.Success {
		t.Errorf("RestartDockerOllama invalide en mode mock : %v", err)
	}

	// Test récupération des logs simulés
	logsRes, err := app.GetDockerLogs(50)
	if err != nil || logsRes == nil || !logsRes.Success {
		t.Errorf("GetDockerLogs invalide en mode mock : %v", err)
	}
	if logsRes != nil && logsRes.Logs == "" {
		t.Errorf("GetDockerLogs ne devrait pas être vide")
	}

	stopRes, err := app.StopDockerOllama()
	if err != nil || stopRes == nil || !stopRes.Success {
		t.Errorf("StopDockerOllama invalide en mode mock : %v", err)
	}
}

func TestAppRealPythonEnvironment(t *testing.T) {
	app := NewApp(".", "", true, false, true, false)
	status, err := app.GetRepoStatus()
	t.Logf("GetRepoStatus: status=%+v, err=%v", status, err)
	if err != nil {
		t.Fatalf("Erreur GetRepoStatus réel : %v", err)
	}
	diff, err := app.GetDiff()
	t.Logf("GetDiff: diff=%+v, err=%v", diff, err)
	if err != nil {
		t.Fatalf("Erreur GetDiff réel : %v", err)
	}

	// Test du diagnostic Python [P] du CLI
	report, errReport := app.VerifyPythonEnvironment()
	if errReport != nil {
		t.Fatalf("VerifyPythonEnvironment a échoué : %v", errReport)
	}
	if report == nil || !report.Success {
		t.Fatalf("Rapport Python non valide : %+v", report)
	}
	t.Logf("Python validé : %s (%s)", report.Executable, report.Version)

	// Test de vérification de candidat
	candReport, errCand := app.VerifyPythonCandidate(report.Executable, false)
	if errCand != nil || candReport == nil || !candReport.Success {
		t.Fatalf("VerifyPythonCandidate a échoué sur exécutable valide : %v", errCand)
	}

	// Test d'imposer un type Conventional Commit [T]
	proposal, errProp := app.ImposeCommitType("fix", "correction du bug d'indexation")
	if errProp != nil {
		t.Fatalf("ImposeCommitType a échoué : %v", errProp)
	}
	if proposal == nil {
		t.Fatal("proposal nil")
	}
}
