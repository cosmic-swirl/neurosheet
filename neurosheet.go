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
	Time time.Time `json:"time"`
	ModificationType ModType `json:"modificationType"`
	Change []Change `json:"change"`
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
)

type IdentityType int

const (
	STORE IdentityType = iota
	CONNECTION
	EVENT
)

var state State

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

func encodeStateSnippetToJson() []byte {
    bytes, err := json.MarshalIndent(state, "", "  ")
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

func getState () string {
	return string(encodeStateSnippetToJson())
}

func getStore () string {
	return string(encodeStateSnippetToJson())
}

func searchStoreForItem(searchItem string) bool {
	for _, v := range state.Store {
		if v.Identity == searchItem {
			return true
		}
	}
	return false
}

func addStoreItem(fileLocation string) {
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
	creationTime := time.Now()
	lastModifiedTime := creationTime
	latestEventID := "ne-0"
	checksum := fmt.Sprintf("%x", hash.Sum(nil))

	item := StoreItem{
		Identity: identity,
		CreationTime: creationTime,
		LastModifiedTime: lastModifiedTime,
		LatestEventID: latestEventID,
		FileLocation: fileLocation,
		Checksum: checksum,
	}
	state.Store = append(state.Store, item)
}

func addConnectionItem(item1 string, item2 string, input_strength float32) {
	if !(1 > input_strength && input_strength > 0) {
		return
	}

	if !( searchStoreForItem(item1) && searchStoreForItem(item2) ) {
		fmt.Println("could not find both items in store for connection")
		return
	}

	identity, idErr := createIdentity(CONNECTION)
	if idErr != nil {
		fmt.Println("could not create id")
		return
	}
	creationTime := time.Now()
	lastModifiedTime := creationTime
	latestEventID := "ne-0"
	strength := input_strength
	items := []string{item1, item2}

	item := ConnectionItem{
		Identity: identity,
		CreationTime: creationTime,
		LastModifiedTime: lastModifiedTime,
		LatestEventID: latestEventID,
		Strength: strength,
		Items: items,
	}
	state.Connections = append(state.Connections, item)
}

func logEvent(currentTime time.Time, modType ModType, changes []Change) {
	identity, idErr := createIdentity(EVENT)
	if idErr != nil {
		fmt.Println("could not create id")
		return
	}

	item := EventLogItem{
		Identity: identity,
		Time: currentTime,
		ModificationType: modType,
		Change: changes,
	}
	state.EventLog = append(state.EventLog, item)
}

func main() {

	// loadCollectionFromJson()
	// // addStoreItem("./test.txt")
	// addConnectionItem("ns-bbmvdq1hb52ct4qc4bdg", "ns-bbmvap9hb52cuucmvolg", 0.5)
	// printState()
	// writeStateToJson(encodeStateSnippetToJson())

	id, err := createIdentity(CONNECTION)
	if err != nil {
		fmt.Println("could not create id")
		return
	}
	fmt.Println(id)

}
