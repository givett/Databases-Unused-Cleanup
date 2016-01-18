package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/starkandwayne/rdpgd/log"
)

type s3Credentials struct {
	awsSecretKey      string
	awsAccessKey      string
	bucketName        string
	awsRegion         string
	configured        bool
	token             string
	endpoint          string
	s3ForcePathStyle  bool
	enabledInManifest string
}

//S3FileMetadata - Basic meta data needed for all file manipulations of backup files
type S3FileMetadata struct {
	Location  string `json:"location"`
	DBName    string `json:"dbname"`
	Node      string `json:"node"`
	ClusterID string `json:"cluster_id"`
}

//S3FileDownload - Meta data needed for copying files from an s3 bucket
type S3FileDownload struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Bucket string `json:"bucket"`
	DBName string `json:"dbname"`
}

//FindFilesToCopyToS3 - Responsible for copying files, such as database backups
//to S3 storage
func (t *Task) FindFilesToCopyToS3() (err error) {
	//Pull S3 Credentials
	s3Creds, err := getS3Credentials()
	if err != nil {
		log.Error(fmt.Sprintf("tasks.FindFilesToCopyToS3() Could not retrieve S3 Credentials ! %s", err))
		return err
	}

	//If S3 creds/bucket aren't set just exit since they aren't configured
	if s3Creds.configured == false {
		log.Error(fmt.Sprintf("tasks.FindFilesToCopyToS3() S3 CONFIGURATION MISSING FOR THIS DEPLOYMENT ! S3 Credentials are not configured, skipping attempt to copy until configured "))
		return
	}

	//Select eligible files
	address := `127.0.0.1`
	sq := `SELECT a.params FROM backups.file_history a WHERE a.removed_at IS NULL AND a.action = 'CreateBackup' AND NOT EXISTS (SELECT b.params FROM backups.file_history b WHERE a.cluster_id = b.cluster_id AND a.dbname = b.dbname AND a.node=b.node AND a.file_name = b.file_name AND b.action='CopyToS3' AND b.status='ok') AND NOT EXISTS (SELECT id FROM tasks.tasks WHERE action = 'CopyFileToS3' AND data = a.params::text)`
	filesToCopy, err := getList(address, sq)
	if err != nil {
		log.Error(fmt.Sprintf(`tasks.Task<%d>#CopyFileToS3() Failed to load list of files ! %s`, t.ID, err))
	}

	log.Trace(fmt.Sprintf("tasks.FindFilesToCopyToS3() > Found %d files to copy", len(filesToCopy)))

	//Loop and add Tasks CopyFileToS3
	for _, fileToCopyParams := range filesToCopy {
		log.Trace(fmt.Sprintf("tasks.FindFilesToCopyToS3() > Attempting to add %s", fileToCopyParams))

		newTask := Task{ClusterID: t.ClusterID, Node: t.Node, Role: t.Role, Action: "CopyFileToS3", Data: fileToCopyParams, TTL: t.TTL, NodeType: t.NodeType}
		err = newTask.Enqueue()
		if err != nil {
			log.Error(fmt.Sprintf(`tasks.FindFilesToCopyToS3() service task schedules ! %s`, err))
		}

	}
	return

}

//CopyFileToS3 - Responsible for copying a file to S3
func (t *Task) CopyFileToS3() (err error) {
	start := time.Now()

	//Pull S3 Credentials
	s3Creds, err := getS3Credentials()
	if err != nil {
		log.Error(fmt.Sprintf("tasks.CopyFileToS3() Could not retrieve S3 Credentials ! %s", err))
		return err
	}
	creds := credentials.NewStaticCredentials(s3Creds.awsAccessKey, s3Creds.awsSecretKey, s3Creds.token)

	config := &aws.Config{
		Region:           &s3Creds.awsRegion,
		Endpoint:         &s3Creds.endpoint,
		S3ForcePathStyle: &s3Creds.s3ForcePathStyle,
		Credentials:      creds,
	}

	s3client := s3.New(config)
	bucketName := s3Creds.bucketName

	taskParams := []byte(t.Data)
	fm := S3FileMetadata{}
	err = json.Unmarshal(taskParams, &fm)
	if err != nil {
		log.Error(fmt.Sprintf("tasks.CopyFileToS3() json.Unmarshal() ! %s", err))
	}

	file, err := os.Open(fm.Location)
	if err != nil {
		log.Error(fmt.Sprintf("tasks.CopyFileToS3() Error attempting to open file %s ! %s", fm.Location, err))
		return err
	}

	defer file.Close()

	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer) // convert to io.ReadSeeker type
	fileType := http.DetectContentType(buffer)

	s3params := &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),    // required
		Key:           aws.String(fm.Location),   // required
		ACL:           aws.String("public-read"), //other values: http://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#CannedACL
		Body:          fileBytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
		Metadata: map[string]*string{
			"Key": aws.String("MetadataValue"), //required
		},
		// see more at http://godoc.org/github.com/aws/aws-sdk-go/service/s3#S3.PutObject
	}

	result, err := s3client.PutObject(s3params)
	log.Trace(fmt.Sprintf("tasks.CopyFileToS3() Copy file to S3 result > %s ", result))

	if err != nil {
		log.Error(fmt.Sprintf("tasks.CopyFileToS3() AWS General Error ! %s", err))
		if awsErr, ok := err.(awserr.Error); ok {
			// Generic AWS Error with Code, Message, and original error (if any)
			log.Error(fmt.Sprintf("tasks.CopyFileToS3() AWS Error %s !! %s ! %s", awsErr.Code(), awsErr.Message(), awsErr.OrigErr()))
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// A service error occurred
				log.Error(fmt.Sprintf("tasks.CopyFileToS3() AWS Service Error %s !!! %s !! %s ! %s", reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID()))
			}
		} else {
			// This case should never be hit, the SDK should always return an
			// error which satisfies the awserr.Error interface.
			log.Error(fmt.Sprintf("tasks.CopyFileToS3() General AWS Error %s ! ", err.Error()))
		}
	}

	status := `ok`
	if err != nil {
		status = `error`
	}

	//Log results to backups.file_history
	f := s3FileHistory{}
	f.source = fm.Location
	f.target = fm.Location
	f.dbname = fm.DBName
	f.size = size
	f.node = fm.Node
	f.status = status
	f.duration = int(time.Since(start).Seconds())
	f.bucket = bucketName
	f.fileName = fileInfo.Name()
	insertErr := insertS3History(f)

	if insertErr != nil {
		return insertErr
	}
	return
}

//CopyFileFromS3 - Responsible for copying a file from S3
func (t *Task) CopyFileFromS3() (err error) {

	start := time.Now()

	//Pull S3 Credentials
	s3Creds, err := getS3Credentials()
	if err != nil {
		log.Error(fmt.Sprintf("tasks.CopyFileToS3() Could not retrieve S3 Credentials ! %s", err))
		return err
	}
	creds := credentials.NewStaticCredentials(s3Creds.awsAccessKey, s3Creds.awsSecretKey, s3Creds.token)

	config := &aws.Config{
		Region:           &s3Creds.awsRegion,
		Endpoint:         &s3Creds.endpoint,
		S3ForcePathStyle: &s3Creds.s3ForcePathStyle,
		Credentials:      creds,
	}

	s3client := s3.New(config)

	taskParams := []byte(t.Data)
	fm := S3FileDownload{}
	err = json.Unmarshal(taskParams, &fm)
	if err != nil {
		log.Error(fmt.Sprintf("tasks.CopyFileFromS3() json.Unmarshal() ! %s", err))
	}

	params := &s3.GetObjectInput{
		Bucket: aws.String(fm.Bucket),
		Key:    aws.String(fm.Source),
	}
	resp, err := s3client.GetObject(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		log.Error(fmt.Sprintf("tasks.CopyFileFromS3() AWS Error: ! %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	downloadFile, err := os.Create(fm.Target)
	if err != nil {
		log.Error(fmt.Sprintf("tasks.CopyFileFromS3() attempting to create file error: ! %s", err))
	}

	defer downloadFile.Close()

	size, err := io.Copy(downloadFile, resp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("tasks.CopyFileFromS3() Failed to copy object to file ! %s", err))
	}

	status := `ok`
	if err != nil {
		status = `error`
	}

	//Log results to backups.file_history
	f := s3FileHistory{}
	f.source = fm.Source
	f.target = fm.Target
	f.dbname = fm.DBName
	f.node = t.Node
	f.status = status
	f.size = size
	f.duration = int(time.Since(start).Seconds())
	f.bucket = fm.Bucket
	f.fileName = downloadFile.Name()
	insertErr := insertS3HistoryCopyFromS3(f)

	if insertErr != nil {
		return insertErr
	}
	return

}

func getS3Credentials() (s s3Credentials, err error) {
	//Initialize values
	s.configured = false
	s.token = ``
	s.s3ForcePathStyle = true

	s.awsAccessKey = os.Getenv(`RDPGD_S3_AWS_ACCESS`)
	s.awsSecretKey = os.Getenv(`RDPGD_S3_AWS_SECRET`)
	s.bucketName = os.Getenv(`RDPGD_S3_BUCKET`)
	s.awsRegion = os.Getenv(`RDPGD_S3_REGION`)
	s.endpoint = os.Getenv(`RDPGD_S3_ENDPOINT`)
	s.enabledInManifest = os.Getenv(`RDPGD_S3_BACKUPS`)

	if s.bucketName != `` && s.awsAccessKey != `` && s.awsSecretKey != `` && s.awsRegion != `` && strings.ToUpper(s.enabledInManifest) == `ENABLED` {
		s.configured = true
	}
	return
	/*

		PSQL="/var/vcap/packages/pgbdr/bin/psql -p7432 -U vcap"; $PSQL rdpg

		PSQL="/var/vcap/packages/postgresql-9.4/bin/psql -p7432 -U vcap"; $PSQL rdpg


			SELECT a.params FROM backups.file_history a WHERE a.removed_at IS NULL AND NOT EXISTS (SELECT b.params FROM backups.file_history b WHERE a.cluster_id = b.cluster_id AND a.dbname = b.dbname AND a.node=b.node AND a.file_name = b.file_name AND b.action='CopyToS3') AND NOT EXISTS (SELECT id FROM tasks.tasks WHERE action = 'CopyFileToS3' AND data = a.params::text);
			update tasks.schedules set last_scheduled_at = current_timestamp - '1 day'::interval where action = 'FindFilesToCopyToS3';
			update tasks.schedules set last_scheduled_at = current_timestamp - '1 day'::interval where action = 'BackupDatabase';
			update tasks.schedules set last_scheduled_at = current_timestamp - '1 day'::interval where action = 'EnforceFileRetention';

	*/
}
