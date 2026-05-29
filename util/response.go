package util

import (
	"fmt"
)

func SuccessResponse() Response {
	return Response{Head: map[string]string{"Code": "1000", "Msg": "Success"}}
}

func BuildError(args ...string) (errResponse Response) {
	if len(args) > 1 {
		errResponse = Response{Head: map[string]string{"Code": args[0], "Msg": args[1]}}
	} else if len(args) == 1 {
		errResponse = Response{Head: map[string]string{"Code": args[0]}}
	} else {
		panic("Need args for errResponse!")
	}
	return
}

func ArrayResponse() Response {
	return Response{
		Head:       map[string]string{"Code": "1000", "Msg": "Success"},
		Params:     map[string]string{},
		Pagination: &Pagination{},
	}
}

type Response struct {
	Head       map[string]string
	Params     map[string]string `json:"-"`
	Pagination *Pagination       `json:"Pagination,omitempty"`
	Body       any               `json:"Body,omitempty"`
}

func (errorResponse Response) Error() string {
	return fmt.Sprintf("Code: %s; Msg: %s", errorResponse.Head["Code"], errorResponse.Head["Msg"])
}

type Pagination struct {
	Order        string `query:"Order"`
	CurrentPage  int64  `query:"CurrentPage"`
	PerPage      int64  `query:"PerPage"`
	Count        int64
	TotalPage    int64
	NextPage     int64
	PreviousPage int64
}

func (pagination *Pagination) Init() {
	if pagination.Order == "" {
		pagination.Order = "id DESC"
	}
	if pagination.CurrentPage < 1 {
		pagination.CurrentPage = 1
	}
	if pagination.PerPage < 1 || pagination.PerPage > 200 {
		pagination.PerPage = 10
	}
	pagination.TotalPage = pagination.Count / pagination.PerPage
	if (pagination.Count % pagination.PerPage) != 0 {
		pagination.TotalPage += 1
	}
	if pagination.TotalPage == 0 {
		pagination.TotalPage = 1
	}
	pagination.NextPage = pagination.CurrentPage + 1
	if pagination.NextPage > pagination.TotalPage {
		pagination.NextPage = pagination.TotalPage
	}
	pagination.PreviousPage = pagination.CurrentPage - 1
	if pagination.PreviousPage < 1 {
		pagination.PreviousPage = 1
	}
}
