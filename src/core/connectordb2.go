package main

import (
	"flag"
	"fmt"
	"github.com/kardianos/osext"
	"github.com/vharitonsky/iniflags"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"plugins/rest"
	"streamdb"
	"streamdb/config"
	"streamdb/dbutil"
	"streamdb/users"
	"strings"
	"time"
)

//ProgramUsage is a string describing how to use the program
const ProgramUsage = `

Usage: connectordb (create|start|stop|shell) directory

directory - the directory holding the configuration and the data for connectordb

create:
    Sets up a new connectordb instance at the given directory.

start:
    Starts connectordb and needed processes running from the given connectordb
    instance directory.

    Subcommands include:
        [blank] - default; start everything
        servers - redis, gnatsd, and postgres if required
        redis   - redis
        gnatsd  - gnatsd
        db      - postgres or none if working from sqlite
        rest    - the rest api
        web     - the web api

stop:
    Stops the connectordb instance running from the directory

`

const (
	//ConnectorDBConfigFileName is the file name to use for the configuration file in the database folder
	ConnectorDBConfigFileName = "cdb.ini"

	//DefaultFolderPermissions is The folder permissions to use when creating a database
	DefaultFolderPermissions = os.FileMode(0755)
)

var (
	createFlags = flag.NewFlagSet("create", flag.ContinueOnError)

	// TODO change this in the future, but until we have the final thing ironed out, leave this as is for testing
	createUsernamePassword = createFlags.String("user", "admin:admin", "The user in username:password format")
	createEmail            = createFlags.String("email", "root@localhost", "The email address for the root user")

	// TODO change this once postgres is fully working/tested and the SQL is up to code
	createDbType = createFlags.String("dbtype", "sqlite", "The type of database to create.")

	startFlags = flag.NewFlagSet("start", flag.ContinueOnError)
)

// The main entrypoint into connectordb
func main() {

	// we parse the flags here to make sure the usage will perform correctly
	flag.Parse()

	if len(flag.Args()) < 2 {
		fmt.Printf(ProgramUsage)
		flag.Usage()
		os.Exit(1)
	}

	// we use flag args just in case the user put flags before the params
	commandName := flag.Args()[0]
	processDirectory := flag.Args()[1]

	switch commandName {
	case "create":
		create(processDirectory)
	case "start":
		start(processDirectory)
	case "stop":
		stop(processDirectory)
	case "shell":
		fmt.Printf("Not yet implemented\n")
	default:
		fmt.Printf("Error: '%v' is not a valid command.\n", commandName)
		fmt.Printf(ProgramUsage)
		os.Exit(1)
	}
}

func waitForPortOpen(hostPort string) {
	var err error

	_, err = net.Dial("tcp", hostPort)

	for err != nil {
		_, err = net.Dial("tcp", hostPort)
	}
}

// Executes a command, redirecting the stdout and stderr to this program's output
func executeCommand(command string, args ...string) error {
	return execCommandRedirect(os.Stdout, os.Stderr, command, args...)
}

// Executes a command, redirecting the stdout and stderr to this program's output
func daemonizeCommand(logpath, command string, args ...string) error {
	// TODO setup infinite loop for restarting crashed processes

	file, err := os.Create(logpath) // For read access.
	if err != nil {
		return err
	}
	go execCommandRedirect(file, file, command, args...)

	return nil
}

// Executes a command doing redirects as necessary
func execCommandRedirect(stdout, stderr *os.File, command string, args ...string) error {
	cmd := exec.Command(command, args...)

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return cmd.Run()

	// TODO setup a global PID structure to track all pending and past operations
}

func createDeviceAndGetKey(udb *users.UserDatabase, user *users.User, devname string) (string, error) {
	err := udb.CreateDevice(devname, user.UserId)
	if err != nil {
		return "", err
	}

	dev, err := udb.ReadDeviceForUserByName(user.UserId, devname)
	if err != nil {
		return "", err
	}

	return dev.ApiKey, nil
}

func create(ProcessDir string) {
	createFlags.Parse(os.Args)

	userPass := strings.Split(*createUsernamePassword, ":")
	if len(userPass) != 2 {
		log.Fatal("Username and password not given in format <username>:<password>\n")
		return
	}
	createUsername := userPass[0]
	createPassword := userPass[1]

	execFolder, _ := osext.ExecutableFolder()

	log.Printf("Initial Setup...\n")

	// Create the initial directory
	log.Printf("> Creating Directory\n")

	if err := os.MkdirAll(ProcessDir, DefaultFolderPermissions); err != nil {
		log.Fatal(err.Error())
	}

	// Copy the config files over to the new folder
	log.Printf("> Copying Config\n")

	err := executeCommand("cp", execFolder+"/config/gnatsd.conf", execFolder+"/config/redis.conf", ProcessDir)

	if err != nil {
		log.Fatal(err.Error())
	}

	databasePath := ""
	switch *createDbType {
	case "sqlite":
		databasePath = filepath.Join(ProcessDir, "connectordb.sqlite3")

		// because sqlite doesn't always like being started on a file that
		// doesn't exist
		executeCommand("touch", databasePath)

	default: // postgres or misconfigured

		// Init the postgres database
		log.Printf("Setting Up Postgres\n")

		postgresPath := filepath.Join(ProcessDir, "connectordb_psql")
		postgresCmd := filepath.Join(execFolder, "config/runpostgres")
		err = executeCommand("bash", postgresCmd, "setup", postgresPath)

		if err != nil {
			log.Fatal(err.Error())
		}

		err = daemonizeCommand("bash", postgresCmd, "run", postgresPath)
		if err != nil {
			log.Fatal(err.Error())
		}

		databasePath = "postgres://localhost:52592/connectordb?sslmode=disable"

		time.Sleep(time.Second * 3)
		log.Printf("Waiting for port to open.")
		waitForPortOpen("localhost:52592")
		time.Sleep(time.Second * 2)
	}

	// Setup the tables
	err = dbutil.UpgradeDatabase(databasePath, true)

	if err != nil {
		log.Fatal("Upgrade failed:" + err.Error())
	}

	// Setup the admin user and two main devices
	log.Printf("Setting up the initial admin user.")
	db, driver, err := dbutil.OpenSqlDatabase(databasePath)

	if err != nil {
		log.Fatal("setup failed:" + err.Error())
	}

	var udb users.UserDatabase
	udb.InitUserDatabase(db, string(driver))

	// create the initial user
	err = udb.CreateUser(createUsername, *createEmail, createPassword)
	if err != nil {
		log.Fatal("create failed:" + err.Error())
	}

	usr, err := udb.ReadUserByName(createUsername)
	if err != nil {
		log.Fatal("read failed:" + err.Error())
	}

	restkey, err := createDeviceAndGetKey(&udb, usr, "rest")
	if err != nil {
		log.Fatal("create failed:" + err.Error())
	}

	webkey, err := createDeviceAndGetKey(&udb, usr, "website")
	if err != nil {
		log.Fatal("create failed:" + err.Error())
	}

	// Setup config
	flag.Set("database.cxn_string", databasePath)
	flag.Set("web.api.key", restkey)
	flag.Set("web.http.key", webkey)
	flag.Set("user", "")

	// Dump config
	config := getConfig()

	configPath := filepath.Join(ProcessDir, ConnectorDBConfigFileName)

	file, err := os.Create(configPath) // For read access.
	if err != nil {
		log.Fatal(err.Error())
	}

	defer file.Close()
	file.WriteString(config)

	log.Println("Finished all setup, exiting.")

	// TODO implement proper system teardown
}

// stolen from iniflags
func quoteValue(v string) string {
	if !strings.ContainsAny(v, "\n#;") && strings.TrimSpace(v) == v {
		return v
	}
	v = strings.Replace(v, "\\", "\\\\", -1)
	v = strings.Replace(v, "\n", "\\n", -1)
	v = strings.Replace(v, "\"", "\\\"", -1)
	return fmt.Sprintf("\"%s\"", v)
}

func escapeUsage(s string) string {
	return strings.Replace(s, "\n", "\n    # ", -1)
}

func getConfig() string {
	config := ""

	flag.VisitAll(func(f *flag.Flag) {
		if f.Name != "config" && f.Name != "dumpflags" {
			config += fmt.Sprintf("%s = %s  # %s\n", f.Name, quoteValue(f.Value.String()), escapeUsage(f.Usage))
		}
	})

	return config
}

func startPostgres(ProcessDir string) error {
	log.Println("Starting postgres")

	execFolder, _ := osext.ExecutableFolder()
	logPath := filepath.Join(ProcessDir, "postgres.log")
	executablePath := filepath.Join(execFolder, "config/runpostgres")
	dbPath := filepath.Join(ProcessDir, "connectordb_psql")

	return daemonizeCommand(logPath, "bash", executablePath, "run", dbPath)
}

func startGnatsd(ProcessDir string) error {
	log.Println("Starting gnatsd")

	execFolder, _ := osext.ExecutableFolder()
	logPath := filepath.Join(ProcessDir, "gnatsd.log")
	binaryPath := filepath.Join(execFolder, "dep/gnatsd")
	configPath := filepath.Join(ProcessDir, "gnatsd.conf")

	return daemonizeCommand(logPath, binaryPath, "-c", configPath)
}

func startRedis(ProcessDir string) error {
	log.Println("Starting redis")

	logPath := filepath.Join(ProcessDir, "redis.log")
	configPath := filepath.Join(ProcessDir, "redis.conf")

	return daemonizeCommand(logPath, "redis-server", configPath)
}

func start(ProcessDir string) {
	execFolder, _ := osext.ExecutableFolder()
	ProcessDir, _ = filepath.Abs(ProcessDir)

	cdbConfigPath := filepath.Join(ProcessDir, ConnectorDBConfigFileName)

	fmt.Printf("Starting connectordb...\n\n")
	fmt.Printf("Exec Folder: %v\n", execFolder)
	fmt.Printf("DB Folder  : %v\n", ProcessDir)
	fmt.Printf("ini path   : %v\n\n\n", cdbConfigPath)

	// load configuration, first we start with the flags library so we can
	// specify the loading path...
	flag.Parse()
	flag.Set("config", cdbConfigPath) // the inipath for iniflags

	iniflags.Parse() // Now we setup the iniflags which handles the sighup stuff

	subSubcommand := ""
	if len(flag.Args()) >= 3 {
		subSubcommand = flag.Args()[2]
	}

	var redisNeeded, dbNeeded, webNeeded, restNeeded, gnatsdNeeded bool
	switch subSubcommand {
	case "servers":
		redisNeeded = true
		dbNeeded = true
		gnatsdNeeded = true

	case "redis":
		redisNeeded = true

	case "gnatsd":
		gnatsdNeeded = true

	case "db":
		dbNeeded = true

	case "rest":
		restNeeded = true

	case "web":
		restNeeded = true
		webNeeded = true

	default:
		redisNeeded = true
		dbNeeded = true
		webNeeded = true
		restNeeded = true
		gnatsdNeeded = true
	}

	// Now start all the services we need
	if gnatsdNeeded {
		startGnatsd(ProcessDir)
	}

	if redisNeeded {
		startRedis(ProcessDir)
	}

	if dbNeeded && !dbutil.UriIsSqlite(*config.DatabaseConnection) {
		startPostgres(ProcessDir)
	}

	if webNeeded || restNeeded {
		time.Sleep(time.Second * 3)

		// start api + webservice
		var err error
		db, err := streamdb.Open(*config.DatabaseConnection, *config.RedisConnection, *config.MessageConnection)

		if err != nil {
			panic(err.Error())
		}

		defer db.Close()
		log.Println("Running REST server")
		r := rest.Router(db, nil)
		http.Handle("/", r)

		serveraddr := fmt.Sprintf(":%d", *config.WebPort)
		err = http.ListenAndServe(serveraddr, nil)
		log.Fatal(err)
	}
}

func stop(ProcessDir string) {
	fmt.Printf("Stopping connectordb...\n")
}
