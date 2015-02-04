package obinary

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

// TODO: pattern this after OStorageRemote ?
type DbClient struct {
	conx                  net.Conn
	buf                   *bytes.Buffer
	sessionId             int
	token                 []byte // orientdb token when not using sessionId
	serializationImpl     string
	binaryProtocolVersion int16
	currDb                *ODatabase
}

//
// NewDbClient creates a new DbClient after contacting the OrientDb server
// specified in the ClientOptions and validating that the server and client
// speak the same binary protocol version.
// The DbClient returned is ready to make calls to the OrientDb but has not
// yet established a database session or a session with the OrientDb server.
// After this, the user needs to call either OpenDatabase or CreateServerSession.
//
func NewDbClient(opts ClientOptions) (*DbClient, error) {
	// binary port range is: 2424-2430
	if opts.ServerHost == "" {
		opts.ServerHost = "0.0.0.0"
	}
	if opts.ServerPort == "" {
		opts.ServerPort = "2424"
	}
	hostport := fmt.Sprintf("%s:%s", opts.ServerHost, opts.ServerPort)
	fmt.Printf("%v\n", hostport) // DEBUG
	conx, err := net.Dial("tcp", hostport)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: %v\n", err)
		return nil, err
	}

	// after connecting the OrientDB server sends back 2 bytes - its binary protocol version
	readbuf := make([]byte, 2)
	n, err := conx.Read(readbuf)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.Write(readbuf[0:n])
	if err != nil {
		return nil, err
	}

	var svrProtocolNum int16
	binary.Read(buf, binary.BigEndian, &svrProtocolNum)
	if svrProtocolNum < MinSupportedBinaryProtocolVersion {
		return nil, UnsupportedVersionError{serverVersion: svrProtocolNum}
	} else if svrProtocolNum > MaxSupportedBinaryProtocolVersion {
		return nil, UnsupportedVersionError{serverVersion: svrProtocolNum}
	}

	serializerImpl := BinarySerialization
	if svrProtocolNum < MinBinarySerializerVersion {
		serializerImpl = CsvSerialization
	}

	// DEBUG
	fmt.Printf("svrProtocolNum: %v\n", svrProtocolNum)
	// END DEBUG

	dbc := &DbClient{
		conx:                  conx,
		buf:                   new(bytes.Buffer),
		serializationImpl:     serializerImpl,
		binaryProtocolVersion: svrProtocolNum,
		sessionId:             NoSessionId,
	}
	return dbc, nil
}

// *DbClient implements Closer
func (dbc *DbClient) Close() error {
	if dbc.currDb != nil {
		// ignoring any error here, since closing the conx also terminates the session
		CloseDatabase(dbc)
	}
	return dbc.conx.Close()
}

func (dbc DbClient) String() string {
	if dbc.currDb == nil {
		return "DbClient[not-connected-to-db]"
	}
	return fmt.Sprintf("DbClient[connected-to: %v of type %v with %d clusters; sessionId: %v\n  CurrDb Details: %v]",
		dbc.currDb.Name, dbc.currDb.Typ, len(dbc.currDb.Clusters), dbc.sessionId, dbc.currDb)
}