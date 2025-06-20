package main

import (
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
	if m.MCPEndpoint != datastore.MapConf.MCPEndpoint || m.MCPTransport != datastore.MapConf.MCPEndpoint {
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
	datastore.NotifyConf = n
	return datastore.SaveNotifyConf() == nil
}

// TestNotifyConf test notfiy
func (a *App) TestNotifyConf(n datastore.NotifyConfEnt) bool {
	return notify.SendTestMail(&n) == nil
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
