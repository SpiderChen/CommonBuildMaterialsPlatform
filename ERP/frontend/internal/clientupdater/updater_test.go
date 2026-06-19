package clientupdater

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunnerApplyTaskDownloadsInstallsAndReports(t *testing.T) {
	token := "probe-site-token"
	reports := []ReportRequest{}
	task := ProductSystemUpdateTask{
		ID: 1, TaskNo: "SU-1", UpdateID: 7, InstanceID: 3, CustomerName: "现场客户", Watermark: "CBMP-SITE",
		Component: "client", Version: "1.2.0", FromVersion: "1.1.0", Action: "apply", Status: "queued",
		ArtifactFileName: "cbmp-client-1.2.0.json", Checksum: "sha256:site-client-120", Signature: "sig:site-client-120",
		DownloadURL: "/api/system/updates/7/download",
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-CBMP-Updater-Token") != token {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		switch {
		case r.URL.Path == "/api/product-ops/system-updates/tasks":
			_ = json.NewEncoder(w).Encode(PollResponse{Accepted: true, Instance: ProductInstance{ID: 3, Watermark: "CBMP-SITE"}, Tasks: []ProductSystemUpdateTask{task}})
		case r.URL.Path == "/api/system/updates/7/download":
			_ = json.NewEncoder(w).Encode(UpdatePackageDownload{
				FileName: "cbmp-client-1.2.0.json", ContentType: "application/json", Verified: true, GeneratedAt: "2026-06-19 12:00:00",
				Package: UpdatePackage{ID: 7, Version: "1.2.0", Component: "client", Checksum: task.Checksum, Signature: task.Signature, FileName: task.ArtifactFileName},
			})
		case strings.HasPrefix(r.URL.Path, "/api/product-ops/system-updates/tasks/SU-1/report"):
			var report ReportRequest
			if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
				t.Fatalf("decode report: %v", err)
			}
			reports = append(reports, report)
			_ = json.NewEncoder(w).Encode(task)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	root := t.TempDir()
	runner, err := NewRunner(Config{BaseURL: server.URL, UpdaterToken: token, RootDir: root, PollIntervalSeconds: 1}, nil)
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	if err := runner.RunOnce(context.Background()); err != nil {
		t.Fatalf("run once: %v", err)
	}
	stateRaw, err := os.ReadFile(filepath.Join(root, "state.json"))
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	var state State
	if err := json.Unmarshal(stateRaw, &state); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	if state.Components["client"].CurrentVersion != "1.2.0" || state.Components["client"].CurrentRelease == "" {
		t.Fatalf("unexpected state: %+v", state)
	}
	if _, err := os.Stat(filepath.Join(state.Components["client"].CurrentRelease, "package.json")); err != nil {
		t.Fatalf("expected installed package: %v", err)
	}
	if len(reports) < 3 || reports[len(reports)-1].Status != "succeeded" || reports[len(reports)-1].CurrentVersion != "1.2.0" {
		t.Fatalf("unexpected reports: %+v", reports)
	}
}

func TestInstallerRollbackPromotesPreviousRelease(t *testing.T) {
	root := t.TempDir()
	installer := NewInstaller(Config{RootDir: root})
	task := ProductSystemUpdateTask{TaskNo: "SU-2", UpdateID: 8, Component: "server", Version: "2.0.0", FromVersion: "1.9.0", Action: "apply", Checksum: "sha256:server-200", Signature: "sig:server-200"}
	download := UpdatePackageDownload{Verified: true, Package: UpdatePackage{ID: 8, Version: "2.0.0", Component: "server", Checksum: task.Checksum, Signature: task.Signature}}
	if _, err := installer.Apply(context.Background(), ProductSystemUpdateTask{TaskNo: "SU-1", UpdateID: 6, Component: "server", Version: "1.9.0", Action: "apply", Checksum: "sha256:server-190", Signature: "sig:server-190"}, UpdatePackageDownload{Verified: true, Package: UpdatePackage{ID: 6, Version: "1.9.0", Component: "server", Checksum: "sha256:server-190", Signature: "sig:server-190"}}, []byte(`{"package":{"version":"1.9.0"}}`)); err != nil {
		t.Fatalf("apply previous: %v", err)
	}
	if _, err := installer.Apply(context.Background(), task, download, []byte(`{"package":{"version":"2.0.0"}}`)); err != nil {
		t.Fatalf("apply current: %v", err)
	}
	if _, err := installer.Rollback(context.Background(), ProductSystemUpdateTask{TaskNo: "SU-3", Component: "server", FromVersion: "1.9.0", Action: "rollback"}); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	state, err := installer.loadState()
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if state.Components["server"].CurrentVersion != "1.9.0" {
		t.Fatalf("expected rollback to 1.9.0: %+v", state.Components["server"])
	}
}

func TestInstallerWritesDownloadedArtifactFileAndStrictChecksum(t *testing.T) {
	root := t.TempDir()
	artifact := []byte("real client package bytes")
	sum := sha256.Sum256(artifact)
	checksum := "sha256:" + hex.EncodeToString(sum[:])
	task := ProductSystemUpdateTask{
		TaskNo:           "SU-ARTIFACT",
		UpdateID:         31,
		Component:        "client",
		Version:          "3.1.0",
		Action:           "apply",
		ArtifactFileName: "cbmp-client-3.1.0.pkg",
		Checksum:         checksum,
		Signature:        "sig:client-310",
	}
	download := UpdatePackageDownload{
		FileName:              "cbmp-client-3.1.0.pkg",
		ContentType:           "application/json",
		Verified:              true,
		GeneratedAt:           "2026-06-19 14:00:00",
		ArtifactFileName:      "cbmp-client-3.1.0.pkg",
		ArtifactContentType:   "application/octet-stream",
		ArtifactSizeBytes:     int64(len(artifact)),
		ArtifactSHA256:        checksum,
		ArtifactContentBase64: base64.StdEncoding.EncodeToString(artifact),
		Package: UpdatePackage{
			ID: 31, Version: "3.1.0", Component: "client", Checksum: checksum, Signature: task.Signature,
			FileName: "cbmp-client-3.1.0.pkg", ArtifactFileName: "cbmp-client-3.1.0.pkg", ArtifactSHA256: checksum, ArtifactSizeBytes: int64(len(artifact)),
		},
	}
	envelope, err := json.Marshal(download)
	if err != nil {
		t.Fatalf("marshal download envelope: %v", err)
	}
	installer := NewInstaller(Config{RootDir: root, StrictChecksum: true})
	releaseDir, err := installer.Apply(context.Background(), task, download, envelope)
	if err != nil {
		t.Fatalf("apply artifact: %v", err)
	}
	installed, err := os.ReadFile(filepath.Join(releaseDir, "cbmp-client-3.1.0.pkg"))
	if err != nil {
		t.Fatalf("read installed artifact: %v", err)
	}
	if !bytes.Equal(installed, artifact) {
		t.Fatalf("installed artifact mismatch")
	}
	if _, err := os.Stat(filepath.Join(releaseDir, "package.json")); err != nil {
		t.Fatalf("expected envelope package.json: %v", err)
	}
	var manifest map[string]interface{}
	rawManifest, err := os.ReadFile(filepath.Join(releaseDir, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if err := json.Unmarshal(rawManifest, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if manifest["artifactSha256"] != checksum || manifest["artifactFileName"] != "cbmp-client-3.1.0.pkg" {
		t.Fatalf("unexpected artifact manifest: %+v", manifest)
	}
}

func TestInstallerAppliesCopyDeltaPatchFromCurrentRelease(t *testing.T) {
	root := t.TempDir()
	baseArtifact := []byte("hello world")
	baseChecksum := artifactDigest(baseArtifact)
	baseTask := ProductSystemUpdateTask{
		TaskNo: "SU-BASE", UpdateID: 51, Component: "server", Version: "1.0.0", Action: "apply",
		ArtifactFileName: "cbmp-server-1.0.0.tar.gz", Checksum: baseChecksum, Signature: "sig:server-100",
	}
	baseDownload := UpdatePackageDownload{
		FileName: "cbmp-server-1.0.0.tar.gz", ContentType: "application/json", Verified: true, GeneratedAt: "2026-06-19 16:00:00",
		ArtifactFileName: "cbmp-server-1.0.0.tar.gz", ArtifactContentType: "application/gzip", ArtifactSizeBytes: int64(len(baseArtifact)),
		ArtifactSHA256: baseChecksum, ArtifactContentBase64: base64.StdEncoding.EncodeToString(baseArtifact),
		Package: UpdatePackage{
			ID: 51, Version: "1.0.0", Component: "server", Channel: "stable", PackageType: "full",
			Checksum: baseChecksum, Signature: baseTask.Signature, FileName: "cbmp-server-1.0.0.tar.gz",
			ArtifactFileName: "cbmp-server-1.0.0.tar.gz", ArtifactSHA256: baseChecksum, ArtifactSizeBytes: int64(len(baseArtifact)),
		},
	}
	baseEnvelope, err := json.Marshal(baseDownload)
	if err != nil {
		t.Fatalf("marshal base envelope: %v", err)
	}
	installer := NewInstaller(Config{RootDir: root, StrictChecksum: true})
	if _, err := installer.Apply(context.Background(), baseTask, baseDownload, baseEnvelope); err != nil {
		t.Fatalf("apply base artifact: %v", err)
	}

	targetArtifact := []byte("hello updater")
	targetChecksum := artifactDigest(targetArtifact)
	patch, err := json.Marshal(copyDeltaPatch{
		Algorithm:      "cbmp-copy-v1",
		BaseSHA256:     baseChecksum,
		TargetSHA256:   targetChecksum,
		TargetFileName: "cbmp-server-1.1.0.tar.gz",
		Ops: []copyDeltaOp{
			{Copy: &copyDeltaCopy{Offset: 0, Length: 6}},
			{Data: base64.StdEncoding.EncodeToString([]byte("updater"))},
		},
	})
	if err != nil {
		t.Fatalf("marshal delta patch: %v", err)
	}
	patchChecksum := artifactDigest(patch)
	deltaTask := ProductSystemUpdateTask{
		TaskNo: "SU-DELTA", UpdateID: 52, Component: "server", Version: "1.1.0", FromVersion: "1.0.0", Action: "apply",
		ArtifactFileName: "cbmp-server-1.1.0.patch.json", Checksum: patchChecksum, Signature: "sig:server-110-delta",
	}
	deltaDownload := UpdatePackageDownload{
		FileName: "cbmp-server-1.1.0.patch.json", ContentType: "application/json", Verified: true, GeneratedAt: "2026-06-19 16:05:00",
		ArtifactFileName: "cbmp-server-1.1.0.patch.json", ArtifactContentType: "application/vnd.cbmp.delta+json", ArtifactSizeBytes: int64(len(patch)),
		ArtifactSHA256: patchChecksum, ArtifactContentBase64: base64.StdEncoding.EncodeToString(patch),
		Package: UpdatePackage{
			ID: 52, Version: "1.1.0", Component: "server", Channel: "stable", PackageType: "delta",
			BaseVersion: "1.0.0", DeltaAlgorithm: "cbmp-copy-v1", BaseArtifactSHA256: baseChecksum, TargetArtifactSHA256: targetChecksum,
			Checksum: patchChecksum, Signature: deltaTask.Signature, FileName: "cbmp-server-1.1.0.patch.json",
			ArtifactFileName: "cbmp-server-1.1.0.patch.json", ArtifactSHA256: patchChecksum, ArtifactSizeBytes: int64(len(patch)),
		},
	}
	deltaEnvelope, err := json.Marshal(deltaDownload)
	if err != nil {
		t.Fatalf("marshal delta envelope: %v", err)
	}
	releaseDir, err := installer.Apply(context.Background(), deltaTask, deltaDownload, deltaEnvelope)
	if err != nil {
		t.Fatalf("apply delta artifact: %v", err)
	}
	installed, err := os.ReadFile(filepath.Join(releaseDir, "cbmp-server-1.1.0.tar.gz"))
	if err != nil {
		t.Fatalf("read installed delta target: %v", err)
	}
	if !bytes.Equal(installed, targetArtifact) {
		t.Fatalf("installed delta target mismatch: %q", string(installed))
	}
	var manifest map[string]interface{}
	rawManifest, err := os.ReadFile(filepath.Join(releaseDir, "manifest.json"))
	if err != nil {
		t.Fatalf("read delta manifest: %v", err)
	}
	if err := json.Unmarshal(rawManifest, &manifest); err != nil {
		t.Fatalf("decode delta manifest: %v", err)
	}
	if manifest["packageType"] != "delta" || manifest["artifactSha256"] != targetChecksum || manifest["patchArtifactSha256"] != patchChecksum {
		t.Fatalf("unexpected delta manifest: %+v", manifest)
	}
}

func TestInstallerVerifiesEd25519SignedUpdatePackage(t *testing.T) {
	root := t.TempDir()
	artifact := []byte("offline signed server package")
	sum := sha256.Sum256(artifact)
	checksum := "sha256:" + hex.EncodeToString(sum[:])
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate ed25519 key: %v", err)
	}
	pkg := UpdatePackage{
		ID: 41, Version: "4.1.0", Component: "server", Channel: "stable", Checksum: checksum,
		FileName: "cbmp-server-4.1.0.tar.gz", ArtifactFileName: "cbmp-server-4.1.0.tar.gz",
		ArtifactSHA256: checksum, ArtifactSizeBytes: int64(len(artifact)),
		SignaturePublicKey: "ed25519:" + base64.RawStdEncoding.EncodeToString(publicKey),
	}
	pkg.Signature = "ed25519:" + base64.RawStdEncoding.EncodeToString(ed25519.Sign(privateKey, []byte(updatePackageSignaturePayload(pkg))))
	task := ProductSystemUpdateTask{
		TaskNo: "SU-ED25519", UpdateID: pkg.ID, Component: pkg.Component, Version: pkg.Version, Action: "apply",
		ArtifactFileName: pkg.ArtifactFileName, Checksum: pkg.Checksum, Signature: pkg.Signature,
	}
	download := UpdatePackageDownload{
		FileName: pkg.FileName, ContentType: "application/json", Verified: true, GeneratedAt: "2026-06-19 15:00:00",
		ArtifactFileName: pkg.ArtifactFileName, ArtifactContentType: "application/gzip", ArtifactSizeBytes: int64(len(artifact)),
		ArtifactSHA256: checksum, ArtifactContentBase64: base64.StdEncoding.EncodeToString(artifact), Package: pkg,
	}
	envelope, err := json.Marshal(download)
	if err != nil {
		t.Fatalf("marshal ed25519 envelope: %v", err)
	}
	installer := NewInstaller(Config{RootDir: root, StrictChecksum: true})
	if _, err := installer.Apply(context.Background(), task, download, envelope); err != nil {
		t.Fatalf("apply ed25519 package: %v", err)
	}

	tampered := download
	tampered.Package.Signature = "ed25519:" + base64.RawStdEncoding.EncodeToString(make([]byte, ed25519.SignatureSize))
	tamperedTask := task
	tamperedTask.Signature = tampered.Package.Signature
	if _, err := installer.Apply(context.Background(), tamperedTask, tampered, envelope); err == nil {
		t.Fatalf("expected tampered ed25519 package to be rejected")
	}
}

func TestClientDownloadTaskResumesPartialFile(t *testing.T) {
	token := "probe-site-token"
	task := ProductSystemUpdateTask{
		TaskNo: "SU-RESUME", UpdateID: 11, Component: "client", Version: "1.3.0",
		Checksum: "sha256:site-client-130", Signature: "sig:site-client-130", DownloadURL: "/api/system/updates/11/download",
	}
	body, err := json.Marshal(UpdatePackageDownload{
		FileName: "cbmp-client-1.3.0.json", ContentType: "application/json", Verified: true, GeneratedAt: "2026-06-19 13:00:00",
		Package: UpdatePackage{ID: 11, Version: "1.3.0", Component: "client", Checksum: task.Checksum, Signature: task.Signature, FileName: "cbmp-client-1.3.0.json"},
	})
	if err != nil {
		t.Fatalf("marshal download: %v", err)
	}
	root := t.TempDir()
	partDir := filepath.Join(root, "downloads", safeName(task.TaskNo))
	if err := os.MkdirAll(partDir, 0o755); err != nil {
		t.Fatalf("mkdir part dir: %v", err)
	}
	offset := len(body) / 2
	if err := os.WriteFile(filepath.Join(partDir, "package.json.part"), body[:offset], 0o644); err != nil {
		t.Fatalf("write part: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-CBMP-Updater-Token") != token {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/api/system/updates/11/download" {
			http.NotFound(w, r)
			return
		}
		wantRange := fmt.Sprintf("bytes=%d-", offset)
		if r.Header.Get("Range") != wantRange {
			t.Fatalf("expected range %s, got %s", wantRange, r.Header.Get("Range"))
		}
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", offset, len(body)-1, len(body)))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(body[offset:])
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, UpdaterToken: token, RootDir: root, ResumeDownloads: true})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	download, raw, err := client.DownloadTask(context.Background(), task)
	if err != nil {
		t.Fatalf("resume download: %v", err)
	}
	if download.Package.ID != task.UpdateID || !bytes.Equal(raw, body) {
		t.Fatalf("unexpected resumed download: %+v raw=%s", download, string(raw))
	}
	if _, err := os.Stat(filepath.Join(partDir, "package.json")); err != nil {
		t.Fatalf("expected final cache file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(partDir, "package.json.part")); !os.IsNotExist(err) {
		t.Fatalf("expected part file removed, err=%v", err)
	}
}

func TestRunnerAutoRollbackOnApplyFailure(t *testing.T) {
	token := "probe-site-token"
	root := t.TempDir()
	installer := NewInstaller(Config{RootDir: root})
	if _, err := installer.Apply(context.Background(), ProductSystemUpdateTask{TaskNo: "SU-OLD", UpdateID: 21, Component: "server", Version: "1.0.0", Action: "apply", Checksum: "sha256:server-100", Signature: "sig:server-100"}, UpdatePackageDownload{Verified: true, Package: UpdatePackage{ID: 21, Version: "1.0.0", Component: "server", Checksum: "sha256:server-100", Signature: "sig:server-100"}}, []byte(`{"package":{"version":"1.0.0"}}`)); err != nil {
		t.Fatalf("seed previous release: %v", err)
	}
	task := ProductSystemUpdateTask{
		ID: 1, TaskNo: "SU-ROLLBACK", UpdateID: 22, InstanceID: 3, CustomerName: "现场客户", Watermark: "CBMP-SITE",
		Component: "server", Version: "2.0.0", FromVersion: "1.0.0", Action: "apply", Status: "queued",
		ArtifactFileName: "cbmp-server-2.0.0.json", Checksum: "sha256:server-200", Signature: "sig:server-200",
		DownloadURL: "/api/system/updates/22/download",
	}
	reports := []ReportRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-CBMP-Updater-Token") != token {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		switch {
		case r.URL.Path == "/api/product-ops/system-updates/tasks":
			_ = json.NewEncoder(w).Encode(PollResponse{Accepted: true, Instance: ProductInstance{ID: 3, Watermark: "CBMP-SITE"}, Tasks: []ProductSystemUpdateTask{task}})
		case r.URL.Path == "/api/system/updates/22/download":
			_ = json.NewEncoder(w).Encode(UpdatePackageDownload{
				FileName: "cbmp-server-2.0.0.json", ContentType: "application/json", Verified: true, GeneratedAt: "2026-06-19 13:30:00",
				Package: UpdatePackage{ID: 22, Version: "2.0.0", Component: "server", Checksum: task.Checksum, Signature: task.Signature, FileName: task.ArtifactFileName},
			})
		case strings.HasPrefix(r.URL.Path, "/api/product-ops/system-updates/tasks/SU-ROLLBACK/report"):
			var report ReportRequest
			if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
				t.Fatalf("decode report: %v", err)
			}
			reports = append(reports, report)
			_ = json.NewEncoder(w).Encode(task)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	cfg := Config{
		BaseURL: server.URL, UpdaterToken: token, RootDir: root, PollIntervalSeconds: 1, AutoRollbackOnFailure: true,
		Components: map[string]ComponentConfig{"server": {HealthCommand: []string{"sh", "-c", "test \"$CBMP_ACTION\" = rollback || exit 1"}, TimeoutSeconds: 5}},
	}
	runner, err := NewRunner(cfg, nil)
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	if err := runner.RunOnce(context.Background()); err != nil {
		t.Fatalf("run once with auto rollback: %v", err)
	}
	if len(reports) == 0 || reports[len(reports)-1].Status != "rolled_back" || reports[len(reports)-1].CurrentVersion != "1.0.0" {
		t.Fatalf("expected rolled_back final report: %+v", reports)
	}
	state, err := installer.loadState()
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if state.Components["server"].CurrentVersion != "1.0.0" {
		t.Fatalf("expected state rolled back to 1.0.0: %+v", state.Components["server"])
	}
}
