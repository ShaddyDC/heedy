package main

import (
	_ "connectordb/plugins/rest"
	_ "connectordb/plugins/run"
	_ "connectordb/plugins/shell"
	_ "connectordb/plugins/webclient"

	"connectordb/config"
	"connectordb/plugins"
	"connectordb/services"
	"connectordb/streamdb"
	"connectordb/streamdb/util"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"

	log "github.com/Sirupsen/logrus"
)

var (
	createFlags            = flag.NewFlagSet("create", flag.ExitOnError)
	createUsernamePassword = createFlags.String("user", "", "The initial user in username:password format")
	createEmail            = createFlags.String("email", "root@localhost", "The email address for the root user")
	createDbType           = createFlags.String("dbtype", "postgres", "The type of database to create.")

	startFlags = flag.NewFlagSet("start", flag.ExitOnError)
	forceStart = startFlags.Bool("force", false, "Force the start despite there being a connectordb pid file")

	stopFlags = flag.NewFlagSet("stop", flag.ExitOnError)

	upgradeFlags = flag.NewFlagSet("upgrade", flag.ExitOnError)

	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

	loglevel = flag.String("log", "INFO", "The log level to run at")
	logfile  = flag.String("logfile", "", "The log file to write to")
)

//PrintUsage gives a nice message of the functionality available from the executable
func PrintUsage() {
	fmt.Printf("ConnectorDB Version %v\nCompiled for %v using %v\n\n", streamdb.Version, runtime.GOARCH, runtime.Version())
	fmt.Printf("Usage:\nconnectordb [command] [path to database folder] [--flags] \n")

	fmt.Printf("\ncreate: Initialize a new database at the given folder\n")
	createFlags.PrintDefaults()
	fmt.Printf("\nstart: Starts the given database\n")
	startFlags.PrintDefaults()
	fmt.Printf("\nstop: Shuts down all processes associated with the given database.\n")
	stopFlags.PrintDefaults()
	fmt.Printf("\nupgrade: Upgrades an existing database to a newer version.\n")
	upgradeFlags.PrintDefaults()
	fmt.Printf("\n")

	// Print all usages of the plugins
	plugins.Usage()

	fmt.Printf("\n")

}

func init() {
	//log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// The main entrypoint into connectordb
func main() {

	// global system stuff
	flag.Parse()

	//Set up the log file
	if *logfile != "" {
		f, err := os.OpenFile(*logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Could not open file %s: %s", *logfile, err.Error())
		}
		defer f.Close()
		log.SetFormatter(new(log.JSONFormatter))
		log.SetOutput(f)
	}

	switch *loglevel {
	default:
		log.Fatalln("Unrecognized log level ", *loglevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	}

	if *cpuprofile != "" {
		log.Debug("Running CPU profile into: ", *cpuprofile)
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

		//It turns out that ctrl+c literally crashes the program, making defers not called.
		//This is a workaround
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for sig := range c {
				log.Printf("caught %v, writing cpu profile...", sig)

				pprof.StopCPUProfile()
				f.Close()
				os.Exit(1)
			}
		}()

	}

	// Make sure we don't go OOB
	if len(flag.Args()) < 2 {
		PrintUsage()
		return
	}

	// Choose our command
	var err error
	commandName := flag.Args()[0]
	dbPath := flag.Args()[1]

	// Make sure this is abs.
	dbPath, _ = filepath.Abs(dbPath)
	log.Debugln("Database path:", dbPath)
	// init and save later
	config.InitConfiguration(dbPath)
	defer config.SaveConfiguration()

	switch commandName {
	case "create":
		err = createDatabase()

	case "start":
		err = startDatabase(dbPath)

	case "stop":
		err = stopDatabase(dbPath)

	case "upgrade":
		err = upgradeDatabase(dbPath)

	default:
		err = runPlugin(commandName, dbPath)
		if err == plugins.ErrNoPlugin {
			PrintUsage()
			return
		}
	}

	if err != nil {
		log.Errorf("A problem occured during %v:\n\n%v\n", commandName, err)
	}
}

// processes the flags and makes sure they're valid, exiting if needed.
func processFlags(fs *flag.FlagSet) {
	fs.Parse(flag.Args()[2:])
}

// Does the creations step
func createDatabase() error {
	processFlags(createFlags)

	//extract the username and password from the formatted string
	usernamePasswordArray := strings.Split(*createUsernamePassword, ":")
	if len(usernamePasswordArray) != 2 {
		log.Errorln("--user: Username and password not given in format <username>:<password>")
		createFlags.PrintDefaults()
		return nil
	}
	username := usernamePasswordArray[0]
	password := usernamePasswordArray[1]

	config.GetConfiguration().DatabaseType = *createDbType
	log.Debugln("CONFIG:", config.GetConfiguration())

	log.Debugln("CONNECTORDB: Doing Init")
	if err := services.Init(config.GetConfiguration()); err != nil {
		return err
	}

	log.Debugln("CONNECTORDB: Creating Files")
	if err := services.Create(config.GetConfiguration(), username, password, *createEmail); err != nil {
		return err
	}

	log.Debugln("CONNECTORDB: Stopping any subsystems")

	services.Stop(config.GetConfiguration())
	//services.Kill(config.GetConfiguration())

	fmt.Printf("\nDatabase created successfully.\n")
	return nil
}

func startDatabase(dbPath string) error {
	processFlags(startFlags)

	dbPath, err := util.ProcessConnectordbDirectory(dbPath)
	if err != nil {
		if err == util.ErrAlreadyRunning && !*forceStart {
			fmt.Println("Use -force to force start the database even with connectordb.pid in there.")
			return err
		} else if err != util.ErrAlreadyRunning {
			return err
		}
	}

	if err := services.Init(config.GetConfiguration()); err != nil {
		return err
	}

	return services.Start(config.GetConfiguration())
}

func stopDatabase(dbPath string) error {
	processFlags(stopFlags)

	dbPath, err := util.ProcessConnectordbDirectory(dbPath)
	if err == nil {
		log.Warningln("Connectordb looks like it isn't already running, but we'll try anyway.")
		return err
	}

	if err := services.Init(config.GetConfiguration()); err != nil {
		return err
	}

	if err := services.Stop(config.GetConfiguration()); err != nil {
		log.Errorln(err.Error())
	}

	return nil
}

func upgradeDatabase(dbPath string) error {
	processFlags(upgradeFlags)

	// get cannonicalized path and make sure we're not already running
	dbPath, err := util.ProcessConnectordbDirectory(dbPath)
	if err != nil {
		return err
	}

	// Start the server

	return services.Upgrade()
}

func runPlugin(cmd, dbPath string) error {
	db, err := streamdb.Open(config.DefaultOptions)
	if err != nil {
		return err
	}
	defer db.Close()

	return plugins.Run(cmd, db, flag.Args()[2:])
}
