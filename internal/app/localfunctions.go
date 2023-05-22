package app

import (
	"fmt"
	"log"

	appstorage "github.com/kormiltsev/item-keeper/internal/app/appstorage"
)

type SearchByParameters struct {
	Mapa          map[string][]string
	Answer        map[string]*appstorage.Item
	FileAddresses map[string][]string //key ItemID, list of file Addresses
}

func NewSearchByParameter() *SearchByParameters {
	return &SearchByParameters{Mapa: map[string][]string{}}
}

// SearchItemByParameters return list of items, that has requestet parameter value
func (searchmapa *SearchByParameters) SearchItemByParameters() error {

	// erase data if exists
	searchmapa.Answer = map[string]*appstorage.Item{}
	searchmapa.FileAddresses = map[string][]string{}

	// prepare to server
	oper, err := appstorage.ReturnOperator(currentuser)
	if err != nil {
		log.Println("user not found in local memory. RegUser before SearchItemByParameters()")
		return nil
	}

	for key, val := range searchmapa.Mapa {
		if _, ok := oper.Search[key]; !ok {
			oper.Search[key] = make([]string, 0)
		}
		oper.Search[key] = append(oper.Search[key], val...)

		log.Printf("will search %v in parameter %s\n", oper.Search[key], key)
	}

	err = oper.FindItemByParameter()
	if err != nil {
		log.Printf("FAIL search error: %v, looking for:%v", err, searchmapa.Mapa)
	}

	ans := "search results:"

	for key, item := range oper.Answer {
		ans = fmt.Sprintf("%s\n%v, Files: %v", ans, item.Parameters, oper.AnswerAddresses[key])

		// copy to answer
		searchmapa.Answer[key] = item
	}
	log.Println(ans)

	return nil
}
