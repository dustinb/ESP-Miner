package service

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"oldbute.com/axego/db"
)

const (
	NetworkScanInterval = 60 * time.Second
	measurementInterval = 20 * time.Second
)

// TakeMeasurements takes measurements for all found Bitaxe
func TakeMeasurements() {
	for {
		var bitaxes []db.Bitaxe
		log.Printf("Taking measurements for Bitaxes")

		// Skip any stale Bitaxes, not updated in 2 scan intervals
		updateAt := time.Now().Add(-NetworkScanInterval * 2)

		db.Database.Find(&bitaxes, "updated_at > ?", updateAt)
		for _, axe := range bitaxes {
			go func() {
				info := GetSystemInfo(axe.IP)
				measurement := db.Measurement{
					IP:        axe.IP,
					MacAddr:   axe.MacAddr,
					Hostname:  info.Hostname,
					Temp:      info.Temp,
					VRTemp:    info.VRTemp,
					Hashrate:  info.HashRate,
					Frequency: info.Frequency,
				}
				db.Database.Create(&measurement)
				log.Printf("Host: %s, IP: %s, Temp: %f, VRTemp: %f, Hashrate: %f, Frequency: %d",
					axe.Hostname, axe.IP, info.Temp, info.VRTemp, info.HashRate, info.Frequency)
			}()
		}
		time.Sleep(measurementInterval)
	}
}

// ScanNetwork scans the network for the Bitaxe
func ScanNetwork(foundAxe chan db.Bitaxe) {
	for {
		log.Printf("Scanning network for Bitaxes")
		addrs, _ := net.InterfaceAddrs()
		for _, address := range addrs {
			host, _ := address.(*net.IPNet)
			if host.IP.IsLoopback() {
				continue
			}
			// Check for IPv4
			if host.IP.To4() == nil {
				continue
			}
			log.Print(host.IP.String())
			octets := strings.Split(host.IP.String(), ".")
			network := octets[0] + "." + octets[1] + "." + octets[2] + ".%d"

			for i := 1; i < 255; i++ {
				ip := fmt.Sprintf(network, i)

				// Do it all at the same time
				go func() {
					info := GetSystemInfo(ip)
					if info.Hostname != "" {
						foundAxe <- db.Bitaxe{IP: ip, Hostname: info.Hostname, MacAddr: info.MacAddr}
					}
				}()
			}
		}
		time.Sleep(NetworkScanInterval)
	}
}
