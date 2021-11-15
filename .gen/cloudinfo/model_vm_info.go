/*
 * Product Info.
 *
 * The product info application uses the cloud provider APIs to asynchronously fetch and parse instance type attributes and prices, while storing the results in an in memory cache and making it available as structured data through a REST API.
 *
 * API version: 0.6.0
 * Contact: info@banzaicloud.com
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package cloudinfo

// VMInfo representation of a virtual machine
type VmInfo struct {
	Attributes map[string]string `json:"attributes,omitempty"`
	Category string `json:"category,omitempty"`
	CpusPerVm float64 `json:"cpusPerVm,omitempty"`
	// CurrentGen signals whether the instance type generation is the current one. Only applies for amazon
	CurrentGen bool `json:"currentGen,omitempty"`
	GpusPerVm float64 `json:"gpusPerVm,omitempty"`
	MemPerVm float64 `json:"memPerVm,omitempty"`
	NtwPerf string `json:"ntwPerf,omitempty"`
	NtwPerfCategory string `json:"ntwPerfCategory,omitempty"`
	OnDemandPrice float64 `json:"onDemandPrice,omitempty"`
	Series string `json:"series,omitempty"`
	SpotPrice []ZonePrice `json:"spotPrice,omitempty"`
	Type string `json:"type,omitempty"`
	Zones []string `json:"zones,omitempty"`
}