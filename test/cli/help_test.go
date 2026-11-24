package cli_test

import (
	"bytes"
	"flag"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"sae-git-cli/flags"
	"sae-git-cli/ui"
	"sae-git-cli/version"
)

// TestVersionPackage vérifie les constantes et le formatage du package dédié version
func TestVersionPackage(t *testing.T) {
	if version.Number != "1.0.0" {
		t.Errorf("attendu version 1.0.0, obtenu %s", version.Number)
	}
	if version.AppName != "Gitelligence" {
		t.Errorf("attendu AppName Gitelligence, obtenu %s", version.AppName)
	}
	if version.BinaryName != "git-generator" {
		t.Errorf("attendu BinaryName git-generator, obtenu %s", version.BinaryName)
	}
	vStr := version.String()
	if !strings.Contains(vStr, "Gitelligence (git-generator) version 1.0.0") {
		t.Errorf("version.String() inattendu : %s", vStr)
	}
}

// TestFlagsPackage vérifie l'initialisation des pointeurs et des valeurs par défaut dans flags
func TestFlagsPackage(t *testing.T) {
	if flags.Dev == nil || *flags.Dev != false {
		t.Errorf("flags.Dev inattendu : attendu false, obtenu %v", flags.Dev)
	}
	if flags.MockInterface == nil || *flags.MockInterface != false {
		t.Errorf("flags.MockInterface inattendu : attendu false, obtenu %v", flags.MockInterface)
	}
	if flags.MockIA == nil || *flags.MockIA != false {
		t.Errorf("flags.MockIA inattendu : attendu false, obtenu %v", flags.MockIA)
	}
	if flags.Repo == nil || *flags.Repo != "." {
		t.Errorf("flags.Repo inattendu : attendu '.', obtenu %v", flags.Repo)
	}
	if flags.AppDir == nil || *flags.AppDir != "" {
		t.Errorf("flags.AppDir inattendu : attendu '', obtenu %v", flags.AppDir)
	}
}

// TestGeneralHelpContent vérifie que l'aide contient toutes les informations requises
func TestGeneralHelpContent(t *testing.T) {
	helpText := ui.GetGeneralHelp()
	rawText := ui.StripANSI(helpText)

	// 1. Nom, binaire et version
	requiredHeaders := []string{
		"Gitelligence",
		"git-generator",
		"v" + ui.AppVersion,
		"Assistant Git intelligent",
	}
	for _, req := range requiredHeaders {
		if !strings.Contains(rawText, req) {
			t.Errorf("L'aide devrait contenir '%s'", req)
		}
	}

	// 2. Sections obligatoires au format standard Linux
	requiredSections := []string{
		"Utilisation :",
		"Options :",
		"Initialiser l'environnement Python (.venv) :",
		"Variables d'environnement (.env) :",
	}
	for _, sec := range requiredSections {
		if !strings.Contains(rawText, sec) {
			t.Errorf("L'aide devrait contenir la section '%s'", sec)
		}
	}

	// 3. Tutoriel d'environnement Python
	expectedPythonCmds := []string{
		"python scripts/setup_python_env.py",
		"source .venv/bin/activate",
	}
	for _, pyCmd := range expectedPythonCmds {
		if !strings.Contains(rawText, pyCmd) {
			t.Errorf("L'aide devrait contenir la commande d'environnement '%s'", pyCmd)
		}
	}

	// 4. Drapeaux CLI et valeurs par défaut
	expectedFlags := []struct {
		flag string
		def  string
	}{
		{"--repo", "."},
		{"--app-dir", "auto-détecté"},
		{"--dev", "false"},
		{"--mockInterface", "false"},
		{"--mockIA", "false"},
		{"--help", ""},
		{"--version", ""},
	}
	for _, f := range expectedFlags {
		if !strings.Contains(rawText, f.flag) {
			t.Errorf("L'aide devrait documenter le drapeau '%s'", f.flag)
		}
		if f.def != "" && !strings.Contains(rawText, f.def) {
			t.Errorf("L'aide devrait documenter la valeur par défaut '%s' pour '%s'", f.def, f.flag)
		}
	}

	// 5. Variables d'environnement et valeurs par défaut
	expectedEnvVars := []struct {
		name string
		def  string
	}{
		{"OLLAMA_BASE_URL", "http://10.22.28.190:11434"},
		{"OLLAMA_MODEL", "gemma4:26b"},
		{"OLLAMA_TIMEOUT_S", "300"},
		{"OLLAMA_NUM_CTX", "32768"},
		{"ENABLE_NOTIFICATIONS", "true"},
		{"PYTHON_BIN", "auto-détecté"},
		{"MOCK_INTERFACE", "false"},
		{"MOCK_AI", "false"},
	}
	for _, ev := range expectedEnvVars {
		if !strings.Contains(rawText, ev.name) {
			t.Errorf("L'aide devrait documenter la variable d'environnement '%s'", ev.name)
		}
		if !strings.Contains(rawText, ev.def) {
			t.Errorf("L'aide devrait documenter la valeur par défaut '%s' pour '%s'", ev.def, ev.name)
		}
	}
}

// TestPrintHelpOutput valide le comportement de PrintHelp et PrintVersion
func TestPrintHelpOutput(t *testing.T) {
	var buf bytes.Buffer
	ui.PrintHelp(&buf)
	if !strings.Contains(buf.String(), "Gitelligence") {
		t.Errorf("PrintHelp devrait afficher l'aide générale")
	}

	buf.Reset()
	ui.PrintVersion(&buf)
	if !strings.Contains(buf.String(), ui.AppVersion) {
		t.Errorf("PrintVersion devrait afficher le numéro de version")
	}
}

// TestHelpAndVersionDetection vérifie la détection robuste des drapeaux d'aide et de version
func TestHelpAndVersionDetection(t *testing.T) {
	tests := []struct {
		args     []string
		wantHelp bool
	}{
		{args: []string{}, wantHelp: false},
		{args: []string{"--help"}, wantHelp: true},
		{args: []string{"-h"}, wantHelp: true},
		{args: []string{"-help"}, wantHelp: true},
		{args: []string{"help"}, wantHelp: true},
		{args: []string{"--repo", ".", "--help"}, wantHelp: true},
		{args: []string{"--dev", "-h"}, wantHelp: true},
		{args: []string{"--repo", "."}, wantHelp: false},
		{args: []string{"--dev"}, wantHelp: false},
		{args: []string{"--repo", "help"}, wantHelp: false},
		{args: []string{"--app-dir", "help"}, wantHelp: false},
		{args: []string{"--repo=help"}, wantHelp: false},
		{args: []string{"--app-dir=help"}, wantHelp: false},
		{args: []string{"--repo", "help", "--help"}, wantHelp: true},
	}

	for _, tt := range tests {
		gotHelp := ui.IsHelpRequested(tt.args)
		if gotHelp != tt.wantHelp {
			t.Errorf("IsHelpRequested(%v): gotHelp=%v, want=%v", tt.args, gotHelp, tt.wantHelp)
		}
	}

	// Tests de détection de version
	if !ui.IsVersionRequested([]string{"-v"}) {
		t.Errorf("IsVersionRequested(['-v']) attendu true")
	}
	if !ui.IsVersionRequested([]string{"--version"}) {
		t.Errorf("IsVersionRequested(['--version']) attendu true")
	}
	if ui.IsVersionRequested([]string{"--repo", "."}) {
		t.Errorf("IsVersionRequested sans flag version attendu false")
	}
	if ui.IsVersionRequested([]string{"--repo", "-v"}) {
		t.Errorf("IsVersionRequested avec '--repo -v' attendu false")
	}
}

// TestBinaryHelpEndToEnd exécute le binaire compilé pour s'assurer du code 0 et de la non-activation de la TUI
func TestBinaryHelpEndToEnd(t *testing.T) {
	cliDir, err := filepath.Abs("../../src/cli")
	if err != nil {
		t.Fatalf("impossible de résoudre src/cli : %v", err)
	}

	cases := []struct {
		name       string
		args       []string
		wantExit   int
		shouldFind []string
	}{
		{
			name:       "Global --help",
			args:       []string{"run", ".", "--help"},
			wantExit:   0,
			shouldFind: []string{"Gitelligence", "Utilisation :", "Options :", "setup_python_env.py"},
		},
		{
			name:       "Global -h",
			args:       []string{"run", ".", "-h"},
			wantExit:   0,
			shouldFind: []string{"Gitelligence", "Variables d'environnement (.env) :"},
		},
		{
			name:       "Global help",
			args:       []string{"run", ".", "help"},
			wantExit:   0,
			shouldFind: []string{"Gitelligence", "Options :"},
		},
		{
			name:       "Version --version",
			args:       []string{"run", ".", "--version"},
			wantExit:   0,
			shouldFind: []string{"Gitelligence", ui.AppVersion},
		},
		{
			name:       "Version -v",
			args:       []string{"run", ".", "-v"},
			wantExit:   0,
			shouldFind: []string{"Gitelligence", ui.AppVersion},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("go", tc.args...)
			cmd.Dir = cliDir

			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out

			runErr := cmd.Run()
			exitCode := 0
			if runErr != nil {
				if exitErr, ok := runErr.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					t.Fatalf("échec de l'exécution de la commande : %v", runErr)
				}
			}

			if exitCode != tc.wantExit {
				t.Errorf("code de sortie inattendu : obtenu %d, attendu %d (sortie : %s)", exitCode, tc.wantExit, out.String())
			}

			outClean := ui.StripANSI(out.String())
			for _, exp := range tc.shouldFind {
				if !strings.Contains(outClean, exp) {
					t.Errorf("texte attendu non trouvé '%s' dans la sortie : %s", exp, outClean)
				}
			}
		})
	}
}

// TestDynamicFlagRegistration vérifie que l'ajout d'une option via flag.CommandLine
// est automatiquement répercuté dans l'aide et reconnu par IsHelpRequested sans modification manuelle.
func TestDynamicFlagRegistration(t *testing.T) {
	// Déclaration dynamique d'une option
	_ = flag.String("dyn-test-flag", "valeur123", "Option de test dynamique")

	helpText := ui.GetGeneralHelp()
	rawText := ui.StripANSI(helpText)

	if !strings.Contains(rawText, "--dyn-test-flag") {
		t.Errorf("L'aide devrait contenir le drapeau dynamique '--dyn-test-flag'")
	}
	if !strings.Contains(rawText, "valeur123") {
		t.Errorf("L'aide devrait contenir la valeur par défaut 'valeur123'")
	}
	if !strings.Contains(rawText, "Option de test dynamique") {
		t.Errorf("L'aide devrait contenir la description de l'option dynamique")
	}

	// Vérifie que la détection d'aide dynamique ignore la valeur d'une option dynamique
	if ui.IsHelpRequested([]string{"--dyn-test-flag", "help"}) {
		t.Errorf("IsHelpRequested(['--dyn-test-flag', 'help']) devrait renvoyer false")
	}
	if !ui.IsHelpRequested([]string{"--dyn-test-flag", "help", "--help"}) {
		t.Errorf("IsHelpRequested(['--dyn-test-flag', 'help', '--help']) devrait renvoyer true")
	}
}
