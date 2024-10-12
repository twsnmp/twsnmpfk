package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/openconfig/gnmic/pkg/api"

	"github.com/twsnmp/twsnmpfk/datastore"
)

type ModelData struct {
	Name         string `json:"name,omitempty"`         // Name of the model.
	Organization string `json:"organization,omitempty"` // Organization publishing the model.
	Version      string `json:"version,omitempty"`      // Semantic version of the model.
}

// GNMICapEnt : gNMI Capabilities responce
type GNMICapEnt struct {
	Version   string       `json:"Version"`
	Encodings string       `json:"Encodings"`
	Models    []*ModelData `json:"Models"`
}

// GNMIGetEnt : gNMI get responce
type GNMIGetEnt struct {
	Path  string `json:"Path"`
	Value string `json:"Value"`
	Index string `json:"Index"`
}

func (a *App) GNMICapabilities(nodeID, target string) *GNMICapEnt {
	ret := &GNMICapEnt{}
	n := datastore.GetNode(nodeID)
	if n == nil {
		log.Println("node not found")
		return ret
	}
	tg, err := api.NewTarget(
		api.Name(n.Name),
		api.Address(target),
		api.Username(n.GNMIUser),
		api.Password(n.GNMIPassword),
		api.SkipVerify(true),
	)
	if err != nil {
		log.Printf("gnmi cap err=%v", err)
		return ret
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = tg.CreateGNMIClient(ctx)
	if err != nil {
		log.Printf("gnmi cap err=%v", err)
		return ret
	}
	defer tg.Close()
	capResp, err := tg.Capabilities(ctx)
	if err != nil {
		log.Printf("gnmi cap err=%v", err)
		return ret
	}
	models := capResp.GetSupportedModels()
	for _, m := range models {
		ret.Models = append(ret.Models, &ModelData{
			Version:      m.Version,
			Organization: m.Organization,
			Name:         m.Name,
		})
	}
	ret.Version = capResp.GetGNMIVersion()
	es := []string{}
	for _, e := range capResp.GetSupportedEncodings() {
		es = append(es, e.String())
	}
	ret.Encodings = strings.Join(es, ",")
	return ret
}

func (a *App) GNMIGet(nodeID, target, path, enc string) []*GNMIGetEnt {
	st := time.Now()
	ret := []*GNMIGetEnt{}
	n := datastore.GetNode(nodeID)
	if n == nil {
		log.Println("node not found")
		return ret
	}
	tg, err := api.NewTarget(
		api.Name(n.Name),
		api.Address(target),
		api.Username(n.GNMIUser),
		api.Password(n.GNMIPassword),
		api.SkipVerify(true),
	)
	if err != nil {
		log.Printf("gnmi get err=%v", err)
		return ret
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = tg.CreateGNMIClient(ctx)
	if err != nil {
		log.Printf("gnmi get err=%v", err)
		return ret
	}
	defer tg.Close()
	if enc == "" {
		enc = "json_ietf"
	}
	getReq, err := api.NewGetRequest(
		api.Path(path),
		api.Encoding(enc))
	if err != nil {
		log.Printf("gnmi get err=%v", err)
		return ret
	}
	getResp, err := tg.Get(ctx, getReq)
	if err != nil {
		ret = append(ret, &GNMIGetEnt{
			Path:  "Error",
			Value: err.Error(),
		})
		log.Printf("gnmi get err=%v", err)
		return ret
	}
	for _, not := range getResp.GetNotification() {
		for _, u := range not.GetUpdate() {
			pa := []string{}
			for _, p := range u.Path.Elem {
				pa = append(pa, p.GetName())
			}
			j := u.Val.GetJsonIetfVal()
			if len(j) < 1 {
				j = u.Val.GetJsonVal()
			}
			var d interface{}
			if err := json.Unmarshal(j, &d); err != nil {
				log.Println(err)
				continue
			}
			path := ""
			if len(pa) > 0 {
				path = "/" + strings.Join(pa, "/")
			}
			ret = append(ret, getPathValue(d, path, "", false)...)
		}
	}
	log.Printf("gNMI get path=%s  dur=%v", path, time.Since(st))
	return ret
}

func getPathValue(d interface{}, path, index string, inArray bool) []*GNMIGetEnt {
	r := []*GNMIGetEnt{}
	switch v := d.(type) {
	case string:
		r = append(r, &GNMIGetEnt{
			Path:  path,
			Value: v,
			Index: index,
		})
		return r
	case float64:
		r = append(r, &GNMIGetEnt{
			Path:  path,
			Value: fmt.Sprintf("%v", v),
			Index: index,
		})
	case bool:
		r = append(r, &GNMIGetEnt{
			Path:  path,
			Value: fmt.Sprintf("%v", v),
			Index: index,
		})
	case map[string]interface{}:
		n := ""
		if in, ok := v["name"]; ok {
			if sn, ok := in.(string); ok {
				n = sn
			}
		}
		for k, vv := range v {
			if inArray && n != "" {
				r = append(r, getPathValue(vv, fmt.Sprintf("%s[name=%s]/%s", path, n, k), index, false)...)
			} else {
				r = append(r, getPathValue(vv, path+"/"+k, index, false)...)
			}
		}
	case []interface{}:
		for i, vv := range v {
			r = append(r, getPathValue(vv, path, fmt.Sprintf("%d", i), true)...)
		}
	default:
		log.Printf("%s=%+v type=%v", path, v, reflect.TypeOf(d))
	}
	return r
}
