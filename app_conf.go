package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	"github.com/twsnmp/twsnmpfk/notify"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

// GetMapConf returns map config
func (a *App) GetMapConf() datastore.MapConfEnt {
	return datastore.MapConf
}

// UpdateMapConf save map config
func (a *App) UpdateMapConf(m datastore.MapConfEnt) bool {
	datastore.RestartSnmpTrapd = datastore.MapConf.SnmpMode != m.SnmpMode ||
		datastore.MapConf.Community != m.Community ||
		datastore.MapConf.SnmpUser != m.SnmpUser ||
		datastore.MapConf.SnmpPassword != m.SnmpPassword
	if m.MapName != datastore.MapConf.MapName {
		a.setMenu()
	}
	if m.OTelRetention < 1 {
		m.OTelRetention = 24
	}
	if m.MCPEndpoint != datastore.MapConf.MCPEndpoint ||
		m.MCPTransport != datastore.MapConf.MCPEndpoint ||
		m.MCPFrom != datastore.MapConf.MCPFrom ||
		m.MCPToken != datastore.MapConf.MCPToken {
		backend.NotifyMCPConfigChanged()
	}
	datastore.MapConf = m

	return datastore.SaveMapConf() == nil
}

// GetMapName returns map name
func (a *App) GetMapName() string {
	return datastore.MapConf.MapName
}

// GetNotifyConf returns notify config
func (a *App) GetNotifyConf() datastore.NotifyConfEnt {
	return datastore.NotifyConf
}

// UpdateNotifyConf save notify config
func (a *App) UpdateNotifyConf(n datastore.NotifyConfEnt) bool {
	if n.Provider != datastore.NotifyConf.Provider ||
		n.ClientID != datastore.NotifyConf.ClientID ||
		n.ClientSecret != datastore.NotifyConf.ClientSecret ||
		n.MSTenant != datastore.NotifyConf.MSTenant {
		datastore.DeleteNotifyOAuth2Token()
	}
	datastore.NotifyConf = n
	return datastore.SaveNotifyConf() == nil
}

// TestNotifyConf test notfiy
func (a *App) TestNotifyConf(n datastore.NotifyConfEnt) bool {
	return notify.SendTestMail(&n) == nil
}

// TestWebhook test webhook of notify conf
func (a *App) TestWebhook(n datastore.NotifyConfEnt) bool {
	return notify.WebHookTest(&n) == nil
}

func (a *App) GetNotifyOAuth2Token() string {
	url, err := notify.GetNotifyOAuth2TokenStep1()
	if err != nil {
		return err.Error()
	}
	wails.BrowserOpenURL(a.ctx, url)
	err = notify.GetNotifyOAuth2TokenStep2()
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) HasValidNotifyOAuth2Token(n datastore.NotifyConfEnt) bool {
	return datastore.HasValidNotifyOAuth2Token(n)
}

type nottifyOAuth2Info struct {
	HasAccessToken  bool
	HasRefreshToken bool
	Expiry          string
	RedirectURL     string
}

func (a *App) GetNotifyOAuth2Info() nottifyOAuth2Info {
	r := nottifyOAuth2Info{
		RedirectURL: fmt.Sprintf("http://localhost:%d/callback", datastore.NotifyOAuth2RedirectPort),
	}
	token := datastore.GetNotifyOAuth2Token()
	if token == nil {
		return r
	}
	r.HasAccessToken = token.Valid()
	r.HasRefreshToken = token.RefreshToken != ""
	r.Expiry = token.Expiry.Format(time.DateTime)
	return r
}

// GetAIConf returns AI config
func (a *App) GetAIConf() datastore.AIConfEnt {
	return datastore.AIConf
}

// UpdateAIConf save AI config
func (a *App) UpdateAIConf(ai datastore.AIConfEnt) bool {
	datastore.AIConf = ai
	return datastore.SaveAIConf() == nil
}

// GetLocConf returns Loc config
func (a *App) GetLocConf() datastore.LocConfEnt {
	return datastore.LocConf
}

// UpdateLocConf save Loc config
func (a *App) UpdateLocConf(loc datastore.LocConfEnt) bool {
	datastore.LocConf = loc
	return datastore.SaveLocConf() == nil
}

// Backup makes a backup file.
func (a *App) Backup() bool {
	f, err := wails.SaveFileDialog(a.ctx, wails.SaveDialogOptions{
		Title:           i18n.Trans("Backup file"),
		DefaultFilename: "twsnmpfk.db",
		Filters: []wails.FileFilter{{
			DisplayName: "*.db",
			Pattern:     "*.db",
		}},
		CanCreateDirectories: true,
	})
	if err != nil || f == "" {
		return false
	}
	err = datastore.BackupDB(f)
	if err != nil {
		log.Printf("backup database err=%v", err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Backup done"),
	})
	return true
}

func (a *App) GetMIBModules() []*datastore.MIBModuleEnt {
	return datastore.MIBModules
}

// GetIcons returns a list of custom icons.
func (a *App) GetIcons() []*datastore.IconEnt {
	return datastore.GetIcons()
}

// UpdateIcon adds or updates a custom icon.
func (a *App) UpdateIcon(icon datastore.IconEnt) bool {
	if err := datastore.AddOrUpdateIcon(&icon); err != nil {
		log.Printf("UpdateIcon err=%v", err)
		return false
	}
	return true
}

// DeleteIcon deletes a custom icon.
func (a *App) DeleteIcon(icon string) bool {
	if err := datastore.DeleteIcon(icon); err != nil {
		log.Printf("UpdateIcon err=%v", err)
		return false
	}
	return true
}

// GetSshdPublicKeys returns the public keys of hosts allowed to access the SSH server.
func (a *App) GetSshdPublicKeys() string {
	return datastore.GetSshdPublicKeys()
}

// SaveSshdPublicKeys saves the public keys of hosts allowed to access the SSH server.
func (a *App) SaveSshdPublicKeys(pk string) bool {
	if err := datastore.SaveSshdPublicKeys(pk); err != nil {
		log.Printf("SaveSshdPublicKeys err=%v", err)
	}
	return true
}

// GetMySSHPublicKey returns the SSH public key of this application.
func (a *App) GetMySSHPublicKey() string {
	r, err := datastore.GetSSHPublicKey()
	if err != nil {
		log.Printf("GetMySSHPublicKeys err=%v", err)
	}
	return r
}

// InitMySSHKey recreates the SSH private key for this application.
func (a *App) InitMySSHKey() bool {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm init ssh key"),
		Message:       i18n.Trans("Do you want to init?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Init ssh private key"),
	})
	return len(datastore.GenSSHPrivateKey()) > 0
}

var regIconCamelCase = regexp.MustCompile("([A-Z])")

type ImportIconEnt struct {
	Icons  []datastore.IconEnt `json:"Icons"`
	Errors []string            `json:"Errors"`
}

func (a *App) ImportIcon(codeList []datastore.IconEnt) ImportIconEnt {
	filter := []wails.FileFilter{}
	ret := ImportIconEnt{}
	filter = append(filter, wails.FileFilter{
		DisplayName: "TWSNMP ICON File",
		Pattern:     "*.csv;*.txt",
	})
	file, err := wails.OpenFileDialog(a.ctx, wails.OpenDialogOptions{
		Title:   "TWSNMP ICON file",
		Filters: filter,
	})
	if err != nil {
		log.Printf("err=%v", err)
		return ret
	}
	if file == "" {
		return ret
	}
	f, err := os.Open(file)
	if err != nil {
		log.Printf("err=%v", err)
		return ret
	}
	defer f.Close()
	iconToCodeMap := make(map[string]int64)
	for _, e := range codeList {
		iconToCodeMap[e.Icon] = e.Code
	}
	addIconList := []*datastore.IconEnt{}
	scanner := bufio.NewScanner(f)
	sep := " "
	if strings.HasSuffix(strings.ToLower(file), ".csv") {
		sep = ","
	}
	for scanner.Scan() {
		l := scanner.Text()
		a := strings.SplitN(l, sep, 2)
		if len(a) != 2 {
			continue
		}
		i := regIconCamelCase.ReplaceAllStringFunc(strings.TrimSpace(a[0]), func(s string) string {
			return "-" + strings.ToLower(s)
		})
		if !strings.HasPrefix(i, "mdi-") {
			ret.Errors = append(ret.Errors, a[0])
			continue
		}
		code, ok := iconToCodeMap[i]
		if !ok {
			ret.Errors = append(ret.Errors, a[0])
			continue
		}
		addIconList = append(addIconList, &datastore.IconEnt{
			Icon: i,
			Code: code,
			Name: strings.TrimSpace(a[1]),
		})
	}
	if len(ret.Errors) > 0 {
		return ret
	}
	for _, i := range addIconList {
		ret.Icons = append(ret.Icons, *i)
		datastore.AddOrUpdateIcon(i)
	}
	return ret
}

func (a *App) ExportIcons() bool {
	icons := datastore.GetIcons()
	if len(icons) < 1 {
		return false
	}
	file, err := wails.SaveFileDialog(a.ctx, wails.SaveDialogOptions{
		DefaultFilename:      "TWSNMP_ICON.csv",
		CanCreateDirectories: true,
		Filters: []wails.FileFilter{{
			DisplayName: "CSV",
			Pattern:     "*.csv",
		}},
	})
	if err != nil {
		wails.LogErrorf(a.ctx, "exportIcon err=%v", err)
		return false
	}
	if file == "" {
		return false
	}
	f, err := os.Create(file)
	if err != nil {
		wails.LogErrorf(a.ctx, "exportIcon err=%v", err)
		return false
	}
	defer f.Close()
	for _, e := range icons {
		f.WriteString(fmt.Sprintf("%s,%s\n", e.Icon, e.Name))
	}
	return true
}
