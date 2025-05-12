package runner

import (
	"fmt"
	"os/exec"
)

func RunDockerJob(workDir string){
	cmd := exec.Command("docker","run","--rm","-v",workDir+":/app","golang:1.21","go","build","/app")
	output,err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Docker run failed:", err)
	}
	fmt.Println(string(output))
}