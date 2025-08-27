// SPDX-FileCopyrightText: 2023 iteratec GmbH
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/secureCodeBox/telemetry/pkg/elasticsearch"
	"github.com/secureCodeBox/telemetry/pkg/log"
)

// officialScanTypes contains the list of official secureCodeBox Scan Types.
// Unofficial Scan Types should be reported as "other" to avoid leakage of confidential data via the scan-types name
var officialScanTypes map[string]bool = map[string]bool{
	"amass":                  true, // deprecated. we'll keep it in this list to still recieve telemetry data from older versions
	"angularjs-csti-scanner": true, // deprecated. we'll keep it in this list to still recieve telemetry data from older versions
	"cmseek":                 true, // deprecated. we'll keep it in this list to still recieve telemetry data from older versions
	"doggo":                  true, // deprecated. we'll keep it in this list to still recieve telemetry data from older versions
	"ffuf":                   true,
	"git-repo-scanner":       true,
	"gitleaks":               true,
	"kube-hunter":            true,
	"kubeaudit":              true, // deprecated. we'll keep it in this list to still recieve telemetry data from older versions
	"ncrack":                 true,
	"nikto":                  true,
	"nmap":                   true,
	"nuclei":                 true,
	"screenshooter":          true,
	"semgrep":                true,
	"ssh-audit":              true,
	"ssh-scan":               true, // deprecated. we'll keep it in this list to still recieve telemetry data from older versions
	"sslyze":                 true,
	"trivy-filesystem":       true,
	"trivy-image":            true,
	"trivy-repo":             true,
	"trivy-sbom-image":       true,
	"trivy":                  true,
	"typo3scan":              true, // deprecated. we'll keep it in this list to still recieve telemetry data from older versions
	"whatweb":                true, // deprecated. we'll keep it in this list to still recieve telemetry data from older versions
	"wpscan":                 true,
	"zap-advanced-scan":      true, // deprecated. we'll keep it in this list to still recieve telemetry data from older versions
	"zap-api-scan":           true, // deprecated. we'll keep it in this list to still recieve telemetry data from older versions
	"zap-automation-scan":    true,
	"zap-baseline-scan":      true, // deprecated. we'll keep it in this list to still recieve telemetry data from older versions
	"zap-full-scan":          true, // deprecated. we'll keep it in this list to still recieve telemetry data from older versions

	"other": true,
}

// TelemetryData submitted by telemetry client in operator
type TelemetryData struct {
	Version            string   `json:"version" binding:"required"`
	InstalledScanTypes []string `json:"installedScanTypes" binding:"required"`
}

// TelemetryDataDocument is TelemetryData including properties required by elasticsearch
type TelemetryDataDocument struct {
	Timestamp          time.Time `json:"@timestamp"`
	Version            string    `json:"version"`
	InstalledScanTypes []string  `json:"installedScanTypes"`
}

var elasticSearchService elasticsearch.Service

func main() {
	service, err := elasticsearch.Setup()
	if err != nil {
		panic("Failed to init Elasticsearch Client")
	}

	router := setupRouter(*service)
	router.Run()
}

func setupRouter(service elasticsearch.Service) *gin.Engine {
	elasticSearchService = service

	router := gin.New()
	router.Use(gin.Recovery(), gin.LoggerWithFormatter(log.AnonymousLogFormatter))

	router.GET("/ready", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	v1 := router.Group("/v1")
	{
		v1.POST("/submit", persistTelemetryData)
	}

	return router
}

// persistTelemetryData saves telemetry data to Elasticsearch
func persistTelemetryData(c *gin.Context) {
	var telemetryDataInput TelemetryData
	if err := c.ShouldBindJSON(&telemetryDataInput); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// Ensure submitted ScanTypes are valid
	for _, scanType := range telemetryDataInput.InstalledScanTypes {
		if _, ok := officialScanTypes[scanType]; !ok {
			c.String(http.StatusBadRequest, fmt.Sprintf("Invalid ScanType '%s'", scanType))
			return
		}
	}

	telemetryData := TelemetryDataDocument{
		Version:            telemetryDataInput.Version,
		InstalledScanTypes: telemetryDataInput.InstalledScanTypes,
		Timestamp:          time.Now(),
	}

	indexName := fmt.Sprintf("telemetry-%s", time.Now().Format("2006"))
	err := elasticSearchService.Create(indexName, telemetryData)

	if err != nil {
		c.String(http.StatusInternalServerError, "elasticsearch connection failed")
		return
	}

	c.String(http.StatusOK, "ok")
}
