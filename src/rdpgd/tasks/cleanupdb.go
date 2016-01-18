package tasks

import (
	"fmt"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/starkandwayne/rdpgd/globals"
	"github.com/starkandwayne/rdpgd/instances"
	"github.com/starkandwayne/rdpgd/log"
	"github.com/starkandwayne/rdpgd/pg"
)

// CleanupUnusedDatabases - Identify Databases that should be decommissioned and decommission them.
func (t *Task) CleanupUnusedDatabases() (err error) {
	// eg. Look for databases that that should have been decommissioned and insert
	// a CleanupUnusedDatabases task to target each database found.

	log.Trace(fmt.Sprintf("tasks.CleanupUnusedDatabases(%s)...", t.Data))

	//SELECT - Identify the databases which should have been dropped.
	address := `127.0.0.1`
	sq := `SELECT dbname FROM cfsb.instances WHERE effective_at IS NOT NULL AND ineffective_at IS NOT NULL AND dbname IN (SELECT datname FROM pg_database WHERE datname LIKE 'd%');`
	log.Trace(fmt.Sprintf("tasks.CleanupUnusedDatabases() > getList(%s, %s", address, sq))
	listCleanupDatabases, err := getList(address, sq)
	log.Trace(fmt.Sprintf("tasks.CleanupUnusedDatabases() - listCleanupDatabases=%s and err=%s", listCleanupDatabases, err))
	if err != nil {
		log.Error(fmt.Sprintf("tasks.Task<%d>#CleanupUnusedDatabases() Failed to load list of databases ! %s", t.ID, err))
		return err
	}

	for _, databaseName := range listCleanupDatabases {
		log.Trace(fmt.Sprintf("tasks.CleanupUnusedDatabases() > Database Name to Cleanup: %s", databaseName))

		err := CleanupDatabase(databaseName, t.ClusterService)
		if err != nil {
			log.Error(fmt.Sprintf("tasks.CleanUpUnusedDatabases() > tasks.LoopThruDBs(): %s", err))
			return err
		}
	}
	return
}

// CleanupDatabase - Decommission the database sent as a paramenter.
func CleanupDatabase(dbname string, clusterService string) (err error) {
	i, err := instances.FindByDatabase(dbname)
	if err != nil {
		log.Error(fmt.Sprintf("tasks.CleanupUnusedDatabases(%s) instances.FindByDatabase() ! %s", i.Database, err))
		return err
	}

	ips, err := i.ClusterIPs()
	if err != nil {
		log.Error(fmt.Sprintf(`tasks.Task#DecommissionDatabase(%s) i.ClusterIPs() ! %s`, i.Database, err))
		return err
	}
	if len(ips) == 0 {
		return fmt.Errorf("tasks.Task#DecommissionDatabase(%s) ! No service cluster nodes found in Consul", i.Database)
	}
	p := pg.NewPG(`127.0.0.1`, pbPort, `rdpg`, `rdpg`, pgPass)
	db, err := p.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("tasks.Task#DecommissionDatabase(%s) p.Connect(%s) ! %s", dbname, p.URI, err))
		return err
	}
	defer db.Close()

	if globals.ServiceRole == "service" {
		// In here we must do everything necessary to physically delete and clean up
		// the database from all service cluster nodes.
		for _, ip := range ips { // Schedule pgbouncer reconfigure on each cluster node.
			newTask := Task{ClusterID: ClusterID, Node: ip, Role: "all", Action: "Reconfigure", Data: "pgbouncer", NodeType: "any"}
			err = newTask.Enqueue()
			if err != nil {
				log.Error(fmt.Sprintf(`tasks.Task#CleanupUnusedDatabases(%s) Reconfigure PGBouncer! %s`, i.Database, err))
			}
		}
		log.Trace(fmt.Sprintf(`tasks.CleanupUnusedDatabases(%s) - Here is where we finally decommission on the service cluster...`, i.Database))

		sq := fmt.Sprintf(`DELETE FROM tasks.tasks WHERE action='BackupDatabase' AND data='%s'`, i.Database)
		_, err = db.Exec(sq)
		if err != nil {
			log.Error(fmt.Sprintf("tasks.Task#CleanupUnusedDatabases(%s) ! %s", i.Database, err))
		}
		sq = fmt.Sprintf(`UPDATE tasks.schedules SET enabled = false WHERE action='BackupDatabase' AND data='%s'`, i.Database)
		_, err = db.Exec(sq)
		if err != nil {
			log.Error(fmt.Sprintf("tasks.Task#CleanupUnusedDatabases(%s) ! %s", i.Database, err))
		}

		if clusterService == "pgbdr" {
			log.Error(fmt.Sprintf("tasks.Task#CleanupUnusedDatabases(%s) ! Cannot cleanup BDR Servers", i.Database))
		} else {
			p.DisableDatabase(i.Database)
			p.DropDatabase(i.Database)

			dbuser := ""
			sq = fmt.Sprintf(`SELECT dbuser FROM cfsb.instances WHERE dbname='%s' LIMIT 1`, i.Database)
			err = db.Get(&dbuser, sq)
			if err != nil {
				log.Error(fmt.Sprintf("tasks.Task#CleanupUnusedDatabases(%s) ! %s", i.Database, err))
			}
			p.DropUser(dbuser)

			sq = fmt.Sprintf(`UPDATE cfsb.instances SET decommissioned_at=CURRENT_TIMESTAMP WHERE dbname='%s'`, i.Database)
			_, err = db.Exec(sq)
			if err != nil {
				log.Error(fmt.Sprintf("tasks.Task#CleanupUnusedDatabases(%s) ! %s", i.Database, err))
			}
		}

		// Notify management cluster that the instance has been decommissioned
		// Find management cluster API address

		client, err := consulapi.NewClient(consulapi.DefaultConfig())
		if err != nil {
			log.Error(fmt.Sprintf("tasks.Task#CleanupUnusedDatabases(%s) consulapi.NewClient() ! %s", i.Database, err))
			return err
		}

		catalog := client.Catalog()
		svcs, _, err := catalog.Service(`rdpgmc`, "", nil)
		if err != nil {
			log.Error(fmt.Sprintf("tasks.Task#CleanupUnusedDatabases(%s) consulapi.Client.Catalog() ! %s", i.Database, err))
			return err
		}
		if len(svcs) == 0 {
			log.Error(fmt.Sprintf("tasks.Task#CleanupUnusedDatabases(%s) ! No services found, no known nodes?!", i.Database))
			return err
		}
		//mgtAPIIPAddress := svcs[0].Address
	}
	return nil
}
