package admin

import (
	"fmt"

	"github.com/starkandwayne/rdpgd/log"
	"github.com/starkandwayne/rdpgd/pg"
)

func execQuery(address string, sq string) (queryRowCount int, err error) {
	p := pg.NewPG(address, "7432", `rdpg`, `rdpg`, pgPass)
	db, err := p.Connect()
	if err != nil {
		return -1, err
	}
	defer db.Close()
	var rowCount []int
	err = db.Select(&rowCount, sq)
	if err != nil {
		return -1, err
	}
	return rowCount[0], nil
}

func getRowCount(sq string) (rowCount int, err error) {

	address := `127.0.0.1`
	rowCount, err = execQuery(address, sq)
	return
	//Expect(nodeRowCount[0]).To(Equal(0))

}

func getQueueDepth() (rowCount int) {
	sq := `SELECT COUNT(*) FROM tasks.tasks;`
	rowCount, err := getRowCount(sq)
	if err != nil {
		log.Error(fmt.Sprintf("admin.getQueueDepth() Could not get row count running query %s ! %s", sq, err))
		return -1
	}
	return
}

func getNumberOfBoundDatabases() (rowCount int) {
	sq := `SELECT COUNT(*) FROM cfsb.instances WHERE instance_id IS NOT NULL AND effective_at IS NOT NULL AND ineffective_at IS NULL AND decommissioned_at IS NULL;`
	rowCount, err := getRowCount(sq)
	if err != nil {
		log.Error(fmt.Sprintf("admin.getNumberOfBoundDatabases() Could not get row count running query %s ! %s", sq, err))
		return -1
	}
	return
}

func getNumberOfFreeDatabases() (rowCount int) {
	sq := `SELECT COUNT(*) FROM cfsb.instances WHERE instance_id IS NULL AND effective_at IS NOT NULL AND ineffective_at IS NULL AND decommissioned_at IS NULL;`
	rowCount, err := getRowCount(sq)
	if err != nil {
		log.Error(fmt.Sprintf("admin.getNumberOfFreeDatabases() Could not get row count running query %s ! %s", sq, err))
		return -1
	}
	return
}

func getNumberOfReplicationSlots() (rowCount int) {
	sq := `SELECT COUNT(*) FROM pg_replication_slots WHERE active=true;`
	rowCount, err := getRowCount(sq)
	if err != nil {
		log.Error(fmt.Sprintf("admin.getNumberOfReplicationSlots() Could not get row count running query %s ! %s", sq, err))
		return -1
	}
	return
}

func getNumberOfDatabaseBackupOnDisk() (rowCount int) {
	sq := `SELECT COUNT(*) FROM backups.file_history WHERE action = 'CreateBackup' AND removed_at IS NULL;`
	rowCount, err := getRowCount(sq)
	if err != nil {
		log.Error(fmt.Sprintf("admin.getNumberOfDatabaseBackupOnDisk() Could not get row count running query %s ! %s", sq, err))
		return -1
	}
	return
}

//func getNumberOfDatabaseBackupOnDisk() (rowCount int) {
//	sq := `SELECT datname AS Name,  pg_catalog.pg_get_userbyid(datdba) AS Owner, pg_catalog.pg_database_size(datname) AS size
//FROM pg_catalog.pg_database
//WHERE datname LIKE 'd%' or datname = 'rdpg'
//ORDER BY 3 DESC;`
//	rowCount, err := getRowCount(sq)
//	if err != nil {
//		log.Error(fmt.Sprintf("admin.getNumberOfDatabaseBackupOnDisk() Could not get row count running query %s ! %s", sq, err))
//		return -1
//	}
//	return
//}
