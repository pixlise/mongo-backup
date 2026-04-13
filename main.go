package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/pixlise/core/v4/api/services"
	"github.com/pixlise/core/v4/core/awsutil"
	"github.com/pixlise/core/v4/core/fileaccess"
	"github.com/pixlise/core/v4/core/logger"
	"github.com/pixlise/core/v4/core/mongoDBConnection"
	"github.com/pixlise/core/v4/core/mongobackup"
	"github.com/pixlise/core/v4/core/timestamper"
)

// Run this to do periodic mongo backups. Dumps and zips the DB and uploads it to an S3 location.
// Provide command line arguments to control operation

func main() {
	// First backup will run this many seconds after starting this process
	startupSecStr := os.Getenv("STARTUP_SEC")

	// Backups will be run on this interval of seconds
	intervalSecStr := os.Getenv("INTERVAL_SEC")

	// Mongo DB host. Can specify multiple and require connection to a secondary
	mongoHost := os.Getenv("DB_HOST")

	// Mongo DB Username
	mongoUsername := os.Getenv("DB_USER")

	// Mongo DB Password
	mongoPassword := os.Getenv("DB_PASSWORD")

	// S3 bucket to write backup to
	backupBucket := os.Getenv("BACKUP_BUCKET")

	// S3 bucket path to write backup to
	backupS3Path := os.Getenv("BACKUP_PATH")

	// Name of database to back up
	dbName := os.Getenv("BACKUP_DB")

	// Validate everything
	startupSec, err := strconv.Atoi(startupSecStr)
	if err != nil || startupSec <= 0 {
		log.Fatalln("STARTUP_SEC must be a positive number")
		return
	}

	intervalSec, err := strconv.Atoi(intervalSecStr)
	if err != nil || intervalSec < 0 {
		log.Fatalln("INTERVAL_SEC must be a positive number")
		return
	}

	if intervalSec == 0 {
		fmt.Printf("NOTE: INTERVAL_SEC is set to 0, so backup will only run once and process will then exit")
	}

	if len(mongoHost) <= 0 {
		log.Fatalln("DB_HOST must not be empty")
		return
	}

	if len(backupBucket) <= 0 {
		log.Fatalln("BACKUP_BUCKET must be set to the name of the s3 bucket to write to")
		return
	}

	if len(backupS3Path) <= 0 {
		log.Fatalln("BACKUP_PATH must be set to the path within the s3 bucket to write to")
		return
	}

	if len(dbName) <= 0 {
		log.Fatalln("BACKUP_DB must be set to the name of the db")
		return
	}

	sess, err := awsutil.GetSession()
	if err != nil {
		log.Fatalf("Failed to create AWS S3 service. Error: %v", err)
	}

	s3svc, err := awsutil.GetS3(sess)
	if err != nil {
		log.Fatalf("Failed to create AWS S3 service. Error: %v", err)
	}

	remoteFS := fileaccess.MakeS3Access(s3svc)

	hostURI := mongoDBConnection.MakeMongoURI(mongoHost, "")

	fmt.Printf("Mongo URI: %v\n", hostURI)

	svcs := &services.APIServices{
		Log: &logger.StdOutLoggerForTest{},
		MongoConnectInfo: mongoDBConnection.MongoConnectionInfo{
			Host:     hostURI,
			Username: mongoUsername,
			Password: mongoPassword,
		},
		TimeStamper: &timestamper.UnixTimeNowStamper{},
		FS:          remoteFS,
	}

	fmt.Printf("Waiting %v seconds to start...\n", startupSec)
	time.Sleep(time.Duration(startupSec) * time.Second)

	errCount := 0
	for {
		fmt.Printf("Running backup...\n")
		err = mongobackup.BackupDB(dbName, backupBucket, backupS3Path, true, svcs)
		if err != nil {
			log.Printf("Error: %v\n", err)
			errCount++
		}

		if intervalSec == 0 {
			fmt.Println("Backup complete, exiting due to interval configuration of 0 meaning don't re-run")
			break
		}

		fmt.Printf("Waiting %v seconds for next backup interval...\n", intervalSec)
		time.Sleep(time.Duration(intervalSec) * time.Second)
	}
}
