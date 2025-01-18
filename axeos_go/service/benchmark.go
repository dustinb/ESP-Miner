package service

import (
	"log"
	"time"

	"oldbute.com/axego/db"
)

func Benchmark(config db.Config) {

	maxTemp := 66.0
	maxVRTemp := 95.0
	voltageStep := 50    // mV
	frequencyStep := 100 // MHz

	info := GetSystemInfo(config.IP)

	go func() {
		sessionStart := time.Now()
		for {
			// Expected
			expectedHashrate := float64(config.Frequency) * ((float64(config.SmallCoreCount) * float64(config.AsicCount)) / 1000)

			// Average Hashrate over session
			var avgHashrate float64
			db.Database.Raw("SELECT AVG(hashrate) FROM measurements WHERE config_id = ?", config.ID).Scan(&avgHashrate)

			log.Printf("Session %d: Expected: %f, Average: %f", config.ID, expectedHashrate, avgHashrate)

			// Temperature back off
			var measurement db.Measurement
			db.Database.Raw("SELECT * FROM measurements WHERE config_id = ? ORDER BY created_at DESC LIMIT 1", config.ID).Scan(&measurement)

			if measurement.Temp > maxTemp || measurement.VRTemp > maxVRTemp {
				config.Passed = false
				db.Database.Save(&config)

				// Good known config
				var goodConfig db.Config
				db.Database.Raw("SELECT * FROM config WHERE passed = ? AND ip = ? ORDER BY frequency DESC LIMIT 1", true, config.IP).Scan(&goodConfig)
				PatchAxe(config.IP, Patch{
					CoreVoltage: goodConfig.CoreVoltage,
					Frequency:   goodConfig.Frequency,
				})
				break
			}

			// Session ended without max temp reached
			if time.Since(sessionStart) > 3*time.Minute {
				config.Passed = true
				db.Database.Save(&config)

				config = db.Config{
					IP:             config.IP,
					Hostname:       config.Hostname,
					Frequency:      config.Frequency,
					CoreVoltage:    config.CoreVoltage,
					AsicCount:      info.AsicCount,
					SmallCoreCount: info.SmallCoreCount,
					Passed:         true,
				}

				if avgHashrate > expectedHashrate {
					config.Frequency += frequencyStep
				} else {
					config.CoreVoltage += voltageStep
				}
				db.Database.Create(&config)
				PatchAxe(config.IP, Patch{
					CoreVoltage: config.CoreVoltage,
					Frequency:   config.Frequency,
				})
				sessionStart = time.Now()
			}
			time.Sleep(10 * time.Second)
		}
	}()
}
