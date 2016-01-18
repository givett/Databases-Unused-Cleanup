package migrations_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/starkandwayne/rdpgd/pg"

	. "github.com/starkandwayne/rdpg-acceptance-tests/rdpg-service/helper-functions"
)

type Schedule struct {
	ID             int64  `db:"id" json:"id"`
	ClusterID      string `db:"cluster_id" json:"cluster_id"`
	ClusterService string `db:"cluster_service" json:"cluster_service"`
	Role           string `db:"role" json:"role"`
	Action         string `db:"action" json:"action"`
	Data           string `db:"data" json:"data"`
	TTL            int64  `db:"ttl" json:"ttl"`
	NodeType       string `db:"node_type" json:"node_type"`
	Frequency      string `db:"frequency" json:"frequency"`
	Enabled        bool   `db:"enabled" json:"enabled"`
}

//Add - Insert a new schedule into tasks.schedules
func (s *Schedule) Add(address string) (err error) {
	p := pg.NewPG(address, "7432", `rdpg`, `rdpg`, "admin")
	p.Set(`database`, `rdpg`)

	scheduleDB, err := p.Connect()
	if err != nil {
		fmt.Printf(`tasks.Schedule.Add() Could not open connection ! %s`, err)
	}

	defer scheduleDB.Close()

	sq := fmt.Sprintf(`INSERT INTO tasks.schedules (cluster_id,role,action,data,frequency,enabled,node_type,cluster_service) SELECT '%s','%s','%s','%s','%s'::interval, %t, '%s', '%s' WHERE NOT EXISTS (SELECT id FROM tasks.schedules WHERE action = '%s' AND node_type = '%s' AND data = '%s') `, s.ClusterID, s.Role, s.Action, s.Data, s.Frequency, s.Enabled, s.NodeType, s.ClusterService, s.Action, s.NodeType, s.Data)

	_, err = scheduleDB.Exec(sq)
	if err != nil {
		fmt.Printf(`tasks.Schedule.Add():  %s`, err)
	}
	return
}

var _ = Describe("RDPG Database Migrations...", func() {

	It("Check cluster_service column in cfsb.plans table exists, otherwise create", func() {

		allNodes := GetAllNodes()

		tableSchema := `tasks`
		tableName := `schedules`

		//If something wiped the tasks.tasks table, rebuild the rows that should be in the table
		for i := 0; i < len(allNodes); i++ {
			address := allNodes[i].Address
			sq := fmt.Sprintf(` SELECT count(*) as rowCount FROM %s.%s `, tableSchema, tableName)
			rowCount, err := GetRowCount(address, sq)
			clusterService := GetClusterServiceType(allNodes[i].ServiceName)
			ClusterID := allNodes[i].ServiceName

			if rowCount == 0 {
				schedules := []Schedule{}
				if clusterService == "pgbdr" {
					schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `all`, Action: `Vacuum`, Data: `tasks.tasks`, NodeType: `read`, Frequency: `5 minutes`, Enabled: true})
					schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `all`, Action: `Vacuum`, Data: `tasks.tasks`, NodeType: `write`, Frequency: `5 minutes`, Enabled: true})
					schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `all`, Action: `DeleteBackupHistory`, Data: ``, NodeType: `read`, Frequency: `1 hour`, Enabled: true})
					schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `all`, Action: `BackupDatabase`, Data: `rdpg`, NodeType: `read`, Frequency: `1 hour`, Enabled: true})
					schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `all`, Action: `EnforceFileRetention`, Data: ``, NodeType: `read`, Frequency: `1 hour`, Enabled: true})
					schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `all`, Action: `EnforceFileRetention`, Data: ``, NodeType: `write`, Frequency: `1 hour`, Enabled: true})
					schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `all`, Action: `FindFilesToCopyToS3`, Data: ``, NodeType: `read`, Frequency: `5 minutes`, Enabled: false})
					schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `all`, Action: `FindFilesToCopyToS3`, Data: ``, NodeType: `write`, Frequency: `5 minutes`, Enabled: false})

					if allNodes[i].ServiceName == "rdpgmc" {
						schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `manager`, Action: `ReconcileAvailableDatabases`, Data: ``, NodeType: `read`, Frequency: `1 minute`, Enabled: true})
						schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `manager`, Action: `ReconcileAllDatabases`, Data: ``, NodeType: `read`, Frequency: `5 minutes`, Enabled: true})
					} else {
						schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `service`, Action: `ScheduleNewDatabaseBackups`, Data: ``, NodeType: `write`, Frequency: `1 minute`, Enabled: true})
						schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `service`, Action: `PrecreateDatabases`, Data: ``, NodeType: `write`, Frequency: `1 minute`, Enabled: true})
						schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `service`, Action: `DecommissionDatabases`, Data: ``, NodeType: `write`, Frequency: `15 minutes`, Enabled: true})
					}

				} else { // Currently else is specifically postgresql only... we'll have to move this to a switch statement later :)
					schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `all`, Action: `Vacuum`, Data: `tasks.tasks`, NodeType: `write`, Frequency: `5 minutes`, Enabled: true})
					schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `all`, Action: `DeleteBackupHistory`, Data: ``, NodeType: `write`, Frequency: `1 hour`, Enabled: true})
					schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `all`, Action: `BackupDatabase`, Data: `rdpg`, NodeType: `write`, Frequency: `1 hour`, Enabled: true})
					schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `all`, Action: `EnforceFileRetention`, Data: ``, NodeType: `write`, Frequency: `1 hour`, Enabled: true})
					schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `all`, Action: `FindFilesToCopyToS3`, Data: ``, NodeType: `write`, Frequency: `5 minutes`, Enabled: false})

					if allNodes[i].ServiceName == "rdpgmc" {
						schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `manager`, Action: `ReconcileAvailableDatabases`, Data: ``, NodeType: `write`, Frequency: `1 minute`, Enabled: true})
						schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `manager`, Action: `ReconcileAllDatabases`, Data: ``, NodeType: `write`, Frequency: `5 minutes`, Enabled: true})
					} else {
						schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `service`, Action: `ScheduleNewDatabaseBackups`, Data: ``, NodeType: `write`, Frequency: `1 minute`, Enabled: true})
						schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `service`, Action: `PrecreateDatabases`, Data: ``, NodeType: `write`, Frequency: `1 minute`, Enabled: true})
						schedules = append(schedules, Schedule{ClusterID: ClusterID, ClusterService: clusterService, Role: `service`, Action: `DecommissionDatabases`, Data: ``, NodeType: `write`, Frequency: `15 minutes`, Enabled: true})
					}
				}

				for index := range schedules {
					err = schedules[index].Add(address)
					if err != nil {
						Expect(err).NotTo(HaveOccurred())
						continue
					}
				}
				fmt.Printf("%s: Populated table %s.%s...\n", allNodes[i].Node, tableSchema, tableName)

			}

			/*

																																PSQL="/var/vcap/packages/pgbdr/bin/psql -p7432 -U vcap"; $PSQL rdpg

																								select service_id, plan_id, created_at from cfsb.instances where service_id is not null order by service_id;
																																service_id              |               plan_id                |         created_at
																									--------------------------------------+--------------------------------------+----------------------------
																									 3a32e505-6320-4001-8b77-8d196d5f8487 | b9d46b6f-8586-467a-84c2-2251321a1c60 | 2015-08-12 20:23:40.636391



																								----R5

																ALTER TABLE cfsb.plans ADD COLUMN cluster_service TEXT;
																ALTER TABLE cfsb.instances ADD COLUMN cluster_service TEXT;
																ALTER TABLE tasks.tasks ADD COLUMN cluster_service TEXT;
																ALTER TABLE tasks.schedules ADD COLUMN cluster_service TEXT;
																DROP TABLE cfsb.services CASCADE;
																DROP TABLE cfsb.plans CASCADE;
																CREATE TABLE cfsb.services (
																id               BIGSERIAL PRIMARY KEY NOT NULL,
																service_id       TEXT UNIQUE NOT NULL DEFAULT gen_random_uuid(),
																name             TEXT NOT NULL,
																description      TEXT NOT NULL,
																bindable         BOOLEAN NOT NULL DEFAULT true,
																dashboard_client json DEFAULT '{}'::json,
																created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
																effective_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
																ineffective_at   TIMESTAMP
																);
																CREATE TABLE cfsb.plans (
																id              BIGSERIAL    PRIMARY KEY NOT NULL,
																service_id      TEXT NOT NULL,
																plan_id         TEXT DEFAULT gen_random_uuid(),
																cluster_service TEXT NOT NULL,
																name            TEXT,
																description     TEXT,
																free            BOOLEAN   DEFAULT true,
																created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
																effective_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
																ineffective_at  TIMESTAMP
																);
																INSERT INTO cfsb.services (service_id, name,description,bindable,dashboard_client)
																VALUES
																('3a32e505-6320-4001-8b77-8d196d5f8487','rdpg', 'Reliable PostgrSQL Service', true, '{}') ;
																INSERT INTO cfsb.plans (service_id,name,description,free,cluster_service)
																VALUES
																('3a32e505-6320-4001-8b77-8d196d5f8487','shared-nr',  'A PostgreSQL database with no replication on a shared server.', true, 'postgresql');
																INSERT INTO cfsb.plans (plan_id,service_id,name,description,free,cluster_service)
																VALUES
																('b9d46b6f-8586-467a-84c2-2251321a1c60','3a32e505-6320-4001-8b77-8d196d5f8487','shared','A Reliable PostgreSQL database on a shared server.', true, 'pgbdr');
																TRUNCATE TABLE tasks.tasks;
																UPDATE cfsb.instances SET cluster_service = 'pgbdr';
																TRUNCATE TABLE tasks.schedules;



																								---R2
																								rdpg=# select service_id, plan_id from cfsb.instances where service_id is not null order by service_id;
																								              service_id              |               plan_id                |
																								--------------------------------------+--------------------------------------+
																								 738508bd-1f4e-4183-b93b-ea7779b15bde | 946ae934-d8ed-4220-a79c-d89b12f7fa3d |

				TRUNCATE TABLE tasks.tasks;
				TRUNCATE TABLE tasks.schedules;
				ALTER TABLE cfsb.plans ADD COLUMN cluster_service TEXT;
				ALTER TABLE cfsb.instances ADD COLUMN cluster_service TEXT;
				ALTER TABLE tasks.tasks ADD COLUMN cluster_service TEXT;
				ALTER TABLE tasks.schedules ADD COLUMN cluster_service TEXT;
				DROP TABLE cfsb.services CASCADE;
				DROP TABLE cfsb.plans CASCADE;
				CREATE TABLE cfsb.services (
				id               BIGSERIAL PRIMARY KEY NOT NULL,
				service_id       TEXT UNIQUE NOT NULL DEFAULT gen_random_uuid(),
				name             TEXT NOT NULL,
				description      TEXT NOT NULL,
				bindable         BOOLEAN NOT NULL DEFAULT true,
				dashboard_client json DEFAULT '{}'::json,
				created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				effective_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				ineffective_at   TIMESTAMP
				);
				CREATE TABLE cfsb.plans (
				id              BIGSERIAL    PRIMARY KEY NOT NULL,
				service_id      TEXT NOT NULL,
				plan_id         TEXT DEFAULT gen_random_uuid(),
				cluster_service TEXT NOT NULL,
				name            TEXT,
				description     TEXT,
				free            BOOLEAN   DEFAULT true,
				created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				effective_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				ineffective_at  TIMESTAMP
				);
				INSERT INTO cfsb.services (service_id, name,description,bindable,dashboard_client)
				VALUES
				('738508bd-1f4e-4183-b93b-ea7779b15bde','postgres', 'Reliable PostgrSQL Service', true, '{}') ;
				INSERT INTO cfsb.plans (service_id,name,description,free,cluster_service)
				VALUES
				('738508bd-1f4e-4183-b93b-ea7779b15bde','shared-nr',  'A PostgreSQL database with no replication on a shared server.', true, 'postgresql');
				INSERT INTO cfsb.plans (plan_id, service_id,name,description,free,cluster_service)
				VALUES
				('946ae934-d8ed-4220-a79c-d89b12f7fa3d','738508bd-1f4e-4183-b93b-ea7779b15bde','shared',     'A Reliable PostgreSQL database on a shared server.', true, 'pgbdr');
				UPDATE cfsb.instances SET cluster_service = 'pgbdr';11100

				PSQL="/var/vcap/packages/pgbdr/bin/psql -p7432 -U vcap"; $PSQL rdpg


								--ThunderChicken
								service_id              |               plan_id
								--------------------------------------+--------------------------------------
								71fba9a4-11c5-4505-9297-439aa8372bfe | c663e924-d517-4eb2-b812-bd9272981fa4

								ALTER TABLE cfsb.plans ADD COLUMN cluster_service TEXT;
								ALTER TABLE cfsb.instances ADD COLUMN cluster_service TEXT;
								ALTER TABLE tasks.tasks ADD COLUMN cluster_service TEXT;
								ALTER TABLE tasks.schedules ADD COLUMN cluster_service TEXT;
								DROP TABLE cfsb.services CASCADE;
								DROP TABLE cfsb.plans CASCADE;
								CREATE TABLE cfsb.services (
								id               BIGSERIAL PRIMARY KEY NOT NULL,
								service_id       TEXT UNIQUE NOT NULL DEFAULT gen_random_uuid(),
								name             TEXT NOT NULL,
								description      TEXT NOT NULL,
								bindable         BOOLEAN NOT NULL DEFAULT true,
								dashboard_client json DEFAULT '{}'::json,
								created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
								effective_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
								ineffective_at   TIMESTAMP
								);
								CREATE TABLE cfsb.plans (
								id              BIGSERIAL    PRIMARY KEY NOT NULL,
								service_id      TEXT NOT NULL,
								plan_id         TEXT DEFAULT gen_random_uuid(),
								cluster_service TEXT NOT NULL,
								name            TEXT,
								description     TEXT,
								free            BOOLEAN   DEFAULT true,
								created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
								effective_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
								ineffective_at  TIMESTAMP
								);
								INSERT INTO cfsb.services (service_id, name,description,bindable,dashboard_client)
								VALUES
								('71fba9a4-11c5-4505-9297-439aa8372bfe','postgres', 'Reliable PostgrSQL Service', true, '{}') ;
								INSERT INTO cfsb.plans (service_id,name,description,free,cluster_service)
								VALUES
								('71fba9a4-11c5-4505-9297-439aa8372bfe','shared-nr',  'A PostgreSQL database with no replication on a shared server.', true, 'postgresql');
								INSERT INTO cfsb.plans (plan_id, service_id,name,description,free,cluster_service)
								VALUES
								('c663e924-d517-4eb2-b812-bd9272981fa4','71fba9a4-11c5-4505-9297-439aa8372bfe','shared',     'A Reliable PostgreSQL database on a shared server.', true, 'pgbdr');
								TRUNCATE TABLE tasks.tasks;
								UPDATE cfsb.instances SET cluster_service = 'pgbdr';
								TRUNCATE TABLE tasks.schedules;



																								--- Generic
																												ALTER TABLE cfsb.plans ADD COLUMN cluster_service TEXT;
																												ALTER TABLE cfsb.instances ADD COLUMN cluster_service TEXT;
																												ALTER TABLE tasks.tasks ADD COLUMN cluster_service TEXT;
																												ALTER TABLE tasks.schedules ADD COLUMN cluster_service TEXT;
																												DROP TABLE cfsb.services CASCADE;
																												DROP TABLE cfsb.plans CASCADE;
																												SELECT * FROM cfsb.services;
																												SELECT * FROM cfsb.plans;
																												CREATE TABLE cfsb.services (
																												id               BIGSERIAL PRIMARY KEY NOT NULL,
																												service_id       TEXT UNIQUE NOT NULL DEFAULT gen_random_uuid(),
																												name             TEXT NOT NULL,
																												description      TEXT NOT NULL,
																												bindable         BOOLEAN NOT NULL DEFAULT true,
																												dashboard_client json DEFAULT '{}'::json,
																												created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
																												effective_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
																												ineffective_at   TIMESTAMP
																												);
																												CREATE TABLE cfsb.plans (
																												id              BIGSERIAL    PRIMARY KEY NOT NULL,
																												service_id      TEXT NOT NULL,
																												plan_id         TEXT DEFAULT gen_random_uuid(),
																												cluster_service TEXT NOT NULL,
																												name            TEXT,
																												description     TEXT,
																												free            BOOLEAN   DEFAULT true,
																												created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
																												effective_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
																												ineffective_at  TIMESTAMP
																												);
																												INSERT INTO cfsb.services (name,description,bindable,dashboard_client)
																												VALUES
																												('postgres', 'Reliable PostgrSQL Service', true, '{}') ;
																												INSERT INTO cfsb.plans (service_id,name,description,free,cluster_service)
																												VALUES
																												((SELECT service_id FROM cfsb.services WHERE name='postgres' LIMIT 1),'shared-nr',  'A PostgreSQL database with no replication on a shared server.', true, 'postgresql'),
																												((SELECT service_id FROM cfsb.services WHERE name='postgres' LIMIT 1),'shared-sr',  'A Streaming Replicated PostgreSQL database on a shared server.', true, 'pgbdr'),
																												((SELECT service_id FROM cfsb.services WHERE name='postgres' LIMIT 1),'shared-bdr', 'A BDR Replicated PostgreSQL database on a shared server.', true, 'pgbdr'),
																												((SELECT service_id FROM cfsb.services WHERE name='postgres' LIMIT 1),'shared',     'A Reliable PostgreSQL database on a shared server.', true, 'pgbdr');
																												TRUNCATE TABLE tasks.tasks;
																												UPDATE cfsb.instances SET cluster_service = 'pgbdr';
																												TRUNCATE TABLE tasks.schedules;


			*/

			/*
							if columnCount == 0 {
								sq = fmt.Sprintf(`ALTER TABLE %s.%s ADD COLUMN %s TEXT ;`, tableSchema, tableName, columnName)
								err = execQuery(address, sq)
								fmt.Printf("%s: Had to create '%s' column in %s.%s...\n", allNodes[i].Node, columnName, tableSchema, tableName)
								Expect(err).NotTo(HaveOccurred())

								tableSchema = `cfsb`
								tableName = `instances`
								columnName = `cluster_service`
								sq = fmt.Sprintf(`ALTER TABLE %s.%s ADD COLUMN %s TEXT ;`, tableSchema, tableName, columnName)
								err = execQuery(address, sq)
								fmt.Printf("%s: Had to create '%s' column in %s.%s...\n", allNodes[i].Node, columnName, tableSchema, tableName)
								Expect(err).NotTo(HaveOccurred())

								tableSchema = `tasks`
								tableName = `tasks`
								columnName = `cluster_service`
								sq = fmt.Sprintf(`ALTER TABLE %s.%s ADD COLUMN %s TEXT;`, tableSchema, tableName, columnName)
								err = execQuery(address, sq)
								fmt.Printf("%s: Had to create '%s' column in %s.%s...\n", allNodes[i].Node, columnName, tableSchema, tableName)
								Expect(err).NotTo(HaveOccurred())

								tableSchema = `tasks`
								tableName = `schedules`
								columnName = `cluster_service`
								sq = fmt.Sprintf(`ALTER TABLE %s.%s ADD COLUMN %s TEXT `, tableSchema, tableName, columnName)
								err = execQuery(address, sq)
								fmt.Printf("%s: Had to create '%s' column in %s.%s...\n", allNodes[i].Node, columnName, tableSchema, tableName)
								Expect(err).NotTo(HaveOccurred())

								//Reload plans and services tables

									tableSchema = `cfsb`
									tableName = `plans`
									sq = fmt.Sprintf(`TRUNCATE TABLE %s.%s CASCADE;`, tableSchema, tableName)
									err = execQuery(address, sq)
									fmt.Printf("%s: Truncated table %s.%s...\n", allNodes[i].Node, tableSchema, tableName)
									Expect(err).NotTo(HaveOccurred())

									tableSchema = `cfsb`
									tableName = `services`
									sq = fmt.Sprintf(`TRUNCATE TABLE %s.%s CASCADE;`, tableSchema, tableName)
									err = execQuery(address, sq)
									fmt.Printf("%s: Truncated table %s.%s...\n", allNodes[i].Node, tableSchema, tableName)
									Expect(err).NotTo(HaveOccurred())


										tableSchema = `cfsb`
										tableName = `plans`
										columnName = `cluster_service`
										sq = fmt.Sprintf(`ALTER TABLE %s.%s ALTER COLUMN %s SET TYPE TEXT;`, tableSchema, tableName, columnName)
										err = execQuery(address, sq)
										fmt.Printf("%s: Had to create '%s' column in %s.%s...\n", allNodes[i].Node, columnName, tableSchema, tableName)
										Expect(err).NotTo(HaveOccurred())

								tableSchema = `cfsb`
								tableName = `services`
								sq = fmt.Sprintf(`DROP TABLE %s.%s CASCADE;`, tableSchema, tableName)
								err = execQuery(address, sq)
								fmt.Printf("%s: Drop table %s.%s...\n", allNodes[i].Node, tableSchema, tableName)
								Expect(err).NotTo(HaveOccurred())

								tableSchema = `cfsb`
								tableName = `plans`
								sq = fmt.Sprintf(`DROP TABLE %s.%s CASCADE;`, tableSchema, tableName)
								err = execQuery(address, sq)
								fmt.Printf("%s: Drop table %s.%s...\n", allNodes[i].Node, tableSchema, tableName)
								Expect(err).NotTo(HaveOccurred())

								sq = `CREATE TABLE cfsb.services (
				id               BIGSERIAL PRIMARY KEY NOT NULL,
				service_id       TEXT UNIQUE NOT NULL DEFAULT gen_random_uuid(),
				name             TEXT NOT NULL,
				description      TEXT NOT NULL,
				bindable         BOOLEAN NOT NULL DEFAULT true,
				dashboard_client json DEFAULT '{}'::json,
				created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				effective_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				ineffective_at   TIMESTAMP
				);`
								err = execQuery(address, sq)
								fmt.Printf("%s: Create table %s.%s...\n", allNodes[i].Node, tableSchema, tableName)
								Expect(err).NotTo(HaveOccurred())

								sq = `CREATE TABLE cfsb.plans (
				id              BIGSERIAL    PRIMARY KEY NOT NULL,
				service_id      TEXT NOT NULL REFERENCES cfsb.services(service_id),
				plan_id         TEXT DEFAULT gen_random_uuid(),
				cluster_service TEXT NOT NULL,
				name            TEXT,
				description     TEXT,
				free            BOOLEAN   DEFAULT true,
				created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				effective_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				ineffective_at  TIMESTAMP
				);`
								err = execQuery(address, sq)
								fmt.Printf("%s: Create table %s.%s...\n", allNodes[i].Node, tableSchema, tableName)
								Expect(err).NotTo(HaveOccurred())

								sq = `INSERT INTO cfsb.services (name,description,bindable,dashboard_client)
				VALUES
				('postgres',       'Reliable PostgrSQL Service', true, '{}'),
				('rdpg',           'Reliable PostgrSQL Service', true, '{}'),
				('postgresql-bdr', 'HA PostgreSQL 9.4 Service',  true, '{}'),
				('postgresql-9.4', 'PostgreSQL 9.4 Service',     true, '{}') ;
								`
								err = execQuery(address, sq)
								fmt.Printf("%s: Populated table %s.%s...\n", allNodes[i].Node, tableSchema, tableName)
								Expect(err).NotTo(HaveOccurred())

								sq = `INSERT INTO cfsb.plans (service_id,name,description,free,cluster_service)
				VALUES
				((SELECT service_id FROM cfsb.services WHERE name='postgresql-bdr' LIMIT 1),'shared', 'A database on a shared server.',                     true, 'pgbdr'),
				((SELECT service_id FROM cfsb.services WHERE name='postgresql-9.4' LIMIT 1),'shared', 'A database on a shared server.',                     true, 'postgresql'),
				((SELECT service_id FROM cfsb.services WHERE name='postgres'       LIMIT 1),'shared', 'A Reliable PostgreSQL database on a shared server.', true, 'pgbdr'),
				((SELECT service_id FROM cfsb.services WHERE name='rdpg'           LIMIT 1),'shared', 'A Reliable PostgreSQL database on a shared server.', true, 'pgbdr');`
								err = execQuery(address, sq)
								fmt.Printf("%s: Populated table %s.%s...\n", allNodes[i].Node, tableSchema, tableName)
								Expect(err).NotTo(HaveOccurred())


									tableSchema = `tasks`
									tableName = `schedules`
									columnName = `cluster_service`
									sq = fmt.Sprintf(`UPDATE %s.%s SET %s = '%s';`, tableSchema, tableName, columnName, clusterService)
									err = execQuery(address, sq)
									fmt.Printf("%s: Updated '%s' column in %s.%s to be '%s'...\n", allNodes[i].Node, columnName, tableSchema, tableName, clusterService)
									Expect(err).NotTo(HaveOccurred())

									tableSchema = `tasks`
									tableName = `tasks`
									columnName = `cluster_service`
									sq = fmt.Sprintf(`UPDATE %s.%s SET %s = '%s';`, tableSchema, tableName, columnName, clusterService)
									err = execQuery(address, sq)
									fmt.Printf("%s: Updated '%s' column in %s.%s to be '%s'...\n", allNodes[i].Node, columnName, tableSchema, tableName, clusterService)
									Expect(err).NotTo(HaveOccurred())


								tableSchema = `tasks`
								tableName = `tasks`
								sq = fmt.Sprintf(`TRUNCATE TABLE %s.%s CASCADE;`, tableSchema, tableName)
								err = execQuery(address, sq)
								fmt.Printf("%s: Truncated table %s.%s...\n", allNodes[i].Node, tableSchema, tableName)
								Expect(err).NotTo(HaveOccurred())

								tableSchema = `cfsb`
								tableName = `instances`
								columnName = `cluster_service`
								sq = fmt.Sprintf(`UPDATE %s.%s SET %s = '%s' WHERE cluster_id = '%s';`, tableSchema, tableName, columnName, clusterService, ClusterID)
								err = execQuery(address, sq)
								fmt.Printf("%s: Updated '%s' column in %s.%s to be '%s'...\n", allNodes[i].Node, columnName, tableSchema, tableName, clusterService)
								Expect(err).NotTo(HaveOccurred())

							}
			*/
			/*
				//Reload tasks.schedules table
				tableSchema = `tasks`
				tableName = `schedules`
				sq = fmt.Sprintf(`TRUNCATE TABLE %s.%s;`, tableSchema, tableName)
				err := execQuery(address, sq)
				fmt.Printf("%s: Truncated table %s.%s...\n", allNodes[i].Node, tableSchema, tableName)
				Expect(err).NotTo(HaveOccurred())
			*/

		}

	})

	/*	It("Check backups.file_history table exists, otherwise create", func() {

				allNodes := GetAllNodes()

				//Check all nodes
				var nodeRowCount []int
				for i := 0; i < len(allNodes); i++ {
					address := allNodes[i].Address
					sq := ` SELECT count(table_name) as rowCount FROM information_schema.tables WHERE table_schema = 'backups' and table_name IN ('file_history'); `
					tableCount, err := GetRowCount(address, sq)

					if tableCount == 0 {
						//Table doesn't exist, create it
						sq = `CREATE TABLE IF NOT EXISTS backups.file_history (
						  id               BIGSERIAL PRIMARY KEY NOT NULL,
							cluster_id        TEXT NOT NULL,
						  dbname            TEXT NOT NULL,
							node							TEXT NOT NULL,
							file_name					TEXT NOT NULL,
							action						TEXT NOT NULL,
							status						TEXT NOT NULL,
							params            json DEFAULT '{}'::json,
							created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
							duration          INT,
							removed_at        TIMESTAMP
						);`
						err = execQuery(address, sq)
						fmt.Printf("%s: Had to create backups.file_history table...\n", allNodes[i].Node)
						Expect(err).NotTo(HaveOccurred())
					}

					//Now rerun and verify the table was created
					sq = ` SELECT count(table_name) as rowCount FROM information_schema.tables WHERE table_schema = 'backups' and table_name IN ('file_history'); `
					rowCount, err := GetRowCount(address, sq)
					nodeRowCount = append(nodeRowCount, rowCount)
					fmt.Printf("%s: Found %d tables in schema 'backups'...\n", allNodes[i].Node, rowCount)
					Expect(err).NotTo(HaveOccurred())
				}

				//Verify each database also sees the same number of records (bdr sanity check)
				for i := 1; i < len(nodeRowCount); i++ {
					Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
				}

				Expect(len(nodeRowCount)).NotTo(Equal(0))
				Expect(nodeRowCount[0]).To(Equal(1))
			})

		It("Check node_type column in tasks.tasks table exists, otherwise create", func() {

			allNodes := GetAllNodes()
			tableSchema := `tasks`
			tableName := `tasks`
			columnName := `node_type`
			defaultValue := `any`

			//Check all nodes
			var nodeRowCount []int
			for i := 0; i < len(allNodes); i++ {
				address := allNodes[i].Address
				sq := fmt.Sprintf(` SELECT count(table_name) as rowCount FROM information_schema.columns WHERE table_schema = '%s' AND table_name = '%s' AND column_name = '%s' `, tableSchema, tableName, columnName)
				columnCount, err := GetRowCount(address, sq)

				if columnCount == 0 {
					//Table doesn't exist, create it

					sq := fmt.Sprintf(`ALTER TABLE %s.%s ADD COLUMN %s text;`, tableSchema, tableName, columnName)
					err = execQuery(address, sq)
					fmt.Printf("%s: Had to create '%s' column in %s.%s...\n", allNodes[i].Node, columnName, tableSchema, tableName)
					Expect(err).NotTo(HaveOccurred())

					sq = fmt.Sprintf(`ALTER TABLE %s.%s ALTER COLUMN %s SET DEFAULT '%s';`, tableSchema, tableName, columnName, defaultValue)
					err = execQuery(address, sq)
					fmt.Printf("%s: Had to create '%s' column in %s.%s, setting default value to '%s'...\n", allNodes[i].Node, columnName, tableSchema, tableName, defaultValue)
					Expect(err).NotTo(HaveOccurred())

				}
				//Now rerun and verify the column was created
				sq = fmt.Sprintf(` SELECT count(table_name) as rowCount FROM information_schema.columns WHERE table_schema = '%s' AND table_name = '%s' AND column_name = '%s' `, tableSchema, tableName, columnName)
				rowCount, err := GetRowCount(address, sq)
				nodeRowCount = append(nodeRowCount, rowCount)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("%s: Found %d '%s' columns in table '%s.%s'...\n", allNodes[i].Node, rowCount, columnName, tableSchema, tableName)
			}

			//Verify each database also sees the same number of records (bdr sanity check)
			for i := 1; i < len(nodeRowCount); i++ {
				Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
			}

			Expect(len(nodeRowCount)).NotTo(Equal(0))
			Expect(nodeRowCount[0]).To(Equal(1))
		})

		It("Check node_type column in tasks.schedules table exists, otherwise create", func() {

			allNodes := GetAllNodes()
			tableSchema := `tasks`
			tableName := `schedules`
			columnName := `node_type`
			defaultValue := `any`

			//Check all nodes
			var nodeRowCount []int
			for i := 0; i < len(allNodes); i++ {
				address := allNodes[i].Address
				sq := fmt.Sprintf(` SELECT count(table_name) as rowCount FROM information_schema.columns WHERE table_schema = '%s' AND table_name = '%s' AND column_name = '%s' `, tableSchema, tableName, columnName)
				columnCount, err := GetRowCount(address, sq)

				if columnCount == 0 {
					//Table doesn't exist, create it

					sq := fmt.Sprintf(`ALTER TABLE %s.%s ADD COLUMN %s text;`, tableSchema, tableName, columnName)
					err = execQuery(address, sq)
					fmt.Printf("%s: Had to create '%s' column in %s.%s...\n", allNodes[i].Node, columnName, tableSchema, tableName)
					Expect(err).NotTo(HaveOccurred())

					sq = fmt.Sprintf(`ALTER TABLE %s.%s ALTER COLUMN %s SET DEFAULT '%s';`, tableSchema, tableName, columnName, defaultValue)
					err = execQuery(address, sq)
					fmt.Printf("%s: Had to create '%s' column in %s.%s, setting default value to '%s'...\n", allNodes[i].Node, columnName, tableSchema, tableName, defaultValue)
					Expect(err).NotTo(HaveOccurred())

				}
				//Now rerun and verify the column was created
				sq = fmt.Sprintf(` SELECT count(table_name) as rowCount FROM information_schema.columns WHERE table_schema = '%s' AND table_name = '%s' AND column_name = '%s' `, tableSchema, tableName, columnName)
				rowCount, err := GetRowCount(address, sq)
				nodeRowCount = append(nodeRowCount, rowCount)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("%s: Found %d '%s' columns in table '%s.%s'...\n", allNodes[i].Node, rowCount, columnName, tableSchema, tableName)
			}

			//Verify each database also sees the same number of records (bdr sanity check)
			for i := 1; i < len(nodeRowCount); i++ {
				Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
			}

			Expect(len(nodeRowCount)).NotTo(Equal(0))
			Expect(nodeRowCount[0]).To(Equal(1))
		})

		It("Check default for defaultDaysToKeepFileHistory added rdpg.config", func() {

			allNodes := GetAllNodes()
			configKey := `defaultDaysToKeepFileHistory`
			configValue := `180`

			var nodeRowCount []int
			for i := 0; i < len(allNodes); i++ {
				address := allNodes[i].Address

				sq := fmt.Sprintf(`SELECT count(key) as rowCount FROM rdpg.config WHERE key IN ('%s') ;  `, configKey)
				configCount, err := GetRowCount(address, sq)

				if configCount == 0 {
					//Table entry doesn't exist, create it
					sq = fmt.Sprintf(`INSERT INTO rdpg.config (key,cluster_id,value) VALUES ('%s', '%s', '%s')`, configKey, allNodes[i].ServiceName, configValue)
					err = execQuery(address, sq)
					fmt.Printf("%s: Had to insert key %s with value %s into 'rdpg.config'...\n", allNodes[i].Node, configKey, configValue)
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(1 * time.Second)
				}

				sq = fmt.Sprintf(`SELECT count(key) as rowCount FROM rdpg.config WHERE key IN ('%s') ;  `, configKey)
				rowCount, err := GetRowCount(address, sq)
				nodeRowCount = append(nodeRowCount, rowCount)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("%s: Found %d default values for key %s in rdpg.config...\n", allNodes[i].Node, rowCount, configKey)
			}
			//Verify each database also sees the same number of records (bdr sanity check)
			for i := 1; i < len(nodeRowCount); i++ {
				Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
			}
			Expect(len(nodeRowCount)).NotTo(Equal(0))
			Expect(nodeRowCount[0]).To(Equal(1))

		})

		It("Check default for BackupPort added rdpg.config", func() {

			allNodes := GetAllNodes()
			configKey := `BackupPort`
			configValue := `7432`

			var nodeRowCount []int
			for i := 0; i < len(allNodes); i++ {
				address := allNodes[i].Address

				sq := fmt.Sprintf(`SELECT count(key) as rowCount FROM rdpg.config WHERE key IN ('%s') ;  `, configKey)
				configCount, err := GetRowCount(address, sq)

				if configCount == 0 {
					//Table entry doesn't exist, create it
					sq = fmt.Sprintf(`INSERT INTO rdpg.config (key,cluster_id,value) VALUES ('%s', '%s', '%s')`, configKey, allNodes[i].ServiceName, configValue)
					err = execQuery(address, sq)
					fmt.Printf("%s: Had to insert key %s with value %s into 'rdpg.config'...\n", allNodes[i].Node, configKey, configValue)
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(1 * time.Second)
				}

				sq = fmt.Sprintf(`SELECT count(key) as rowCount FROM rdpg.config WHERE key IN ('%s') ;  `, configKey)
				rowCount, err := GetRowCount(address, sq)
				nodeRowCount = append(nodeRowCount, rowCount)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("%s: Found %d default values for key %s in rdpg.config...\n", allNodes[i].Node, rowCount, configKey)
			}
			//Verify each database also sees the same number of records (bdr sanity check)
			for i := 1; i < len(nodeRowCount); i++ {
				Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
			}
			Expect(len(nodeRowCount)).NotTo(Equal(0))
			Expect(nodeRowCount[0]).To(Equal(1))

		})

		It("Check default for BackupsPath added rdpg.config", func() {

			allNodes := GetAllNodes()
			configKey := `BackupsPath`
			configValue := `/var/vcap/store/pgbdr/backups`

			var nodeRowCount []int
			for i := 0; i < len(allNodes); i++ {
				address := allNodes[i].Address

				sq := fmt.Sprintf(`SELECT count(key) as rowCount FROM rdpg.config WHERE key IN ('%s') ;  `, configKey)
				configCount, err := GetRowCount(address, sq)

				if configCount == 0 {
					//Table entry doesn't exist, create it
					sq = fmt.Sprintf(`INSERT INTO rdpg.config (key,cluster_id,value) VALUES ('%s', '%s', '%s')`, configKey, allNodes[i].ServiceName, configValue)
					err = execQuery(address, sq)
					fmt.Printf("%s: Had to insert key %s with value %s into 'rdpg.config'...\n", allNodes[i].Node, configKey, configValue)
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(1 * time.Second)
				}

				sq = fmt.Sprintf(`SELECT count(key) as rowCount FROM rdpg.config WHERE key IN ('%s') ;  `, configKey)
				rowCount, err := GetRowCount(address, sq)
				nodeRowCount = append(nodeRowCount, rowCount)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("%s: Found %d default values for key %s in rdpg.config...\n", allNodes[i].Node, rowCount, configKey)
			}
			//Verify each database also sees the same number of records (bdr sanity check)
			for i := 1; i < len(nodeRowCount); i++ {
				Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
			}
			Expect(len(nodeRowCount)).NotTo(Equal(0))
			Expect(nodeRowCount[0]).To(Equal(1))

		})

		It("Check default for pgDumpBinaryLocation added rdpg.config", func() {

			allNodes := GetAllNodes()
			configKey := `pgDumpBinaryLocation`
			configValue := `/var/vcap/packages/pgbdr/bin/pg_dump`

			var nodeRowCount []int
			for i := 0; i < len(allNodes); i++ {
				address := allNodes[i].Address

				sq := fmt.Sprintf(`SELECT count(key) as rowCount FROM rdpg.config WHERE key IN ('%s') ;  `, configKey)
				configCount, err := GetRowCount(address, sq)

				if configCount == 0 {
					//Table entry doesn't exist, create it
					sq = fmt.Sprintf(`INSERT INTO rdpg.config (key,cluster_id,value) VALUES ('%s', '%s', '%s')`, configKey, allNodes[i].ServiceName, configValue)
					err = execQuery(address, sq)
					fmt.Printf("%s: Had to insert key %s with value %s into 'rdpg.config'...\n", allNodes[i].Node, configKey, configValue)
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(1 * time.Second)
				}

				sq = fmt.Sprintf(`SELECT count(key) as rowCount FROM rdpg.config WHERE key IN ('%s') ;  `, configKey)
				rowCount, err := GetRowCount(address, sq)
				nodeRowCount = append(nodeRowCount, rowCount)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("%s: Found %d default values for key %s in rdpg.config...\n", allNodes[i].Node, rowCount, configKey)
			}
			//Verify each database also sees the same number of records (bdr sanity check)
			for i := 1; i < len(nodeRowCount); i++ {
				Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
			}
			Expect(len(nodeRowCount)).NotTo(Equal(0))
			Expect(nodeRowCount[0]).To(Equal(1))

		})

		It("Check default for pgDumpBinaryLocation added rdpg.config", func() {

			allNodes := GetAllNodes()
			configKey := `pgDumpBinaryLocation`
			configValue := `/var/vcap/packages/pgbdr/bin/pg_dump`

			var nodeRowCount []int
			for i := 0; i < len(allNodes); i++ {
				address := allNodes[i].Address

				sq := fmt.Sprintf(`SELECT count(key) as rowCount FROM rdpg.config WHERE key IN ('%s') ;  `, configKey)
				configCount, err := GetRowCount(address, sq)

				if configCount == 0 {
					//Table entry doesn't exist, create it
					sq = fmt.Sprintf(`INSERT INTO rdpg.config (key,cluster_id,value) VALUES ('%s', '%s', '%s')`, configKey, allNodes[i].ServiceName, configValue)
					err = execQuery(address, sq)
					fmt.Printf("%s: Had to insert key %s with value %s into 'rdpg.config'...\n", allNodes[i].Node, configKey, configValue)
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(1 * time.Second)
				}

				sq = fmt.Sprintf(`SELECT count(key) as rowCount FROM rdpg.config WHERE key IN ('%s') ;  `, configKey)
				rowCount, err := GetRowCount(address, sq)
				nodeRowCount = append(nodeRowCount, rowCount)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("%s: Found %d default values for key %s in rdpg.config...\n", allNodes[i].Node, rowCount, configKey)
			}
			//Verify each database also sees the same number of records (bdr sanity check)
			for i := 1; i < len(nodeRowCount); i++ {
				Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
			}
			Expect(len(nodeRowCount)).NotTo(Equal(0))
			Expect(nodeRowCount[0]).To(Equal(1))

		})

		It("Check DeleteBackupHistory job exists in tasks.schedules", func() {

			allNodes := GetAllNodes()

			var nodeRowCount []int
			for i := 0; i < len(allNodes); i++ {
				address := allNodes[i].Address

				sq := `SELECT count(action) as rowCount FROM tasks.schedules WHERE action = 'DeleteBackupHistory'; `
				taskCount, err := GetRowCount(address, sq)

				fmt.Printf("%s: Found %d taskCount'...\n", allNodes[i].Node, taskCount)

				if taskCount == 0 {
					//Table entry doesn't exist, create it
					sq = fmt.Sprintf(`INSERT INTO tasks.schedules (cluster_id,role,action,data,frequency,enabled,node_type) VALUES ('%s','all','DeleteBackupHistory','','1 hour'::interval, true, 'read')`, allNodes[i].ServiceName)

					err = execQuery(address, sq)
					fmt.Printf("%s: Had to add DeleteBackupHistory into 'task.schedules'...\n", allNodes[i].Node)

					Expect(err).NotTo(HaveOccurred())
					time.Sleep(1 * time.Second)
				}

				sq = `SELECT count(action) as rowCount FROM tasks.schedules WHERE action = 'DeleteBackupHistory'; `
				rowCount, err := GetRowCount(address, sq)
				nodeRowCount = append(nodeRowCount, rowCount)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("%s: Found %d scheduled tasks for DeleteBackupHistory in tasks.schedules...\n", allNodes[i].Node, rowCount)
			}
			//Verify each database also sees the same number of records (bdr sanity check)
			for i := 1; i < len(nodeRowCount); i++ {
				Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
			}
			Expect(len(nodeRowCount)).NotTo(Equal(0))
			Expect(nodeRowCount[0]).To(Equal(1))

		})

		It("Check ScheduleNewDatabaseBackups job exists in tasks.schedules", func() {

			allNodes := GetServiceNodes()

			var nodeRowCount []int
			for i := 0; i < len(allNodes); i++ {
				address := allNodes[i].Address

				sq := `SELECT count(action) as rowCount FROM tasks.schedules WHERE action = 'ScheduleNewDatabaseBackups'; `
				taskCount, err := GetRowCount(address, sq)

				fmt.Printf("%s: Found %d taskCount'...\n", allNodes[i].Node, taskCount)

				if taskCount == 0 {
					//Table entry doesn't exist, create it
					sq = fmt.Sprintf(`INSERT INTO tasks.schedules (cluster_id,role,action,data,frequency,enabled,node_type) VALUES ('%s','service','ScheduleNewDatabaseBackups','','1 minute'::interval, true, 'write')`, allNodes[i].ServiceName)

					err = execQuery(address, sq)
					fmt.Printf("%s: Had to add ScheduleNewDatabaseBackups into 'task.schedules'...\n", allNodes[i].Node)
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(1 * time.Second)
				}

				sq = `SELECT count(action) as rowCount FROM tasks.schedules WHERE action = 'ScheduleNewDatabaseBackups'; `
				rowCount, err := GetRowCount(address, sq)
				nodeRowCount = append(nodeRowCount, rowCount)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("%s: Found %d scheduled tasks for ScheduleNewDatabaseBackups in tasks.schedules...\n", allNodes[i].Node, rowCount)
			}
			//Verify each database also sees the same number of records (bdr sanity check)
			for i := 1; i < len(nodeRowCount); i++ {
				Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
			}
			Expect(len(nodeRowCount)).NotTo(Equal(0))
			Expect(nodeRowCount[0]).To(Equal(1))

		})

		It("Check no null values in node_type for tasks.schedules", func() {

			allNodes := GetAllNodes()

			var nodeRowCount []int
			for i := 0; i < len(allNodes); i++ {
				address := allNodes[i].Address

				sq := `SELECT count(node_type) as rowCount FROM tasks.schedules WHERE node_type IS NULL; `
				taskCount, err := GetRowCount(address, sq)

				if taskCount == 0 {
					//Table entry doesn't exist, create it
					sq = `UPDATE tasks.schedules SET node_type='write' WHERE node_type IS NULL;`

					err = execQuery(address, sq)
					fmt.Printf("%s: Had to update %d task with a default in 'task.schedules'...\n", allNodes[i].Node, taskCount)

					Expect(err).NotTo(HaveOccurred())
				}

				sq = `SELECT count(node_type) as rowCount FROM tasks.schedules WHERE node_type IS NULL; `
				rowCount, err := GetRowCount(address, sq)
				nodeRowCount = append(nodeRowCount, rowCount)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("%s: Found %d rows with null values in node_type column of tasks.schedules...\n", allNodes[i].Node, rowCount)
			}
			//Verify each database also sees the same number of records (bdr sanity check)
			for i := 1; i < len(nodeRowCount); i++ {
				Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
			}
			Expect(len(nodeRowCount)).NotTo(Equal(0))
			Expect(nodeRowCount[0]).To(Equal(0))

		})

		It("Check BackupDatabase job exists in tasks.schedules for the rdpg system database", func() {

			allNodes := GetAllNodes()

			var nodeRowCount []int
			for i := 0; i < len(allNodes); i++ {
				address := allNodes[i].Address

				sq := `SELECT count(action) as rowCount FROM tasks.schedules WHERE action = 'BackupDatabase' AND data = 'rdpg'; `
				taskCount, err := GetRowCount(address, sq)

				fmt.Printf("%s: Found %d taskCount'...\n", allNodes[i].Node, taskCount)

				if taskCount == 0 {
					//Table entry doesn't exist, create it
					sq = fmt.Sprintf(`INSERT INTO tasks.schedules (cluster_id,role,action,data,frequency,enabled,node_type) VALUES ('%s','all','BackupDatabase','rdpg','1 day'::interval, true, 'read')`, allNodes[i].ServiceName)

					err = execQuery(address, sq)
					fmt.Printf("%s: Had to add BackupDatabase for 'rdpg' into 'task.schedules'...\n", allNodes[i].Node)
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(1 * time.Second)
				}

				sq = `SELECT count(action) as rowCount FROM tasks.schedules WHERE action = 'BackupDatabase' AND data = 'rdpg'; `
				rowCount, err := GetRowCount(address, sq)
				nodeRowCount = append(nodeRowCount, rowCount)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("%s: Found %d scheduled tasks for BackupDatabase for database 'rdpg' in tasks.schedules...\n", allNodes[i].Node, rowCount)
			}
			//Verify each database also sees the same number of records (bdr sanity check)
			for i := 1; i < len(nodeRowCount); i++ {
				Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
			}
			Expect(len(nodeRowCount)).NotTo(Equal(0))
			Expect(nodeRowCount[0]).To(Equal(1))

		})

		It("Check Vacuum job in tasks.schedules no longer is scheduled for 'any'", func() {

			allNodes := GetAllNodes()

			var nodeRowCount []int
			for i := 0; i < len(allNodes); i++ {
				address := allNodes[i].Address

				sq := `SELECT count(action) as rowCount FROM tasks.schedules WHERE action = 'Vacuum' AND node_type = 'any'; `
				taskCount, err := GetRowCount(address, sq)

				fmt.Printf("%s: Found %d taskCount'...\n", allNodes[i].Node, taskCount)

				if taskCount == 1 {
					//Table entry doesn't exist, create it
					sq = fmt.Sprintf(`UPDATE tasks.schedules SET node_type = 'read' WHERE action = 'Vacuum' AND node_type = 'any';`)

					err = execQuery(address, sq)
					fmt.Printf("%s: Had to add change Vacuum node_type to 'read' from 'any' in 'task.schedules'...\n", allNodes[i].Node)
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(1 * time.Second)
				}

				sq = `SELECT count(action) as rowCount FROM tasks.schedules WHERE action = 'Vacuum' AND node_type = 'read'; `
				rowCount, err := GetRowCount(address, sq)
				nodeRowCount = append(nodeRowCount, rowCount)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("%s: Found %d scheduled tasks for Vacuum with node_type 'read' in tasks.schedules...\n", allNodes[i].Node, rowCount)
			}
			//Verify each database also sees the same number of records (bdr sanity check)
			for i := 1; i < len(nodeRowCount); i++ {
				Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
			}
			Expect(len(nodeRowCount)).NotTo(Equal(0))
			Expect(nodeRowCount[0]).To(Equal(1))

		})

		It("Check Vacuum job exists in tasks.schedules is scheduled for node_type 'write'", func() {

			allNodes := GetAllNodes()

			var nodeRowCount []int
			for i := 0; i < len(allNodes); i++ {
				address := allNodes[i].Address

				sq := `SELECT count(action) as rowCount FROM tasks.schedules WHERE action = 'Vacuum' AND node_type = 'write'; `
				taskCount, err := GetRowCount(address, sq)

				fmt.Printf("%s: Found %d taskCount'...\n", allNodes[i].Node, taskCount)

				if taskCount == 0 {
					//Table entry doesn't exist, create it
					sq = fmt.Sprintf(`INSERT INTO tasks.schedules (cluster_id,role,action,data,frequency,enabled,node_type) VALUES ('%s','all','Vacuum','tasks.tasks','5 minutes'::interval, true, 'write')`, allNodes[i].ServiceName)

					err = execQuery(address, sq)
					fmt.Printf("%s: Had to add Vacuum job for the node_type 'write' into 'task.schedules'...\n", allNodes[i].Node)
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(1 * time.Second)
				}

				sq = `SELECT count(action) as rowCount FROM tasks.schedules WHERE action = 'Vacuum' AND node_type = 'write'; `
				rowCount, err := GetRowCount(address, sq)
				nodeRowCount = append(nodeRowCount, rowCount)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("%s: Found %d scheduled tasks for Vacuum with node_type 'write' in tasks.schedules...\n", allNodes[i].Node, rowCount)
			}
			//Verify each database also sees the same number of records (bdr sanity check)
			for i := 1; i < len(nodeRowCount); i++ {
				Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
			}
			Expect(len(nodeRowCount)).NotTo(Equal(0))
			Expect(nodeRowCount[0]).To(Equal(1))

		})

		It("Check FindFilesToCopyToS3 job exists in tasks.schedules is scheduled for node_type 'write'", func() {

			allNodes := GetAllNodes()

			var nodeRowCount []int
			for i := 0; i < len(allNodes); i++ {
				address := allNodes[i].Address

				sq := `SELECT count(action) as rowCount FROM tasks.schedules WHERE action = 'FindFilesToCopyToS3' AND node_type = 'write'; `
				taskCount, err := GetRowCount(address, sq)

				fmt.Printf("%s: Found %d taskCount'...\n", allNodes[i].Node, taskCount)

				if taskCount == 0 {
					//Table entry doesn't exist, create it
					sq = fmt.Sprintf(`INSERT INTO tasks.schedules (cluster_id,role,action,data,frequency,enabled,node_type) VALUES ('%s','all','FindFilesToCopyToS3','tasks.tasks','5 minutes'::interval, false, 'write')`, allNodes[i].ServiceName)

					err = execQuery(address, sq)
					fmt.Printf("%s: Had to add FindFilesToCopyToS3 job for the node_type 'write' into 'task.schedules'...\n", allNodes[i].Node)
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(1 * time.Second)
				}

				sq = `SELECT count(action) as rowCount FROM tasks.schedules WHERE action = 'FindFilesToCopyToS3' AND node_type = 'write'; `
				rowCount, err := GetRowCount(address, sq)
				nodeRowCount = append(nodeRowCount, rowCount)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("%s: Found %d scheduled tasks for FindFilesToCopyToS3 with node_type 'write' in tasks.schedules...\n", allNodes[i].Node, rowCount)
			}
			//Verify each database also sees the same number of records (bdr sanity check)
			for i := 1; i < len(nodeRowCount); i++ {
				Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
			}
			Expect(len(nodeRowCount)).NotTo(Equal(0))
			Expect(nodeRowCount[0]).To(Equal(1))

		})

		It("Check FindFilesToCopyToS3 job exists in tasks.schedules is scheduled for node_type 'read'", func() {

			allNodes := GetAllNodes()

			var nodeRowCount []int
			for i := 0; i < len(allNodes); i++ {
				address := allNodes[i].Address

				sq := `SELECT count(action) as rowCount FROM tasks.schedules WHERE action = 'FindFilesToCopyToS3' AND node_type = 'read'; `
				taskCount, err := GetRowCount(address, sq)

				fmt.Printf("%s: Found %d taskCount'...\n", allNodes[i].Node, taskCount)

				if taskCount == 0 {
					//Table entry doesn't exist, create it
					sq = fmt.Sprintf(`INSERT INTO tasks.schedules (cluster_id,role,action,data,frequency,enabled,node_type) VALUES ('%s','all','FindFilesToCopyToS3','tasks.tasks','5 minutes'::interval, false, 'read')`, allNodes[i].ServiceName)

					err = execQuery(address, sq)
					fmt.Printf("%s: Had to add FindFilesToCopyToS3 job for the node_type 'read' into 'task.schedules'...\n", allNodes[i].Node)
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(1 * time.Second)
				}

				sq = `SELECT count(action) as rowCount FROM tasks.schedules WHERE action = 'FindFilesToCopyToS3' AND node_type = 'read'; `
				rowCount, err := GetRowCount(address, sq)
				nodeRowCount = append(nodeRowCount, rowCount)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("%s: Found %d scheduled tasks for FindFilesToCopyToS3 with node_type 'read' in tasks.schedules...\n", allNodes[i].Node, rowCount)
			}
			//Verify each database also sees the same number of records (bdr sanity check)
			for i := 1; i < len(nodeRowCount); i++ {
				Expect(nodeRowCount[0]).To(Equal(nodeRowCount[i]))
			}
			Expect(len(nodeRowCount)).NotTo(Equal(0))
			Expect(nodeRowCount[0]).To(Equal(1))

		})
	*/
})
