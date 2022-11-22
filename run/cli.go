package run

import (
	"fmt"
)

var (
	inputUi  = inputPromptUi{}
	selectUi = selectPromptUi{}
)

func Start() {
	keyword := inputUi.searchKeyword()
	fmt.Println(keyword)
	project := selectUi.Project([]string{"openapi-oss", "cmp-instance"})
	fmt.Println(project)

}
