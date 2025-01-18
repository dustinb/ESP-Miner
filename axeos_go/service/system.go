package service

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

type Info struct {
	Temp              float64 `json:"temp"`
	VRTemp            float64 `json:"vrTemp"`
	HashRate          float64 `json:"hashRate"`
	CoreVoltage       int     `json:"coreVoltage"`
	CoreVoltageActual int     `json:"coreVoltageActual"`
	Frequency         int     `json:"frequency"`
	Hostname          string  `json:"hostname"`
	AsicCount         int     `json:"asicCount"`
	SmallCoreCount    int     `json:"smallCoreCount"`
	OverHeatMode      bool    `json:"overheat_mode"`
	FanSpeed          int     `json:"fanSpeed"`
	MacAddr           string  `json:"macAddr"`
}

type Patch struct {
	CoreVoltage int `json:"coreVoltage"`
	Frequency   int `json:"frequency"`
}

func GetSystemInfo(ip string) Info {
	client := http.Client{}
	client.Timeout = 2 * time.Second
	resp, err := client.Get("http://" + ip + "/api/system/info")
	if err != nil {
		return Info{}
	}
	body, _ := io.ReadAll(resp.Body)
	info := Info{}
	json.Unmarshal(body, &info)
	return info
}

func PatchAxe(ip string, patch Patch) {
	client := http.Client{}
	json, _ := json.Marshal(patch)
	log.Printf("Patching Axe: %s", json)
	req, _ := http.NewRequest("PATCH", "http://"+ip+"/api/system", bytes.NewBuffer(json))
	client.Do(req)
	time.Sleep(3 * time.Second)
	client.Post("http://"+ip+"/api/system/restart", "application/json", nil)
	time.Sleep(10 * time.Second)

}
