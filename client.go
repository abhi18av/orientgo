package main

import (
	"fmt"
	"log"
	"ogonori/obinary"
	"os"
)

func serverCommands(dbc *obinary.DbClient) {
	fmt.Println("\n-------- server commands --------")

	err := obinary.CreateServerSession(dbc, "root", "A406A900E578DC7094FBA78001A45BE611DB06B25F593026CCDF737A31A9D0E9")
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN s1: %v\n", err)
		return
	}

	// err = obinary.RequestDbList(dbc)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "WARN s2: %v\n", err)
	// 	return
	// }

	var status bool
	status, err = obinary.DatabaseExists(dbc, "cars", obinary.PersistentStorageType)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("`cars` persistent db exists?: %v\n", status)
	}

	dbexists, err := obinary.DatabaseExists(dbc, "cars", obinary.VolatileStorageType)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("`cars` volatile db exists?: %v\n", dbexists)

	dbexists, err = obinary.DatabaseExists(dbc, "clammy", obinary.VolatileStorageType)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("`clammy` volatile db exists?: %v\n", dbexists)

	if !dbexists {
		fmt.Println("attemping to create clammy db ... ")
		err = obinary.CreateDatabase(dbc, "clammy", obinary.DocumentDbType, obinary.VolatileStorageType)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("clammy db created ... ")
	}

	status, err = obinary.DatabaseExists(dbc, "clammy", obinary.VolatileStorageType)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("`clammy` volatile db exists?: %v\n", status)

	err = obinary.DropDatabase(dbc, "clammy", obinary.PersistentStorageType)
	if err != nil {
		log.Fatal(err)
	}
}

func dbCommands(dbc *obinary.DbClient) {
	fmt.Println("\n-------- database-level commands --------")

	fmt.Println("OpenDatabase")
	err := obinary.OpenDatabase(dbc, "cars", obinary.DocumentDbType, "admin", "admin")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", dbc) // DEBUG

	clusterId, err := obinary.AddCluster(dbc, "bigapple")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("bigapple cluster added => clusterId in `cars`: %v\n", clusterId)

	cnt, err := obinary.GetClusterCount(dbc, "bigapple")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("bigapple cluster count = %d\n", cnt)

	for _, name := range []string{"bigApple"} {
		err = obinary.DropCluster(dbc, name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("cluster %v dropped successfully\n", name)
	}

	cnt, err = obinary.GetClusterCountIncludingDeleted(dbc, "person", "v", "ouser")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ClusterCount w/ deleted: %v\n", cnt)

	cnt, err = obinary.GetClusterCount(dbc, "person", "v", "ouser")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ClusterCount w/o deleted: %v\n", cnt)

	begin, end, err := obinary.GetClusterDataRange(dbc, "ouser")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ClusterDataRange for ouser: %d-%d\n", begin, end)

	fmt.Printf("\n+++ Attempting to fetch record now +++\n cmd num = %v\n", obinary.REQUEST_RECORD_LOAD)
	err = obinary.GetRecordByRID(dbc, "11:0", "", false, false)
	if err != nil {
		log.Fatal(err)
	}

	obinary.CloseDatabase(dbc)
}

func main() {
	var (
		dbc *obinary.DbClient
		err error
	)
	fmt.Println("ConnectToServer")
	dbc, err = obinary.NewDbClient(obinary.ClientOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer dbc.Close()

	// serverCommands(dbc)
	dbCommands(dbc)

	fmt.Println("DONE")
}