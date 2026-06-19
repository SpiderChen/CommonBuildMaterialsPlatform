package appliance

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
)

type UpdatePackageDownload struct {
	FileName              string            `json:"fileName"`
	ContentType           string            `json:"contentType"`
	Verified              bool              `json:"verified"`
	GeneratedAt           string            `json:"generatedAt"`
	ArtifactFileName      string            `json:"artifactFileName"`
	ArtifactContentType   string            `json:"artifactContentType"`
	ArtifactSizeBytes     int64             `json:"artifactSizeBytes"`
	ArtifactSHA256        string            `json:"artifactSha256"`
	ArtifactContentBase64 string            `json:"artifactContentBase64"`
	Manifest              map[string]string `json:"manifest"`
	Package               UpdatePackage     `json:"package"`
}

func buildUpdatePackageDownload(item UpdatePackage) UpdatePackageDownload {
	_, contentBase64, contentType, artifactSHA256, artifactSize := updatePackageArtifact(item)
	fileName := fallback(item.ArtifactFileName, fallback(item.FileName, updatePackageFileName(item)))
	return UpdatePackageDownload{
		FileName:              fileName,
		ContentType:           "application/json",
		Verified:              updatePackageVerified(item) && artifactVerified(item, artifactSHA256),
		GeneratedAt:           nowString(),
		ArtifactFileName:      fileName,
		ArtifactContentType:   contentType,
		ArtifactSizeBytes:     artifactSize,
		ArtifactSHA256:        artifactSHA256,
		ArtifactContentBase64: contentBase64,
		Manifest: map[string]string{
			"id":                      numberString(item.ID),
			"version":                 item.Version,
			"component":               item.Component,
			"channel":                 item.Channel,
			"packageType":             fallback(item.PackageType, "full"),
			"baseVersion":             item.BaseVersion,
			"deltaAlgorithm":          item.DeltaAlgorithm,
			"checksum":                item.Checksum,
			"signature":               item.Signature,
			"signaturePublicKey":      item.SignaturePublicKey,
			"signatureKeyFingerprint": item.SignatureKeyFingerprint,
			"baseArtifactSha256":      item.BaseArtifactSHA256,
			"targetArtifactSha256":    item.TargetArtifactSHA256,
		},
		Package: sanitizeUpdatePackageForResponse(item),
	}
}

func updatePackageArtifact(item UpdatePackage) ([]byte, string, string, string, int64) {
	contentType := fallback(item.ArtifactContentType, "application/json")
	if item.ArtifactContentBase64 != "" {
		artifact, err := decodeUpdateArtifact(item.ArtifactContentBase64)
		if err == nil {
			digest := updateArtifactSHA256(artifact)
			return artifact, base64.StdEncoding.EncodeToString(artifact), contentType, digest, int64(len(artifact))
		}
		return nil, item.ArtifactContentBase64, contentType, item.ArtifactSHA256, item.ArtifactSizeBytes
	}
	artifact := synthesizeUpdateArtifact(item)
	digest := updateArtifactSHA256(artifact)
	return artifact, base64.StdEncoding.EncodeToString(artifact), "application/json", digest, int64(len(artifact))
}

func artifactVerified(item UpdatePackage, artifactSHA256 string) bool {
	if item.ArtifactContentBase64 == "" {
		return true
	}
	if item.ArtifactSHA256 != "" && item.ArtifactSHA256 != artifactSHA256 {
		return false
	}
	if expected, ok := strictUpdateChecksum(item.Checksum); ok {
		return expected == artifactSHA256[len("sha256:"):]
	}
	return false
}

func synthesizeUpdateArtifact(item UpdatePackage) []byte {
	payload := map[string]interface{}{
		"id":              item.ID,
		"version":         item.Version,
		"component":       item.Component,
		"channel":         item.Channel,
		"status":          item.Status,
		"packageType":     fallback(item.PackageType, "full"),
		"baseVersion":     item.BaseVersion,
		"deltaAlgorithm":  item.DeltaAlgorithm,
		"checksum":        item.Checksum,
		"signature":       item.Signature,
		"targetSha256":    item.TargetArtifactSHA256,
		"fileName":        fallback(item.FileName, updatePackageFileName(item)),
		"rollbackVersion": item.RollbackVersion,
		"publishedBy":     item.PublishedBy,
		"publishedAt":     item.PublishedAt,
		"remark":          item.Remark,
	}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return []byte("{}")
	}
	raw = append(raw, '\n')
	return raw
}

func numberString(value int64) string {
	if value == 0 {
		return ""
	}
	return strconv.FormatInt(value, 10)
}
