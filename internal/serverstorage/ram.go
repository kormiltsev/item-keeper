package serverstorage

import "sync"

type changeRow struct {
	updated int64
	userid  string
	itemid  int64
	fileid  int64
}

// list of changes
var (
	muchanges   sync.Mutex
	changesRows = []changeRow{}
)

// logNewChange save new change in RAM
func logNewChange(userid string, itemid, fileid, changesid int64) {
	muchanges.Lock()
	defer muchanges.Unlock()

	changesRows = append(changesRows, changeRow{
		updated: changesid,
		userid:  userid,
		itemid:  itemid,
		fileid:  fileid,
	})
}
