package serverupdater

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type ServiceFiles struct {
	SystemdUnit       string `json:"systemdUnit"`
	LaunchdPlist      string `json:"launchdPlist"`
	WindowsPowerShell string `json:"windowsPowerShell"`
}

func BuildServiceFiles(binaryPath, configPath string) ServiceFiles {
	binaryPath = fallbackText(strings.TrimSpace(binaryPath), "/opt/cbmp/server-updater/cbmp-server-updater")
	configPath = fallbackText(strings.TrimSpace(configPath), "/etc/cbmp/server-updater.json")
	return ServiceFiles{
		SystemdUnit: fmt.Sprintf(`[Unit]
Description=CBMP Server Updater
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%s -config %s
Restart=always
RestartSec=10
WorkingDirectory=%s
User=cbmp

[Install]
WantedBy=multi-user.target
`, binaryPath, configPath, filepath.Dir(binaryPath)),
		LaunchdPlist: fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>com.cbmp.server-updater</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>-config</string>
    <string>%s</string>
  </array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>WorkingDirectory</key><string>%s</string>
  <key>StandardOutPath</key><string>/var/log/cbmp-server-updater.log</string>
  <key>StandardErrorPath</key><string>/var/log/cbmp-server-updater.err.log</string>
</dict>
</plist>
`, binaryPath, configPath, filepath.Dir(binaryPath)),
		WindowsPowerShell: fmt.Sprintf(`$Binary = "%s"
$Config = "%s"
sc.exe create CBMPServerUpdater binPath= "$Binary -config $Config" start= auto DisplayName= "CBMP Server Updater"
sc.exe failure CBMPServerUpdater reset= 60 actions= restart/10000/restart/10000/restart/10000
sc.exe start CBMPServerUpdater
`, binaryPath, configPath),
	}
}

func WriteServiceFiles(out io.Writer, binaryPath, configPath string) error {
	files := BuildServiceFiles(binaryPath, configPath)
	_, err := fmt.Fprintf(out, "# Linux systemd: /etc/systemd/system/cbmp-server-updater.service\n%s\n# macOS launchd: /Library/LaunchDaemons/com.cbmp.server-updater.plist\n%s\n# Windows PowerShell\n%s", files.SystemdUnit, files.LaunchdPlist, files.WindowsPowerShell)
	return err
}
