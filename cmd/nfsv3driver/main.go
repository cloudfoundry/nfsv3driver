package main

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"time"

	cf_http "code.cloudfoundry.org/cfhttp"
	cf_debug_server "code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/nfsdriver"
	"code.cloudfoundry.org/nfsdriver/oshelper"

	"strconv"

	"code.cloudfoundry.org/goshims/bufioshim"
	"code.cloudfoundry.org/goshims/execshim"
	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/ioutilshim"
	"code.cloudfoundry.org/goshims/ldapshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagerflags"
	"code.cloudfoundry.org/nfsv3driver"
	"code.cloudfoundry.org/nfsv3driver/driveradmin/driveradminhttp"
	"code.cloudfoundry.org/nfsv3driver/driveradmin/driveradminlocal"
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"code.cloudfoundry.org/voldriver/invoker"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var atAddress = flag.String(
	"listenAddr",
	"127.0.0.1:7589",
	"host:port to serve volume management functions",
)

var adminAddress = flag.String(
	"adminAddr",
	"127.0.0.1:7590",
	"host:port to serve process admin functions",
)

var driversPath = flag.String(
	"driversPath",
	"",
	"Path to directory where drivers are installed",
)

var transport = flag.String(
	"transport",
	"tcp",
	"Transport protocol to transmit HTTP over",
)

var mapfsPath = flag.String(
	"mapfsPath",
	"/var/vcap/packages/mapfs/bin/mapfs",
	"Path to the mapfs binary",
)

var mountDir = flag.String(
	"mountDir",
	"/tmp/volumes",
	"Path to directory where NFS v3 volumes are created",
)

var requireSSL = flag.Bool(
	"requireSSL",
	false,
	"whether the NFS v3 driver should require ssl-secured communication",
)

var caFile = flag.String(
	"caFile",
	"",
	"the certificate authority public key file to use with ssl authentication",
)

var certFile = flag.String(
	"certFile",
	"",
	"the public key file to use with ssl authentication",
)

var keyFile = flag.String(
	"keyFile",
	"",
	"the private key file to use with ssl authentication",
)
var clientCertFile = flag.String(
	"clientCertFile",
	"",
	"the public key file to use with client ssl authentication",
)

var clientKeyFile = flag.String(
	"clientKeyFile",
	"",
	"the private key file to use with client ssl authentication",
)

var insecureSkipVerify = flag.Bool(
	"insecureSkipVerify",
	false,
	"whether SSL communication should skip verification of server IP addresses in the certificate",
)

var sourceFlagAllowed = flag.String(
	"allowed-in-source",
	"",
	"This is a comma separted list of parameters allowed to be send in share url. Each of this parameters can be specify by brokers",
)

var sourceFlagDefault = flag.String(
	"default-in-source",
	"",
	"This is a comma separted list of like params:value. This list specify default value of parameters. If parameters has default value and is not in allowed list, this default value become a forced value who's cannot be override",
)

var mountFlagAllowed = flag.String(
	"allowed-in-mount",
	"",
	"This is a comma separted list of parameters allowed to be send in extra config. Each of this parameters can be specify by brokers",
)

var mountFlagDefault = flag.String(
	"default-in-mount",
	"",
	"This is a comma separted list of like params:value. This list specify default value of parameters. If parameters has default value and is not in allowed list, this default value become a forced value who's cannot be override",
)

var useMockMounter = flag.Bool(
	"useMockMounter",
	false,
	"Whether to use a mock mounter for integration test",
)

var mockMountSeconds = flag.Int64(
	"mockMountSeconds",
	11,
	"How many seconds it takes to mount simulated by mock mounter",
)

var uniqueVolumeIds = flag.Bool(
	"uniqueVolumeIds",
	false,
	"whether the NFS v3 driver should opt-in to unique volumes",
)

const fsType = "nfs"
const mountOptions = "rsize=1048576,wsize=1048576,hard,intr,timeo=600,retrans=2,actimeo=0"

// static variables pulled from the environment
var (
	ldapSvcUser  string
	ldapSvcPass  string
	ldapUserFqdn string
	ldapHost     string
	ldapPort     int
	ldapCACert   string
	ldapProto    string
	ldapTimeout  int
)

func main() {
	parseCommandLine()
	parseEnvironment()

	var localDriverServer ifrit.Runner
	var idResolver nfsv3driver.IdResolver
	var mounter, legacyMounter nfsdriver.Mounter

	logger, logTap := newLogger()
	logger.Info("start")
	defer logger.Info("end")

	source := nfsv3driver.NewNfsV3ConfigDetails()
	source.ReadConf(*sourceFlagAllowed, *sourceFlagDefault, []string{})

	mounts := nfsv3driver.NewNfsV3ConfigDetails()
	mounts.ReadConf(*mountFlagAllowed, *mountFlagDefault, []string{})

	if ldapHost != "" {
		idResolver = nfsv3driver.NewLdapIdResolver(
			ldapSvcUser,
			ldapSvcPass,
			ldapHost,
			ldapPort,
			ldapProto,
			ldapUserFqdn,
			ldapCACert,
			&ldapshim.LdapShim{},
			time.Duration(ldapTimeout)*time.Second,
		)
	}

	if *useMockMounter {
		mounter = nfsv3driver.NewMockMounter(time.Duration(*mockMountSeconds)*time.Second, logger)
	} else {
		config := nfsv3driver.NewNfsV3Config(source, mounts)
		legacyMounter = nfsv3driver.NewNfsV3Mounter(
			invoker.NewRealInvoker(),
			&osshim.OsShim{},
			&ioutilshim.IoutilShim{},
			&bufioshim.BufioShim{},
			config,
			idResolver,
		)
		mounter = nfsv3driver.NewMapfsMounter(
			invoker.NewRealInvoker(),
			nfsv3driver.NewBackgroundInvoker(&execshim.ExecShim{}),
			legacyMounter,
			&osshim.OsShim{},
			&ioutilshim.IoutilShim{},
			&bufioshim.BufioShim{},
			fsType,
			mountOptions,
			idResolver,
			config,
			*mapfsPath,
		)
	}

	client := nfsdriver.NewNfsDriver(
		logger,
		&osshim.OsShim{},
		&filepathshim.FilepathShim{},
		&ioutilshim.IoutilShim{},
		*mountDir,
		mounter,
		oshelper.NewOsHelper(),
	)

	if *transport == "tcp" {
		localDriverServer = createNfsDriverServer(logger, client, *atAddress, *driversPath, false, false)
	} else if *transport == "tcp-json" {
		localDriverServer = createNfsDriverServer(logger, client, *atAddress, *driversPath, true, *uniqueVolumeIds)
	} else {
		localDriverServer = createNfsDriverUnixServer(logger, client, *atAddress)
	}

	servers := grouper.Members{
		{Name: "localdriver-server", Runner: localDriverServer},
	}

	if dbgAddr := cf_debug_server.DebugAddress(flag.CommandLine); dbgAddr != "" {
		servers = append(grouper.Members{
			{Name: "debug-server", Runner: cf_debug_server.Runner(dbgAddr, logTap)},
		}, servers...)
	}

	adminClient := driveradminlocal.NewDriverAdminLocal()
	adminHandler, _ := driveradminhttp.NewHandler(logger, adminClient)
	// TODO handle error
	adminServer := http_server.New(*adminAddress, adminHandler)

	servers = append(grouper.Members{
		{Name: "driveradmin", Runner: adminServer},
	}, servers...)

	process := ifrit.Invoke(processRunnerFor(servers))
	logger.Info("started")

	adminClient.SetServerProc(process)
	adminClient.RegisterDrainable(client)

	untilTerminated(logger, process)
}

func exitOnFailure(logger lager.Logger, err error) {
	if err != nil {
		logger.Fatal("fatal-err-aborting", err)
	}
}

func untilTerminated(logger lager.Logger, process ifrit.Process) {
	err := <-process.Wait()
	exitOnFailure(logger, err)
}

func processRunnerFor(servers grouper.Members) ifrit.Runner {
	return sigmon.New(grouper.NewOrdered(os.Interrupt, servers))
}

func createNfsDriverServer(logger lager.Logger, client voldriver.Driver, atAddress, driversPath string, jsonSpec bool, uniqueVolumeIds bool) ifrit.Runner {
	advertisedUrl := "http://" + atAddress
	logger.Info("writing-spec-file", lager.Data{"location": driversPath, "name": "nfsv3driver", "address": advertisedUrl})
	if jsonSpec {
		driverJsonSpec := voldriver.DriverSpec{Name: "nfsv3driver", Address: advertisedUrl, UniqueVolumeIds: uniqueVolumeIds}

		if *requireSSL {
			absCaFile, err := filepath.Abs(*caFile)
			exitOnFailure(logger, err)
			absClientCertFile, err := filepath.Abs(*clientCertFile)
			exitOnFailure(logger, err)
			absClientKeyFile, err := filepath.Abs(*clientKeyFile)
			exitOnFailure(logger, err)
			driverJsonSpec.TLSConfig = &voldriver.TLSConfig{InsecureSkipVerify: *insecureSkipVerify, CAFile: absCaFile, CertFile: absClientCertFile, KeyFile: absClientKeyFile}
			driverJsonSpec.Address = "https://" + atAddress
		}

		jsonBytes, err := json.Marshal(driverJsonSpec)

		exitOnFailure(logger, err)
		err = voldriver.WriteDriverSpec(logger, driversPath, "nfsv3driver", "json", jsonBytes)
		exitOnFailure(logger, err)
	} else {
		err := voldriver.WriteDriverSpec(logger, driversPath, "nfsv3driver", "spec", []byte(advertisedUrl))
		exitOnFailure(logger, err)
	}

	handler, err := driverhttp.NewHandler(logger, client)

	exitOnFailure(logger, err)

	var server ifrit.Runner
	if *requireSSL {
		tlsConfig, err := cf_http.NewTLSConfig(*certFile, *keyFile, *caFile)
		if err != nil {
			logger.Fatal("tls-configuration-failed", err)
		}
		server = http_server.NewTLSServer(atAddress, handler, tlsConfig)
	} else {
		server = http_server.New(atAddress, handler)
	}

	return server
}

func createNfsDriverUnixServer(logger lager.Logger, client voldriver.Driver, atAddress string) ifrit.Runner {
	handler, err := driverhttp.NewHandler(logger, client)
	exitOnFailure(logger, err)
	return http_server.NewUnixServer(atAddress, handler)
}

func newLogger() (lager.Logger, *lager.ReconfigurableSink) {
	sink, err := lager.NewRedactingWriterSink(os.Stdout, lager.DEBUG, nil, nil)
	if err != nil {
		panic(err)
	}
	logger, reconfigurableSink := lagerflags.NewFromSink("nfs-driver-server", sink)
	return logger, reconfigurableSink
}

func parseCommandLine() {
	lagerflags.AddFlags(flag.CommandLine)
	cf_debug_server.AddFlags(flag.CommandLine)
	flag.Parse()
}

func parseEnvironment() {
	ldapSvcUser, _ = os.LookupEnv("LDAP_SVC_USER")
	ldapSvcPass, _ = os.LookupEnv("LDAP_SVC_PASS")
	ldapUserFqdn, _ = os.LookupEnv("LDAP_USER_FQDN")
	ldapHost, _ = os.LookupEnv("LDAP_HOST")
	port, _ := os.LookupEnv("LDAP_PORT")
	ldapPort, _ = strconv.Atoi(port)
	ldapCACert, _ = os.LookupEnv("LDAP_CA_CERT")
	ldapProto, _ = os.LookupEnv("LDAP_PROTO")
	timeout, _ := os.LookupEnv("LDAP_TIMEOUT")
	ldapTimeout, _ = strconv.Atoi(timeout)

	if ldapProto == "" {
		ldapProto = "tcp"
	}

	if ldapHost != "" && (ldapSvcUser == "" || ldapSvcPass == "" || ldapUserFqdn == "" || ldapPort == 0) {
		panic("LDAP is enabled but required LDAP parameters are not set.")
	}

	if ldapTimeout < 0 {
		panic("LDAP_TIMEOUT is set to negtive value")
	}

	// if ldapTimeout is not set, use default value
	if ldapTimeout == 0 {
		ldapTimeout = 120
	}
}
