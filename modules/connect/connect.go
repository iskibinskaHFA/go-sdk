// Package connect enables connection to Postgres Database and Returns Connection handle.
package connect

import (
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/jinzhu/gorm"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
)

//GetDBConnection returns Postgres connection handle
func GetDBConnection() *sql.DB {
	psqlInfo := getDBInfo()
	fmt.Println(psqlInfo)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)

	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func getSession() (session.Session, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String("us-east-1")},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		panic(err)
	}
	return *sess, nil
}

func getValue(keyname string) (string, error) {
	sess, _ := getSession()
	ssmsvc := ssm.New(&sess, aws.NewConfig().WithRegion("us-east-1"))
	withDecryption := true
	param, err := ssmsvc.GetParameter(&ssm.GetParameterInput{
		Name:           &keyname,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	return strings.TrimSpace(aws.StringValue(param.Parameter.Value)), nil
}

func getDBInfo() string {
	host, _ := getValue("/royalties_sys/MLCPRC_HOST")
	password, _ := getValue("/royalties_sys/MLCPRC_PASSWORD")
	user, _ := getValue("/royalties_sys/MLCPRC_USER")
	ports, _ := getValue("/royalties_sys/MLCPRC_PORT")
	database, _ := getValue("/royalties_sys/MLCPRC_DATABASE")
	port, _ := strconv.Atoi(ports)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, database)
	return psqlInfo
}

//NewTestConnection returns Postgres connection handle for test environment
func NewTestConnection() (*gorm.DB, *sql.DB) {
	host := os.Getenv("HOST")
	user := os.Getenv("USER")
	password := os.Getenv("PASSWORD")

	port := os.Getenv("PORT")
	database := os.Getenv("DATABASE")
	i, _ := strconv.Atoi(port)
	dsn := url.URL{

		User:     url.UserPassword(user, password),
		Scheme:   "postgres",
		Host:     fmt.Sprintf("%s:%d", host, i),
		Path:     database,
		RawQuery: (&url.Values{"sslmode": []string{"disable"}}).Encode(),
	}

	db, err := gorm.Open("postgres", dsn.String())

	if err != nil {
		log.Println(err.Error())
	}

	dbase := db.DB()
	err = dbase.Ping()
	if err != nil {
		log.Println(err.Error())
	}

	defineHandler(db)

	return db, dbase
}

func defineHandler(*gorm.DB) {
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		switch defaultTableName {
		case "headers":
			fallthrough
		case "usage_summaries":
			fallthrough
		case "releases":
			fallthrough
		case "resources":
			fallthrough
		case "works":
			fallthrough
		case "works_writer":
			fallthrough
		case "writers":
			return "usage." + defaultTableName
		}
		return "royalty." + defaultTableName
	}
}


