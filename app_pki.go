package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	"github.com/twsnmp/twsnmpfk/pki"
)

// IsCAValid: CAが構築済みかを返す
func (a *App) IsCAValid() bool {
	return pki.IsCAValid()
}

func (a *App) GetDefaultCreateCAReq() datastore.CreateCAReq {
	ret := datastore.CreateCAReq{
		RootCAKeyType: datastore.PKIConf.RootCAKeyType,
		Name:          datastore.PKIConf.Name,
		SANs:          datastore.PKIConf.SANs,
		AcmeBaseURL:   datastore.PKIConf.AcmeBaseURL,
		AcmePort:      datastore.PKIConf.AcmePort,
		HttpBaseURL:   datastore.PKIConf.HttpBaseURL,
		HttpPort:      datastore.PKIConf.HttpPort,
		RootCATerm:    datastore.PKIConf.RootCATerm,
		CrlInterval:   datastore.PKIConf.CrlInterval,
		CertTerm:      datastore.PKIConf.CertTerm,
	}
	return ret
}

func (a *App) CreateCA(req datastore.CreateCAReq) string {
	if a.IsCAValid() {
		return i18n.Trans("CA is already valid")
	}
	if err := pki.CreateCA(req); err != nil {
		return fmt.Sprintf(i18n.Trans("Can not create CA err=%v"), err)
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:  time.Now().UnixNano(),
		Type:  "ca",
		Level: "info",
		Event: i18n.Trans("Create CA"),
	})
	return ""
}

func (a *App) DestroyCA() {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm destroy CA"),
		Message:       i18n.Trans("Do you want to destroy CA?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return
	}
	pki.DestroyCA()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:  time.Now().UnixNano(),
		Type:  "ca",
		Level: "info",
		Event: i18n.Trans("Destroy CA"),
	})
}

type CertEnt struct {
	Status  string `json:"Status"`
	ID      string `json:"ID"`
	Subject string `json:"Subject"`
	Node    string `json:"Node"`
	Created int64  `json:"Created"`
	Revoked int64  `json:"Revoked"`
	Expire  int64  `json:"Expire"`
	Type    string `json:"Type"`
}

// GetCerts: 証明書のリストを返す
func (a *App) GetCerts() []*CertEnt {
	ret := []*CertEnt{}
	now := time.Now().UnixNano()
	datastore.ForEachCert(func(c *datastore.CertEnt) bool {
		status := "valid"
		if c.Revoked > 0 {
			status = "revoked"
		} else if c.Expire < now {
			status = "expired"
		}
		node := ""
		if c.NodeID != "" {
			if n := datastore.GetNode(c.NodeID); n != nil {
				node = n.Name
			}
		}
		ret = append(ret, &CertEnt{
			Status:  status,
			ID:      c.ID,
			Subject: c.Subject,
			Node:    node,
			Created: c.Created,
			Revoked: c.Revoked,
			Expire:  c.Expire,
			Type:    c.Type,
		})
		return true
	})
	return ret
}

func (a *App) CreateCertificateRequest(req pki.CSRReqEnt) string {
	fileBase := strings.ReplaceAll(req.CommonName, " ", "_") + ".csr"
	file, err := wails.SaveFileDialog(a.ctx, wails.SaveDialogOptions{
		DefaultFilename:      fileBase,
		CanCreateDirectories: true,
		Filters: []wails.FileFilter{{
			DisplayName: "Certificate Request File",
			Pattern:     "*.csr",
		}},
	})
	if err != nil || file == "" {
		return ""
	}
	if err := pki.CreateCertificateRequest(&req, file); err != nil {
		log.Printf("create CSR err=%v", err)
		return fmt.Sprintf(i18n.Trans("Can not create CSR err=%v"), err)
	}
	return ""
}

func (a *App) CreateCertificate() string {
	file, err := wails.OpenFileDialog(a.ctx, wails.OpenDialogOptions{
		Title: "Certificate Request File",
		Filters: []wails.FileFilter{{
			DisplayName: "Certificate Request File",
			Pattern:     "*.csr",
		}},
	})
	if err != nil || file == "" {
		return ""
	}
	b, err := os.ReadFile(file)
	if err != nil {
		log.Printf("read CSR err=%v", err)
		return fmt.Sprintf(i18n.Trans("Can not create Certificate err=%v"), err)
	}
	if err := pki.CreateCertificate(b, file); err != nil {
		log.Printf("create Certificate err=%v", err)
		return fmt.Sprintf(i18n.Trans("Can not create Certificate err=%v"), err)
	}
	return ""
}

// RevokeCert: 証明書の失効
func (a *App) RevokeCert(id string) {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm revoke cert"),
		Message:       i18n.Trans("Do you want to revoke selected Cert?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return
	}
	datastore.RevokeCertByID(id)
}

// ExportCert: 証明書をエクスポート
func (a *App) ExportCert(id string) {
	cert := datastore.FindCert(id)
	if cert == nil {
		return
	}
	d := time.Now().Format("20060102150405")
	file, err := wails.SaveFileDialog(a.ctx, wails.SaveDialogOptions{
		DefaultFilename:      "crt_" + d + ".crt",
		CanCreateDirectories: true,
		Filters: []wails.FileFilter{{
			DisplayName: "Certificate",
			Pattern:     "*.crt",
		}},
	})
	if err != nil || file == "" {
		return
	}
	os.WriteFile(file, []byte(cert.Certificate), 0660)
}

func (a *App) GetPKIControl() datastore.PKIControlEnt {
	return datastore.PKIControlEnt{
		EnableAcme:  datastore.PKIConf.EnableAcme,
		EnableHttp:  datastore.PKIConf.EnableHttp,
		AcmeBaseURL: datastore.PKIConf.AcmeBaseURL,
		CertTerm:    datastore.PKIConf.CertTerm,
		CrlInterval: datastore.PKIConf.CrlInterval,
		AcmeStatus:  pki.GetAcmeServerStatus(),
		HttpStatus:  pki.GetHttpServerStatus(),
	}
}

func (a *App) SetPKIControl(req datastore.PKIControlEnt) {
	datastore.PKIConf.EnableAcme = req.EnableAcme
	datastore.PKIConf.EnableHttp = req.EnableHttp
	datastore.PKIConf.AcmeBaseURL = req.AcmeBaseURL
	datastore.PKIConf.CertTerm = req.CertTerm
	datastore.PKIConf.CrlInterval = req.CrlInterval
	datastore.SavePKIConf()
}
