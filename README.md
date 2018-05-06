# Neurosheet

## Description

A registry of data items, and the connections between them, stored in json, available through RESTful requests.
	
State:
	Store:
		List of informational items.
		Items can include, but are not limited to:
			* Text files
			* Images
			* Videos
			* Web links
		Each item is labeled with a unique id and their checksums, as well as metadata.
	Connections:
		List of links between items in the store, and the strength of their connection (0.0-1.0).
		Each link has a unique id, as well as metadata.
		A links id is deterministically generated from the two items being linked ids.

	Metadata:
		* Creation Time
		* Time of last modification
		* Last modification's event ID

EventLog:
	The state is modified through events, and can be replayed via event sourcing.
	Any change is made by triggering an event, this type and information contained in this event will be used to modify the **State**.

## Example

```json
{
	store: [
		{
			identity: 'NSx12as89df7a9',
			fileLocation: '../file.txt',
			checksum: '1l2k34jkl12j'
			creationTime: 1829374012347, // epoch
			lastModified: 1829739128738, // epoch
			modificationEventID: 'NEx24293842093'
		},
		{
			identity: 'NSx25kj457hl423l4j',
			fileLocation: '../img.txt',
			checksum: '2sadf234klj',
			creationTime: 2389238948932,
			lastModified: 2389428934822,
			modificationEventID: 'NEx28934289349'
		}
	],
	connections: [
		{
			identity: 'NCx234g2k34hjlhl',
			strength: '0.5',
			items: ['NSx12as89df7a9', 'NSx25kj457hl423l4j'],
			creationTime: 234234234234234,
			lastModified: 234238942834828,
			modificationEventID: 'NEx28423482j34j'
		},
	]
	eventLog: [
		{
			time: 928739179239
			modificationType: 'append', // append (new identity), revert, progressive_modification, destructive_modification
			identity: 'NSx12as89df7a9',
			lastIdentityEvent: 'none', // set to last eventID at this location or none, if identity is new
			change: {} // actual changes made to field
		},
		{},
		{}
	]
}
```

## API

	getState() // return entire state tree (store, connections)
	getStore() // return all items in store
	getConnections() // get all connected items
	getEventLog() // get log of all events that have occurred on this sheet

	updateConnection() // modify a connection
	updateStoreItem() // add an item to the store
		
	addConnection()
	addStoreItem()

	deleteConnection()
	deleteStoreItem()
	
	validateConnection()
	validateStoreItem() // ensure store item still matches its checksum


### TODO
JSON read // done
JSON write // not done
HTTP SERV // not done

	
