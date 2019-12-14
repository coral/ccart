package controller

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"github.com/roffe/ccart/pkg/caddycfg"
)

var (
	caddyURL = "http://localhost:2019/config/"
)

func updateServer(name string) {
	jsonStr, err := srv.ParseJSON()
	if err != nil {
		panic(err)
	}
	reqURL := caddyURL + "apps/http/servers/" + name
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		body, _ := ioutil.ReadAll(resp.Body)
		glog.Infoln("response Status:", resp.Status)
		glog.Infoln("response Headers:", resp.Header)
		glog.Infoln("response Body:", string(body))
	}
}

func setInitialConfig(cfg caddycfg.Config) {
	jsonStr, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	glog.Info(string(jsonStr))

	req, err := http.NewRequest("POST", caddyURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		body, _ := ioutil.ReadAll(resp.Body)
		glog.Infoln("response Status:", resp.Status)
		glog.Infoln("response Headers:", resp.Header)
		glog.Infoln("response Body:", string(body))
	}
}
