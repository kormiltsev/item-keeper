package storage

import "time"

type queue struct {
	IDofNewUploadedFiles chan Item
}

var queueOfTasks = queue{
	IDofNewUploadedFiles: make(chan Item, 100),
}

// RunCollector collect new files addresses and push storager to update in DB
func RunCollector(stop chan struct{}) {
	ticker := time.NewTicker(3000 * time.Millisecond)
	listItemsToUpdateFotoLinks := map[string]Item{}
	for {
		select {
		case <-stop:
			return
		case item := <-queueOfTasks.IDofNewUploadedFiles:
			if itm, ok := listItemsToUpdateFotoLinks[item.ID]; ok {
				itm.ImageLink = append(itm.ImageLink, item.ImageLink[0])
				listItemsToUpdateFotoLinks[item.ID] = itm
			}
		case <-ticker.C:
			if len(listItemsToUpdateFotoLinks) != 0 {
				// make new items list
				u := NewItem()

				// copy all items from map
				for _, item := range listItemsToUpdateFotoLinks {
					u.List = append(u.List, item)
				}

				// create interface
				u.DB = NewToStorage(u)

				// update links
				u.DB.UpdateItemsImageLinks()
			}
		}
	}
}

func chanIDofNewUploadedFiles() chan Item {
	return queueOfTasks.IDofNewUploadedFiles
}
