package tasks

import (
	"fmt"

	"github.com/starkandwayne/rdpgd/log"
)

type backupFileHistory struct {
	backupFile        string
	backupPathAndFile string
	dbname            string
	node              string
	status            string
	duration          int
}

type s3FileHistory struct {
	fileName string
	source   string
	target   string
	dbname   string
	size     int64
	node     string
	status   string
	duration int
	bucket   string
}

//DeleteBackupHistory - Responsible for deleting records from backups.file_history
//older than the value in rdpg.config.key = defaultDaysToKeepFileHistory
func (t *Task) DeleteBackupHistory() (err error) {

	daysToKeep, err := getConfigKeyValue(`defaultDaysToKeepFileHistory`)
	log.Trace(fmt.Sprintf("tasks.DeleteBackupHistory() Keeping %s days of file history in backups.file_history", daysToKeep))

	address := `127.0.0.1`
	sq := fmt.Sprintf(`DELETE FROM backups.file_history WHERE created_at < NOW() - '%s days'::interval; `, daysToKeep)

	err = execQuery(address, sq)
	if err != nil {
		log.Error(fmt.Sprintf(`tasks.DeleteBackupHistory() Error when running query %s ! %s`, sq, err))
	}

	return

}

func insertBackupHistory(f backupFileHistory) (err error) {
	address := `127.0.0.1`
	sq := fmt.Sprintf(`INSERT INTO backups.file_history(cluster_id, dbname, node, file_name, action, status, duration, params) VALUES ('%s','%s','%s','%s','%s','%s',%d,'{"location":"%s","dbname":"%s","node":"%s","cluster_id":"%s"}')`, ClusterID, f.dbname, f.node, f.backupFile, `CreateBackup`, f.status, f.duration, f.backupPathAndFile, f.dbname, f.node, ClusterID)
	err = execQuery(address, sq)
	if err != nil {
		log.Error(fmt.Sprintf("tasks.insertHistory() Error inserting record into backups.file_history, running query: %s ! %s", sq, err))
	}
	return
}

func insertS3History(f s3FileHistory) (err error) {
	address := `127.0.0.1`
	sq := fmt.Sprintf(`INSERT INTO backups.file_history(cluster_id, dbname, node, file_name, action, status, duration, params) VALUES ('%s','%s','%s','%s','%s','%s',%d,'{"source":"%s", "target":"%s", "size":"%d", "bucket":"%s"}')`, ClusterID, f.dbname, f.node, f.fileName, `CopyToS3`, f.status, f.duration, f.source, f.target, f.size, f.bucket)
	err = execQuery(address, sq)
	if err != nil {
		log.Error(fmt.Sprintf("tasks.insertHistory() s3FileHistory  Error inserting record into backups.file_history, running query: %s ! %s", sq, err))
	}
	return
}

func insertS3HistoryCopyFromS3(f s3FileHistory) (err error) {
	address := `127.0.0.1`
	sq := fmt.Sprintf(`INSERT INTO backups.file_history(cluster_id, dbname, node, file_name, action, status, duration, params) VALUES ('%s','%s','%s','%s','%s','%s',%d,'{"source":"%s", "target":"%s", "size":"%d", "bucket":"%s"}')`, ClusterID, f.dbname, f.node, f.fileName, `CopyFromS3`, f.status, f.duration, f.source, f.target, f.size, f.bucket)
	err = execQuery(address, sq)
	if err != nil {
		log.Error(fmt.Sprintf("tasks.insertHistory() s3FileHistory  Error inserting record into backups.file_history, running query: %s ! %s", sq, err))
	}
	return
}
