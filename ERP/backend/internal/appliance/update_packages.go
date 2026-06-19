package appliance

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

func normalizeUpdatePackage(req UpdatePackage, actor string) (UpdatePackage, error) {
	req.Version = strings.TrimSpace(req.Version)
	req.Component = strings.ToLower(strings.TrimSpace(req.Component))
	req.Channel = strings.ToLower(strings.TrimSpace(req.Channel))
	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	req.PackageType = strings.ToLower(strings.TrimSpace(req.PackageType))
	req.BaseVersion = strings.TrimSpace(req.BaseVersion)
	req.DeltaAlgorithm = strings.ToLower(strings.TrimSpace(req.DeltaAlgorithm))
	req.Checksum = strings.TrimSpace(req.Checksum)
	req.Signature = strings.TrimSpace(req.Signature)
	req.SignaturePublicKey = strings.TrimSpace(req.SignaturePublicKey)
	req.SignatureKeyFingerprint = strings.TrimSpace(req.SignatureKeyFingerprint)
	req.FileName = strings.TrimSpace(req.FileName)
	req.ArtifactFileName = strings.TrimSpace(req.ArtifactFileName)
	req.ArtifactContentType = strings.TrimSpace(req.ArtifactContentType)
	req.ArtifactContentBase64 = strings.TrimSpace(req.ArtifactContentBase64)
	req.ArtifactSHA256 = strings.TrimSpace(req.ArtifactSHA256)
	req.BaseArtifactSHA256 = strings.TrimSpace(req.BaseArtifactSHA256)
	req.TargetArtifactSHA256 = strings.TrimSpace(req.TargetArtifactSHA256)
	req.RollbackVersion = strings.TrimSpace(req.RollbackVersion)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.Version == "" {
		return req, fmt.Errorf("更新包版本不能为空")
	}
	if req.Component == "" {
		req.Component = "server"
	}
	switch req.Component {
	case "client", "server", "all":
	default:
		return req, fmt.Errorf("更新包组件必须是 client、server 或 all")
	}
	if req.Channel == "" {
		req.Channel = "stable"
	}
	if req.Status == "" {
		req.Status = "available"
	}
	if req.PackageType == "" {
		req.PackageType = "full"
	}
	switch req.PackageType {
	case "full":
	case "delta":
		if req.BaseVersion == "" {
			return req, fmt.Errorf("差分更新包必须包含 baseVersion")
		}
		if req.DeltaAlgorithm == "" {
			req.DeltaAlgorithm = "cbmp-copy-v1"
		}
		if req.DeltaAlgorithm != "cbmp-copy-v1" {
			return req, fmt.Errorf("差分更新算法必须是 cbmp-copy-v1")
		}
		if req.ArtifactContentBase64 == "" {
			return req, fmt.Errorf("差分更新包必须包含 artifactContentBase64 patch")
		}
		if _, ok := strictUpdateChecksum(req.TargetArtifactSHA256); !ok {
			return req, fmt.Errorf("差分更新包必须包含 targetArtifactSha256")
		}
		if req.BaseArtifactSHA256 != "" {
			if _, ok := strictUpdateChecksum(req.BaseArtifactSHA256); !ok {
				return req, fmt.Errorf("baseArtifactSha256 必须使用 sha256:<64位hex>")
			}
		}
	default:
		return req, fmt.Errorf("更新包类型必须是 full 或 delta")
	}
	if err := normalizeUpdateArtifact(&req); err != nil {
		return req, err
	}
	if req.ArtifactContentBase64 != "" && (req.Signature == "" || strings.HasPrefix(req.Signature, "sig:")) {
		signed, err := signUpdatePackage(req)
		if err != nil {
			return req, err
		}
		req = signed
	} else if strings.HasPrefix(strings.ToLower(req.Signature), "ed25519:") && req.SignaturePublicKey == "" {
		if privateKey, ok, err := updateSigningPrivateKey(); err != nil {
			return req, err
		} else if ok {
			req.SignaturePublicKey = encodeUpdatePublicKey(privateKey.Public().(ed25519.PublicKey))
		}
	}
	if req.SignaturePublicKey != "" {
		req.SignatureKeyFingerprint = updatePublicKeyFingerprint(req.SignaturePublicKey)
	}
	if req.Checksum == "" || req.Signature == "" {
		return req, fmt.Errorf("更新包必须包含 checksum 和 signature")
	}
	if !updatePackageVerified(req) {
		return req, fmt.Errorf("更新包验签失败")
	}
	now := nowString()
	if req.CreatedAt == "" {
		req.CreatedAt = now
	}
	if req.PublishedAt == "" {
		req.PublishedAt = now
	}
	if req.PublishedBy == "" {
		req.PublishedBy = actor
	}
	if req.FileName == "" {
		req.FileName = updatePackageFileName(req)
	}
	if req.ArtifactFileName == "" {
		req.ArtifactFileName = req.FileName
	}
	return req, nil
}

func updatePackageFileName(item UpdatePackage) string {
	component := item.Component
	if component == "" {
		component = "server"
	}
	return fmt.Sprintf("cbmp-%s-update-%s.json", component, item.Version)
}

func normalizeUpdateArtifact(req *UpdatePackage) error {
	if req.ArtifactContentBase64 == "" {
		if req.ArtifactFileName == "" {
			req.ArtifactFileName = req.FileName
		}
		if req.ArtifactContentType == "" {
			req.ArtifactContentType = "application/json"
		}
		return nil
	}
	artifact, err := decodeUpdateArtifact(req.ArtifactContentBase64)
	if err != nil {
		return fmt.Errorf("更新包 artifactContentBase64 无法解码")
	}
	digest := updateArtifactSHA256(artifact)
	req.ArtifactSHA256 = digest
	req.ArtifactSizeBytes = int64(len(artifact))
	if req.SizeBytes <= 0 {
		req.SizeBytes = req.ArtifactSizeBytes
	}
	if req.FileName == "" {
		req.FileName = updatePackageFileName(*req)
	}
	if req.ArtifactFileName == "" {
		req.ArtifactFileName = req.FileName
	}
	if req.ArtifactContentType == "" {
		req.ArtifactContentType = "application/octet-stream"
	}
	if req.Checksum == "" {
		req.Checksum = digest
	}
	if expected, ok := strictUpdateChecksum(req.Checksum); ok && expected != strings.TrimPrefix(digest, "sha256:") {
		return fmt.Errorf("更新包 checksum 与 artifact 内容不一致")
	}
	if _, ok := strictUpdateChecksum(req.Checksum); !ok {
		return fmt.Errorf("真实更新包 checksum 必须使用 sha256:<64位hex>")
	}
	return nil
}

func decodeUpdateArtifact(value string) ([]byte, error) {
	if value == "" {
		return nil, nil
	}
	if out, err := base64.StdEncoding.DecodeString(value); err == nil {
		return out, nil
	}
	return base64.RawStdEncoding.DecodeString(value)
}

func updateArtifactSHA256(artifact []byte) string {
	sum := sha256.Sum256(artifact)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func strictUpdateChecksum(value string) (string, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	digest := strings.TrimPrefix(value, "sha256:")
	if len(digest) != 64 {
		return "", false
	}
	for _, ch := range digest {
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
			return "", false
		}
	}
	return digest, strings.HasPrefix(value, "sha256:")
}

func signUpdatePackage(item UpdatePackage) (UpdatePackage, error) {
	if privateKey, ok, err := updateSigningPrivateKey(); err != nil {
		return item, err
	} else if ok {
		item.SignaturePublicKey = encodeUpdatePublicKey(privateKey.Public().(ed25519.PublicKey))
		item.SignatureKeyFingerprint = updatePublicKeyFingerprint(item.SignaturePublicKey)
		payload := []byte(updatePackageSignaturePayload(item))
		item.Signature = "ed25519:" + base64.RawStdEncoding.EncodeToString(ed25519.Sign(privateKey, payload))
		return item, nil
	}
	item.SignaturePublicKey = ""
	item.SignatureKeyFingerprint = ""
	item.Signature = signUpdatePackageHMAC(item)
	return item, nil
}

func signUpdatePackageHMAC(item UpdatePackage) string {
	mac := hmac.New(sha256.New, []byte(updateSigningSecret()))
	_, _ = mac.Write([]byte(updatePackageSignaturePayload(item)))
	return "hmac-sha256:" + hex.EncodeToString(mac.Sum(nil))
}

func updatePackageVerified(item UpdatePackage) bool {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(item.Signature)), "hmac-sha256:") {
		expected := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(item.Signature)), "hmac-sha256:")
		if _, ok := strictUpdateChecksum("sha256:" + expected); !ok {
			return false
		}
		return hmac.Equal([]byte(expected), []byte(strings.TrimPrefix(signUpdatePackageHMAC(item), "hmac-sha256:")))
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(item.Signature)), "ed25519:") {
		publicKey, err := decodeUpdatePublicKey(item.SignaturePublicKey)
		if err != nil {
			return false
		}
		signature, err := decodeUpdateSignature(item.Signature)
		if err != nil {
			return false
		}
		return ed25519.Verify(publicKey, []byte(updatePackageSignaturePayload(item)), signature)
	}
	return checksumVerified(item.Checksum, item.Signature)
}

func updatePackageSignaturePayload(item UpdatePackage) string {
	return strings.Join([]string{
		item.Version,
		item.Component,
		item.Channel,
		fallback(item.PackageType, "full"),
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

func updateSigningSecret() string {
	if value := strings.TrimSpace(os.Getenv("CBMP_UPDATE_SIGNING_SECRET")); value != "" {
		return value
	}
	return "cbmp-demo-update-signing-secret"
}

func updateSigningPrivateKey() (ed25519.PrivateKey, bool, error) {
	value := strings.TrimSpace(os.Getenv("CBMP_UPDATE_SIGNING_PRIVATE_KEY"))
	if value == "" {
		return nil, false, nil
	}
	raw := strings.TrimPrefix(value, "ed25519:")
	decoded, err := base64.RawStdEncoding.DecodeString(raw)
	if err != nil {
		return nil, true, fmt.Errorf("更新包 Ed25519 私钥无效")
	}
	switch len(decoded) {
	case ed25519.PrivateKeySize:
		return ed25519.PrivateKey(decoded), true, nil
	case ed25519.SeedSize:
		return ed25519.NewKeyFromSeed(decoded), true, nil
	default:
		return nil, true, fmt.Errorf("更新包 Ed25519 私钥长度无效")
	}
}

func encodeUpdatePublicKey(publicKey ed25519.PublicKey) string {
	return "ed25519:" + base64.RawStdEncoding.EncodeToString(publicKey)
}

func decodeUpdatePublicKey(value string) (ed25519.PublicKey, error) {
	raw := strings.TrimPrefix(strings.TrimSpace(value), "ed25519:")
	decoded, err := base64.RawStdEncoding.DecodeString(raw)
	if err != nil || len(decoded) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("更新包 Ed25519 公钥无效")
	}
	return ed25519.PublicKey(decoded), nil
}

func decodeUpdateSignature(value string) ([]byte, error) {
	raw := strings.TrimPrefix(strings.TrimSpace(value), "ed25519:")
	decoded, err := base64.RawStdEncoding.DecodeString(raw)
	if err != nil || len(decoded) != ed25519.SignatureSize {
		return nil, fmt.Errorf("更新包 Ed25519 签名无效")
	}
	return decoded, nil
}

func updatePublicKeyFingerprint(value string) string {
	publicKey, err := decodeUpdatePublicKey(value)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(publicKey)
	return hex.EncodeToString(sum[:8])
}

func sameUpdatePackage(a, b UpdatePackage) bool {
	return a.Version == b.Version && a.Component == b.Component && a.Channel == b.Channel
}

func sanitizeUpdatePackageForResponse(item UpdatePackage) UpdatePackage {
	item.ArtifactContentBase64 = ""
	return item
}

func sanitizeUpdatePackagesForResponse(items []UpdatePackage) []UpdatePackage {
	out := make([]UpdatePackage, len(items))
	for i, item := range items {
		out[i] = sanitizeUpdatePackageForResponse(item)
	}
	return out
}

func mergeUpdatePackageArtifact(next, existing UpdatePackage) UpdatePackage {
	if next.ArtifactContentBase64 != "" || next.ArtifactSHA256 != "" || next.ArtifactSizeBytes > 0 {
		return next
	}
	next.ArtifactFileName = existing.ArtifactFileName
	next.ArtifactContentType = existing.ArtifactContentType
	next.ArtifactContentBase64 = existing.ArtifactContentBase64
	next.ArtifactSHA256 = existing.ArtifactSHA256
	next.ArtifactSizeBytes = existing.ArtifactSizeBytes
	if next.SizeBytes <= 0 {
		next.SizeBytes = existing.SizeBytes
	}
	return next
}
