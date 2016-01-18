/* Handlers for API calls to interact with the backup system */
package admin

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"errors"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/starkandwayne/rdpgd/consul"
	"github.com/starkandwayne/rdpgd/globals"
	"github.com/starkandwayne/rdpgd/instances"
	"github.com/starkandwayne/rdpgd/log"
	"github.com/starkandwayne/rdpgd/tasks"

	//"strings"
)

const localBackupPath string = "/var/vcap/store/pgbdr/backups"
const backupFileSuffix string = ".sql"

/* Should contain a form value dbname which equals the database name
   e.g. curl www.hostname.com/backup/now -X POST -d "dbname=nameofdatabase"
   The {how} should be either "now" or "enqueue" */
func BackupHandler(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	dbname := request.FormValue("dbname")
	t := tasks.NewTask()
	t.Action = "BackupDatabase"
	t.Data = dbname
	t.Node = globals.MyIP
	t.Role = globals.ServiceRole
	t.TTL = 3600
	t.ClusterService = globals.ClusterService
	t.NodeType = "read"
	if consul.IsWriteNode(globals.MyIP) {
		t.NodeType = "write"
	}

	var err error
	if dbname != "rdpg" {
		//Using FindByDatabase to determine if the database actually exists to be backed up.
		inst, err := instances.FindByDatabase(dbname)
		if err != nil {
			log.Error(fmt.Sprintf("admin.BackupHandler() instances.FindByDatabase(%s) Error occurred when searching for database.", dbname))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error encountered while searching for database"))
			return
		}
		if inst == nil {
			//...then the database doesn't exist on this cluster.
			log.Debug(fmt.Sprintf("admin.BackupHandler() Attempt to initiate backup on non-existant database with name: %s", dbname))
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Database not found"))
			return
		}
	}

	switch vars[`how`] {
	//Immediately calls Backup() and performs the backup
	case "now":
		err = t.BackupDatabase()
		if err != nil {
			log.Error(fmt.Sprintf(`api.BackupHandler() Task.BackupDatabase() %+v ! %s`, t, err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error encountered while trying to perform backup"))
			return
		}
		w.Write([]byte("Backup completed."))
	case "enqueue":
		// Queues up a backup to be done with a worker thread gets around to it.
		// This call returns after the queuing process is done; not after the backup is done.
		err = t.Enqueue()
		if err != nil {
			log.Error(fmt.Sprintf(`api.BackupHandler() Task.Enqueue() %+v ! %s`, t, err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error while trying to queue"))
			return
		}
		w.Write([]byte("Backup successfully queued."))
	}
}

type databaseBackupList struct {
	Database string
	Backups  []dbBackup
}

type dbBackup struct {
	Name string
	Size string
}

/* Lists all the backup files at the desired location.
   The {where} should be "local" or "remote". The "local" option finds all the backup
   files on the local filesystem. The "remote" option will display all of the
   backup files on the remote storage, such as S3, but this feature is not yet
   implemented. Backups are returned in json format.
   The request must be a POST.
   Form values:
     "fmt", "filename" is the only supported value at present. Defaults to "filename"
            if absent or if left blank
     "dbname": the name of the database for which to query backups of. Will eventually
            allow being left blank to return all database backups, but not yet. */
func BackupListHandler(w http.ResponseWriter, request *http.Request) {
	//Default printing format to print pretty timestamps. So pretty.
	printFormat := "filename"
	if request.Method == "POST" && request.FormValue("fmt") != "" {
		printFormat = request.FormValue("fmt")
	}
	backupList := []databaseBackupList{}
	// If the dbname wasn't specified of if the field is blank, then return the backups of
	// all databases.
	dbname := request.FormValue("dbname")
	// Where are we getting the files from?
	vars := mux.Vars(request)
	var err error
	switch vars["where"] {
	case "local":
		backupList, err = handleLocalListing(dbname)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	case "remote":
		log.Warn(fmt.Sprintf(`api.BackupListHandler() Remote Backup Listing not yet supported.`))
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Not Yet Implemented"))
		return
	}

	switch printFormat {
	case "filename":
		outputString := "{ "
		var separator string
		for i, d := range backupList {
			if i == 0 {
				separator = ""
			} else {
				separator = ", "
			}
			outputString = outputString + fmt.Sprintf("%s\"%s\": [", separator, d.Database)
			for j, v := range d.Backups {
				if j == 0 {
					separator = ""
				} else {
					separator = ", "
				}
				outputString = outputString + fmt.Sprintf("%s{ \"Name\": \"%s\", \"Bytes\": \"%s\" }", separator, v.Name, v.Size)
			}
			outputString = outputString + "]"
		}
		outputString = outputString + "}"
		w.Write([]byte(outputString))
	//case "timestamp": TODO
	default:
		log.Debug(fmt.Sprintf(`api.BackupListHandler() Requested unsupported format.`))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported Format Requested"))
	}
}

func handleLocalListing(dbname string) (backupList []databaseBackupList, err error) {
	backupList = []databaseBackupList{}
	dirListing, err := ioutil.ReadDir(localBackupLocation())
	if err != nil {
		log.Trace(fmt.Sprintf("admin.backup.handleLocalListing() No backups present on this cluster."))
		err = nil
		return
	}
	matchingString := fmt.Sprintf(".+%s\\z", regexp.QuoteMeta(backupFileSuffix))
	for _, dir := range dirListing {
		if (dir.Name() == dbname || dbname == "") && dir.IsDir() {
			thisDatabase := databaseBackupList{dir.Name(), []dbBackup{}}
			backupFiles, err := ioutil.ReadDir(localBackupLocation() + dir.Name())
			// This could result from a folder getting deleted between the original search and
			// trying to access it. This shouldn't happen, as even moving backups to S3 shouldn't cause the backup
			// folder to get deleted... so treat it as a true error.
			if err != nil {
				log.Error(fmt.Sprintf("Error when attempting to open directory: %s ! %s", localBackupLocation()+dir.Name(), err))
				return backupList, errors.New("An error occurred when trying to open a backup directory")
			}
			for _, f := range backupFiles {
				matched, err := regexp.Match(matchingString, []byte(f.Name()))
				if err != nil {
					log.Error(fmt.Sprintf(`api.BackupListHander() Error when attempting regexp: %s / %s ! %s`, matchingString, f.Name(), err))
					return backupList, errors.New("A regexp error occurred")
				}
				//The match matches on "<basename><backupFileSuffix>" e.g "asdf.sql"
				if f.Mode().IsRegular() && matched {
					thisDatabase.Backups = append(thisDatabase.Backups, dbBackup{f.Name(), strconv.FormatInt(f.Size(), 10)})
				}
			}
			backupList = append(backupList, thisDatabase)
		}
	}
	return
}

func localBackupLocation() string {
	return localBackupPath + "/"
}
