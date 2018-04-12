package main

import (
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/executor"
	"github.com/percona/dcos-mongo-tools/executor/metrics"
	"github.com/percona/dcos-mongo-tools/executor/pmm"
	log "github.com/sirupsen/logrus"
)

var (
	mongod                     = kingpin.Command("mongod", "run a mongod instance")
	mongos                     = kingpin.Command("mongos", "run a mongos instance")
	DefaultDelayBackgroundJob  = "15s"
	DefaultMetricsIntervalSecs = "10"
)

func handleMetrics(cnf *executor.Config) {
	kingpin.Flag(
		"metrics.enable",
		"Enable DC/OS Metrics monitoring for MongoDB, defaults to "+common.EnvMetricsEnabled+" env var",
	).Envar(common.EnvMetricsEnabled).BoolVar(&cnf.Metrics.Enabled)
	kingpin.Flag(
		"metrics.intervalSecs",
		"The frequency (in seconds) to send metrics to DC/OS Metrics service, defaults to "+common.EnvMetricsIntervalSecs+" env var",
	).Default(DefaultMetricsIntervalSecs).Envar(common.EnvMetricsIntervalSecs).UintVar(&cnf.Metrics.IntervalSecs)
}

func handlePmm(cnf *executor.Config) {
	kingpin.Flag(
		"pmm.configDir",
		"Directory containing the PMM client config file (pmm.yml), defaults to "+common.EnvMesosSandbox+" env var",
	).Envar(common.EnvMesosSandbox).StringVar(&cnf.PMM.ConfigDir)
	kingpin.Flag(
		"pmm.enable",
		"Enable Percona PMM monitoring for OS and MongoDB, defaults to "+common.EnvPMMEnabled+" env var",
	).Envar(common.EnvPMMEnabled).BoolVar(&cnf.PMM.Enabled)
	kingpin.Flag(
		"pmm.enableQueryAnalytics",
		"Enable Percona PMM query analytics (QAN) client/agent, defaults to "+common.EnvPMMEnableQueryAnalytics+" env var",
	).Envar(common.EnvPMMEnableQueryAnalytics).BoolVar(&cnf.PMM.EnableQueryAnalytics)
	kingpin.Flag(
		"pmm.serverAddress",
		"Percona PMM server address, defaults to "+common.EnvPMMServerAddress+" env var",
	).Envar(common.EnvPMMServerAddress).StringVar(&cnf.PMM.ServerAddress)
	kingpin.Flag(
		"pmm.clientName",
		"Percona PMM client address, defaults to "+common.EnvTaskName+" env var",
	).Envar(common.EnvTaskName).StringVar(&cnf.PMM.ClientName)
	kingpin.Flag(
		"pmm.serverSSL",
		"Enable SSL communication between Percona PMM client and server, defaults to "+common.EnvPMMServerSSL+" env var",
	).Envar(common.EnvPMMServerSSL).BoolVar(&cnf.PMM.ServerSSL)
	kingpin.Flag(
		"pmm.serverInsecureSSL",
		"Enable insecure SSL communication between Percona PMM client and server, defaults to "+common.EnvPMMServerInsecureSSL+" env var",
	).Envar(common.EnvPMMServerInsecureSSL).BoolVar(&cnf.PMM.ServerInsecureSSL)
	kingpin.Flag(
		"pmm.linuxMetricsExporterPort",
		"Port number for bind Percona PMM Linux Metrics exporter to, defaults to "+common.EnvPMMLinuxMetricsExporterPort+" env var",
	).Envar(common.EnvPMMLinuxMetricsExporterPort).UintVar(&cnf.PMM.LinuxMetricsExporterPort)
	kingpin.Flag(
		"pmm.mongodbMetricsExporterPort",
		"Port number for bind Percona PMM MongoDB Metrics exporter to, defaults to "+common.EnvPMMMongoDBMetricsExporterPort+" env var",
	).Envar(common.EnvPMMMongoDBMetricsExporterPort).UintVar(&cnf.PMM.MongoDBMetricsExporterPort)
	kingpin.Flag(
		"pmm.mongodb.clusterName",
		"Percona PMM client mongodb cluster name, defaults to "+common.EnvFrameworkName+" env var",
	).Envar(common.EnvFrameworkName).StringVar(&cnf.PMM.MongoDB.ClusterName)
}

func main() {
	dbConfig := common.NewDBConfig(
		common.EnvMongoDBClusterMonitorUser,
		common.EnvMongoDBClusterMonitorPassword,
	)
	cnf := &executor.Config{
		DB: dbConfig,
		Metrics: &metrics.Config{
			DB: dbConfig,
		},
		PMM: &pmm.Config{
			DB:      dbConfig,
			MongoDB: &pmm.ConfigMongoDB{},
		},
		Tool: common.NewToolConfig(os.Args[0]),
	}

	kingpin.Flag(
		"framework",
		"dcos framework name, overridden by env var "+common.EnvFrameworkName,
	).Default(common.DefaultFrameworkName).Envar(common.EnvFrameworkName).StringVar(&cnf.FrameworkName)
	kingpin.Flag(
		"configDir",
		"path to mongodb instance config file, defaults to $"+common.EnvMesosSandbox+" if available, otherwise "+executor.DefaultMongoConfigDirFallback,
	).Default(executor.DefaultMongoConfigDirFallback).Envar(common.EnvMesosSandbox).StringVar(&cnf.ConfigDir)
	kingpin.Flag(
		"binDir",
		"path to mongodb binary directory",
	).Default(executor.DefaultBinDir).StringVar(&cnf.BinDir)
	kingpin.Flag(
		"tmpDir",
		"path to mongodb temporary directory, defaults to $"+common.EnvMesosSandbox+"/tmp if available, otherwise "+executor.DefaultTmpDirFallback,
	).Default(executor.MesosSandboxPathOrFallback(
		"tmp",
		executor.DefaultTmpDirFallback,
	)).StringVar(&cnf.TmpDir)
	kingpin.Flag(
		"user",
		"user to run mongodb instance as",
	).Default(executor.DefaultUser).StringVar(&cnf.User)
	kingpin.Flag(
		"group",
		"group to run mongodb instance as",
	).Default(executor.DefaultGroup).StringVar(&cnf.Group)
	kingpin.Flag(
		"connectTries",
		"number of times to retry the connection/ping to mongodb",
	).Default(executor.DefaultConnectTries).UintVar(&cnf.ConnectTries)
	kingpin.Flag(
		"connectRetrySleep",
		"duration to wait between retries of the connection/ping to mongodb",
	).Default(executor.DefaultConnectRetrySleep).DurationVar(&cnf.ConnectRetrySleep)
	kingpin.Flag(
		"delayBackgroundJobs",
		"Amount of time to delay running of executor background jobs",
	).Default(DefaultDelayBackgroundJob).DurationVar(&cnf.DelayBackgroundJob)

	handleMetrics(cnf)
	handlePmm(cnf)

	cnf.NodeType = kingpin.Parse()
	e := executor.New(cnf)

	if cnf.Tool.PrintVersion {
		cnf.Tool.PrintVersionAndExit()
	}

	common.SetupLogger(cnf.Tool)

	switch cnf.NodeType {
	case executor.NodeTypeMongod:
		mongod := executor.NewMongod(cnf)
		err := e.Run(mongod)
		if err != nil {
			log.Errorf("Failed to start mongod: %s", err)
			return
		}
	case executor.NodeTypeMongos:
		log.Error("mongos nodes are not supported yet!")
		return
	default:
		log.Error("did not start anything, this is unexpected")
		return
	}
}
