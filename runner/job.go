package runner

import (
	"fmt"
	"os"
	"git-runner/utils"
)

func RunJob(repoURL, commitSHA string){
	dir := "/tmp/build-job"
	os.RemoveAll(dir)

err:= utils.CloneAndCheckout(repoURL, commitSHA, dir)
if err != nil {
	fmt.Println("Git error", err)
	return
}

fmt.Println("Cloned. Starting Docker job...")
RunDockerJob(dir)
}