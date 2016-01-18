package tasks

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/starkandwayne/rdpgd/log"
)

/*EnforceFileRetention - Responsible for adding removing files which are no longer
needed on the local file system.  For example, backup files which have been created
successfully locally and copied to S3 successfully can be deleted to preserve
local disk storage */
func (t *Task) EnforceFileRetention() (err error) {

	/*
	   If s3 copy is enabled you cannot delete files until they have been copied to s3
	   otherwise keep the most recent backups, say the last 48 hours worth and delete all others

	*/

	//Select eligible files
	address := `127.0.0.1`
	sq := ``
	if isS3FileCopyEnabled() {
		sq = `SELECT a.params FROM backups.file_history a WHERE a.removed_at IS NULL AND a.action = 'CreateBackup' AND EXISTS (SELECT b.params FROM backups.file_history b WHERE a.cluster_id = b.cluster_id AND a.dbname = b.dbname AND a.node=b.node AND a.file_name = b.file_name AND b.action='CopyToS3' AND b.status='ok') AND NOT EXISTS (SELECT id FROM tasks.tasks WHERE action = 'DeleteFile' AND data = a.params::text)`
	} else {
		sq = `SELECT a.params FROM backups.file_history a WHERE a.removed_at IS NULL AND a.action = 'CreateBackup' AND a.status = 'ok' AND created_at < current_timestamp - '48 hours'::interval  AND NOT EXISTS (SELECT id FROM tasks.tasks WHERE action = 'DeleteFile' AND data = a.params::text)`
	}
	filesToDelete, err := getList(address, sq)
	if err != nil {
		log.Error(fmt.Sprintf(`tasks.Task<%d>#EnforceFileRetention() Failed to load list of files ! %s`, t.ID, err))
	}

	log.Trace(fmt.Sprintf("tasks.EnforceFileRetention() > Found %d files to delete", len(filesToDelete)))

	for _, fileToDeleteParams := range filesToDelete {
		log.Trace(fmt.Sprintf("tasks.EnforceFileRetention() > Attempting to add %s", fileToDeleteParams))

		newTask := Task{ClusterID: t.ClusterID, Node: t.Node, Role: t.Role, Action: "DeleteFile", Data: fileToDeleteParams, TTL: t.TTL, NodeType: t.NodeType}
		err = newTask.Enqueue()
		if err != nil {
			log.Error(fmt.Sprintf(`tasks.FindFilesToCopyToS3() service task schedules ! %s`, err))
		}
	}

	return
}

//DeleteFile - Delete a file from the operating system
func (t *Task) DeleteFile() (err error) {

	taskParams := []byte(t.Data)
	fm := S3FileMetadata{}
	err = json.Unmarshal(taskParams, &fm)
	if err != nil {
		log.Error(fmt.Sprintf("tasks.DeleteFile() json.Unmarshal() ! %s", err))
	}
	log.Error(fmt.Sprintf(`tasks.DeleteFile() Attempting to delete file "%s" `, fm.Location))
	err = os.Remove(fm.Location)
	if err != nil {
		log.Error(fmt.Sprintf(`tasks.DeleteFile() Attempted to delete file "%s" ! %s`, fm.Location, err))
	} else {
		//As long as the file was deleted, update the history record
		address := `127.0.0.1`
		sq := fmt.Sprintf(`UPDATE backups.file_history SET removed_at = CURRENT_TIMESTAMP WHERE params::text = '%s'`, t.Data)
		err = execQuery(address, sq)
		if err != nil {
			log.Error(fmt.Sprintf(`tasks.DeleteFile() Attempted to update backups.file_history using query <<<%s>>> ! %s`, sq, err))
		}

	}

	return
}

func isS3FileCopyEnabled() (isEnabled bool) {
	isEnabled = false
	if strings.ToUpper(os.Getenv(`RDPGD_S3_BACKUPS`)) == "ENABLED" {
		isEnabled = true
	}
	return
}
