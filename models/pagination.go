package models

type PaginationQuery struct {
	PageNum  int  `form:"pageNum" json:"pageNum"`
	PageSize int  `form:"pageSize" json:"pageSize"`
	Desc     bool `form:"desc" json:"desc"`
}

type PaginationResult struct {
	Total    int `form:"total" json:"total"`
	PageNum  int `form:"pageNum" json:"pageNum"`
	PageSize int `form:"pageSize" json:"pageSize"`
}

// NewPaginationResult 创建一个新的分页结果对象
func NewPaginationResult(total, pageNum, pageSize int) PaginationResult {
	return PaginationResult{
		Total:    total,
		PageNum:  pageNum,
		PageSize: pageSize,
	}
}
