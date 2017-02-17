package nfsv3driver_test

import (
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "code.cloudfoundry.org/nfsv3driver"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strconv"
	"strings"
)

func map2string(entry map[string]string, joinKeyVal string, prefix string, joinElemnts string) string {
	return strings.Join(map2slice(entry, joinKeyVal, prefix), joinElemnts)
}

func mapstring2mapinterface(entry map[string]string) map[string]interface{} {
	result := make(map[string]interface{}, 0)

	for k, v := range entry {
		result[k] = v
	}

	return result
}

func map2slice(entry map[string]string, joinKeyVal string, prefix string) []string {
	result := make([]string, 0)

	for k, v := range entry {
		result = append(result, fmt.Sprintf("%s%s%s%s", prefix, k, joinKeyVal, v))
	}

	return result
}

func mapint2slice(entry map[string]interface{}, joinKeyVal string, prefix string) []string {
	result := make([]string, 0)

	for k, v := range entry {
		switch v.(type) {
		case int:
			result = append(result, fmt.Sprintf("%s%s%s%s", prefix, k, joinKeyVal, strconv.FormatInt(int64(v.(int)), 10)))

		case string:
			result = append(result, fmt.Sprintf("%s%s%s%s", prefix, k, joinKeyVal, v.(string)))

		case bool:
			result = append(result, fmt.Sprintf("%s%s%s%s", prefix, k, joinKeyVal, strconv.FormatBool(v.(bool))))
		}

	}

	return result
}

func inSliceString(list []string, val string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}

	return false
}

func inMapInt(list map[string]interface{}, key string, val interface{}) bool {
	for k, v := range list {
		if k != key {
			continue
		}

		if v == val {
			return true
		} else {
			return false
		}
	}

	return false
}

var _ = Describe("BrokerConfigDetails", func() {
	var (
		logger lager.Logger

		ClientShare     string
		AbitraryConfig  map[string]interface{}
		IgnoreConfigKey []string

		SourceAllowed   []string
		SourceOptions   map[string]string
		SourceMandatory []string

		MountsAllowed   []string
		MountsOptions   map[string]string
		MountsMandatory []string

		source *ConfigDetails
		mounts *ConfigDetails
		config *Config

		errorEntries error
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test-broker-config")
	})

	Context("Given no mandatory and empty params", func() {
		BeforeEach(func() {
			ClientShare = "nfs://1.2.3.4"
			AbitraryConfig = make(map[string]interface{}, 0)
			IgnoreConfigKey = make([]string, 0)

			SourceAllowed = make([]string, 0)
			SourceOptions = make(map[string]string, 0)
			SourceMandatory = make([]string, 0)

			MountsAllowed = make([]string, 0)
			MountsOptions = make(map[string]string, 0)
			MountsMandatory = make([]string, 0)

			source = NewNfsV3ConfigDetails()
			source.ReadConf(strings.Join(SourceAllowed, ","), map2string(SourceOptions, ":", "", ","), SourceMandatory)

			mounts = NewNfsV3ConfigDetails()
			mounts.ReadConf(strings.Join(MountsAllowed, ","), map2string(MountsOptions, ":", "", ","), MountsMandatory)

			config = NewNfsV3Config(source, mounts)
			logger.Debug("Debug config Initiated", lager.Data{"source": source, "mount": mounts})

			errorEntries = config.SetEntries(ClientShare, AbitraryConfig, IgnoreConfigKey)
			logger.Debug("Debug config updated", lager.Data{"source": source, "mount": mounts})
		})

		It("should returns empty allowed list", func() {
			Expect(len(source.Allowed)).To(Equal(0))
			Expect(len(mounts.Allowed)).To(Equal(0))
		})

		It("should returns empty forced list", func() {
			Expect(len(source.Forced)).To(Equal(0))
			Expect(len(mounts.Forced)).To(Equal(0))
		})

		It("should returns empty options list", func() {
			Expect(len(source.Options)).To(Equal(0))
			Expect(len(mounts.Options)).To(Equal(0))
		})

		It("should flow sloppy_mount as disabled", func() {
			Expect(source.IsSloppyMount()).To(BeFalse())
			Expect(mounts.IsSloppyMount()).To(BeFalse())
		})

		It("should returns no missing mandatory fields", func() {
			Expect(len(source.CheckMandatory())).To(Equal(0))
			Expect(len(mounts.CheckMandatory())).To(Equal(0))
		})

		It("should returns no error on given client abitrary config", func() {
			Expect(errorEntries).To(BeNil())
		})

		It("should returns no mount command parameters", func() {
			Expect(len(config.Mount())).To(Equal(0))
		})

		It("should returns no MountOptions struct", func() {
			Expect(len(config.MountConfig())).To(Equal(0))
		})

		It("returns no added parameters to the client share", func() {
			Expect(config.Share(ClientShare)).To(Equal(ClientShare))
		})
	})

	Context("Given source mandatory, no mount mandatory and empty params", func() {
		BeforeEach(func() {
			ClientShare = "nfs://1.2.3.4"
			AbitraryConfig = make(map[string]interface{}, 0)
			IgnoreConfigKey = make([]string, 0)

			SourceAllowed = make([]string, 0)
			SourceOptions = make(map[string]string, 0)
			SourceMandatory = []string{"uid", "gid"}

			MountsAllowed = make([]string, 0)
			MountsOptions = make(map[string]string, 0)
			MountsMandatory = make([]string, 0)

			source = NewNfsV3ConfigDetails()
			source.ReadConf(strings.Join(SourceAllowed, ","), map2string(SourceOptions, ":", "", ","), SourceMandatory)

			mounts = NewNfsV3ConfigDetails()
			mounts.ReadConf(strings.Join(MountsAllowed, ","), map2string(MountsOptions, ":", "", ","), MountsMandatory)

			config = NewNfsV3Config(source, mounts)
			logger.Debug("Debug config Initiated", lager.Data{"source": source, "mount": mounts})

			errorEntries = config.SetEntries(ClientShare, AbitraryConfig, IgnoreConfigKey)
			logger.Debug("Debug config updated", lager.Data{"config": config, "source": source, "mount": mounts})
		})

		It("should returns empty allowed list", func() {
			Expect(len(source.Allowed)).To(Equal(0))
			Expect(len(mounts.Allowed)).To(Equal(0))
		})

		It("should returns empty forced list", func() {
			Expect(len(source.Forced)).To(Equal(0))
			Expect(len(mounts.Forced)).To(Equal(0))
		})

		It("should returns empty options list", func() {
			Expect(len(source.Options)).To(Equal(0))
			Expect(len(mounts.Options)).To(Equal(0))
		})

		It("should flow sloppy_mount as disabled", func() {
			Expect(source.IsSloppyMount()).To(BeFalse())
			Expect(mounts.IsSloppyMount()).To(BeFalse())
		})

		It("should returns no missing mount mandatory fields", func() {
			Expect(len(mounts.CheckMandatory())).To(Equal(0))
		})

		It("should flow the mandatory config as missing mandatory field", func() {
			Expect(len(source.CheckMandatory())).To(Equal(2))
			Expect(source.CheckMandatory()).To(Equal(SourceMandatory))
		})

		It("should occures an error because there are missing fields", func() {
			Expect(errorEntries).To(HaveOccurred())
		})

		It("should returns no mount command parameters", func() {
			Expect(len(config.Mount())).To(Equal(0))
		})

		It("should returns no MountOptions struct", func() {
			Expect(len(config.MountConfig())).To(Equal(0))
		})

		It("should returns no added parameters to the client share", func() {
			Expect(config.Share(ClientShare)).To(Equal(ClientShare))
		})
	})

	Context("Given mandatory, allowed and default params", func() {
		BeforeEach(func() {
			ClientShare = "nfs://1.2.3.4"
			AbitraryConfig = make(map[string]interface{}, 0)
			IgnoreConfigKey = make([]string, 0)

			SourceAllowed = []string{"uid", "gid"}
			SourceMandatory = []string{"uid", "gid"}
			SourceOptions = map[string]string{
				"uid": "1004",
				"gid": "1002",
			}

			MountsAllowed = []string{"sloppy_mount", "nfs_uid", "nfs_gid"}
			MountsMandatory = make([]string, 0)
			MountsOptions = map[string]string{
				"nfs_uid": "1003",
				"nfs_gid": "1001",
			}

			source = NewNfsV3ConfigDetails()
			source.ReadConf(strings.Join(SourceAllowed, ","), map2string(SourceOptions, ":", "", ","), SourceMandatory)

			mounts = NewNfsV3ConfigDetails()
			mounts.ReadConf(strings.Join(MountsAllowed, ","), map2string(MountsOptions, ":", "", ","), MountsMandatory)

			config = NewNfsV3Config(source, mounts)
			logger.Debug("Debug config Initiated", lager.Data{"config": config, "source": source, "mount": mounts})
		})

		It("should flow the allowed list", func() {
			Expect(source.Allowed).To(Equal(SourceAllowed))
			Expect(mounts.Allowed).To(Equal(MountsAllowed))
		})

		It("should return empty forced list", func() {
			Expect(len(source.Forced)).To(Equal(0))
			Expect(len(mounts.Forced)).To(Equal(0))
		})

		It("should flow the default params as options list", func() {
			Expect(source.Options).To(Equal(SourceOptions))
			Expect(mounts.Options).To(Equal(MountsOptions))
		})

		It("should flow sloppy_mount as disabled", func() {
			Expect(source.IsSloppyMount()).To(BeFalse())
			Expect(mounts.IsSloppyMount()).To(BeFalse())
		})

		It("should return empty missing mandatory field", func() {
			Expect(len(source.CheckMandatory())).To(Equal(0))
			Expect(len(mounts.CheckMandatory())).To(Equal(0))
		})

		Context("Given empty abitrary params and share without any params", func() {
			BeforeEach(func() {
				errorEntries = config.SetEntries(ClientShare, AbitraryConfig, IgnoreConfigKey)
				logger.Debug("Debug config updated", lager.Data{"config": config, "source": source, "mount": mounts})
			})

			It("should return nil result on setting end users'entries", func() {
				Expect(errorEntries).To(BeNil())
			})

			It("flow the mount default options into the mount command parameters ", func() {
				actualRes := config.Mount()
				expectRes := map2slice(MountsOptions, "=", "--")

				for _, exp := range expectRes {
					logger.Debug("Checking actualRes contain part", lager.Data{"actualRes": actualRes, "part": exp})
					Expect(inSliceString(actualRes, exp)).To(BeTrue())
				}

				for _, exp := range actualRes {
					logger.Debug("Checking expectRes contain part", lager.Data{"expectRes": expectRes, "part": exp})
					Expect(inSliceString(expectRes, exp)).To(BeTrue())
				}
			})

			It("should flow the mount default options into the MountOptions struct", func() {
				actualRes := config.MountConfig()
				expectRes := mapstring2mapinterface(MountsOptions)

				for k, exp := range expectRes {
					logger.Debug("Checking expectRes contain part", lager.Data{"expectRes": expectRes, "key": k, "val": exp})
					Expect(inMapInt(actualRes, k, exp)).To(BeTrue())
				}

				for k, exp := range actualRes {
					logger.Debug("Checking expectRes contain part", lager.Data{"expectRes": expectRes, "key": k, "val": exp})
					Expect(inMapInt(expectRes, k, exp)).To(BeTrue())
				}
			})

			It("should flow the source default options as parameters into the client share url", func() {

				share := config.Share(ClientShare)
				Expect(share).To(ContainSubstring(ClientShare + "?"))

				for _, exp := range map2slice(SourceOptions, "=", "") {
					logger.Debug("Checking Share contain part", lager.Data{"share": share, "part": exp})
					Expect(share).To(ContainSubstring(exp))
				}
			})
		})

		Context("Given bad abitrary params and bad share params", func() {

			BeforeEach(func() {
				ClientShare = "nfs://1.2.3.4?err=true&test=err"
				AbitraryConfig = map[string]interface{}{
					"missing": true,
					"wrong":   1234,
					"search":  "notfound",
				}
				IgnoreConfigKey = make([]string, 0)

				errorEntries = config.SetEntries(ClientShare, AbitraryConfig, IgnoreConfigKey)
				logger.Debug("Debug config updated", lager.Data{"config": config, "source": source, "mount": mounts})
			})

			It("should occured an error", func() {
				Expect(errorEntries).To(HaveOccurred())
				logger.Debug("Debug config updated with entry", lager.Data{"config": config, "source": source, "mount": mounts})
			})

			It("should flow the mount default options into the mount command parameters ", func() {
				actualRes := config.Mount()
				expectRes := map2slice(MountsOptions, "=", "--")

				for _, exp := range expectRes {
					logger.Debug("Checking actualRes contain part", lager.Data{"actualRes": actualRes, "part": exp})
					Expect(inSliceString(actualRes, exp)).To(BeTrue())
				}

				for _, exp := range actualRes {
					logger.Debug("Checking expectRes contain part", lager.Data{"expectRes": expectRes, "part": exp})
					Expect(inSliceString(expectRes, exp)).To(BeTrue())
				}
			})

			It("should flow the mount default options into the MountOptions struct", func() {
				actualRes := config.MountConfig()
				expectRes := mapstring2mapinterface(MountsOptions)

				for k, exp := range expectRes {
					logger.Debug("Checking expectRes contain part", lager.Data{"expectRes": expectRes, "key": k, "val": exp})
					Expect(inMapInt(actualRes, k, exp)).To(BeTrue())
				}

				for k, exp := range actualRes {
					logger.Debug("Checking expectRes contain part", lager.Data{"expectRes": expectRes, "key": k, "val": exp})
					Expect(inMapInt(expectRes, k, exp)).To(BeTrue())
				}
			})

			It("should flow the source default options as parameters into the client share url", func() {

				share := config.Share(ClientShare)
				Expect(share).To(ContainSubstring("nfs://1.2.3.4?"))

				for _, exp := range map2slice(SourceOptions, "=", "") {
					logger.Debug("Checking Share contain part", lager.Data{"share": share, "part": exp})
					Expect(share).To(ContainSubstring(exp))
				}
			})
		})

		Context("Given bad abitrary params and bad share params with sloppy_mount mode", func() {

			BeforeEach(func() {
				ClientShare = "nfs://1.2.3.4?err=true&test=err"
				AbitraryConfig = map[string]interface{}{
					"sloppy_mount": true,
					"missing":      true,
					"wrong":        1234,
					"search":       "notfound",
				}
				IgnoreConfigKey = make([]string, 0)

				errorEntries = config.SetEntries(ClientShare, AbitraryConfig, IgnoreConfigKey)
				logger.Debug("Debug config updated", lager.Data{"config": config, "source": source, "mount": mounts})
			})

			It("should not occured an error, return nil", func() {
				Expect(errorEntries).To(BeNil())
				logger.Debug("Debug config updated with entry", lager.Data{"config": config, "source": source, "mount": mounts})
			})

			It("should flow the mount default options into the mount command parameters ", func() {
				actualRes := config.Mount()
				expectRes := map2slice(MountsOptions, "=", "--")

				for _, exp := range expectRes {
					logger.Debug("Checking actualRes contain part", lager.Data{"actualRes": actualRes, "part": exp})
					Expect(inSliceString(actualRes, exp)).To(BeTrue())
				}

				for _, exp := range actualRes {
					logger.Debug("Checking expectRes contain part", lager.Data{"expectRes": expectRes, "part": exp})
					Expect(inSliceString(expectRes, exp)).To(BeTrue())
				}
			})

			It("should flow the mount default options into the MountOptions struct", func() {
				actualRes := config.MountConfig()
				expectRes := mapstring2mapinterface(MountsOptions)

				for k, exp := range expectRes {
					logger.Debug("Checking expectRes contain part", lager.Data{"expectRes": expectRes, "key": k, "val": exp})
					Expect(inMapInt(actualRes, k, exp)).To(BeTrue())
				}

				for k, exp := range actualRes {
					logger.Debug("Checking expectRes contain part", lager.Data{"expectRes": expectRes, "key": k, "val": exp})
					Expect(inMapInt(expectRes, k, exp)).To(BeTrue())
				}
			})

			It("should flow the source default options as parameters into the client share url", func() {

				share := config.Share(ClientShare)
				Expect(share).To(ContainSubstring("nfs://1.2.3.4?"))

				for _, exp := range map2slice(SourceOptions, "=", "") {
					logger.Debug("Checking Share contain part", lager.Data{"share": share, "part": exp})
					Expect(share).To(ContainSubstring(exp))
				}
			})
		})

		Context("Given good abitrary params and good share params", func() {

			BeforeEach(func() {
				ClientShare = "nfs://1.2.3.4?uid=2999&gid=1999"
				AbitraryConfig = map[string]interface{}{
					"nfs_uid": "1234",
					"nfs_gid": "5678",
				}
				IgnoreConfigKey = make([]string, 0)

				errorEntries = config.SetEntries(ClientShare, AbitraryConfig, IgnoreConfigKey)
				logger.Debug("Debug config updated", lager.Data{"config": config, "source": source, "mount": mounts})
			})

			It("should not occured an error, return nil", func() {
				Expect(errorEntries).To(BeNil())
			})

			It("should flow the arbitrary config into the mount command parameters ", func() {
				actualRes := config.Mount()
				expectRes := mapint2slice(AbitraryConfig, "=", "--")

				for _, exp := range expectRes {
					logger.Debug("Checking actualRes contain part", lager.Data{"actualRes": actualRes, "part": exp})
					Expect(inSliceString(actualRes, exp)).To(BeTrue())
				}

				for _, exp := range actualRes {
					logger.Debug("Checking expectRes contain part", lager.Data{"expectRes": expectRes, "part": exp})
					Expect(inSliceString(expectRes, exp)).To(BeTrue())
				}
			})

			It("should flow the arbitrary config into the MountOptions struct", func() {
				actualRes := config.MountConfig()
				expectRes := AbitraryConfig

				for k, exp := range expectRes {
					logger.Debug("Checking expectRes contain part", lager.Data{"expectRes": expectRes, "key": k, "val": exp})
					Expect(inMapInt(actualRes, k, exp)).To(BeTrue())
				}

				for k, exp := range actualRes {
					logger.Debug("Checking expectRes contain part", lager.Data{"expectRes": expectRes, "key": k, "val": exp})
					Expect(inMapInt(expectRes, k, exp)).To(BeTrue())
				}
			})

			It("should flow the source default options overrided by the share good params into the mount share url", func() {
				share := config.Share(ClientShare)

				Expect(share).To(ContainSubstring("nfs://1.2.3.4?"))
				Expect(share).To(ContainSubstring("uid=2999"))
				Expect(share).To(ContainSubstring("gid=1999"))
			})
		})
	})

	Context("Given mandatory and default params but with empty allowed", func() {
		BeforeEach(func() {
			ClientShare = "nfs://1.2.3.4"
			AbitraryConfig = make(map[string]interface{}, 0)
			IgnoreConfigKey = make([]string, 0)

			SourceAllowed = make([]string, 0)
			SourceMandatory = []string{"uid", "gid"}
			SourceOptions = map[string]string{
				"uid": "1004",
				"gid": "1002",
			}

			MountsAllowed = make([]string, 0)
			MountsMandatory = make([]string, 0)
			MountsOptions = map[string]string{
				"auto_cache": "true",
			}

			source = NewNfsV3ConfigDetails()
			source.ReadConf(strings.Join(SourceAllowed, ","), map2string(SourceOptions, ":", "", ","), SourceMandatory)

			mounts = NewNfsV3ConfigDetails()
			mounts.ReadConf(strings.Join(MountsAllowed, ","), map2string(MountsOptions, ":", "", ","), MountsMandatory)

			config = NewNfsV3Config(source, mounts)
			logger.Debug("Debug config Initiated", lager.Data{"config": config, "source": source, "mount": mounts})
			logger.Debug("Debug config updated", lager.Data{"config": config, "source": source, "mount": mounts})
		})

		It("should return empty allowed list", func() {
			Expect(len(source.Allowed)).To(Equal(0))
			Expect(len(mounts.Allowed)).To(Equal(0))
		})

		It("should flow the default list as forced", func() {
			Expect(source.Forced).To(Equal(SourceOptions))
			Expect(mounts.Forced).To(Equal(MountsOptions))
		})

		It("should return empty options list", func() {
			Expect(len(source.Options)).To(Equal(0))
			Expect(len(mounts.Options)).To(Equal(0))
		})

		It("should flow sloppy_mount as disabled", func() {
			Expect(source.IsSloppyMount()).To(BeFalse())
			Expect(mounts.IsSloppyMount()).To(BeFalse())
		})

		It("should return empty missing mandatory field", func() {
			Expect(len(source.CheckMandatory())).To(Equal(0))
			Expect(len(mounts.CheckMandatory())).To(Equal(0))
		})
	})
})
