package app

import (
	"log"

	appstorage "github.com/kormiltsev/item-keeper/internal/app/appstorage"
)

type SearchByParameters struct {
	Mapa          map[string][]string
	Answer        map[int64]*appstorage.Item
	FileAddresses map[int64][]string //key ItemID, list of file Addresses, for UI uses
}

func NewSearchByParameter() *SearchByParameters {
	return &SearchByParameters{Mapa: map[string][]string{}}
}

// SearchItemByParameters return list of items, that has requestet parameter value
func (searchmapa *SearchByParameters) SearchItemByParameters() error {

	// erase data if exists
	searchmapa.Answer = map[int64]*appstorage.Item{}
	searchmapa.FileAddresses = map[int64][]string{}

	// prepare to server
	oper, err := appstorage.ReturnOperator(currentuser)
	if err != nil {
		log.Println("user not found in local memory. RegUser before SearchItemByParameters(): ", err)
		return nil
	}

	for key, val := range searchmapa.Mapa {
		if _, ok := oper.Search[key]; !ok {
			oper.Search[key] = make([]string, 0)
		}
		oper.Search[key] = append(oper.Search[key], val...)

		log.Printf("will search %v in parameter [%s]\n", oper.Search[key], key)
	}

	err = oper.FindItemByParameter()
	if err != nil {
		log.Printf("FAIL search error: %v, looking for:%v", err, searchmapa.Mapa)
	}

	// ans := "search results:"

	for key, item := range oper.Answer {
		// ans = fmt.Sprintf("%s\n%v, Files: %v", ans, item.Parameters, oper.AnswerAddresses[key])

		// copy to answer
		searchmapa.Answer[key] = item
	}
	// log.Println(ans)

	return nil
}
