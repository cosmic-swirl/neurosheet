package main

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
	"github.com/rs/xid"
	"crypto/sha256"
	"math"
	"io"
	"errors"
	"strconv"
)


type State struct {
	Store []StoreItem `json:"store"`
	Connections []ConnectionItem `json:"connections"`
	EventLog []EventLogItem `json:"eventLog"`
}

type StoreItem struct {
	Identity string `json:"identity"`
	CreationTime time.Time `json:"creationTime"`
	LastModifiedTime time.Time `json:"lastModifiedTime"`
	LatestEventID string `json:"latestEventID"`
	FileLocation string `json:"fileLocation"`
	Checksum string `json:"checksum"`
}

type ConnectionItem struct {
	Identity string `json:"identity"`
	CreationTime time.Time `json:"creationTime"`
	LastModifiedTime time.Time `json:"lastModifiedTime"`
	LatestEventID string `json:"latestEventID"`
	Strength float32 `json:"strength"`
	Items []string `json:"items"`
}

type EventLogItem struct {
	Identity string `json:"identity"`
	PreviousEvent string `json:previousEvent`
	Time time.Time `json:"time"`
	ModificationType ModType `json:"modificationType"`
	Changes []Change `json:"changes"`
}

type Change struct {
	Field string
	Value string
}

type ModType int

const (
	INITIAL ModType = iota
	REVERT
	APPEND
	REMOVE
	DELETE
)

type IdentityType int

const (
	STORE IdentityType = iota
	CONNECTION
	EVENT
)

var state State

func floatToString(input_num float64) string {
    // to convert a float number to a string
    return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func createIdentity(idType IdentityType) (string, error) {
	switch (idType) {
		case STORE:
			return fmt.Sprintf("ns-%s", xid.New().String()), nil
		case CONNECTION:
			return fmt.Sprintf("nc-%s", xid.New().String()), nil
		case EVENT:
			return fmt.Sprintf("ne-%s", xid.New().String()), nil
		default:
			return "", errors.New("idType does not exist")
	}
}

func encodeStateSnippetToJson(input_state interface{}) []byte {
    bytes, err := json.MarshalIndent(input_state, "", "  ")
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }

    return bytes
}

func loadCollectionFromJson() {
	raw, err := ioutil.ReadFile("./collection.json")
	if err != nil {
		fmt.Println("error reading state from json", err.Error())
		os.Exit(1)
	}

	json.Unmarshal(raw, &state)
}

func writeStateToJson(bytes []byte) {
	err := ioutil.WriteFile("./collection.json", bytes, 0777)
	if err != nil {
		fmt.Println("error writing state to json", err.Error())
		os.Exit(1)
	}
}

func getState() string {
	return string(encodeStateSnippetToJson(state))
}

func getStore() string {
	return string(encodeStateSnippetToJson(state.Store))
}

func getConnections() string {
	return string(encodeStateSnippetToJson(state.Connections))
}

func getEventLog() string {
	return string(encodeStateSnippetToJson(state.EventLog))
}

func searchStoreForItem(searchItem string) (bool, int, StoreItem) {
	for i, v := range state.Store {
		if v.Identity == searchItem {
			return true, i, v
		}
	}
	return false, -1, StoreItem{}
}

func searchConnectionsForItem(searchItem string) (bool, int, ConnectionItem) {
	for i, v := range state.Connections {
		if v.Identity == searchItem {
			return true, i, v
		}
	}
	return false, -1, ConnectionItem{}
}

func logEvent(identity string, currentTime time.Time, modType ModType, previousEvent string, changes []Change) string {

	item := EventLogItem{
		Identity: identity,
		Time: currentTime,
		PreviousEvent: previousEvent,
		ModificationType: modType,
		Changes: changes,
	}
	state.EventLog = append(state.EventLog, item)
	return identity
}

func addStoreItem(fileLocation string) {
	creationTime := time.Now()
	file, openErr := os.Open(fileLocation)
	if openErr != nil {
		fmt.Println("could not open file")
		return
	}

	defer file.Close()

	const filechunk = 8192

	info, _ := file.Stat()
	filesize := info.Size()
	blocks := uint64(math.Ceil(float64(filesize) / float64(filechunk)))
	hash := sha256.New()

	for i := uint64(0); i < blocks; i++ {
		blocksize := int(math.Min(filechunk, float64(filesize-int64(i*filechunk))))
		buf := make([]byte, blocksize)

		file.Read(buf)
		io.WriteString(hash, string(buf))
	}

	identity, idErr := createIdentity(STORE)
	if idErr != nil {
		fmt.Println("could not create id")
		return
	}
	lastModifiedTime := creationTime
	checksum := fmt.Sprintf("%x", hash.Sum(nil))

	eventId, eventIdErr := createIdentity(EVENT)
	if eventIdErr != nil {
		fmt.Println("error when creating event identity")
		return
	}

	changes := []Change{
		Change{Field: "Identity", Value: identity},
		Change{Field: "CreationTime", Value: creationTime.String()},
		Change{Field: "LastModifiedTime", Value: creationTime.String()},
		Change{Field: "FileLocation", Value: fileLocation},
		Change{Field: "Checksum", Value: checksum},
	}
	logEvent(eventId, creationTime, INITIAL, "nn", changes)

	item := StoreItem{
		Identity: identity,
		CreationTime: creationTime,
		LastModifiedTime: lastModifiedTime,
		LatestEventID: eventId,
		FileLocation: fileLocation,
		Checksum: checksum,
	}
	state.Store = append(state.Store, item)
}

func addConnectionItem(item1 string, item2 string, input_strength float32) {
	creationTime := time.Now()
	if !(1 > input_strength && input_strength > 0) {
		return
	}

	search1, _, _ := searchStoreForItem(item1)
	search2, _, _ := searchStoreForItem(item2)

	if !( search1 && search2 ) {
		fmt.Println("could not find both items in store for connection")
		return
	}

	identity, idErr := createIdentity(CONNECTION)
	if idErr != nil {
		fmt.Println("could not create id")
		return
	}
	eventId, eventIdErr := createIdentity(EVENT)
	if eventIdErr != nil {
		fmt.Println("could not create event id")
		return
	}
	lastModifiedTime := creationTime
	strength := input_strength
	items := []string{item1, item2}

	changes := []Change{
		Change{Field: "Identity", Value: identity},
		Change{Field: "CreationTime", Value: creationTime.String()},
		Change{Field: "LastModifiedTime", Value: creationTime.String()},
		Change{Field: "Strength", Value: floatToString( float64(strength) )},
		Change{Field: "Items", Value: string( encodeStateSnippetToJson(items) )},
	}
	logEvent(eventId, creationTime, INITIAL, "nn", changes)

	item := ConnectionItem{
		Identity: identity,
		CreationTime: creationTime,
		LastModifiedTime: lastModifiedTime,
		LatestEventID: eventId,
		Strength: strength,
		Items: items,
	}
	state.Connections = append(state.Connections, item)
}

func deleteStoreItem(identity string) {
	currentTime := time.Now()
	_, index, v := searchStoreForItem(identity)
	if !(index > 0) {
		return
	}
	eventId, idErr := createIdentity(EVENT)
	if idErr != nil {
		return
	}
	changes := []Change{}
	logEvent(eventId, currentTime, DELETE, v.LatestEventID, changes)
	state.Store = append(state.Store[:index], state.Store[index+1:]...)
	return
}

func deleteConnection(identity string) {
	currentTime := time.Now()
	changes := []Change{}
	eventId, idErr := createIdentity(EVENT)
	if idErr != nil {
		return
	}

	_, index, item := searchConnectionsForItem(identity)
	if !(index > 0) {
		return
	}
	logEvent(eventId, currentTime, DELETE, item.LatestEventID, changes)
	state.Connections = append(state.Connections[:index], state.Connections[index+1:]...)
	return
}

func main() {

	loadCollectionFromJson()
	addStoreItem("./test.txt")
	// addConnectionItem("ns-bbmvdq1hb52ct4qc4bdg", "ns-bbmvap9hb52cuucmvolg", 0.5)
	// printState()
	writeStateToJson(encodeStateSnippetToJson(state))

	// id, err := createIdentity(CONNECTION)
	// if err != nil {
	// 	fmt.Println("could not create id")
	// 	return
	// }
	// fmt.Println(id)

	fmt.Println(getEventLog())
	// fmt.Println(getState()) 

}
