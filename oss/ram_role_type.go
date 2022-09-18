package oss

import "time"

type RamRole struct {
	AccessKeyId     string    `json:"AccessKeyId"`
	AccessKeySecret string    `json:"AccessKeySecret"`
	SecurityToken   string    `json:"SecurityToken"`
	Code            string    `json:"Code"`
	Expiration      time.Time `json:"Expiration"`
	LastUpdated     time.Time `json:"LastUpdated"`
}
