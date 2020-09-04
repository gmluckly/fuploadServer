package main

import (
	"flag"
	"fmt"
	"fuploadServer/config"
	"fuploadServer/service"
	"log"
)

func main() {
	fmt.Println("fupload server start........")
	var c = flag.String("c", "", "configure path")
	flag.Parse()
	config, err := config.NewConfig(*c)
	if err != nil {
		log.Fatal("can not find config file")
	}
	fmt.Println("configInfo", config)
	/*
		if strings.EqualFold(config.MysqlDB.Host, "") {
			//no db config,save the tasks info to temp file
		} else {
			host := config.MysqlDB.Host
			database := config.MysqlDB.Database
			userName := config.MysqlDB.UserName
			password := config.MysqlDB.Password
			dao.InitDB(host, database, userName, password)
		}
	*/
	port := config.Server.Port
	service.Run(port)
}
