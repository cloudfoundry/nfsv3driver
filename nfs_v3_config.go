package nfsv3driver

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ConfigDetails struct {
	Allowed   []string
	Mandatory []string

	Forced  map[string]string
	Options map[string]string
}

type Config struct {
	source      ConfigDetails
	mount       ConfigDetails
	sloppyMount bool
}

func inArray(list []string, key string) bool {
	for _, k := range list {
		if k == key {
			return true
		}
	}

	return false
}

func NewNfsV3Config(sourceDetails *ConfigDetails, mountDetails *ConfigDetails) *Config {
	myConf := new(Config)

	myConf.source = *sourceDetails
	myConf.mount = *mountDetails
	myConf.sloppyMount = false

	return myConf
}

func NewNfsV3ConfigDetails() *ConfigDetails {
	myConf := new(ConfigDetails)

	myConf.Allowed = make([]string, 0)
	myConf.Options = make(map[string]string, 0)
	myConf.Forced = make(map[string]string, 0)
	myConf.Mandatory = make([]string, 0)

	return myConf
}

func (m *Config) SetEntries(share string, opts map[string]interface{}, ignoreList []string) error {

	m.source.parseMap(opts, ignoreList)
	m.mount.parseMap(opts, ignoreList)

	allowed := append(ignoreList, m.source.Allowed...)
	allowed = append(allowed, m.mount.Allowed...)
	errorList := m.source.parseUrl(share, ignoreList)
	m.sloppyMount = m.mount.IsSloppyMount()

	for k, _ := range opts {
		if !inArray(allowed, k) {
			errorList = append(errorList, k)
		}
	}

	if len(errorList) > 0 && m.sloppyMount != true {
		err := errors.New("Not allowed options : " + strings.Join(errorList, ", "))
		return err
	}

	if mdtErr := append(m.source.CheckMandatory(), m.mount.CheckMandatory()...); len(mdtErr) > 0 {
		err := errors.New("Missing mandatory options : " + strings.Join(mdtErr, ", "))
		return err
	}

	return nil
}

func (m Config) Share(share string) string {

	srcPart := strings.SplitN(share, "?", 2)

	if len(srcPart) < 2 {
		srcPart = append(srcPart, "")
	}

	srcPart[1] = strings.Join(m.source.makeParams(""), "&")

	if len(srcPart[1]) < 1 {
		srcPart = srcPart[:len(srcPart)-1]
	}

	return strings.Join(srcPart, "?")
}

func (m Config) Mount() []string {

	return m.mount.makeParams("--")
}

func (m Config) MountConfig() map[string]interface{} {

	return m.mount.makeConfig()
}

func (m *ConfigDetails) readConfAllowed(flagString string) error {
	if len(flagString) > 0 {
		m.Allowed = strings.Split(flagString, ",")
	}

	return nil
}

func (m *ConfigDetails) readConfDefault(flagString string) error {
	if len(flagString) < 1 {
		return nil
	}

	m.Options = m.parseConfig(strings.Split(flagString, ","))
	m.Forced = make(map[string]string)

	for k, v := range m.Options {
		if !inArray(m.Allowed, k) {
			m.Forced[k] = v
			delete(m.Options, k)
		}
	}

	return nil
}

func (m *ConfigDetails) ReadConf(allowedFlag string, defaultFlag string, mandatoryFields []string) error {
	if err := m.readConfAllowed(allowedFlag); err != nil {
		return err
	}

	if err := m.readConfDefault(defaultFlag); err != nil {
		return err
	}

	if len(mandatoryFields) > 0 {
		m.Mandatory = mandatoryFields
	}

	return nil
}

func (m ConfigDetails) CheckMandatory() []string {
	var result []string

	for _, k := range m.Mandatory {

		_, oko := m.Options[k]
		_, okf := m.Forced[k]

		if !okf && !oko {
			result = append(result, k)
		}
	}

	return result
}

func (m ConfigDetails) parseConfig(listEntry []string) map[string]string {

	result := map[string]string{}

	for _, opt := range listEntry {

		key := strings.SplitN(opt, ":", 2)

		if len(key[0]) < 1 {
			continue
		}

		if len(key[1]) < 1 {
			result[key[0]] = ""
		} else {
			result[key[0]] = key[1]
		}
	}

	return result
}

func (m *ConfigDetails) IsSloppyMount() bool {

	spm := ""
	ok := false

	if _, ok = m.Options["sloppy_mount"]; ok {
		spm = m.Options["sloppy_mount"]
		delete(m.Options, "sloppy_mount")
	}

	if _, ok = m.Forced["sloppy_mount"]; ok {
		spm = m.Forced["sloppy_mount"]
		delete(m.Forced, "sloppy_mount")
	}

	if len(spm) > 0 {
		if val, err := strconv.ParseBool(spm); err == nil {
			return val
		}
	}

	return false
}

func (m ConfigDetails) uniformKeyData(key string, data interface{}) string {
	switch key {
	case "auto-traverse-mounts":
		return m.uniformData(data, true)

	case "dircache":
		return m.uniformData(data, true)

	}

	return m.uniformData(data, false)
}

func (m ConfigDetails) uniformData(data interface{}, boolAsInt bool) string {

	switch data.(type) {
	case int:
		return strconv.FormatInt(int64(data.(int)), 10)

	case string:
		return data.(string)

	case bool:
		if boolAsInt {
			if data.(bool) {
				return "1"
			} else {
				return "0"
			}
		} else {
			return strconv.FormatBool(data.(bool))
		}
	}

	return ""
}

func (m *ConfigDetails) parseUrl(url string, ignoreList []string) []string {

	var errorList []string

	part := strings.SplitN(url, "?", 2)

	if len(part) < 2 {
		part = append(part, "")
	}

	for _, p := range strings.Split(part[1], "&") {
		if key := m.parseUrlParams(p, ignoreList); len(key) > 0 {
			errorList = append(errorList, key)
		}

	}

	return errorList
}

func (m *ConfigDetails) parseUrlParams(urlParams string, ignoreList []string) string {

	op := strings.SplitN(urlParams, "=", 2)

	if len(op) < 2 || len(op[1]) < 1 || op[1] == "" || inArray(ignoreList, op[0]) {
		return ""
	}

	if inArray(m.Allowed, op[0]) {
		m.Options[op[0]] = m.uniformKeyData(op[0], op[1])
		return ""
	}

	return op[0]
}

func (m *ConfigDetails) parseMap(entryList map[string]interface{}, ignoreList []string) []string {

	var errorList []string

	for k, v := range entryList {

		value := m.uniformData(v, false)

		if value == "" || len(k) < 1 || inArray(ignoreList, k) {
			continue
		}

		if inArray(m.Allowed, k) {
			m.Options[k] = value
		} else {
			errorList = append(errorList, k)
		}
	}

	return errorList
}

func (m ConfigDetails) makeParams(prefix string) []string {
	params := []string{}

	for k, v := range m.makeConfig() {

		if k == "sloppy_mount" {
			continue
		}

		if val, err := strconv.ParseBool(v.(string)); err == nil {
			if val {
				params = append(params, fmt.Sprintf("%s%s", prefix, k))
			}
			continue
		}

		if val, err := strconv.ParseInt(v.(string), 10, 16); err == nil {
			params = append(params, fmt.Sprintf("%s%s=%d", prefix, k, val))
			continue
		}

		params = append(params, fmt.Sprintf("%s%s=%s", prefix, k, v.(string)))
	}

	return params
}

func (m *ConfigDetails) makeConfig() map[string]interface{} {

	params := map[string]interface{}{}

	for k, v := range m.Options {
		params[k] = v
	}

	for k, v := range m.Forced {
		params[k] = v
	}

	return params
}
