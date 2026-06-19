package clientupdater

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Installer struct {
	cfg Config
}

func NewInstaller(cfg Config) Installer {
	return Installer{cfg: cfg}
}

func downloadArtifact(download UpdatePackageDownload, envelope []byte) (string, []byte, error) {
	fileName := safeName(fallbackText(download.ArtifactFileName, fallbackText(download.FileName, fallbackText(download.Package.ArtifactFileName, download.Package.FileName))))
	if fileName == "" {
		fileName = "artifact.bin"
	}
	if fileName == "package.json" {
		fileName = "artifact-package.json"
	}
	if strings.TrimSpace(download.ArtifactContentBase64) == "" {
		return fileName, envelope, nil
	}
	artifact, err := base64.StdEncoding.DecodeString(download.ArtifactContentBase64)
	if err != nil {
		artifact, err = base64.RawStdEncoding.DecodeString(download.ArtifactContentBase64)
	}
	if err != nil {
		return "", nil, fmt.Errorf("decode artifact content: %w", err)
	}
	return fileName, artifact, nil
}

func (i Installer) Apply(ctx context.Context, task ProductSystemUpdateTask, download UpdatePackageDownload, artifact []byte) (string, error) {
	if i.cfg.TargetComponent != "" && !strings.EqualFold(strings.TrimSpace(task.Component), i.cfg.TargetComponent) {
		return "", fmt.Errorf("%s updater cannot apply %s component", i.cfg.TargetComponent, task.Component)
	}
	artifactFileName, artifactBytes, err := downloadArtifact(download, artifact)
	if err != nil {
		return "", err
	}
	if err := i.verifyArtifact(task, download, artifactBytes); err != nil {
		return "", err
	}
	patchArtifactSHA256 := ""
	if strings.EqualFold(fallbackText(download.Package.PackageType, "full"), "delta") {
		patchArtifactSHA256 = download.ArtifactSHA256
		artifactFileName, artifactBytes, err = i.applyDeltaPatch(task, download, artifactBytes)
		if err != nil {
			return "", err
		}
	}
	component := safeName(task.Component)
	if component == "" {
		return "", fmt.Errorf("task component is required")
	}
	slot, err := i.nextSlot(component)
	if err != nil {
		return "", err
	}
	releaseDir := filepath.Join(i.componentDir(component), "releases", safeName(task.Version)+"-"+slot+"-"+safeName(task.TaskNo))
	if err := os.MkdirAll(releaseDir, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(releaseDir, "package.json"), artifact, 0o644); err != nil {
		return "", err
	}
	artifactPath := filepath.Join(releaseDir, artifactFileName)
	if err := os.WriteFile(artifactPath, artifactBytes, 0o644); err != nil {
		return "", err
	}
	if err := writeJSONFile(filepath.Join(releaseDir, "task.json"), task); err != nil {
		return "", err
	}
	manifest := map[string]any{
		"taskNo":               task.TaskNo,
		"component":            task.Component,
		"version":              task.Version,
		"checksum":             task.Checksum,
		"signature":            task.Signature,
		"downloaded":           download.GeneratedAt,
		"installedAt":          time.Now().Format(time.RFC3339),
		"slot":                 slot,
		"artifactFileName":     artifactFileName,
		"artifactPath":         artifactPath,
		"artifactSha256":       artifactDigest(artifactBytes),
		"patchArtifactSha256":  patchArtifactSHA256,
		"artifactContentType":  download.ArtifactContentType,
		"artifactSizeBytes":    len(artifactBytes),
		"packageType":          fallbackText(download.Package.PackageType, "full"),
		"baseVersion":          download.Package.BaseVersion,
		"deltaAlgorithm":       download.Package.DeltaAlgorithm,
		"targetArtifactSha256": download.Package.TargetArtifactSHA256,
	}
	if err := writeJSONFile(filepath.Join(releaseDir, "manifest.json"), manifest); err != nil {
		return "", err
	}
	if err := i.runHooks(ctx, task, releaseDir, artifactPath, "apply"); err != nil {
		return "", err
	}
	if err := i.promote(component, task.Version, slot, releaseDir, task.TaskNo); err != nil {
		return "", err
	}
	return releaseDir, nil
}

func (i Installer) Rollback(ctx context.Context, task ProductSystemUpdateTask) (string, error) {
	if i.cfg.TargetComponent != "" && !strings.EqualFold(strings.TrimSpace(task.Component), i.cfg.TargetComponent) {
		return "", fmt.Errorf("%s updater cannot rollback %s component", i.cfg.TargetComponent, task.Component)
	}
	component := safeName(task.Component)
	if component == "" {
		return "", fmt.Errorf("task component is required")
	}
	state, err := i.loadState()
	if err != nil {
		return "", err
	}
	current := state.Components[component]
	targetVersion := strings.TrimSpace(task.FromVersion)
	if targetVersion == "" {
		targetVersion = current.PreviousVersion
	}
	if targetVersion == "" {
		return "", fmt.Errorf("no previous version is available for rollback")
	}
	releaseDir := current.PreviousRelease
	if releaseDir == "" || !strings.Contains(filepath.Base(releaseDir), safeName(targetVersion)) {
		found, err := i.findRelease(component, targetVersion)
		if err != nil {
			return "", err
		}
		releaseDir = found
	}
	if releaseDir == "" {
		return "", fmt.Errorf("release for rollback version %s not found", targetVersion)
	}
	if err := i.runHooks(ctx, task, releaseDir, "", "rollback"); err != nil {
		return "", err
	}
	slot := current.ActiveSlot
	if slot == "" {
		slot = "blue"
	}
	if err := i.promote(component, targetVersion, slot, releaseDir, task.TaskNo); err != nil {
		return "", err
	}
	return releaseDir, nil
}

func (i Installer) verifyArtifact(task ProductSystemUpdateTask, download UpdatePackageDownload, artifact []byte) error {
	if !download.Verified {
		return fmt.Errorf("server reported package verification failed")
	}
	if download.Package.ID != task.UpdateID {
		return fmt.Errorf("downloaded package id %d does not match task update id %d", download.Package.ID, task.UpdateID)
	}
	if download.Package.Component != "" && task.Component != "" && download.Package.Component != task.Component {
		return fmt.Errorf("downloaded package component %s does not match task component %s", download.Package.Component, task.Component)
	}
	if download.Package.Version != "" && task.Version != "" && download.Package.Version != task.Version {
		return fmt.Errorf("downloaded package version %s does not match task version %s", download.Package.Version, task.Version)
	}
	if download.Package.Checksum != task.Checksum || download.Package.Signature != task.Signature {
		return fmt.Errorf("downloaded package checksum/signature does not match task")
	}
	signature := strings.ToLower(strings.TrimSpace(task.Signature))
	if !strings.HasPrefix(task.Checksum, "sha256:") || (!strings.HasPrefix(signature, "sig:") && !strings.HasPrefix(signature, "hmac-sha256:") && !strings.HasPrefix(signature, "ed25519:")) {
		return fmt.Errorf("task checksum/signature format is invalid")
	}
	if strings.HasPrefix(signature, "ed25519:") {
		if err := verifyEd25519UpdatePackage(download.Package); err != nil {
			return err
		}
	}
	if i.cfg.StrictChecksum {
		expected := strings.TrimPrefix(task.Checksum, "sha256:")
		if !hex64(expected) {
			return fmt.Errorf("strict checksum requires sha256 hex digest")
		}
		sum := sha256.Sum256(artifact)
		if !bytes.Equal([]byte(hex.EncodeToString(sum[:])), []byte(strings.ToLower(expected))) {
			return fmt.Errorf("downloaded package sha256 mismatch")
		}
	}
	if download.ArtifactSHA256 != "" {
		expected := strings.TrimPrefix(strings.ToLower(download.ArtifactSHA256), "sha256:")
		if !hex64(expected) {
			return fmt.Errorf("download artifact sha256 format is invalid")
		}
		sum := sha256.Sum256(artifact)
		if !bytes.Equal([]byte(hex.EncodeToString(sum[:])), []byte(expected)) {
			return fmt.Errorf("downloaded artifact sha256 mismatch")
		}
	}
	return nil
}

func verifyEd25519UpdatePackage(item UpdatePackage) error {
	publicKey, err := decodeEd25519PublicKey(item.SignaturePublicKey)
	if err != nil {
		return err
	}
	signature, err := decodeEd25519Signature(item.Signature)
	if err != nil {
		return err
	}
	if !ed25519.Verify(publicKey, []byte(updatePackageSignaturePayload(item)), signature) {
		return fmt.Errorf("downloaded package ed25519 signature mismatch")
	}
	return nil
}

func updatePackageSignaturePayload(item UpdatePackage) string {
	return strings.Join([]string{
		item.Version,
		item.Component,
		item.Channel,
		fallbackText(item.PackageType, "full"),
		item.BaseVersion,
		item.DeltaAlgorithm,
		item.Checksum,
		item.ArtifactSHA256,
		item.ArtifactFileName,
		fmt.Sprintf("%d", item.ArtifactSizeBytes),
		item.BaseArtifactSHA256,
		item.TargetArtifactSHA256,
	}, "\n")
}

type copyDeltaPatch struct {
	Algorithm      string        `json:"algorithm"`
	BaseSHA256     string        `json:"baseSha256"`
	TargetSHA256   string        `json:"targetSha256"`
	TargetFileName string        `json:"targetFileName"`
	Ops            []copyDeltaOp `json:"ops"`
}

type copyDeltaOp struct {
	Copy *copyDeltaCopy `json:"copy,omitempty"`
	Data string         `json:"data,omitempty"`
}

type copyDeltaCopy struct {
	Offset int64 `json:"offset"`
	Length int64 `json:"length"`
}

func (i Installer) applyDeltaPatch(task ProductSystemUpdateTask, download UpdatePackageDownload, patchBytes []byte) (string, []byte, error) {
	var patch copyDeltaPatch
	if err := json.Unmarshal(patchBytes, &patch); err != nil {
		return "", nil, fmt.Errorf("decode delta patch: %w", err)
	}
	algorithm := fallbackText(strings.TrimSpace(patch.Algorithm), download.Package.DeltaAlgorithm)
	if algorithm == "" {
		algorithm = "cbmp-copy-v1"
	}
	if algorithm != "cbmp-copy-v1" {
		return "", nil, fmt.Errorf("unsupported delta algorithm %s", algorithm)
	}
	baseBytes, err := i.baseArtifactBytes(task.Component, download.Package.BaseVersion)
	if err != nil {
		return "", nil, err
	}
	baseSHA := firstNonEmptyText(download.Package.BaseArtifactSHA256, patch.BaseSHA256)
	if baseSHA != "" {
		if err := verifySHA256Digest(baseBytes, baseSHA, "base artifact"); err != nil {
			return "", nil, err
		}
	}
	target := make([]byte, 0, len(baseBytes)+len(patchBytes))
	for _, op := range patch.Ops {
		if op.Copy != nil {
			if op.Copy.Offset < 0 || op.Copy.Length < 0 || op.Copy.Offset+op.Copy.Length > int64(len(baseBytes)) {
				return "", nil, fmt.Errorf("delta copy range is invalid")
			}
			start := int(op.Copy.Offset)
			end := int(op.Copy.Offset + op.Copy.Length)
			target = append(target, baseBytes[start:end]...)
		}
		if strings.TrimSpace(op.Data) != "" {
			data, err := decodePatchData(op.Data)
			if err != nil {
				return "", nil, err
			}
			target = append(target, data...)
		}
	}
	targetSHA := firstNonEmptyText(download.Package.TargetArtifactSHA256, patch.TargetSHA256)
	if targetSHA == "" {
		return "", nil, fmt.Errorf("delta package target sha256 is required")
	}
	if err := verifySHA256Digest(target, targetSHA, "target artifact"); err != nil {
		return "", nil, err
	}
	fileName := safeName(firstNonEmptyText(patch.TargetFileName, download.Package.FileName, task.ArtifactFileName))
	if fileName == "" {
		fileName = safeName(task.Component + "-" + task.Version + ".bin")
	}
	return fileName, target, nil
}

func (i Installer) baseArtifactBytes(component, baseVersion string) ([]byte, error) {
	component = safeName(component)
	state, err := i.loadState()
	if err != nil {
		return nil, err
	}
	current := state.Components[component]
	releaseDir := ""
	if current.CurrentRelease != "" && (baseVersion == "" || current.CurrentVersion == baseVersion) {
		releaseDir = current.CurrentRelease
	}
	if releaseDir == "" && baseVersion != "" {
		releaseDir, err = i.findRelease(component, baseVersion)
		if err != nil {
			return nil, err
		}
	}
	if releaseDir == "" {
		return nil, fmt.Errorf("base release for %s %s not found", component, baseVersion)
	}
	manifestPath := filepath.Join(releaseDir, "manifest.json")
	var manifest map[string]any
	if raw, err := os.ReadFile(manifestPath); err == nil {
		_ = json.Unmarshal(raw, &manifest)
	}
	artifactPath, _ := manifest["artifactPath"].(string)
	if artifactPath == "" {
		if artifactFileName, _ := manifest["artifactFileName"].(string); artifactFileName != "" {
			artifactPath = filepath.Join(releaseDir, safeName(artifactFileName))
		}
	}
	if artifactPath == "" {
		return nil, fmt.Errorf("base release artifact metadata is missing")
	}
	raw, err := os.ReadFile(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("read base artifact: %w", err)
	}
	return raw, nil
}

func decodePatchData(value string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(value)
	}
	if err != nil {
		return nil, fmt.Errorf("decode delta data op: %w", err)
	}
	return decoded, nil
}

func verifySHA256Digest(data []byte, expected string, label string) error {
	expected = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(expected)), "sha256:")
	if !hex64(expected) {
		return fmt.Errorf("%s sha256 format is invalid", label)
	}
	if strings.TrimPrefix(artifactDigest(data), "sha256:") != expected {
		return fmt.Errorf("%s sha256 mismatch", label)
	}
	return nil
}

func artifactDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func firstNonEmptyText(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func decodeEd25519PublicKey(value string) (ed25519.PublicKey, error) {
	raw := strings.TrimPrefix(strings.TrimSpace(value), "ed25519:")
	decoded, err := base64.RawStdEncoding.DecodeString(raw)
	if err != nil || len(decoded) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("downloaded package ed25519 public key is invalid")
	}
	return ed25519.PublicKey(decoded), nil
}

func decodeEd25519Signature(value string) ([]byte, error) {
	raw := strings.TrimPrefix(strings.TrimSpace(value), "ed25519:")
	decoded, err := base64.RawStdEncoding.DecodeString(raw)
	if err != nil || len(decoded) != ed25519.SignatureSize {
		return nil, fmt.Errorf("downloaded package ed25519 signature is invalid")
	}
	return decoded, nil
}

func (i Installer) runHooks(ctx context.Context, task ProductSystemUpdateTask, releaseDir, artifactPath, phase string) error {
	hooks := i.cfg.Components[task.Component]
	timeout := time.Duration(hooks.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	for _, command := range [][]string{hooks.StopCommand, hooks.InstallCommand, hooks.StartCommand, hooks.HealthCommand} {
		if len(command) == 0 {
			continue
		}
		runCtx, cancel := context.WithTimeout(ctx, timeout)
		err := runCommand(runCtx, command, map[string]string{
			"CBMP_RELEASE_DIR":   releaseDir,
			"CBMP_COMPONENT":     task.Component,
			"CBMP_VERSION":       task.Version,
			"CBMP_ACTION":        phase,
			"CBMP_TASK_NO":       task.TaskNo,
			"CBMP_ARTIFACT_PATH": artifactPath,
		})
		cancel()
		if err != nil {
			return err
		}
	}
	return nil
}

func runCommand(ctx context.Context, command []string, env map[string]string) error {
	if len(command) == 0 {
		return nil
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Env = os.Environ()
	for key, value := range env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command %q failed: %w: %s", strings.Join(command, " "), err, string(output))
	}
	return nil
}

func (i Installer) promote(component, version, slot, releaseDir, taskNo string) error {
	state, err := i.loadState()
	if err != nil {
		return err
	}
	current := state.Components[component]
	next := ComponentState{
		CurrentVersion:  version,
		PreviousVersion: current.CurrentVersion,
		ActiveSlot:      slot,
		CurrentRelease:  releaseDir,
		PreviousRelease: current.CurrentRelease,
		LastTaskNo:      taskNo,
		UpdatedAt:       time.Now().Format(time.RFC3339),
	}
	state.Components[component] = next
	if err := os.MkdirAll(i.componentDir(component), 0o755); err != nil {
		return err
	}
	if err := writeJSONFile(filepath.Join(i.componentDir(component), "current.json"), next); err != nil {
		return err
	}
	return i.saveState(state)
}

func (i Installer) nextSlot(component string) (string, error) {
	state, err := i.loadState()
	if err != nil {
		return "", err
	}
	if state.Components[component].ActiveSlot == "blue" {
		return "green", nil
	}
	return "blue", nil
}

func (i Installer) findRelease(component, version string) (string, error) {
	root := filepath.Join(i.componentDir(component), "releases")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	prefix := safeName(version)
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix+"-") {
			return filepath.Join(root, entry.Name()), nil
		}
	}
	return "", nil
}

func (i Installer) loadState() (State, error) {
	state := State{Components: map[string]ComponentState{}}
	raw, err := os.ReadFile(i.statePath())
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return state, err
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		return state, err
	}
	if state.Components == nil {
		state.Components = map[string]ComponentState{}
	}
	return state, nil
}

func (i Installer) saveState(state State) error {
	if err := os.MkdirAll(i.cfg.RootDir, 0o755); err != nil {
		return err
	}
	return writeJSONFile(i.statePath(), state)
}

func (i Installer) statePath() string {
	return filepath.Join(i.cfg.RootDir, "state.json")
}

func (i Installer) componentDir(component string) string {
	return filepath.Join(i.cfg.RootDir, "components", component)
}

func writeJSONFile(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o644)
}

var safeNamePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func safeName(value string) string {
	value = strings.TrimSpace(value)
	value = safeNamePattern.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}

func hex64(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, ch := range strings.ToLower(value) {
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
			return false
		}
	}
	return true
}
